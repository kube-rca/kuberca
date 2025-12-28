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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kube-rca/backend/internal/config"
)

func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn, err := buildPostgresURL(cfg)
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}

func buildPostgresURL(cfg config.PostgresConfig) (string, error) {
	if cfg.DatabaseURL != "" {
		return cfg.DatabaseURL, nil
	}

	user := cfg.User
	dbName := cfg.Database
	if user == "" || dbName == "" {
		return "", fmt.Errorf("missing required env: DATABASE_URL or PGUSER/PGDATABASE")
	}

	host := defaultIfEmpty(cfg.Host, "localhost")
	port := defaultIfEmpty(cfg.Port, "5432")
	password := cfg.Password
	sslmode := defaultIfEmpty(cfg.SSLMode, "disable")

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

func defaultIfEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
