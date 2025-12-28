package db

import "testing"

func TestEmbeddingInsertSQL(t *testing.T) {
	query := embeddingInsertQuery()
	if query == "" {
		t.Fatalf("expected query")
	}
}
