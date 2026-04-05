package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kube-rca/backend/internal/model"
)

func (db *Postgres) EnsureAlertNotificationDeliverySchema() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS alert_notification_deliveries (
			delivery_id BIGSERIAL PRIMARY KEY,
			alert_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL DEFAULT '',
			incident_id TEXT,
			notifier_type TEXT NOT NULL DEFAULT 'slack',
			webhook_config_id INT,
			route_key TEXT NOT NULL DEFAULT '',
			channel_id TEXT NOT NULL DEFAULT '',
			root_message_ts TEXT NOT NULL DEFAULT '',
			thread_ts TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'firing',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			last_used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE UNIQUE INDEX IF NOT EXISTS alert_notification_deliveries_alert_route_uniq ON alert_notification_deliveries(alert_id, route_key)`,
		`CREATE INDEX IF NOT EXISTS alert_notification_deliveries_alert_id_idx ON alert_notification_deliveries(alert_id)`,
		`CREATE INDEX IF NOT EXISTS alert_notification_deliveries_fingerprint_idx ON alert_notification_deliveries(fingerprint) WHERE fingerprint != ''`,
		`CREATE INDEX IF NOT EXISTS alert_notification_deliveries_incident_id_idx ON alert_notification_deliveries(incident_id) WHERE incident_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS alert_notification_deliveries_active_idx ON alert_notification_deliveries(is_active) WHERE is_active = TRUE`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

func (db *Postgres) UpsertAlertNotificationDeliveries(deliveries []model.AlertNotificationDelivery) error {
	if len(deliveries) == 0 {
		return nil
	}

	ctx := context.Background()
	query := `
		INSERT INTO alert_notification_deliveries (
			alert_id, fingerprint, incident_id, notifier_type, webhook_config_id, route_key,
			channel_id, root_message_ts, thread_ts, status, is_active, last_used_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (alert_id, route_key) DO UPDATE SET
			fingerprint = EXCLUDED.fingerprint,
			incident_id = EXCLUDED.incident_id,
			notifier_type = EXCLUDED.notifier_type,
			webhook_config_id = EXCLUDED.webhook_config_id,
			channel_id = EXCLUDED.channel_id,
			root_message_ts = EXCLUDED.root_message_ts,
			thread_ts = EXCLUDED.thread_ts,
			status = EXCLUDED.status,
			is_active = EXCLUDED.is_active,
			last_used_at = EXCLUDED.last_used_at,
			updated_at = NOW()
	`

	for _, delivery := range deliveries {
		if delivery.AlertID == "" || delivery.RouteKey == "" || delivery.ChannelID == "" || delivery.ThreadTS == "" {
			return fmt.Errorf("invalid notification delivery: alert_id=%q route_key=%q channel_id=%q thread_ts=%q", delivery.AlertID, delivery.RouteKey, delivery.ChannelID, delivery.ThreadTS)
		}
		if delivery.RootMessageTS == "" {
			delivery.RootMessageTS = delivery.ThreadTS
		}
		if _, err := db.Pool.Exec(
			ctx,
			query,
			delivery.AlertID,
			delivery.Fingerprint,
			delivery.IncidentID,
			delivery.NotifierType,
			delivery.WebhookConfigID,
			delivery.RouteKey,
			delivery.ChannelID,
			delivery.RootMessageTS,
			delivery.ThreadTS,
			delivery.Status,
			delivery.IsActive,
			delivery.LastUsedAt,
		); err != nil {
			return fmt.Errorf("failed to upsert alert notification delivery: %w", err)
		}
	}
	return nil
}

func (db *Postgres) GetAlertNotificationDeliveries(alertID string) ([]model.AlertNotificationDelivery, error) {
	query := `
		SELECT
			delivery_id, alert_id, fingerprint, incident_id, notifier_type, webhook_config_id,
			route_key, channel_id, root_message_ts, thread_ts, status, is_active,
			created_at, updated_at, last_used_at
		FROM alert_notification_deliveries
		WHERE alert_id = $1 AND is_active = TRUE
		ORDER BY created_at ASC
	`

	rows, err := db.Pool.Query(context.Background(), query, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert notification deliveries: %w", err)
	}
	defer rows.Close()

	deliveries := make([]model.AlertNotificationDelivery, 0)
	for rows.Next() {
		var delivery model.AlertNotificationDelivery
		var incidentID sql.NullString
		var webhookConfigID sql.NullInt32
		var lastUsedAt sql.NullTime
		if err := rows.Scan(
			&delivery.DeliveryID,
			&delivery.AlertID,
			&delivery.Fingerprint,
			&incidentID,
			&delivery.NotifierType,
			&webhookConfigID,
			&delivery.RouteKey,
			&delivery.ChannelID,
			&delivery.RootMessageTS,
			&delivery.ThreadTS,
			&delivery.Status,
			&delivery.IsActive,
			&delivery.CreatedAt,
			&delivery.UpdatedAt,
			&lastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert notification delivery: %w", err)
		}
		if incidentID.Valid {
			delivery.IncidentID = &incidentID.String
		}
		if webhookConfigID.Valid {
			id := int(webhookConfigID.Int32)
			delivery.WebhookConfigID = &id
		}
		if lastUsedAt.Valid {
			ts := lastUsedAt.Time
			delivery.LastUsedAt = &ts
		}
		deliveries = append(deliveries, delivery)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate alert notification deliveries: %w", rows.Err())
	}
	return deliveries, nil
}

func (db *Postgres) TouchAlertNotificationDeliveries(alertID string, at time.Time) error {
	query := `
		UPDATE alert_notification_deliveries
		SET last_used_at = $2, updated_at = NOW()
		WHERE alert_id = $1 AND is_active = TRUE
	`
	if _, err := db.Pool.Exec(context.Background(), query, alertID, at); err != nil {
		return fmt.Errorf("failed to update alert notification delivery last_used_at: %w", err)
	}
	return nil
}
