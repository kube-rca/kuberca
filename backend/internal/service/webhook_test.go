package service

import (
	"context"
	"testing"

	"github.com/kube-rca/backend/internal/model"
)

type webhookRepoMock struct {
	createdCfg model.WebhookConfig
	updatedCfg model.WebhookConfig
	updatedID  int
}

func (m *webhookRepoMock) GetWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error) {
	return nil, nil
}

func (m *webhookRepoMock) GetWebhookConfigByID(ctx context.Context, id int) (*model.WebhookConfig, error) {
	return nil, nil
}

func (m *webhookRepoMock) CreateWebhookConfig(ctx context.Context, cfg model.WebhookConfig) (int, error) {
	m.createdCfg = cfg
	return 1, nil
}

func (m *webhookRepoMock) UpdateWebhookConfig(ctx context.Context, id int, cfg model.WebhookConfig) error {
	m.updatedID = id
	m.updatedCfg = cfg
	return nil
}

func (m *webhookRepoMock) DeleteWebhookConfig(ctx context.Context, id int) error {
	return nil
}

func TestCreateWebhookConfig_MapsAllWebhookFields(t *testing.T) {
	repo := &webhookRepoMock{}
	svc := NewWebhookService(repo)

	req := model.WebhookConfigRequest{
		URL:     "  https://slack.com/api/chat.postMessage  ",
		Type:    "slack",
		Token:   "xoxb-test",
		Channel: "C0123456789",
	}

	if _, err := svc.CreateWebhookConfig(context.Background(), req); err != nil {
		t.Fatalf("CreateWebhookConfig() error = %v", err)
	}

	if repo.createdCfg.URL != "https://slack.com/api/chat.postMessage" {
		t.Fatalf("created url = %q, want %q", repo.createdCfg.URL, "https://slack.com/api/chat.postMessage")
	}
	if repo.createdCfg.Type != req.Type {
		t.Fatalf("created type = %q, want %q", repo.createdCfg.Type, req.Type)
	}
	if repo.createdCfg.Token != req.Token {
		t.Fatalf("created token = %q, want %q", repo.createdCfg.Token, req.Token)
	}
	if repo.createdCfg.Channel != req.Channel {
		t.Fatalf("created channel = %q, want %q", repo.createdCfg.Channel, req.Channel)
	}
}

func TestUpdateWebhookConfig_MapsAllWebhookFields(t *testing.T) {
	repo := &webhookRepoMock{}
	svc := NewWebhookService(repo)

	req := model.WebhookConfigRequest{
		URL:     " https://example.com/webhook ",
		Type:    "http",
		Token:   "token-123",
		Channel: "",
	}

	const updateID = 77
	if err := svc.UpdateWebhookConfig(context.Background(), updateID, req); err != nil {
		t.Fatalf("UpdateWebhookConfig() error = %v", err)
	}

	if repo.updatedID != updateID {
		t.Fatalf("updated id = %d, want %d", repo.updatedID, updateID)
	}
	if repo.updatedCfg.URL != "https://example.com/webhook" {
		t.Fatalf("updated url = %q, want %q", repo.updatedCfg.URL, "https://example.com/webhook")
	}
	if repo.updatedCfg.Type != req.Type {
		t.Fatalf("updated type = %q, want %q", repo.updatedCfg.Type, req.Type)
	}
	if repo.updatedCfg.Token != req.Token {
		t.Fatalf("updated token = %q, want %q", repo.updatedCfg.Token, req.Token)
	}
	if repo.updatedCfg.Channel != req.Channel {
		t.Fatalf("updated channel = %q, want %q", repo.updatedCfg.Channel, req.Channel)
	}
}
