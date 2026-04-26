package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// RcaHandler tests
//
// RcaHandler holds *service.RcaService and *service.AlertService — concrete
// types that require a real DB for most operations.  We exercise only the
// validation paths that return early (400) before any service call.
// ---------------------------------------------------------------------------

func newRcaRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Pass nil services — only bad-input paths are tested (no service call).
	h := NewRcaHandler(nil, nil)
	r.POST("/api/v1/incidents/:id/resolve", h.ResolveIncident)
	r.PUT("/api/v1/incidents/:id", h.UpdateIncident)
	r.POST("/api/v1/alerts/bulk-resolve", h.BulkResolveAlerts)
	return r
}

// ---------------------------------------------------------------------------
// ResolveIncident — bad JSON bind (returns early with bad resolvedBy → 200
// because handler accepts empty resolvedBy gracefully). We test that the
// handler does NOT crash and handles the path. Since nil svc panics on
// service call, only paths that bypass the service are exercised.
//
// Note: The handler's ShouldBindJSON failure resets req.ResolvedBy to ""; it
// does NOT return 400 — it falls through to the service call. With a nil svc
// this would panic, so we skip that test and focus on other handlers.
// ---------------------------------------------------------------------------

// TestUpdateIncident_BadJSON_Returns400 exercises UpdateIncident's JSON parse gate.
func TestUpdateIncident_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	router := newRcaRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/incidents/inc-1",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestBulkResolveAlerts_BadJSON_Returns400 exercises BulkResolveAlerts JSON gate.
func TestBulkResolveAlerts_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	router := newRcaRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/bulk-resolve",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestBulkResolveAlerts_EmptyIDs_Returns400 exercises the empty-list check.
func TestBulkResolveAlerts_EmptyIDs_Returns400(t *testing.T) {
	t.Parallel()
	router := newRcaRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/bulk-resolve",
		strings.NewReader(`{"alert_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// ChatHandler — uses concrete *service.ChatService; we test bad-JSON path.
// ---------------------------------------------------------------------------

func newChatHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewChatHandler(nil) // nil service — only JSON decode path tested
	r.POST("/api/v1/chat/:incidentId", h.Chat)
	return r
}

func TestChatHandler_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	router := newChatHandlerRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/inc-1",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
