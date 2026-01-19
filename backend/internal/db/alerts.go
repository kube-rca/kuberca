package db

import (
	"context"
	"time"

	"github.com/kube-rca/backend/internal/model"
)

// EnsureAlertSchema - alerts 테이블 생성
func (db *Postgres) EnsureAlertSchema() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS alerts (
			alert_id TEXT PRIMARY KEY,
			incident_id TEXT,
			alarm_title TEXT NOT NULL DEFAULT '',
			severity TEXT NOT NULL DEFAULT 'warning',
			status TEXT NOT NULL DEFAULT 'firing',
			fired_at TIMESTAMPTZ,
			resolved_at TIMESTAMPTZ,
			analysis_summary TEXT NOT NULL DEFAULT '',
			analysis_detail TEXT NOT NULL DEFAULT '',
			fingerprint TEXT NOT NULL DEFAULT '',
			thread_ts TEXT NOT NULL DEFAULT '',
			labels JSONB NOT NULL DEFAULT '{}',
			annotations JSONB NOT NULL DEFAULT '{}',
			is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS alerts_incident_id_idx ON alerts(incident_id) WHERE incident_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS alerts_fingerprint_idx ON alerts(fingerprint) WHERE fingerprint != ''`,
		`CREATE INDEX IF NOT EXISTS alerts_thread_ts_idx ON alerts(thread_ts) WHERE thread_ts != ''`,
		`CREATE INDEX IF NOT EXISTS alerts_status_idx ON alerts(status)`,
		`CREATE INDEX IF NOT EXISTS alerts_fired_at_idx ON alerts(fired_at DESC)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

// SaveAlert - Alertmanager 알림을 alerts 테이블에 저장
func (db *Postgres) SaveAlert(alert model.Alert, incidentID string) error {
	alertName := alert.Labels["alertname"]
	severity := alert.Labels["severity"]
	if severity == "" {
		severity = "warning"
	}

	var incidentIDPtr *string
	if incidentID != "" {
		incidentIDPtr = &incidentID
	}

	query := `
		INSERT INTO alerts (
			alert_id, incident_id, alarm_title, severity, status, fired_at,
			fingerprint, labels, annotations, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (alert_id) DO UPDATE SET
			incident_id = COALESCE(EXCLUDED.incident_id, alerts.incident_id),
			status = EXCLUDED.status,
			updated_at = NOW()
	`

	_, err := db.Pool.Exec(context.Background(), query,
		alert.Fingerprint, // alert_id == fingerprint
		incidentIDPtr,
		alertName,
		severity,
		alert.Status,
		alert.StartsAt,
		alert.Fingerprint,
		alert.Labels,
		alert.Annotations,
	)
	return err
}

// GetAlertList - Alert 목록 조회
func (db *Postgres) GetAlertList() ([]model.AlertListResponse, error) {
	query := `
		SELECT alert_id, incident_id, alarm_title, labels->>'namespace' as namespace, severity, status, fired_at, resolved_at
		FROM alerts
		WHERE is_enabled = TRUE
		ORDER BY fired_at DESC`

	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AlertListResponse
	for rows.Next() {
		var a model.AlertListResponse
		if err := rows.Scan(&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Namespace, &a.Severity, &a.Status, &a.FiredAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	if list == nil {
		list = []model.AlertListResponse{}
	}
	return list, nil
}

// GetAlertsByIncidentID - 특정 Incident에 속한 Alert 목록 조회
func (db *Postgres) GetAlertsByIncidentID(incidentID string) ([]model.AlertListResponse, error) {
	query := `
		SELECT alert_id, incident_id, alarm_title, severity, status, fired_at, resolved_at
		FROM alerts
		WHERE incident_id = $1 AND is_enabled = TRUE
		ORDER BY fired_at DESC`

	rows, err := db.Pool.Query(context.Background(), query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AlertListResponse
	for rows.Next() {
		var a model.AlertListResponse
		if err := rows.Scan(&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Severity, &a.Status, &a.FiredAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	if list == nil {
		list = []model.AlertListResponse{}
	}
	return list, nil
}

// GetAlertsWithAnalysisByIncidentID - 특정 Incident에 속한 Alert 목록 조회 (분석 내용 포함)
func (db *Postgres) GetAlertsWithAnalysisByIncidentID(incidentID string) ([]model.AlertDetailResponse, error) {
	query := `
		SELECT
			alert_id, incident_id, alarm_title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail,
			fingerprint, thread_ts, labels, annotations
		FROM alerts
		WHERE incident_id = $1 AND is_enabled = TRUE
		ORDER BY fired_at DESC`

	rows, err := db.Pool.Query(context.Background(), query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AlertDetailResponse
	for rows.Next() {
		var a model.AlertDetailResponse
		if err := rows.Scan(
			&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Severity, &a.Status,
			&a.FiredAt, &a.ResolvedAt, &a.AnalysisSummary, &a.AnalysisDetail,
			&a.Fingerprint, &a.ThreadTS, &a.Labels, &a.Annotations,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	if list == nil {
		list = []model.AlertDetailResponse{}
	}
	return list, nil
}

// GetAlertDetail - Alert 상세 조회
func (db *Postgres) GetAlertDetail(alertID string) (*model.AlertDetailResponse, error) {
	query := `
		SELECT
			alert_id, incident_id, alarm_title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail,
			fingerprint, thread_ts, labels, annotations
		FROM alerts
		WHERE alert_id = $1
	`

	var a model.AlertDetailResponse
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(
		&a.AlertID,
		&a.IncidentID,
		&a.AlarmTitle,
		&a.Severity,
		&a.Status,
		&a.FiredAt,
		&a.ResolvedAt,
		&a.AnalysisSummary,
		&a.AnalysisDetail,
		&a.Fingerprint,
		&a.ThreadTS,
		&a.Labels,
		&a.Annotations,
	)

	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAlertThreadTS - Alert에 Slack thread_ts 저장
func (db *Postgres) UpdateAlertThreadTS(alertID, threadTS string) error {
	query := `
		UPDATE alerts
		SET thread_ts = $2, updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, threadTS)
	return err
}

// GetAlertThreadTS - Alert의 thread_ts 조회
func (db *Postgres) GetAlertThreadTS(alertID string) (string, bool) {
	query := `
		SELECT thread_ts FROM alerts
		WHERE alert_id = $1 AND thread_ts != ''
	`

	var threadTS string
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(&threadTS)
	if err != nil || threadTS == "" {
		return "", false
	}
	return threadTS, true
}

// UpdateAlertResolved - Alert resolved 상태로 업데이트
func (db *Postgres) UpdateAlertResolved(alertID string, resolvedAt time.Time) error {
	query := `
		UPDATE alerts
		SET status = 'resolved', resolved_at = $2, updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, resolvedAt)
	return err
}

// UpdateAlertAnalysis - Alert 분석 결과 저장
func (db *Postgres) UpdateAlertAnalysis(alertID, summary, detail string) error {
	query := `
		UPDATE alerts
		SET analysis_summary = $2, analysis_detail = $3, updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, summary, detail)
	return err
}

// UpdateAlertIncidentID - Alert의 Incident ID 변경
func (db *Postgres) UpdateAlertIncidentID(alertID, incidentID string) error {
	query := `
		UPDATE alerts
		SET incident_id = $2, updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, incidentID)
	return err
}
