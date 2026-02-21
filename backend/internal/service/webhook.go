package service

import (
	"context"

	"github.com/kube-rca/backend/internal/model"
)

// webhookRepo - DB 인터페이스
type webhookRepo interface {
	GetWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error)
	GetWebhookConfigByID(ctx context.Context, id int) (*model.WebhookConfig, error)
	CreateWebhookConfig(ctx context.Context, cfg model.WebhookConfig) (int, error)
	UpdateWebhookConfig(ctx context.Context, id int, cfg model.WebhookConfig) error
	DeleteWebhookConfig(ctx context.Context, id int) error
}

// WebhookService - 웹훅 설정 비즈니스 로직
type WebhookService struct {
	db webhookRepo
}

func NewWebhookService(db webhookRepo) *WebhookService {
	return &WebhookService{db: db}
}

func (s *WebhookService) ListWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error) {
	return s.db.GetWebhookConfigs(ctx)
}

func (s *WebhookService) GetWebhookConfig(ctx context.Context, id int) (*model.WebhookConfig, error) {
	return s.db.GetWebhookConfigByID(ctx, id)
}

func (s *WebhookService) CreateWebhookConfig(ctx context.Context, req model.WebhookConfigRequest) (int, error) {
	cfg := model.WebhookConfig{
		URL:    req.URL,
		Method: req.Method,
		Body:   req.Body,
	}
	if req.Headers != nil {
		cfg.Headers = req.Headers
	} else {
		cfg.Headers = []model.WebhookHeader{}
	}
	return s.db.CreateWebhookConfig(ctx, cfg)
}

func (s *WebhookService) UpdateWebhookConfig(ctx context.Context, id int, req model.WebhookConfigRequest) error {
	cfg := model.WebhookConfig{
		URL:    req.URL,
		Method: req.Method,
		Body:   req.Body,
	}
	if req.Headers != nil {
		cfg.Headers = req.Headers
	} else {
		cfg.Headers = []model.WebhookHeader{}
	}
	return s.db.UpdateWebhookConfig(ctx, id, cfg)
}

func (s *WebhookService) DeleteWebhookConfig(ctx context.Context, id int) error {
	return s.db.DeleteWebhookConfig(ctx, id)
}
