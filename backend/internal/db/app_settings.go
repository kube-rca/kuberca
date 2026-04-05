package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kube-rca/backend/internal/model"
)

// EnsureAppSettingsSchema - app_settings 테이블 생성
func (p *Postgres) EnsureAppSettingsSchema() error {
	ctx := context.Background()
	_, err := p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS app_settings (
			key        TEXT        PRIMARY KEY,
			value      JSONB       NOT NULL DEFAULT '{}',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create app_settings table: %w", err)
	}
	return nil
}

// GetAppSetting - key로 단건 조회
func (p *Postgres) GetAppSetting(ctx context.Context, key string) (*model.AppSetting, error) {
	row := p.Pool.QueryRow(ctx, `
		SELECT key, value, updated_at
		FROM app_settings
		WHERE key = $1;
	`, key)

	var setting model.AppSetting
	if err := row.Scan(&setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get app setting: %w", err)
	}
	return &setting, nil
}

// UpsertAppSetting - INSERT ON CONFLICT UPDATE
func (p *Postgres) UpsertAppSetting(ctx context.Context, key string, value json.RawMessage) error {
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE
		SET value = $2, updated_at = NOW();
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to upsert app setting: %w", err)
	}
	return nil
}

// GetAllAppSettings - 전체 설정 조회
func (p *Postgres) GetAllAppSettings(ctx context.Context) ([]model.AppSetting, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT key, value, updated_at
		FROM app_settings
		ORDER BY key;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query app settings: %w", err)
	}
	defer rows.Close()

	var settings []model.AppSetting
	for rows.Next() {
		var s model.AppSetting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan app setting: %w", err)
		}
		settings = append(settings, s)
	}
	if settings == nil {
		settings = []model.AppSetting{}
	}
	return settings, nil
}
