package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
)

// ---------------------------------------------------------------------------
// Stub for webhookHandlerService (narrow interface matching handler layer)
// ---------------------------------------------------------------------------

// webhookHandlerService allows injecting a fake into testAlertHandler below.
type webhookHandlerService interface {
	ProcessWebhook(webhook model.AlertmanagerWebhook) (sent, failed int)
}

// testAlertHandler mirrors AlertHandler.Webhook so we can drive the HTTP layer
// without a real *service.AlertService. The production Webhook handler logic is
// also tested via the actual AlertHandler in TestAlertHandler_* tests.
type testAlertHandler struct {
	svc webhookHandlerService
}

func (h *testAlertHandler) webhook(c *gin.Context) {
	var webhook model.AlertmanagerWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	sent, failed := h.svc.ProcessWebhook(webhook)
	c.JSON(http.StatusOK, model.AlertWebhookResponse{
		Status:      "received",
		AlertCount:  len(webhook.Alerts),
		SlackSent:   sent,
		SlackFailed: failed,
	})
}

type fakeWebhookSvc struct{ sent, failed int }

func (f *fakeWebhookSvc) ProcessWebhook(_ model.AlertmanagerWebhook) (int, int) {
	return f.sent, f.failed
}

func newWebhookRouter(svc webhookHandlerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &testAlertHandler{svc: svc}
	r.POST("/webhook/alertmanager", h.webhook)
	return r
}

// validWebhookPayload returns a minimal valid Alertmanager payload.
func validWebhookPayload(status string, alertCount int) []byte {
	alerts := make([]model.Alert, alertCount)
	for i := range alerts {
		alerts[i] = model.Alert{
			Status:      status,
			Fingerprint: "abc123",
			Labels: map[string]string{
				"alertname": "TestAlert",
				"severity":  "warning",
				"namespace": "default",
			},
		}
	}
	body, _ := json.Marshal(model.AlertmanagerWebhook{
		Status:   status,
		Receiver: "test-receiver",
		Alerts:   alerts,
	})
	return body
}

// ---------------------------------------------------------------------------
// Production AlertHandler.Webhook tests (actual production code path)
// ---------------------------------------------------------------------------

// newProductionWebhookRouter wires the real AlertHandler with a nil service so
// only the JSON-decode path (400) is exercised without a DB dependency.
func newProductionWebhookRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// NewAlertHandler accepts nil safely; Webhook panics only if ProcessWebhook
	// is reached on a nil service. We only test the bad-payload (400) path here.
	h := NewAlertHandler(nil)
	r.POST("/webhook/alertmanager", h.Webhook)
	return r
}

func TestAlertHandler_Webhook_BadPayload_Returns400(t *testing.T) {
	// TODO: re-enable after W2 HMAC middleware merged
	t.Parallel()
	router := newProductionWebhookRouter()

	cases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"not json", "this is not json"},
		{"truncated json", `{"version":"`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/webhook/alertmanager",
				strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d (body=%q)", w.Code, tc.body)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Stub-based webhook tests (additional logic coverage)
// ---------------------------------------------------------------------------

func TestWebhookHandler_ValidPayload_Returns200(t *testing.T) {
	// TODO: re-enable after W2 HMAC middleware merged
	t.Parallel()
	router := newWebhookRouter(&fakeWebhookSvc{sent: 1, failed: 0})

	body := validWebhookPayload("firing", 2)
	req := httptest.NewRequest(http.MethodPost, "/webhook/alertmanager",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.AlertWebhookResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "received" {
		t.Errorf("status = %q, want %q", resp.Status, "received")
	}
	if resp.AlertCount != 2 {
		t.Errorf("alertCount = %d, want 2", resp.AlertCount)
	}
	if resp.SlackSent != 1 {
		t.Errorf("slackSent = %d, want 1", resp.SlackSent)
	}
}

func TestWebhookHandler_FiringAndResolved_AlertCount(t *testing.T) {
	// TODO: re-enable after W2 HMAC middleware merged
	t.Parallel()
	cases := []struct {
		name   string
		status string
		count  int
		sent   int
	}{
		{"firing single", "firing", 1, 1},
		{"resolved zero", "resolved", 0, 0},
		{"firing three", "firing", 3, 3},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			router := newWebhookRouter(&fakeWebhookSvc{sent: tc.sent})
			body := validWebhookPayload(tc.status, tc.count)
			req := httptest.NewRequest(http.MethodPost, "/webhook/alertmanager",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", w.Code)
			}
			var resp model.AlertWebhookResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.AlertCount != tc.count {
				t.Errorf("alertCount = %d, want %d", resp.AlertCount, tc.count)
			}
		})
	}
}
