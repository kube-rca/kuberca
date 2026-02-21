package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kube-rca/backend/internal/model"
)

// EnsureWebhookSchema - webhook_configs 테이블 생성 (없으면)
func (p *Postgres) EnsureWebhookSchema() error {
	ctx := context.Background()
	_, err := p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS webhook_configs (
			id         SERIAL       PRIMARY KEY,
			url        TEXT         NOT NULL DEFAULT '',
			method     TEXT         NOT NULL DEFAULT 'POST',
			headers    JSONB        NOT NULL DEFAULT '[]',
			body       TEXT         NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create webhook_configs table: %w", err)
	}
	return nil
}

// GetWebhookConfigs - 웹훅 설정 전체 목록 조회 (최신순)
func (p *Postgres) GetWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT id, url, method, headers, body, updated_at
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
		var headersJSON []byte
		if err := rows.Scan(&cfg.ID, &cfg.URL, &cfg.Method, &headersJSON, &cfg.Body, &cfg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook config: %w", err)
		}
		if err := json.Unmarshal(headersJSON, &cfg.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
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
		SELECT id, url, method, headers, body, updated_at
		FROM webhook_configs
		WHERE id = $1;
	`, id)

	var cfg model.WebhookConfig
	var headersJSON []byte
	if err := row.Scan(&cfg.ID, &cfg.URL, &cfg.Method, &headersJSON, &cfg.Body, &cfg.UpdatedAt); err != nil {
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}
	if err := json.Unmarshal(headersJSON, &cfg.Headers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
	}
	return &cfg, nil
}

// CreateWebhookConfig - 신규 웹훅 설정 저장
func (p *Postgres) CreateWebhookConfig(ctx context.Context, cfg model.WebhookConfig) (int, error) {
	headersJSON, err := json.Marshal(cfg.Headers)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal headers: %w", err)
	}

	var id int
	err = p.Pool.QueryRow(ctx, `
		INSERT INTO webhook_configs (url, method, headers, body, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id;
	`, cfg.URL, cfg.Method, headersJSON, cfg.Body).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook config: %w", err)
	}
	return id, nil
}

// UpdateWebhookConfig - ID로 웹훅 설정 수정
func (p *Postgres) UpdateWebhookConfig(ctx context.Context, id int, cfg model.WebhookConfig) error {
	headersJSON, err := json.Marshal(cfg.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	tag, err := p.Pool.Exec(ctx, `
		UPDATE webhook_configs
		SET url = $1, method = $2, headers = $3, body = $4, updated_at = NOW()
		WHERE id = $5;
	`, cfg.URL, cfg.Method, headersJSON, cfg.Body, id)
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
