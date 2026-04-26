package config

import (
	"testing"
)

// TestLoad_Defaults verifies that Load() runs without panicking when env vars
// are unset and returns a Config with expected fallback values.
func TestLoad_Defaults(t *testing.T) {
	// JWTAccessTTL has a hardcoded fallback of "15m".
	t.Setenv("JWT_ACCESS_TTL", "")
	t.Setenv("JWT_REFRESH_TTL", "")

	cfg := Load()

	if cfg.Auth.JWTAccessTTL != "15m" {
		t.Errorf("JWTAccessTTL = %q, want %q", cfg.Auth.JWTAccessTTL, "15m")
	}
	if cfg.Auth.JWTRefreshTTL != "168h" {
		t.Errorf("JWTRefreshTTL = %q, want %q", cfg.Auth.JWTRefreshTTL, "168h")
	}
	if cfg.Auth.CookieSameSite != "Lax" {
		t.Errorf("CookieSameSite = %q, want %q", cfg.Auth.CookieSameSite, "Lax")
	}
	if cfg.Auth.CookiePath != "/" {
		t.Errorf("CookiePath = %q, want %q", cfg.Auth.CookiePath, "/")
	}
}

// TestLoad_EnvOverrides verifies that env vars override defaults.
func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ACCESS_TTL", "30m")
	t.Setenv("JWT_REFRESH_TTL", "72h")
	t.Setenv("ALLOW_SIGNUP", "true")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")

	cfg := Load()

	if cfg.Auth.JWTSecret != "test-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.Auth.JWTSecret, "test-secret")
	}
	if cfg.Auth.JWTAccessTTL != "30m" {
		t.Errorf("JWTAccessTTL = %q, want %q", cfg.Auth.JWTAccessTTL, "30m")
	}
	if cfg.Auth.JWTRefreshTTL != "72h" {
		t.Errorf("JWTRefreshTTL = %q, want %q", cfg.Auth.JWTRefreshTTL, "72h")
	}
	if cfg.Auth.AllowSignup != "true" {
		t.Errorf("AllowSignup = %q, want %q", cfg.Auth.AllowSignup, "true")
	}
	if cfg.Postgres.DatabaseURL != "postgres://localhost/testdb" {
		t.Errorf("Postgres.DatabaseURL = %q, want %q", cfg.Postgres.DatabaseURL, "postgres://localhost/testdb")
	}
}

// TestLoad_OIDCEnabled verifies OIDC config is populated from env.
func TestLoad_OIDCEnabled(t *testing.T) {
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_CLIENT_ID", "my-client")
	t.Setenv("OIDC_REDIRECT_URI", "https://app.example.com/callback")

	cfg := Load()

	if !cfg.OIDC.Enabled {
		t.Error("OIDC.Enabled should be true")
	}
	if cfg.OIDC.ClientID != "my-client" {
		t.Errorf("OIDC.ClientID = %q, want %q", cfg.OIDC.ClientID, "my-client")
	}
	if cfg.OIDC.RedirectURI != "https://app.example.com/callback" {
		t.Errorf("OIDC.RedirectURI = %q, want %q", cfg.OIDC.RedirectURI, "https://app.example.com/callback")
	}
}

// TestLoad_AgentConfig verifies agent URL default.
func TestLoad_AgentConfig(t *testing.T) {
	t.Setenv("AGENT_URL", "")
	t.Setenv("AGENT_HTTP_TIMEOUT_SECONDS", "")

	cfg := Load()

	if cfg.Agent.BaseURL != "http://kube-rca-agent.kube-rca.svc:8000" {
		t.Errorf("Agent.BaseURL = %q, want default", cfg.Agent.BaseURL)
	}
	if cfg.Agent.HTTPTimeoutSeconds != 240 {
		t.Errorf("Agent.HTTPTimeoutSeconds = %d, want 240", cfg.Agent.HTTPTimeoutSeconds)
	}
}

// TestLoad_AgentURLOverride verifies agent URL can be overridden.
func TestLoad_AgentURLOverride(t *testing.T) {
	t.Setenv("AGENT_URL", "http://localhost:8000")
	cfg := Load()
	if cfg.Agent.BaseURL != "http://localhost:8000" {
		t.Errorf("Agent.BaseURL = %q, want %q", cfg.Agent.BaseURL, "http://localhost:8000")
	}
}

// TestLoad_FlappingDefaults verifies flapping detection defaults.
func TestLoad_FlappingDefaults(t *testing.T) {
	t.Setenv("FLAP_ENABLED", "")
	t.Setenv("FLAP_CYCLE_THRESHOLD", "")
	t.Setenv("FLAP_DETECTION_WINDOW_MINUTES", "")

	cfg := Load()

	if !cfg.Flapping.Enabled {
		t.Error("Flapping.Enabled should default to true")
	}
	if cfg.Flapping.CycleThreshold != 3 {
		t.Errorf("CycleThreshold = %d, want 3", cfg.Flapping.CycleThreshold)
	}
	if cfg.Flapping.DetectionWindowMinutes != 30 {
		t.Errorf("DetectionWindowMinutes = %d, want 30", cfg.Flapping.DetectionWindowMinutes)
	}
}

// TestLoad_FlappingOverride verifies flapping can be disabled via env.
func TestLoad_FlappingOverride(t *testing.T) {
	t.Setenv("FLAP_ENABLED", "false")
	t.Setenv("FLAP_CYCLE_THRESHOLD", "5")

	cfg := Load()

	if cfg.Flapping.Enabled {
		t.Error("Flapping.Enabled should be false")
	}
	if cfg.Flapping.CycleThreshold != 5 {
		t.Errorf("CycleThreshold = %d, want 5", cfg.Flapping.CycleThreshold)
	}
}

// TestLoad_WebhookDefaults verifies webhook hardening defaults.
func TestLoad_WebhookDefaults(t *testing.T) {
	t.Setenv("WEBHOOK_HMAC_HEADER", "")
	t.Setenv("WEBHOOK_RATE_LIMIT", "")

	cfg := Load()

	if cfg.Webhook.HMACHeader != "X-Webhook-Signature" {
		t.Errorf("Webhook.HMACHeader = %q, want %q", cfg.Webhook.HMACHeader, "X-Webhook-Signature")
	}
	if cfg.Webhook.RateLimitPerMinute != 100 {
		t.Errorf("Webhook.RateLimitPerMinute = %d, want 100", cfg.Webhook.RateLimitPerMinute)
	}
}

// TestLoad_EmbeddingDefaults verifies embedding defaults.
func TestLoad_EmbeddingDefaults(t *testing.T) {
	t.Setenv("EMBEDDING_PROVIDER", "")
	t.Setenv("EMBEDDING_MODEL", "")

	cfg := Load()

	if cfg.Embedding.Provider != "google" {
		t.Errorf("Embedding.Provider = %q, want %q", cfg.Embedding.Provider, "google")
	}
	if cfg.Embedding.Model != "text-embedding-004" {
		t.Errorf("Embedding.Model = %q, want %q", cfg.Embedding.Model, "text-embedding-004")
	}
}

// TestLoad_AIProvider verifies AI provider default.
func TestLoad_AIProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "")
	cfg := Load()
	if cfg.AI.Provider != "gemini" {
		t.Errorf("AI.Provider = %q, want %q", cfg.AI.Provider, "gemini")
	}
}

// TestLoad_AIProviderOverride verifies supported AI providers can be set.
func TestLoad_AIProviderOverride(t *testing.T) {
	for _, provider := range []string{"gemini", "openai", "anthropic"} {
		t.Setenv("AI_PROVIDER", provider)
		cfg := Load()
		if cfg.AI.Provider != provider {
			t.Errorf("AI.Provider = %q, want %q", cfg.AI.Provider, provider)
		}
	}
}

// TestLoad_PostgresDefaults verifies postgres connection defaults.
func TestLoad_PostgresDefaults(t *testing.T) {
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGSSLMODE", "")

	cfg := Load()

	if cfg.Postgres.Host != "localhost" {
		t.Errorf("Postgres.Host = %q, want localhost", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != "5432" {
		t.Errorf("Postgres.Port = %q, want 5432", cfg.Postgres.Port)
	}
	if cfg.Postgres.SSLMode != "disable" {
		t.Errorf("Postgres.SSLMode = %q, want disable", cfg.Postgres.SSLMode)
	}
}
