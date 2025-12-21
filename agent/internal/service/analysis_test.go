package service

import (
	"testing"

	"github.com/kube-rca/agent/internal/model"
)

func TestAnalyzeAlertRequestMessage(t *testing.T) {
	service := NewAnalysisService()
	request := model.AlertAnalysisRequest{
		Alert: model.Alert{
			Status: "firing",
			Labels: map[string]string{"severity": "critical"},
		},
		ThreadTS:    "1234567890.123456",
		CallbackURL: "http://kube-rca-backend.kube-rca.svc:8080/callback/agent",
	}

	got := service.AnalyzeAlertRequest(request)

	if got != analysisCompleteMessage {
		t.Fatalf("message = %q, want %q", got, analysisCompleteMessage)
	}
}
