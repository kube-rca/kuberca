// PostgreSQL 연결 초기화 유틸
//
// 환경변수:
//   - DATABASE_URL: postgres://user:pass@host:port/dbname?sslmode=disable
//   - PGHOST (default: localhost)
//   - PGPORT (default: 5432)
//   - PGUSER
//   - PGPASSWORD
//   - PGDATABASE
//   - PGSSLMODE (default: disable)

package db

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn, err := buildPostgresURL()
	if err != nil {
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}

func buildPostgresURL() (string, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}

	user := os.Getenv("PGUSER")
	dbName := os.Getenv("PGDATABASE")
	if user == "" || dbName == "" {
		return "", fmt.Errorf("missing required env: DATABASE_URL or PGUSER/PGDATABASE")
	}

	host := getenv("PGHOST", "localhost")
	port := getenv("PGPORT", "5432")
	password := os.Getenv("PGPASSWORD")
	sslmode := getenv("PGSSLMODE", "disable")

	u := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, port),
		Path:   dbName,
	}
	if password == "" {
		u.User = url.User(user)
	} else {
		u.User = url.UserPassword(user, password)
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}