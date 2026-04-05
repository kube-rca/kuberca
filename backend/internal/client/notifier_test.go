package client

import (
	"strings"
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/model"
)

func TestSlackClient_ImplementsNotifier(t *testing.T) {
	var _ Notifier = (*SlackClient)(nil)
	var _ ThreadRefStore = (*SlackClient)(nil)
}

func TestSlackClientNotify_AlertStatusChanged_NotConfigured(t *testing.T) {
	c := &SlackClient{}

	err := c.Notify(AlertStatusChangedEvent{
		Alert: model.Alert{
			Status:      "firing",
			Fingerprint: "fp-1",
			Labels: map[string]string{
				"alertname": "HighCPU",
				"severity":  "warning",
				"namespace": "default",
			},
			Annotations: map[string]string{
				"description": "cpu is high",
			},
			StartsAt: time.Now(),
		},
		IncidentID: "incident-1",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type unknownNotifierEvent struct{}

func (unknownNotifierEvent) EventType() string { return "unknown" }

func TestSlackClientNotify_UnknownEvent(t *testing.T) {
	c := &SlackClient{}

	err := c.Notify(unknownNotifierEvent{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported notifier event") {
		t.Fatalf("unexpected error: %v", err)
	}
}
