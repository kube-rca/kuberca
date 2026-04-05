package client

import (
	"testing"

	"github.com/kube-rca/backend/internal/config"
)

func TestEmbeddingClientConfig(t *testing.T) {
	cfg := config.EmbeddingConfig{APIKey: "key", Model: "text-embedding-004"}
	if cfg.Model == "" || cfg.APIKey == "" {
		t.Fatalf("expected model and api key")
	}
}
