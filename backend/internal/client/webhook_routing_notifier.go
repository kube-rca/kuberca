package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

type webhookConfigSource interface {
	GetWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error)
}

type webhookRoutingNotifier struct {
	cfgSource           webhookConfigSource
	fallback            Notifier
	fallbackThreadStore ThreadRefStore
	frontendURL         string
	httpClient          *http.Client

	mu           sync.RWMutex
	slackClients map[int]*SlackClient
	threadStore  sync.Map // 모든 Slack 클라이언트에 독립적인 thread ref 저장소
}

var _ Notifier = (*webhookRoutingNotifier)(nil)
var _ ThreadRefStore = (*webhookRoutingNotifier)(nil)
var _ ThreadRefRequirement = (*webhookRoutingNotifier)(nil)

type webhookEventEnvelope struct {
	EventType string      `json:"event_type"`
	SentAt    time.Time   `json:"sent_at"`
	Data      interface{} `json:"data"`
}

// NewWebhookRoutingNotifier는 DB webhook_configs(type)을 기반으로 notifier를 라우팅한다.
// DB 설정이 없으면 fallback notifier를 사용한다.
func NewWebhookRoutingNotifier(
	cfgSource webhookConfigSource,
	fallback Notifier,
	fallbackThreadStore ThreadRefStore,
	frontendURL string,
) ThreadAwareNotifier {
	return &webhookRoutingNotifier{
		cfgSource:           cfgSource,
		fallback:            fallback,
		fallbackThreadStore: fallbackThreadStore,
		frontendURL:         strings.TrimRight(frontendURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		slackClients: make(map[int]*SlackClient),
	}
}

func (n *webhookRoutingNotifier) Notify(event NotifierEvent) error {
	// thread 기반 이벤트(분석 결과, flapping 해제)는 thread를 소유한 Slack 클라이언트에만 전송.
	switch e := event.(type) {
	case AnalysisResultPostedEvent:
		return n.notifyByThread(e.ThreadRef, event)
	case *AnalysisResultPostedEvent:
		return n.notifyByThread(e.ThreadRef, event)
	case FlappingClearedEvent:
		return n.notifyByThread(e.ThreadRef, event)
	case *FlappingClearedEvent:
		return n.notifyByThread(e.ThreadRef, event)
	}

	severity := extractEventSeverity(event)
	targets, err := n.resolveNotifiers(severity)
	if err != nil {
		log.Printf("Failed to load webhook configs, falling back to default notifier: %v", err)
		if n.fallback != nil {
			return n.fallback.Notify(event)
		}
		return err
	}

	if len(targets) == 0 {
		if n.fallback != nil {
			return n.fallback.Notify(event)
		}
		return fmt.Errorf("no notifier target configured")
	}

	var errs []error
	success := 0
	for _, target := range targets {
		if err := target.Notify(event); err != nil {
			errs = append(errs, err)
			continue
		}
		success++
	}

	if success > 0 {
		return nil
	}

	if n.fallback != nil {
		if err := n.fallback.Notify(event); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Errorf("fallback notify failed: %w", err))
		}
	}

	if len(errs) == 0 {
		return fmt.Errorf("no notifier target configured")
	}
	return errors.Join(errs...)
}

// notifyByThread는 thread_ts를 소유한 Slack 클라이언트에만 이벤트를 전송한다.
// HTTP/Teams 엔드포인트는 thread 개념이 없으므로 제외한다.
func (n *webhookRoutingNotifier) notifyByThread(threadRef string, event NotifierEvent) error {
	if threadRef == "" {
		return nil
	}

	var found bool
	var sendErr error
	n.forEachSlackClient(func(c *SlackClient) {
		if found {
			return
		}
		if !c.HasThreadValue(threadRef) {
			return
		}
		found = true
		if err := c.Notify(event); err != nil {
			sendErr = err
		}
	})

	if found {
		return sendErr
	}

	// threadMap에서 못 찾은 경우(재시작 후 등): fallback에 위임
	if n.fallback != nil {
		return n.fallback.Notify(event)
	}
	return nil
}

// extractEventSeverity - 이벤트에서 severity를 추출한다.
// severity가 없는 이벤트(FlappingCleared, AnalysisResult)는 "" 반환 → 모든 채널로 전송.
func extractEventSeverity(event NotifierEvent) string {
	switch e := event.(type) {
	case AlertStatusChangedEvent:
		return e.Alert.Labels["severity"]
	case *AlertStatusChangedEvent:
		return e.Alert.Labels["severity"]
	case FlappingDetectedEvent:
		return e.Alert.Labels["severity"]
	case *FlappingDetectedEvent:
		return e.Alert.Labels["severity"]
	default:
		return ""
	}
}

// severityMatches - 명시적 severity 필터(비어있지 않은 배열)와 이벤트 severity가 매칭되는지 확인.
// eventSeverity가 빈 문자열(FlappingCleared 등)이면 항상 true.
// cfgSeverities가 비어있는 경우는 resolveNotifiers에서 별도 분기하므로 여기서는 비어있지 않음.
func severityMatches(cfgSeverities []string, eventSeverity string) bool {
	if eventSeverity == "" {
		return true
	}
	for _, s := range cfgSeverities {
		if s == eventSeverity {
			return true
		}
	}
	return false
}

// forEachSlackClient DB에 등록된 모든 유효한 Slack 클라이언트에 fn을 적용한다.
func (n *webhookRoutingNotifier) forEachSlackClient(fn func(*SlackClient)) {
	if n.cfgSource == nil {
		return
	}
	configs, err := n.cfgSource.GetWebhookConfigs(context.Background())
	if err != nil {
		return
	}
	for _, cfg := range configs {
		if normalizeWebhookType(cfg.Type) != "slack" {
			continue
		}
		client, ok := n.slackClientForConfig(cfg)
		if !ok {
			continue
		}
		fn(client)
	}
}

func (n *webhookRoutingNotifier) StoreThreadRef(alertKey, threadRef string) {
	// 자체 threadStore에 저장
	n.threadStore.Store(alertKey, threadRef)
	// 모든 Slack 클라이언트에도 전파 (post-restart recovery 지원)
	n.forEachSlackClient(func(c *SlackClient) {
		c.StoreThreadRef(alertKey, threadRef)
	})
	// fallback store에도 전파
	if n.fallbackThreadStore != nil {
		n.fallbackThreadStore.StoreThreadRef(alertKey, threadRef)
	}
}

func (n *webhookRoutingNotifier) GetThreadRef(alertKey string) (string, bool) {
	// 각 Slack 클라이언트의 인메모리 threadMap 우선 확인
	var found string
	n.forEachSlackClient(func(c *SlackClient) {
		if found != "" {
			return
		}
		if ref, ok := c.GetThreadRef(alertKey); ok {
			found = ref
		}
	})
	if found != "" {
		return found, true
	}
	// 자체 threadStore 확인
	if val, ok := n.threadStore.Load(alertKey); ok {
		if ref, ok := val.(string); ok {
			return ref, true
		}
	}
	// fallback
	if n.fallbackThreadStore != nil {
		return n.fallbackThreadStore.GetThreadRef(alertKey)
	}
	return "", false
}

func (n *webhookRoutingNotifier) DeleteThreadRef(alertKey string) {
	n.threadStore.Delete(alertKey)
	n.forEachSlackClient(func(c *SlackClient) {
		c.DeleteThreadRef(alertKey)
	})
	if n.fallbackThreadStore != nil {
		n.fallbackThreadStore.DeleteThreadRef(alertKey)
	}
}

func (n *webhookRoutingNotifier) RequiresThreadRef() bool {
	if n.cfgSource == nil {
		return n.fallbackThreadStore != nil
	}
	configs, err := n.cfgSource.GetWebhookConfigs(context.Background())
	if err != nil {
		return n.fallbackThreadStore != nil
	}
	for _, cfg := range configs {
		if normalizeWebhookType(cfg.Type) == "slack" {
			if _, ok := n.slackClientForConfig(cfg); ok {
				return true
			}
		}
	}
	return false
}

func (n *webhookRoutingNotifier) resolveNotifiers(severity string) ([]Notifier, error) {
	if n.cfgSource == nil {
		return nil, nil
	}

	configs, err := n.cfgSource.GetWebhookConfigs(context.Background())
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, nil
	}

	// 어떤 웹훅이든 명시적으로 claim한 severity 집합 계산.
	// severity가 없는 이벤트(FlappingCleared 등)는 이 로직을 건너뜀.
	claimed := make(map[string]bool)
	if severity != "" {
		for _, cfg := range configs {
			for _, s := range cfg.Severities {
				claimed[s] = true
			}
		}
	}

	targets := make([]Notifier, 0, len(configs))
	for _, cfg := range configs {
		var matches bool
		if len(cfg.Severities) == 0 {
			// 빈 배열(catch-all): 아무도 claim하지 않은 severity만 수신.
			// severity == ""인 이벤트(FlappingCleared, AnalysisResult)는 항상 수신.
			matches = severity == "" || !claimed[severity]
		} else {
			matches = severityMatches(cfg.Severities, severity)
		}
		if !matches {
			continue
		}
		switch normalizeWebhookType(cfg.Type) {
		case "slack":
			slackNotifier, ok := n.slackClientForConfig(cfg)
			if !ok {
				continue
			}
			targets = append(targets, slackNotifier)
		case "http", "teams":
			if strings.TrimSpace(cfg.URL) == "" {
				continue
			}
			targets = append(targets, &webhookEndpointNotifier{
				cfg:        cfg,
				httpClient: n.httpClient,
			})
		default:
			continue
		}
	}

	return targets, nil
}

func (n *webhookRoutingNotifier) slackClientForConfig(cfg model.WebhookConfig) (*SlackClient, bool) {
	token := strings.TrimSpace(cfg.Token)
	channel := strings.TrimSpace(cfg.Channel)
	if token == "" || channel == "" {
		return nil, false
	}

	n.mu.RLock()
	existing, ok := n.slackClients[cfg.ID]
	n.mu.RUnlock()

	if ok && existing.botToken == token && existing.channelID == channel {
		return existing, true
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if existing, ok := n.slackClients[cfg.ID]; ok {
		if existing.botToken == token && existing.channelID == channel {
			return existing, true
		}
	}

	clientCfg := config.SlackConfig{
		BotToken:    token,
		ChannelID:   channel,
		FrontendURL: n.frontendURL,
	}
	newClient := NewSlackClient(clientCfg)
	n.slackClients[cfg.ID] = newClient
	return newClient, true
}

func normalizeWebhookType(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

type webhookEndpointNotifier struct {
	cfg        model.WebhookConfig
	httpClient *http.Client
}

func (n *webhookEndpointNotifier) Notify(event NotifierEvent) error {
	payload, err := json.Marshal(webhookEventEnvelope{
		EventType: event.EventType(),
		SentAt:    time.Now().UTC(),
		Data:      event,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", n.cfg.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(n.cfg.Token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	return nil
}
