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
	targets, err := n.resolveNotifiers()
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

func (n *webhookRoutingNotifier) StoreThreadRef(alertKey, threadRef string) {
	store := n.resolveThreadStore()
	if store == nil {
		return
	}
	store.StoreThreadRef(alertKey, threadRef)
}

func (n *webhookRoutingNotifier) GetThreadRef(alertKey string) (string, bool) {
	store := n.resolveThreadStore()
	if store == nil {
		return "", false
	}
	return store.GetThreadRef(alertKey)
}

func (n *webhookRoutingNotifier) DeleteThreadRef(alertKey string) {
	store := n.resolveThreadStore()
	if store == nil {
		return
	}
	store.DeleteThreadRef(alertKey)
}

func (n *webhookRoutingNotifier) RequiresThreadRef() bool {
	return n.resolveThreadStore() != nil
}

func (n *webhookRoutingNotifier) resolveNotifiers() ([]Notifier, error) {
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

	targets := make([]Notifier, 0, len(configs))
	for _, cfg := range configs {
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

func (n *webhookRoutingNotifier) resolveThreadStore() ThreadRefStore {
	if n.cfgSource == nil {
		return n.fallbackThreadStore
	}

	configs, err := n.cfgSource.GetWebhookConfigs(context.Background())
	if err != nil {
		log.Printf("Failed to load webhook configs for thread store: %v", err)
		return n.fallbackThreadStore
	}
	if len(configs) == 0 {
		return n.fallbackThreadStore
	}

	for _, cfg := range configs {
		if normalizeWebhookType(cfg.Type) != "slack" {
			continue
		}
		slackNotifier, ok := n.slackClientForConfig(cfg)
		if !ok {
			continue
		}
		return slackNotifier
	}
	return nil
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
