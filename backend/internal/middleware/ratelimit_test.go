package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddleware_AllowsBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(RateLimitConfig{RequestsPerMinute: 60, Burst: 3}))
	router.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("burst request %d: got %d, want 200", i+1, rec.Code)
		}
	}
}

func TestRateLimitMiddleware_RejectsOverBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// 1 request per minute with burst 1 -> first request OK, immediate second is rejected.
	router.Use(RateLimitMiddleware(RateLimitConfig{RequestsPerMinute: 1, Burst: 1}))
	router.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	first := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	req1.RemoteAddr = "10.0.0.2:1234"
	router.ServeHTTP(first, req1)
	if first.Code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", first.Code)
	}

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	router.ServeHTTP(second, req2)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: got %d, want 429 (body=%s)", second.Code, second.Body.String())
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header on 429 response")
	}
}

func TestRateLimitMiddleware_PerIPIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(RateLimitConfig{RequestsPerMinute: 1, Burst: 1}))
	router.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for _, ip := range []string{"10.0.0.3:1", "10.0.0.4:1", "10.0.0.5:1"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("ip %s: got %d, want 200", ip, rec.Code)
		}
	}
}
