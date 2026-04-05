package model

import (
	"fmt"
	"time"
)

// AlertNotificationDelivery는 alert별 Slack thread delivery source of truth다.
// route_key는 같은 alert가 같은 목적지(channel/config)로 재전송될 때 최신 root thread를 덮어쓰기 위해 사용한다.
type AlertNotificationDelivery struct {
	DeliveryID      int64      `json:"delivery_id"`
	AlertID         string     `json:"alert_id"`
	Fingerprint     string     `json:"fingerprint"`
	IncidentID      *string    `json:"incident_id,omitempty"`
	NotifierType    string     `json:"notifier_type"`
	WebhookConfigID *int       `json:"webhook_config_id,omitempty"`
	RouteKey        string     `json:"route_key"`
	ChannelID       string     `json:"channel_id"`
	RootMessageTS   string     `json:"root_message_ts"`
	ThreadTS        string     `json:"thread_ts"`
	Status          string     `json:"status"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
}

func BuildNotificationRouteKey(notifierType string, webhookConfigID *int, channelID string) string {
	if webhookConfigID == nil {
		return fmt.Sprintf("%s:fallback:%s", notifierType, channelID)
	}
	return fmt.Sprintf("%s:cfg:%d:%s", notifierType, *webhookConfigID, channelID)
}
