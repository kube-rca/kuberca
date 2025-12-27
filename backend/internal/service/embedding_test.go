package service

import (
	"context"
	"testing"
)

type fakeEmbeddingClient struct{}

type fakeEmbeddingRepo struct{}

func (f *fakeEmbeddingClient) EmbedText(ctx context.Context, text string) ([]float32, string, error) {
	return []float32{0.1}, "text-embedding-004", nil
}

func (f *fakeEmbeddingRepo) InsertEmbedding(ctx context.Context, incidentID, summary, model string, vector []float32) (int64, error) {
	return 1, nil
}

func TestCreateEmbedding(t *testing.T) {
	svc := NewEmbeddingService(&fakeEmbeddingRepo{}, &fakeEmbeddingClient{})
	id, model, err := svc.CreateEmbedding(context.Background(), "INC-1", "summary")
	if err != nil || id == 0 || model == "" {
		t.Fatalf("expected success")
	}
}
