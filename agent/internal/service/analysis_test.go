package service

import (
	"testing"

	"github.com/kube-rca/agent/internal/model"
)

func TestAnalyzeWebhookMessage(t *testing.T) {
	service := NewAnalysisService()
	webhook := model.AlertmanagerWebhook{
		Alerts: []model.Alert{
			{Status: "firing", Labels: map[string]string{"severity": "critical"}},
			{Status: "resolved", Labels: map[string]string{"severity": "warning"}},
			{Status: "", Labels: map[string]string{}},
		},
	}

	got := service.AnalyzeWebhook(webhook)

	if got != analysisCompleteMessage {
		t.Fatalf("message = %q, want %q", got, analysisCompleteMessage)
	}
}
