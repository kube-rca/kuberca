package db

import (
	"context"
	"math"
	"testing"

	"github.com/kube-rca/backend/internal/db/dbtest"
)

func setupEmbeddingDB(t *testing.T) *Postgres {
	t.Helper()
	pool := dbtest.StartPostgres(t)
	pg := &Postgres{Pool: pool}
	if err := pg.EnsureEmbeddingSchema(context.Background()); err != nil {
		t.Fatalf("EnsureEmbeddingSchema: %v", err)
	}
	return pg
}

// makeVec returns a unit vector of dimension d with only the i-th element set.
func makeVec(d, i int) []float32 {
	v := make([]float32, d)
	v[i] = 1.0
	return v
}

func TestInsertEmbedding_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pg := setupEmbeddingDB(t)

	vec := makeVec(3072, 0)
	id, err := pg.InsertEmbedding(ctx, "INC-test-001", "OOMKilled in prod", "gemini-embedding-004", vec)
	if err != nil {
		t.Fatalf("InsertEmbedding: %v", err)
	}
	if id <= 0 {
		t.Fatalf("InsertEmbedding returned id=%d; want >0", id)
	}
}

func TestSearchEmbeddings_CosineSimilarity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pg := setupEmbeddingDB(t)

	// Insert two embeddings: one identical to query, one orthogonal.
	// Cosine similarity: identical=1.0, orthogonal=0.0.
	queryVec := makeVec(3072, 0)
	orthVec := makeVec(3072, 1)

	if _, err := pg.InsertEmbedding(ctx, "INC-cos-A", "Relevant incident", "model", queryVec); err != nil {
		t.Fatalf("InsertEmbedding A: %v", err)
	}
	if _, err := pg.InsertEmbedding(ctx, "INC-cos-B", "Unrelated incident", "model", orthVec); err != nil {
		t.Fatalf("InsertEmbedding B: %v", err)
	}

	results, err := pg.SearchEmbeddings(ctx, queryVec, 2)
	if err != nil {
		t.Fatalf("SearchEmbeddings: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d; want 2", len(results))
	}

	// Results are ordered by cosine distance (closest first).
	// The first result must be the identical vector (similarity ≈ 1.0).
	if results[0].IncidentID != "INC-cos-A" {
		t.Errorf("top result IncidentID = %q; want INC-cos-A", results[0].IncidentID)
	}
	if math.Abs(results[0].Similarity-1.0) > 0.001 {
		t.Errorf("similarity = %f; want ≈1.0", results[0].Similarity)
	}
	// Second result: orthogonal vector → similarity ≈ 0.0.
	if math.Abs(results[1].Similarity) > 0.001 {
		t.Errorf("orthogonal similarity = %f; want ≈0.0", results[1].Similarity)
	}
}

func TestSearchEmbeddings_DefaultLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pg := setupEmbeddingDB(t)

	// limit=0 should fall back to 10 (internal default).
	vec := makeVec(3072, 5)
	if _, err := pg.InsertEmbedding(ctx, "INC-lim-001", "Test", "model", vec); err != nil {
		t.Fatalf("InsertEmbedding: %v", err)
	}

	results, err := pg.SearchEmbeddings(ctx, vec, 0)
	if err != nil {
		t.Fatalf("SearchEmbeddings(limit=0): %v", err)
	}
	if len(results) == 0 {
		t.Error("SearchEmbeddings returned 0 results with limit=0 fallback")
	}
}

func TestInsertEmbedding_DimensionMismatch_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pg := setupEmbeddingDB(t)

	// The schema defines vector(3072); a 128-dim vector should fail.
	smallVec := makeVec(128, 0)
	_, err := pg.InsertEmbedding(ctx, "INC-dim-err", "Wrong dim", "model", smallVec)
	if err == nil {
		t.Error("expected error for dimension mismatch, got nil")
	}
}
