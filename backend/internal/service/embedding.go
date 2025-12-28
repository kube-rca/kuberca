package service

import (
	"context"
	"fmt"
)

type EmbeddingRepo interface {
	InsertEmbedding(ctx context.Context, incidentID, summary, model string, vector []float32) (int64, error)
}

type EmbeddingClient interface {
	EmbedText(ctx context.Context, text string) ([]float32, string, error)
}

type EmbeddingService struct {
	repo   EmbeddingRepo
	client EmbeddingClient
}

func NewEmbeddingService(repo EmbeddingRepo, client EmbeddingClient) *EmbeddingService {
	return &EmbeddingService{repo: repo, client: client}
}

func (s *EmbeddingService) CreateEmbedding(ctx context.Context, incidentID, summary string) (int64, string, error) {
	if incidentID == "" || summary == "" {
		return 0, "", fmt.Errorf("incident_id and incident_summary are required")
	}
	vector, model, err := s.client.EmbedText(ctx, summary)
	if err != nil {
		return 0, model, err
	}
	id, err := s.repo.InsertEmbedding(ctx, incidentID, summary, model, vector)
	return id, model, err
}
