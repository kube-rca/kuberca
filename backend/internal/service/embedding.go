package service

import (
	"context"
	"fmt"
	"log"

	"github.com/kube-rca/backend/internal/db"
)

type EmbeddingRepo interface {
	InsertEmbedding(ctx context.Context, incidentID, summary, model string, vector []float32) (int64, error)
	SearchEmbeddings(ctx context.Context, vector []float32, limit int) ([]db.EmbeddingSearchResult, error)
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
		log.Printf("Failed to embed text: %v", err)
		return 0, model, err
	}
	id, err := s.repo.InsertEmbedding(ctx, incidentID, summary, model, vector)
	return id, model, err
}

func (s *EmbeddingService) SearchSimilar(ctx context.Context, query string, limit int) ([]db.EmbeddingSearchResult, string, error) {
	if query == "" {
		return nil, "", fmt.Errorf("query is required")
	}
	vector, model, err := s.client.EmbedText(ctx, query)
	if err != nil {
		log.Printf("Failed to embed query: %v", err)
		return nil, model, err
	}
	results, err := s.repo.SearchEmbeddings(ctx, vector, limit)
	return results, model, err
}
