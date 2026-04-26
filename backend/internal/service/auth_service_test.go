package service

import (
	"net/http"
	"testing"

	"github.com/kube-rca/backend/internal/config"
)

// ---------------------------------------------------------------------------
// parseBool
// ---------------------------------------------------------------------------

func TestParseBool_EmptyUsesDefault(t *testing.T) {
	t.Parallel()
	for _, dflt := range []bool{true, false} {
		got, err := parseBool("", dflt)
		if err != nil {
			t.Fatalf("parseBool(%q, %v) error = %v", "", dflt, err)
		}
		if got != dflt {
			t.Errorf("parseBool(%q, %v) = %v, want %v", "", dflt, got, dflt)
		}
	}
}

func TestParseBool_TrueValues(t *testing.T) {
	t.Parallel()
	for _, v := range []string{"true", "True", "TRUE", "1"} {
		got, err := parseBool(v, false)
		if err != nil {
			t.Fatalf("parseBool(%q) error = %v", v, err)
		}
		if !got {
			t.Errorf("parseBool(%q) = false, want true", v)
		}
	}
}

func TestParseBool_FalseValues(t *testing.T) {
	t.Parallel()
	for _, v := range []string{"false", "False", "FALSE", "0"} {
		got, err := parseBool(v, true)
		if err != nil {
			t.Fatalf("parseBool(%q) error = %v", v, err)
		}
		if got {
			t.Errorf("parseBool(%q) = true, want false", v)
		}
	}
}

func TestParseBool_InvalidReturnsError(t *testing.T) {
	t.Parallel()
	_, err := parseBool("yes", false)
	if err == nil {
		t.Error("parseBool('yes') should return an error")
	}
}

// ---------------------------------------------------------------------------
// parseSameSite
// ---------------------------------------------------------------------------

func TestParseSameSite_ValidValues(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  http.SameSite
	}{
		{"lax", http.SameSiteLaxMode},
		{"Lax", http.SameSiteLaxMode},
		{"strict", http.SameSiteStrictMode},
		{"Strict", http.SameSiteStrictMode},
		{"none", http.SameSiteNoneMode},
		{"None", http.SameSiteNoneMode},
		{"", http.SameSiteLaxMode}, // empty → default Lax
	}
	for _, tc := range cases {
		got, err := parseSameSite(tc.input)
		if err != nil {
			t.Errorf("parseSameSite(%q) error = %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseSameSite(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseSameSite_InvalidReturnsError(t *testing.T) {
	t.Parallel()
	_, err := parseSameSite("invalid")
	if err == nil {
		t.Error("parseSameSite('invalid') should return error")
	}
}

// ---------------------------------------------------------------------------
// validateCredentials
// ---------------------------------------------------------------------------

func TestValidateCredentials_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct{ id, pw string }{
		{"admin", "password1"},
		{"usr", "12345678"}, // minimum valid
		{"a-b-c", "p@ssword1"},
	}
	for _, tc := range cases {
		if err := validateCredentials(tc.id, tc.pw); err != nil {
			t.Errorf("validateCredentials(%q, %q) error = %v", tc.id, tc.pw, err)
		}
	}
}

func TestValidateCredentials_TooShortLoginID(t *testing.T) {
	t.Parallel()
	err := validateCredentials("ab", "password123")
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for short loginID, got %v", err)
	}
}

func TestValidateCredentials_TooShortPassword(t *testing.T) {
	t.Parallel()
	err := validateCredentials("validuser", "short")
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for short password, got %v", err)
	}
}

func TestValidateCredentials_TooLongLoginID(t *testing.T) {
	t.Parallel()
	longID := string(make([]byte, 65))
	for i := range longID {
		_ = i
	}
	id := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 67 chars
	err := validateCredentials(id, "password123")
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for loginID len=%d, got %v", len(id), err)
	}
}

// ---------------------------------------------------------------------------
// normalizePreferredLanguage
// ---------------------------------------------------------------------------

func TestNormalizePreferredLanguage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"en", "en"},
		{"EN", "en"},
		{"En", "en"},
		{"ko", "ko"},
		{"", "ko"},
		{"fr", "ko"},
		{"ja", "ko"},
	}
	for _, tc := range cases {
		got := normalizePreferredLanguage(tc.input)
		if got != tc.want {
			t.Errorf("normalizePreferredLanguage(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// NewAuthService — configuration validation
// ---------------------------------------------------------------------------

func TestNewAuthService_MissingJWTSecret_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET, got nil")
	}
}

func TestNewAuthService_InvalidAccessTTL_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "secret",
		JWTAccessTTL:   "not-a-duration",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err == nil {
		t.Fatal("expected error for invalid JWT_ACCESS_TTL, got nil")
	}
}

func TestNewAuthService_InvalidRefreshTTL_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "secret",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "not-valid",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err == nil {
		t.Fatal("expected error for invalid JWT_REFRESH_TTL, got nil")
	}
}

func TestNewAuthService_InvalidAllowSignup_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "secret",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		AllowSignup:    "yes-please", // invalid bool
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err == nil {
		t.Fatal("expected error for invalid ALLOW_SIGNUP, got nil")
	}
}

func TestNewAuthService_SameSiteNoneRequiresSecure(t *testing.T) {
	t.Parallel()
	_, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "secret",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "None", // requires Secure=true
	})
	if err == nil {
		t.Fatal("expected error for SameSite=None with Secure=false, got nil")
	}
}

func TestNewAuthService_ValidConfig_ReturnsService(t *testing.T) {
	t.Parallel()
	svc, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "test-secret-at-least-32-bytes-pad",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		AllowSignup:    "true",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
		CookiePath:     "/",
	})
	if err != nil {
		t.Fatalf("NewAuthService error = %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if !svc.AllowSignup() {
		t.Error("AllowSignup() should be true")
	}
}

// ---------------------------------------------------------------------------
// ParseAccessToken — invalid token returns error
// ---------------------------------------------------------------------------

func TestParseAccessToken_InvalidToken_ReturnsError(t *testing.T) {
	t.Parallel()
	svc, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "test-secret-at-least-32-bytes-pad",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	_, err = svc.ParseAccessToken("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseAccessToken_EmptyToken_ReturnsError(t *testing.T) {
	t.Parallel()
	svc, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "test-secret-at-least-32-bytes-pad",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "Lax",
	})
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	_, err = svc.ParseAccessToken("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

// ---------------------------------------------------------------------------
// CookieConfig — verify field population
// ---------------------------------------------------------------------------

func TestAuthService_CookieConfig_Populated(t *testing.T) {
	t.Parallel()
	svc, err := NewAuthService(nil, config.AuthConfig{
		JWTSecret:      "test-secret-at-least-32-bytes-pad",
		JWTAccessTTL:   "15m",
		JWTRefreshTTL:  "168h",
		CookieSecure:   "false",
		CookieSameSite: "Strict",
		CookiePath:     "/app",
	})
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}
	cfg := svc.CookieConfig()
	if cfg.Name != "kube_rca_refresh" {
		t.Errorf("cookie name = %q, want %q", cfg.Name, "kube_rca_refresh")
	}
	if cfg.Path != "/app" {
		t.Errorf("cookie path = %q, want %q", cfg.Path, "/app")
	}
	if cfg.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie SameSite = %v, want Strict", cfg.SameSite)
	}
}
