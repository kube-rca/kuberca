package db

import (
	"context"
	"fmt"

	"github.com/kube-rca/backend/internal/model"
)

// EnsureWebhookSchema - webhook_configs 테이블 생성/정규화
func (p *Postgres) EnsureWebhookSchema() error {
	ctx := context.Background()
	_, err := p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS webhook_configs (
			id         SERIAL       PRIMARY KEY,
			url        TEXT         NOT NULL DEFAULT '',
			type       TEXT         NOT NULL DEFAULT 'http',
			token      TEXT         NOT NULL DEFAULT '',
			channel    TEXT         NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create webhook_configs table: %w", err)
	}

	_, err = p.Pool.Exec(ctx, `
		ALTER TABLE webhook_configs
		ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'http',
		ADD COLUMN IF NOT EXISTS token TEXT NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS channel TEXT NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS severities TEXT[] NOT NULL DEFAULT '{}';
	`)
	if err != nil {
		return fmt.Errorf("failed to alter webhook_configs table(add columns): %w", err)
	}

	_, err = p.Pool.Exec(ctx, `
		ALTER TABLE webhook_configs
		DROP COLUMN IF EXISTS method,
		DROP COLUMN IF EXISTS headers,
		DROP COLUMN IF EXISTS body;
	`)
	if err != nil {
		return fmt.Errorf("failed to alter webhook_configs table(drop old columns): %w", err)
	}

	_, err = p.Pool.Exec(ctx, `
		UPDATE webhook_configs
		SET
			url = COALESCE(TRIM(url), ''),
			type = CASE
				WHEN LOWER(TRIM(type)) IN ('slack', 'teams', 'http') THEN LOWER(TRIM(type))
				ELSE 'http'
			END,
			token = COALESCE(TRIM(token), ''),
			channel = COALESCE(TRIM(channel), '');
	`)
	if err != nil {
		return fmt.Errorf("failed to normalize webhook_configs rows: %w", err)
	}

	return nil
}

// GetWebhookConfigs - 웹훅 설정 전체 목록 조회 (최신순)
func (p *Postgres) GetWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT id, url, type, token, channel, severities, updated_at
		FROM webhook_configs
		ORDER BY updated_at DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhook configs: %w", err)
	}
	defer rows.Close()

	var configs []model.WebhookConfig
	for rows.Next() {
		var cfg model.WebhookConfig
		if err := rows.Scan(&cfg.ID, &cfg.URL, &cfg.Type, &cfg.Token, &cfg.Channel, &cfg.Severities, &cfg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook config: %w", err)
		}
		if cfg.Severities == nil {
			cfg.Severities = []string{}
		}
		configs = append(configs, cfg)
	}
	if configs == nil {
		configs = []model.WebhookConfig{}
	}
	return configs, nil
}

// GetWebhookConfigByID - ID로 단건 조회
func (p *Postgres) GetWebhookConfigByID(ctx context.Context, id int) (*model.WebhookConfig, error) {
	row := p.Pool.QueryRow(ctx, `
		SELECT id, url, type, token, channel, severities, updated_at
		FROM webhook_configs
		WHERE id = $1;
	`, id)

	var cfg model.WebhookConfig
	if err := row.Scan(&cfg.ID, &cfg.URL, &cfg.Type, &cfg.Token, &cfg.Channel, &cfg.Severities, &cfg.UpdatedAt); err != nil {
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}
	if cfg.Severities == nil {
		cfg.Severities = []string{}
	}
	return &cfg, nil
}

// CreateWebhookConfig - 신규 웹훅 설정 저장
func (p *Postgres) CreateWebhookConfig(ctx context.Context, cfg model.WebhookConfig) (int, error) {
	severities := cfg.Severities
	if severities == nil {
		severities = []string{}
	}
	var id int
	err := p.Pool.QueryRow(ctx, `
		INSERT INTO webhook_configs (url, type, token, channel, severities, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id;
	`, cfg.URL, cfg.Type, cfg.Token, cfg.Channel, severities).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook config: %w", err)
	}
	return id, nil
}

// UpdateWebhookConfig - ID로 웹훅 설정 수정
func (p *Postgres) UpdateWebhookConfig(ctx context.Context, id int, cfg model.WebhookConfig) error {
	severities := cfg.Severities
	if severities == nil {
		severities = []string{}
	}
	tag, err := p.Pool.Exec(ctx, `
		UPDATE webhook_configs
		SET url = $1, type = $2, token = $3, channel = $4, severities = $5, updated_at = NOW()
		WHERE id = $6;
	`, cfg.URL, cfg.Type, cfg.Token, cfg.Channel, severities, id)
	if err != nil {
		return fmt.Errorf("failed to update webhook config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webhook config not found: id=%d", id)
	}
	return nil
}

// DeleteWebhookConfig - ID로 웹훅 설정 삭제
func (p *Postgres) DeleteWebhookConfig(ctx context.Context, id int) error {
	tag, err := p.Pool.Exec(ctx, `DELETE FROM webhook_configs WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webhook config not found: id=%d", id)
	}
	return nil
}
