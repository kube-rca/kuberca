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
	"log"
	"math/rand"
	"net"
	"net/url"
	"time"

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

	maxAttempts := cfg.RetryMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	initialBackoff := time.Duration(cfg.RetryInitialBackoffSecs) * time.Second
	if initialBackoff <= 0 {
		initialBackoff = time.Second
	}
	maxBackoff := time.Duration(cfg.RetryMaxBackoffSecs) * time.Second
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}

	var pool *pgxpool.Pool
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while connecting to postgres: %w", ctx.Err())
		default:
		}

		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				pool.Close()
				err = pingErr
			}
		}

		if attempt < maxAttempts-1 {
			backoff := initialBackoff * (1 << uint(attempt))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			jitter := time.Duration(float64(backoff) * (0.75 + rand.Float64()*0.5))
			log.Printf("postgres connection attempt %d/%d failed: %v — retrying in %s", attempt+1, maxAttempts, err, jitter)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled while waiting to retry postgres: %w", ctx.Err())
			case <-time.After(jitter):
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to postgres after %d attempts: %w", maxAttempts, err)
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
