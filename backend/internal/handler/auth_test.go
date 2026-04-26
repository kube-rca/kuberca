package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

// ---------------------------------------------------------------------------
// Helper: build a minimal AuthService with no real DB (for config/logout paths
// that don't touch the repo, and for JWT parsing tests).
// ---------------------------------------------------------------------------

func mustAuthService(t *testing.T, allowSignup bool) *service.AuthService {
	t.Helper()
	signup := "false"
	if allowSignup {
		signup = "true"
	}
	svc, err := service.NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "test-secret-32-bytes-padding-ok!",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		AllowSignup:    signup,
		CookieSecure:   "false",
		CookieSameSite: "Lax",
		CookiePath:     "/",
	})
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}
	return svc
}

// ---------------------------------------------------------------------------
// AuthHandler.Config
// ---------------------------------------------------------------------------

func TestAuthHandler_Config_OIDCDisabled(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, true)
	h := NewAuthHandler(svc)
	h.SetOIDCConfig(false, "", "")

	r := gin.New()
	r.GET("/api/v1/auth/config", h.Config)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.AuthConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.OIDCEnabled {
		t.Error("oidcEnabled should be false")
	}
	if !resp.AllowSignup {
		t.Error("allowSignup should be true")
	}
}

func TestAuthHandler_Config_OIDCEnabled(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	h := NewAuthHandler(svc)
	h.SetOIDCConfig(true, "https://accounts.example.com/auth", "google")

	r := gin.New()
	r.GET("/api/v1/auth/config", h.Config)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.AuthConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.OIDCEnabled {
		t.Error("oidcEnabled should be true")
	}
	if resp.OIDCProvider != "google" {
		t.Errorf("oidcProvider = %q, want %q", resp.OIDCProvider, "google")
	}
}

// ---------------------------------------------------------------------------
// AuthHandler.Login - bad-JSON path (no DB needed)
// ---------------------------------------------------------------------------

func TestAuthHandler_Login_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	h := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/api/v1/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// AuthHandler.Register - bad-JSON path (no DB needed)
// ---------------------------------------------------------------------------

func TestAuthHandler_Register_BadJSON_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, true)
	h := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// AuthHandler.Logout - no DB needed (clears cookie)
// ---------------------------------------------------------------------------

func TestAuthHandler_Logout_Returns200(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	h := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/api/v1/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.AuthLogoutResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "logged_out" {
		t.Errorf("status = %q, want %q", resp.Status, "logged_out")
	}
}

// ---------------------------------------------------------------------------
// AuthHandler.Me - unauthorized path (no token set)
// ---------------------------------------------------------------------------

func TestAuthHandler_Me_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)
	h := NewAuthHandler(svc)

	r := gin.New()
	r.GET("/api/v1/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// AuthMiddleware unit tests (production code)
// ---------------------------------------------------------------------------

func TestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)

	r := gin.New()
	r.Use(AuthMiddleware(svc))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_EmptyBearer_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)

	r := gin.New()
	r.Use(AuthMiddleware(svc))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)

	r := gin.New()
	r.Use(AuthMiddleware(svc))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_OptionsPassthrough(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := mustAuthService(t, false)

	r := gin.New()
	r.Use(AuthMiddleware(svc))
	r.OPTIONS("/preflight", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/preflight", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// CORSMiddleware unit tests (production code)
// ---------------------------------------------------------------------------

func TestCORSMiddleware_AllowedOrigin_SetsHeaders(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware([]string{"https://app.example.com"}, true))
	r.GET("/api", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO header = %q, want %q", got, "https://app.example.com")
	}
}

func TestCORSMiddleware_DisallowedOrigin_NoHeader(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware([]string{"https://app.example.com"}, true))
	r.GET("/api", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO header should be empty for disallowed origin, got %q", got)
	}
}

func TestCORSMiddleware_PreflightOptions(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware([]string{"https://app.example.com"}, false))
	// Register OPTIONS handler so gin doesn't 404
	r.OPTIONS("/api", func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// WebhookSettingsHandler (production code — no DB needed for validation paths)
// ---------------------------------------------------------------------------

func TestWebhookSettingsHandler_CreateBlankName_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	// webhookService interface is in handler.go — we can instantiate with a nil
	// service; the handler validates request before calling the service.
	r := gin.New()
	h := NewWebhookSettingsHandler(nil)
	r.POST("/api/v1/settings/webhooks", h.CreateWebhookConfig)

	body := `{"name":"   ","url":"https://example.com/hook","type":"slack"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/webhooks",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWebhookSettingsHandler_UpdateBlankName_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWebhookSettingsHandler(nil)
	r.PUT("/api/v1/settings/webhooks/:id", h.UpdateWebhookConfig)

	body := `{"name":"   ","url":"https://example.com/hook","type":"slack"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/webhooks/1",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWebhookSettingsHandler_GetBadID_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWebhookSettingsHandler(nil)
	r.GET("/api/v1/settings/webhooks/:id", h.GetWebhookConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/webhooks/notanint", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWebhookSettingsHandler_DeleteBadID_Returns400(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWebhookSettingsHandler(nil)
	r.DELETE("/api/v1/settings/webhooks/:id", h.DeleteWebhookConfig)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/settings/webhooks/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
