package client

import (
	"context"
	"fmt"
	"os"

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

func NewEmbeddingClient() (*EmbeddingClient, error) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing AI_API_KEY")
	}
	cfg := EmbeddingClientConfig{APIKey: apiKey, Model: "text-embedding-004"}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: cfg.APIKey})
	if err != nil {
		return nil, err
	}
	return &EmbeddingClient{client: client, model: cfg.Model}, nil
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
