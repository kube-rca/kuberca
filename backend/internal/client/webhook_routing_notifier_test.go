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

	err := n.Notify(AnalysisResultPostedEvent{Content: "analysis"})
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

	err := n.Notify(FlappingClearedEvent{Fingerprint: "fp-1", ThreadRef: "t1"})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if gotType != NotifierEventFlappingCleared {
		t.Fatalf("event type = %q, want %q", gotType, NotifierEventFlappingCleared)
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

func TestWebhookRoutingNotifier_ThreadRefStoreDisabledForThreadlessTargets(t *testing.T) {
	repo := webhookConfigRepoStub{
		configs: []model.WebhookConfig{
			{ID: 1, Type: "http", URL: "https://example.com/webhook"},
		},
	}
	fallback := &fallbackNotifierStub{}
	n := NewWebhookRoutingNotifier(repo, fallback, fallback, "")

	n.StoreThreadRef("alert-1", "thread-1")
	if _, ok := n.GetThreadRef("alert-1"); ok {
		t.Fatal("GetThreadRef() ok = true, want false for threadless targets")
	}
}
