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
		Name:    "  Primary Slack Alerts  ",
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
	if repo.createdCfg.Name != "Primary Slack Alerts" {
		t.Fatalf("created name = %q, want %q", repo.createdCfg.Name, "Primary Slack Alerts")
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
		Name:    "  Incident Webhook  ",
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
	if repo.updatedCfg.Name != "Incident Webhook" {
		t.Fatalf("updated name = %q, want %q", repo.updatedCfg.Name, "Incident Webhook")
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

func TestCreateWebhookConfig_RejectsBlankName(t *testing.T) {
	repo := &webhookRepoMock{}
	svc := NewWebhookService(repo)

	_, err := svc.CreateWebhookConfig(context.Background(), model.WebhookConfigRequest{
		Name: "   ",
		URL:  "https://example.com/webhook",
		Type: "http",
	})
	if err == nil {
		t.Fatal("CreateWebhookConfig() error = nil, want error for blank name")
	}
	if repo.createdCfg.Name != "" || repo.createdCfg.URL != "" || repo.createdCfg.Type != "" {
		t.Fatalf("repo should not be called when name is blank, got %+v", repo.createdCfg)
	}
}

func TestUpdateWebhookConfig_RejectsBlankName(t *testing.T) {
	repo := &webhookRepoMock{}
	svc := NewWebhookService(repo)

	err := svc.UpdateWebhookConfig(context.Background(), 77, model.WebhookConfigRequest{
		Name: "   ",
		URL:  "https://example.com/webhook",
		Type: "http",
	})
	if err == nil {
		t.Fatal("UpdateWebhookConfig() error = nil, want error for blank name")
	}
	if repo.updatedID != 0 || repo.updatedCfg.Name != "" || repo.updatedCfg.URL != "" || repo.updatedCfg.Type != "" {
		t.Fatalf("repo should not be called when name is blank, got id=%d cfg=%+v", repo.updatedID, repo.updatedCfg)
	}
}
