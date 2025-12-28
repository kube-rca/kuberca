package client

import (
	"context"
	"fmt"

	"github.com/kube-rca/backend/internal/config"
	"google.golang.org/genai"
)

type EmbeddingClientConfig struct {
	APIKey string
	Model  string
}

type EmbeddingClient struct {
	client *genai.Client
	model  string
}

func NewEmbeddingClient(cfg config.EmbeddingConfig) (*EmbeddingClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("missing AI_API_KEY")
	}
	clientCfg := EmbeddingClientConfig{APIKey: cfg.APIKey, Model: "text-embedding-004"}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: clientCfg.APIKey})
	if err != nil {
		return nil, err
	}
	return &EmbeddingClient{client: client, model: clientCfg.Model}, nil
}

func (c *EmbeddingClient) EmbedText(ctx context.Context, text string) ([]float32, string, error) {
	res, err := c.client.Models.EmbedContent(ctx, c.model, genai.Text(text), nil)
	if err != nil {
		return nil, c.model, err
	}
	if res == nil || len(res.Embeddings) == 0 || res.Embeddings[0] == nil {
		return nil, c.model, fmt.Errorf("empty embedding result")
	}
	return res.Embeddings[0].Values, c.model, nil
}
