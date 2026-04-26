package model

import (
	"testing"
)

func TestBuildNotificationRouteKey_NilConfigID(t *testing.T) {
	t.Parallel()
	got := BuildNotificationRouteKey("slack", nil, "C0123456789")
	want := "slack:fallback:C0123456789"
	if got != want {
		t.Errorf("BuildNotificationRouteKey = %q, want %q", got, want)
	}
}

func TestBuildNotificationRouteKey_WithConfigID(t *testing.T) {
	t.Parallel()
	id := 42
	got := BuildNotificationRouteKey("slack", &id, "C0123456789")
	want := "slack:cfg:42:C0123456789"
	if got != want {
		t.Errorf("BuildNotificationRouteKey = %q, want %q", got, want)
	}
}

func TestBuildNotificationRouteKey_DifferentNotifierTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		notifierType string
		cfgID        *int
		channelID    string
		want         string
	}{
		{"slack", nil, "C1", "slack:fallback:C1"},
		{"http", nil, "https://example.com/hook", "http:fallback:https://example.com/hook"},
		{"slack", intPtr(1), "C2", "slack:cfg:1:C2"},
		{"slack", intPtr(99), "C99", "slack:cfg:99:C99"},
	}
	for _, tc := range cases {
		got := BuildNotificationRouteKey(tc.notifierType, tc.cfgID, tc.channelID)
		if got != tc.want {
			t.Errorf("BuildNotificationRouteKey(%q, %v, %q) = %q, want %q",
				tc.notifierType, tc.cfgID, tc.channelID, got, tc.want)
		}
	}
}

func intPtr(v int) *int { return &v }
