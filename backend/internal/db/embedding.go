package db

import (
	"context"

	"github.com/pgvector/pgvector-go"
)

// EnsureEmbeddingSchema creates the embeddings table with pgvector extension
func (db *Postgres) EnsureEmbeddingSchema(ctx context.Context) error {
	queries := []string{
		// pgvector 확장 활성화
		`CREATE EXTENSION IF NOT EXISTS vector`,
		// embeddings 테이블 생성
		`CREATE TABLE IF NOT EXISTS embeddings (
			id BIGSERIAL PRIMARY KEY,
			incident_id TEXT NOT NULL,
			incident_summary TEXT NOT NULL,
			embedding vector(3072),
			model TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// 인덱스 생성
		`CREATE INDEX IF NOT EXISTS embeddings_incident_id_idx ON embeddings(incident_id)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func embeddingInsertQuery() string {
	return `
		INSERT INTO embeddings (incident_id, incident_summary, embedding, model)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
}

func (db *Postgres) InsertEmbedding(ctx context.Context, incidentID, summary, model string, vector []float32) (int64, error) {
	var id int64
	query := embeddingInsertQuery()
	err := db.Pool.QueryRow(ctx, query, incidentID, summary, pgvector.NewVector(vector), model).Scan(&id)
	return id, err
}
