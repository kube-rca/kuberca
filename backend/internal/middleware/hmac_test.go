package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// signBody returns the canonical hex-encoded HMAC-SHA256 of body.
func signBody(t *testing.T, secret, body string) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

func newTestRouter(cfg HMACConfig) *gin.Engine {
	r := gin.New()
	r.POST("/webhook", HMACMiddleware(cfg), func(c *gin.Context) {
		// Drain the body to verify the middleware restored it for handlers.
		body, _ := io.ReadAll(c.Request.Body)
		c.JSON(http.StatusOK, gin.H{"received": len(body)})
	})
	return r
}

func TestHMACMiddleware(t *testing.T) {
	const secret = "super-secret-key"
	const body = `{"alerts":[{"status":"firing"}]}`

	tests := []struct {
		name        string
		cfg         HMACConfig
		body        string
		header      string
		signature   string
		wantStatus  int
		wantBodyLen int
	}{
		{
			name:        "valid signature returns 200",
			cfg:         HMACConfig{Secret: secret},
			body:        body,
			header:      DefaultWebhookSignatureHeader,
			signature:   signBody(t, secret, body),
			wantStatus:  http.StatusOK,
			wantBodyLen: len(body),
		},
		{
			name:        "valid signature with sha256= prefix returns 200",
			cfg:         HMACConfig{Secret: secret},
			body:        body,
			header:      DefaultWebhookSignatureHeader,
			signature:   "sha256=" + signBody(t, secret, body),
			wantStatus:  http.StatusOK,
			wantBodyLen: len(body),
		},
		{
			name:       "invalid signature returns 401",
			cfg:        HMACConfig{Secret: secret},
			body:       body,
			header:     DefaultWebhookSignatureHeader,
			signature:  signBody(t, "different-secret", body),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing header with secret configured returns 401",
			cfg:        HMACConfig{Secret: secret},
			body:       body,
			header:     "",
			signature:  "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "missing header with empty secret allows request",
			cfg:         HMACConfig{Secret: ""},
			body:        body,
			header:      "",
			signature:   "",
			wantStatus:  http.StatusOK,
			wantBodyLen: len(body),
		},
		{
			name:       "tampered body returns 401",
			cfg:        HMACConfig{Secret: secret},
			body:       body + "tampered",
			header:     DefaultWebhookSignatureHeader,
			signature:  signBody(t, secret, body), // signature for original body
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-hex signature returns 401",
			cfg:        HMACConfig{Secret: secret},
			body:       body,
			header:     DefaultWebhookSignatureHeader,
			signature:  "not-hex-zzzz",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "custom header name accepted",
			cfg:         HMACConfig{Secret: secret, HeaderName: "X-Hub-Signature-256"},
			body:        body,
			header:      "X-Hub-Signature-256",
			signature:   "sha256=" + signBody(t, secret, body),
			wantStatus:  http.StatusOK,
			wantBodyLen: len(body),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(tc.cfg)
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.header != "" {
				req.Header.Set(tc.header, tc.signature)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d (body=%s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.wantStatus == http.StatusOK && tc.wantBodyLen > 0 {
				// Sanity check: handler saw the original body length.
				if !bytes.Contains(rec.Body.Bytes(), []byte(`"received"`)) {
					t.Fatalf("expected handler to drain body, got: %s", rec.Body.String())
				}
			}
		})
	}
}

func TestHMACMiddleware_BodyRestoredForHandler(t *testing.T) {
	const secret = "abc"
	const body = `{"hello":"world"}`

	router := gin.New()
	router.POST("/echo", HMACMiddleware(HMACConfig{Secret: secret}), func(c *gin.Context) {
		var payload map[string]string
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, payload)
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(DefaultWebhookSignatureHeader, signBody(t, secret, body))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"hello":"world"`)) {
		t.Fatalf("handler did not receive original body: %s", rec.Body.String())
	}
}
