package db

import (
	"context"

	"github.com/pgvector/pgvector-go"
)

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
