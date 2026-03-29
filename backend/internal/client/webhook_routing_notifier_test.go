package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

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

func TestWebhookRoutingNotifier_UsesFallbackWhenNoWebhookConfig(t *testing.T) {
	repo := webhookConfigRepoStub{}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")

	err := n.Notify(AnalysisResultPostedEvent{ThreadRef: "t1", Content: "analysis"})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if fallback.notifyCount != 1 {
		t.Fatalf("fallback notify count = %d, want 1", fallback.notifyCount)
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
