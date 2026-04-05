package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

// --- isRetryableAgentError tests ---

func TestIsRetryableAgentError_Nil(t *testing.T) {
	if isRetryableAgentError(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryableAgentError_5xx(t *testing.T) {
	for _, code := range []int{500, 502, 503, 504} {
		err := &AgentError{StatusCode: code, Message: "server error"}
		if !isRetryableAgentError(err) {
			t.Errorf("status %d should be retryable", code)
		}
	}
}

func TestIsRetryableAgentError_4xx(t *testing.T) {
	for _, code := range []int{400, 401, 404, 422} {
		err := &AgentError{StatusCode: code, Message: "client error"}
		if isRetryableAgentError(err) {
			t.Errorf("status %d should not be retryable", code)
		}
	}
}

func TestIsRetryableAgentError_NetworkError(t *testing.T) {
	err := errors.New("dial tcp: connection refused")
	if !isRetryableAgentError(err) {
		t.Error("network error should be retryable")
	}
}

type timeoutErr struct{}

func (e *timeoutErr) Error() string { return "timeout" }
func (e *timeoutErr) Timeout() bool { return true }

func TestIsRetryableAgentError_Timeout(t *testing.T) {
	err := &timeoutErr{}
	if isRetryableAgentError(err) {
		t.Error("timeout error should not be retryable")
	}
}

// --- RequestAnalysis retry integration tests ---

func newTestClient(url string, retryMaxAttempts, baseBackoffMs, maxBackoffMs int) *AgentClient {
	cfg := config.AgentConfig{
		BaseURL:              url,
		HTTPTimeoutSeconds:   5,
		RetryMaxAttempts:     retryMaxAttempts,
		RetryBaseBackoffSecs: 0, // will be overridden below
		RetryMaxBackoffSecs:  0,
	}
	c := NewAgentClient(cfg)
	// Override to millisecond-level backoffs for fast tests
	c.retryBaseBackoff = time.Duration(baseBackoffMs) * time.Millisecond
	c.retryMaxBackoff = time.Duration(maxBackoffMs) * time.Millisecond
	return c
}

func dummyAlert() model.Alert {
	return model.Alert{
		Fingerprint: "test-fp",
		Status:      "firing",
		Labels:      map[string]string{"alertname": "TestAlert"},
	}
}

func TestRequestAnalysis_RetriesOn500(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}
		resp := AgentAnalysisResponse{
			Status:          "ok",
			AnalysisSummary: "test summary",
			AnalysisDetail:  "test detail",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server.URL, 3, 10, 50)
	resp, err := client.RequestAnalysis(dummyAlert(), "", "", "firing", nil)

	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if resp.AnalysisSummary != "test summary" {
		t.Errorf("unexpected summary: %s", resp.AnalysisSummary)
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}

func TestRequestAnalysis_NoRetryOn400(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := newTestClient(server.URL, 3, 10, 50)
	_, err := client.RequestAnalysis(dummyAlert(), "", "", "firing", nil)

	if err == nil {
		t.Fatal("expected error for 400")
	}
	var agentErr *AgentError
	if !errors.As(err, &agentErr) {
		t.Fatalf("expected AgentError, got: %T", err)
	}
	if agentErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", agentErr.StatusCode)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call (no retry for 4xx), got %d", callCount.Load())
	}
}

func TestRequestAnalysis_ExhaustsRetries(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("always failing"))
	}))
	defer server.Close()

	client := newTestClient(server.URL, 3, 10, 50)
	_, err := client.RequestAnalysis(dummyAlert(), "", "", "firing", nil)

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}
