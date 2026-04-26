package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// incidentsRouter wires the real RcaHandler routes for handler-level tests.
// Services are nil; only validation paths (400/401) that return before calling
// service are exercised — any path that reaches h.svc.* would panic.
func incidentsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRcaHandler(nil, nil)

	// Incidents
	r.PUT("/api/v1/incidents/:id", h.UpdateIncident)
	r.PATCH("/api/v1/incidents/:id", h.HideIncident)
	r.PATCH("/api/v1/incidents/:id/unhide", h.UnhideIncident)
	r.POST("/api/v1/incidents/:id/resolve", h.ResolveIncident)

	// Feedback — all require auth; nil auth user → 401
	r.GET("/api/v1/incidents/:id/feedback", h.GetIncidentFeedback)
	r.POST("/api/v1/incidents/:id/comments", h.CreateIncidentComment)
	r.PUT("/api/v1/incidents/:id/comments/:commentId", h.UpdateIncidentComment)
	r.DELETE("/api/v1/incidents/:id/comments/:commentId", h.DeleteIncidentComment)
	r.POST("/api/v1/incidents/:id/vote", h.VoteIncidentFeedback)

	// Alert feedback
	r.GET("/api/v1/alerts/:id/feedback", h.GetAlertFeedback)
	r.POST("/api/v1/alerts/:id/comments", h.CreateAlertComment)
	r.PUT("/api/v1/alerts/:id/comments/:commentId", h.UpdateAlertComment)
	r.DELETE("/api/v1/alerts/:id/comments/:commentId", h.DeleteAlertComment)
	r.POST("/api/v1/alerts/:id/vote", h.VoteAlertFeedback)

	// Bulk resolve
	r.POST("/api/v1/alerts/bulk-resolve", h.BulkResolveAlerts)

	return r
}

// ---------------------------------------------------------------------------
// UpdateIncident validation
// ---------------------------------------------------------------------------

func TestIncidentHandler_UpdateIncident_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/incidents/inc-1",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// BulkResolveAlerts validation
// ---------------------------------------------------------------------------

func TestIncidentHandler_BulkResolveAlerts_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/bulk-resolve",
		strings.NewReader("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIncidentHandler_BulkResolveAlerts_EmptyIDs_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/bulk-resolve",
		strings.NewReader(`{"alert_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Feedback / Comment — no auth user → 401
// ---------------------------------------------------------------------------

func TestIncidentHandler_GetFeedback_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1/feedback", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestIncidentHandler_CreateComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/comments",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestIncidentHandler_UpdateComment_BadCommentID_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/incidents/inc-1/comments/notanint",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIncidentHandler_UpdateComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/incidents/inc-1/comments/42",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestIncidentHandler_DeleteComment_BadCommentID_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/inc-1/comments/notanint", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIncidentHandler_DeleteComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/inc-1/comments/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestIncidentHandler_VoteFeedback_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/vote",
		strings.NewReader(`{"vote_type":"up"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Alert feedback — no auth user → 401
// ---------------------------------------------------------------------------

func TestAlertHandler_GetFeedback_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/a1/feedback", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAlertHandler_CreateComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/a1/comments",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAlertHandler_UpdateComment_BadCommentID_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts/a1/comments/notanint",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAlertHandler_UpdateComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts/a1/comments/42",
		strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAlertHandler_DeleteComment_BadCommentID_Returns400(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/a1/comments/notanint", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAlertHandler_DeleteComment_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/a1/comments/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAlertHandler_VoteFeedback_Unauthorized(t *testing.T) {
	t.Parallel()
	r := incidentsRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/a1/vote",
		strings.NewReader(`{"vote_type":"down"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// SSEAuthMiddleware paths (production code)
// ---------------------------------------------------------------------------

func TestSSEAuthMiddleware_NoTokenOrHeader_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	r := gin.New()
	r.Use(SSEAuthMiddleware(svc))
	r.GET("/events", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSSEAuthMiddleware_InvalidQueryToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	r := gin.New()
	r.Use(SSEAuthMiddleware(svc))
	r.GET("/events", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/events?token=bad.token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSSEAuthMiddleware_OptionsPassthrough(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	r := gin.New()
	r.Use(SSEAuthMiddleware(svc))
	r.OPTIONS("/events", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodOptions, "/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// UpdatePreferredLanguage — no auth user → 401
// ---------------------------------------------------------------------------

func TestAuthHandler_UpdatePreferredLanguage_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	h := NewAuthHandler(svc)
	r := gin.New()
	r.PUT("/api/v1/auth/language", h.UpdatePreferredLanguage)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/language",
		strings.NewReader(`{"preferredLanguage":"en"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
