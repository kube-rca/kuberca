package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kube-rca/backend/internal/model"
)

func (db *Postgres) EnsureAuthSchema(ctx context.Context) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			login_id TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (db *Postgres) CreateUser(ctx context.Context, loginID, passwordHash string) (*model.User, error) {
	query := `
		INSERT INTO users (login_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, login_id, password_hash, created_at, updated_at
	`
	var user model.User
	err := db.Pool.QueryRow(ctx, query, loginID, passwordHash).Scan(
		&user.ID,
		&user.LoginID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Postgres) GetUserByLoginID(ctx context.Context, loginID string) (*model.User, error) {
	query := `
		SELECT id, login_id, password_hash, created_at, updated_at
		FROM users
		WHERE login_id = $1
	`
	var user model.User
	err := db.Pool.QueryRow(ctx, query, loginID).Scan(
		&user.ID,
		&user.LoginID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Postgres) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	query := `
		SELECT id, login_id, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user model.User
	err := db.Pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.LoginID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Postgres) InsertRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := db.Pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (db *Postgres) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var token model.RefreshToken
	err := db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (db *Postgres) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	_, err := db.Pool.Exec(ctx, query, tokenHash)
	return err
}

func (db *Postgres) RotateRefreshToken(ctx context.Context, oldTokenID int64, userID int64, newTokenHash string, newExpiresAt time.Time) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`, oldTokenID); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, newTokenHash, newExpiresAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
