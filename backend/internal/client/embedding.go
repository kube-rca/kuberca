package client

import (
	"context"
	"fmt"

	"github.com/kube-rca/backend/internal/config"
	"google.golang.org/genai"
)

type EmbeddingClient struct {
	client *genai.Client
	model  string
}

func NewEmbeddingClient(cfg config.EmbeddingConfig) (*EmbeddingClient, error) {
	fmt.Println("api key: [REDACTED]")
	fmt.Printf("model: %s\n", cfg.Model)
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("missing AI_API_KEY")
	}
	clientConfig := genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI}
	client, err := genai.NewClient(context.Background(), &clientConfig)
	if err != nil {
		return nil, err
	}
	return &EmbeddingClient{client: client, model: cfg.Model}, nil
}

func (c *EmbeddingClient) EmbedText(ctx context.Context, text string) ([]float32, string, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}
	res, err := c.client.Models.EmbedContent(ctx, c.model, contents, nil)
	if err != nil {
		return nil, c.model, err
	}
	if res == nil || len(res.Embeddings) == 0 || res.Embeddings[0] == nil {
		return nil, c.model, fmt.Errorf("empty embedding result")
	}
	return res.Embeddings[0].Values, c.model, nil
}
