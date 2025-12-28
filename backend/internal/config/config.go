package config

import "os"

type Config struct {
	Slack     SlackConfig
	Agent     AgentConfig
	Embedding EmbeddingConfig
	Postgres  PostgresConfig
	Auth      AuthConfig
}

type SlackConfig struct {
	BotToken  string
	ChannelID string
}

type AgentConfig struct {
	BaseURL string
}

type EmbeddingConfig struct {
	APIKey string
}

type PostgresConfig struct {
	DatabaseURL string
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	SSLMode     string
}

type AuthConfig struct {
	JWTSecret          string
	JWTAccessTTL       string
	JWTRefreshTTL      string
	AllowSignup        string
	AdminUsername      string
	AdminPassword      string
	CookieSecure       string
	CookieSameSite     string
	CookieDomain       string
	CookiePath         string
	CorsAllowedOrigins string
}

func Load() Config {
	return Config{
		Slack: SlackConfig{
			BotToken:  os.Getenv("SLACK_BOT_TOKEN"),
			ChannelID: os.Getenv("SLACK_CHANNEL_ID"),
		},
		Agent: AgentConfig{
			BaseURL: getenv("AGENT_URL", "http://kube-rca-agent.kube-rca.svc:8000"),
		},
		Embedding: EmbeddingConfig{
			APIKey: os.Getenv("AI_API_KEY"),
		},
		Postgres: PostgresConfig{
			DatabaseURL: os.Getenv("DATABASE_URL"),
			Host:        getenv("PGHOST", "localhost"),
			Port:        getenv("PGPORT", "5432"),
			User:        os.Getenv("PGUSER"),
			Password:    os.Getenv("PGPASSWORD"),
			Database:    os.Getenv("PGDATABASE"),
			SSLMode:     getenv("PGSSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:          os.Getenv("JWT_SECRET"),
			JWTAccessTTL:       getenv("JWT_ACCESS_TTL", "15m"),
			JWTRefreshTTL:      getenv("JWT_REFRESH_TTL", "168h"),
			AllowSignup:        getenv("ALLOW_SIGNUP", "false"),
			AdminUsername:      os.Getenv("ADMIN_USERNAME"),
			AdminPassword:      os.Getenv("ADMIN_PASSWORD"),
			CookieSecure:       getenv("AUTH_COOKIE_SECURE", "true"),
			CookieSameSite:     getenv("AUTH_COOKIE_SAMESITE", "Lax"),
			CookieDomain:       os.Getenv("AUTH_COOKIE_DOMAIN"),
			CookiePath:         getenv("AUTH_COOKIE_PATH", "/"),
			CorsAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
		},
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
