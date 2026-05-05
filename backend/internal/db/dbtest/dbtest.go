// Package dbtest provides a shared helper for spinning up a real PostgreSQL
// container (pgvector/pgvector:pg16) for integration tests.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    pool := dbtest.StartPostgres(t)
//	    pg := &db.Postgres{Pool: pool}
//	    // ...
//	}
package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// StartPostgres launches a pgvector/pgvector:pg16 container, applies the
// vector extension, and returns a connected *pgxpool.Pool.  The container is
// stopped automatically via t.Cleanup.
func StartPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx,
		"pgvector/pgvector:pg16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("dbtest: failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if termErr := ctr.Terminate(ctx); termErr != nil {
			t.Logf("dbtest: warning — failed to terminate container: %v", termErr)
		}
	})

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dbtest: failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("dbtest: failed to create pool: %v", err)
	}

	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("dbtest: failed to ping postgres: %v", err)
	}

	// Ensure pgvector extension is available.
	if _, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		t.Fatalf("dbtest: failed to create vector extension: %v", err)
	}

	return pool
}
