package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

type webhookConfigRepoStub struct {
	configs []model.WebhookConfig
	err     error
}

func (s webhookConfigRepoStub) GetWebhookConfigs(_ context.Context) ([]model.WebhookConfig, error) {
	return s.configs, s.err
}

type fallbackNotifierStub struct {
	mu          sync.Mutex
	notifyCount int
	store       map[string]string
}

func (f *fallbackNotifierStub) Notify(_ NotifierEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifyCount++
	return nil
}

func (f *fallbackNotifierStub) StoreThreadRef(alertKey, threadRef string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.store == nil {
		f.store = make(map[string]string)
	}
	f.store[alertKey] = threadRef
}

func (f *fallbackNotifierStub) GetThreadRef(alertKey string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.store == nil {
		return "", false
	}
	val, ok := f.store[alertKey]
	return val, ok
}

func (f *fallbackNotifierStub) DeleteThreadRef(alertKey string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.store == nil {
		return
	}
	delete(f.store, alertKey)
}

func TestWebhookRoutingNotifier_StrictThreadRoutingSkipsFallbackWhenNoWebhookConfig(t *testing.T) {
	repo := webhookConfigRepoStub{}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")

	err := n.Notify(AnalysisResultPostedEvent{ThreadRef: "t1", Content: "analysis"})
	if err == nil {
		t.Fatal("Notify() error = nil, want strict routing error")
	}
	if fallback.notifyCount != 0 {
		t.Fatalf("fallback notify count = %d, want 0", fallback.notifyCount)
	}
}

func TestWebhookRoutingNotifier_SendsHTTPWebhook(t *testing.T) {
	var gotType string
	var gotAuth string

	repo := webhookConfigRepoStub{
		configs: []model.WebhookConfig{
			{
				ID:    1,
				Type:  "http",
				URL:   "https://example.com/webhook",
				Token: "token-1",
			},
		},
	}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")
	impl, ok := n.(*webhookRoutingNotifier)
	if !ok {
		t.Fatalf("notifier type = %T, want *webhookRoutingNotifier", n)
	}
	impl.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAuth = req.Header.Get("Authorization")
			defer req.Body.Close()

			body, _ := io.ReadAll(req.Body)
			var payload struct {
				EventType string `json:"event_type"`
			}
			_ = json.Unmarshal(body, &payload)
			gotType = payload.EventType

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	// AlertStatusChangedEvent는 resolveNotifiers 경로로 HTTP 타겟에 도달한다.
	// (FlappingClearedEvent는 notifyByThread 경로를 사용하므로 Slack 전용)
	err := n.Notify(AlertStatusChangedEvent{
		Alert: model.Alert{
			Status: "firing",
			Labels: map[string]string{"severity": "warning", "alertname": "test"},
		},
	})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if gotType != NotifierEventAlertStatusChanged {
		t.Fatalf("event type = %q, want %q", gotType, NotifierEventAlertStatusChanged)
	}
	if gotAuth != "Bearer token-1" {
		t.Fatalf("authorization = %q, want %q", gotAuth, "Bearer token-1")
	}
	if fallback.notifyCount != 0 {
		t.Fatalf("fallback notify count = %d, want 0", fallback.notifyCount)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWebhookRoutingNotifier_ThreadRefStoreFallsBackToThreadStore(t *testing.T) {
	repo := webhookConfigRepoStub{
		configs: []model.WebhookConfig{
			{ID: 1, Type: "http", URL: "https://example.com/webhook"},
		},
	}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")

	n.StoreThreadRef("alert-1", "thread-1")
	// HTTP-only 타겟이라도 fallbackThreadStore가 thread 지원하면 조회 가능
	ref, ok := n.GetThreadRef("alert-1")
	if !ok {
		t.Fatal("GetThreadRef() ok = false, want true (fallback thread store)")
	}
	if ref != "thread-1" {
		t.Fatalf("GetThreadRef() = %q, want %q", ref, "thread-1")
	}
}

func TestWebhookRoutingNotifier_NotifyRootWithReceipts_ReturnsSlackReceipt(t *testing.T) {
	repo := webhookConfigRepoStub{
		configs: []model.WebhookConfig{
			{
				ID:         1,
				Type:       "slack",
				Token:      "token-1",
				Channel:    "C123",
				Severities: []string{"warning"},
			},
		},
	}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")
	impl := n.(*webhookRoutingNotifier)
	impl.slackClients[1] = NewSlackClient(config.SlackConfig{
		BotToken:  "token-1",
		ChannelID: "C123",
	})
	impl.slackClients[1].httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"1712345678.000100"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	receipts, err := n.NotifyRootWithReceipts(AlertStatusChangedEvent{
		Alert: model.Alert{
			Status: "firing",
			Labels: map[string]string{
				"severity":  "warning",
				"alertname": "test",
			},
		},
	})
	if err != nil {
		t.Fatalf("NotifyRootWithReceipts() error = %v", err)
	}
	if len(receipts) != 1 {
		t.Fatalf("receipt count = %d, want 1", len(receipts))
	}
	if receipts[0].WebhookConfigID == nil || *receipts[0].WebhookConfigID != 1 {
		t.Fatalf("receipt webhook_config_id = %v, want 1", receipts[0].WebhookConfigID)
	}
	if receipts[0].ChannelID != "C123" {
		t.Fatalf("receipt channel_id = %q, want %q", receipts[0].ChannelID, "C123")
	}
	if receipts[0].ThreadTS != "1712345678.000100" {
		t.Fatalf("receipt thread_ts = %q, want %q", receipts[0].ThreadTS, "1712345678.000100")
	}
}

func TestWebhookRoutingNotifier_NotifyThreadEvent_UsesExplicitDelivery(t *testing.T) {
	repo := webhookConfigRepoStub{
		configs: []model.WebhookConfig{
			{
				ID:      7,
				Type:    "slack",
				Token:   "token-7",
				Channel: "C999",
			},
		},
	}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")
	impl := n.(*webhookRoutingNotifier)
	impl.slackClients[7] = NewSlackClient(config.SlackConfig{
		BotToken:  "token-7",
		ChannelID: "C999",
	})

	var captured struct {
		Channel  string `json:"channel"`
		ThreadTS string `json:"thread_ts"`
	}
	impl.slackClients[7].httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			defer req.Body.Close()
			body, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(body, &captured)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"1712345678.000200"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	configID := 7
	err := n.NotifyThreadEvent(
		AnalysisResultPostedEvent{Content: "analysis"},
		[]model.AlertNotificationDelivery{
			{
				AlertID:         "ALR-1",
				Fingerprint:     "fp-1",
				NotifierType:    "slack",
				WebhookConfigID: &configID,
				RouteKey:        model.BuildNotificationRouteKey("slack", &configID, "C777"),
				ChannelID:       "C777",
				ThreadTS:        "1712345678.000123",
				RootMessageTS:   "1712345678.000123",
				IsActive:        true,
			},
		},
	)
	if err != nil {
		t.Fatalf("NotifyThreadEvent() error = %v", err)
	}
	if captured.Channel != "C777" {
		t.Fatalf("channel = %q, want %q", captured.Channel, "C777")
	}
	if captured.ThreadTS != "1712345678.000123" {
		t.Fatalf("thread_ts = %q, want %q", captured.ThreadTS, "1712345678.000123")
	}
}

func TestWebhookRoutingNotifier_Notify_UsesFallbackSlackWhenThreadKnown(t *testing.T) {
	fallback := NewSlackClient(config.SlackConfig{
		BotToken:  "token-fallback",
		ChannelID: "C-fallback",
	})
	var captured struct {
		Channel  string `json:"channel"`
		ThreadTS string `json:"thread_ts"`
	}
	fallback.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			defer req.Body.Close()
			body, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(body, &captured)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"1712345678.000300"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	fallback.StoreThreadRef("fp-1", "1712345678.000111")

	n := NewWebhookRoutingNotifier(webhookConfigRepoStub{}, fallback, fallback, "")
	err := n.Notify(AnalysisResultPostedEvent{
		ThreadRef: "1712345678.000111",
		Content:   "analysis",
	})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if captured.Channel != "C-fallback" {
		t.Fatalf("channel = %q, want %q", captured.Channel, "C-fallback")
	}
	if captured.ThreadTS != "1712345678.000111" {
		t.Fatalf("thread_ts = %q, want %q", captured.ThreadTS, "1712345678.000111")
	}
}
