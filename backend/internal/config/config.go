package config

import "os"

type Config struct {
	Slack     SlackConfig
	Agent     AgentConfig
	Embedding EmbeddingConfig
	Postgres  PostgresConfig
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
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
