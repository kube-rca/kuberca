package client

import "testing"

func TestEmbeddingClientConfig(t *testing.T) {
	cfg := EmbeddingClientConfig{APIKey: "key", Model: "text-embedding-004"}
	if cfg.Model == "" || cfg.APIKey == "" {
		t.Fatalf("expected model and api key")
	}
}
