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
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS is_flapping BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS flap_cycle_count INT NOT NULL DEFAULT 0`,
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS flap_window_start TIMESTAMPTZ`,
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS last_flap_notification_at TIMESTAMPTZ`,
		`
		CREATE TABLE IF NOT EXISTS alert_state_transitions (
			transition_id SERIAL PRIMARY KEY,
			alert_id TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			transitioned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS alerts_incident_id_idx ON alerts(incident_id) WHERE incident_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS alerts_fingerprint_idx ON alerts(fingerprint) WHERE fingerprint != ''`,
		`CREATE INDEX IF NOT EXISTS alerts_thread_ts_idx ON alerts(thread_ts) WHERE thread_ts != ''`,
		`CREATE INDEX IF NOT EXISTS alerts_status_idx ON alerts(status)`,
		`CREATE INDEX IF NOT EXISTS alerts_fired_at_idx ON alerts(fired_at DESC)`,
		`CREATE INDEX IF NOT EXISTS alerts_is_flapping_idx ON alerts(is_flapping) WHERE is_flapping = TRUE`,
		`CREATE INDEX IF NOT EXISTS alert_state_transitions_alert_id_idx ON alert_state_transitions(alert_id, transitioned_at DESC)`,
		`CREATE INDEX IF NOT EXISTS alert_state_transitions_time_idx ON alert_state_transitions(transitioned_at DESC)`,
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
		SELECT alert_id, incident_id, alarm_title, labels->>'namespace' as namespace, severity, status, fired_at, resolved_at, labels
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
		if err := rows.Scan(&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Namespace, &a.Severity, &a.Status, &a.FiredAt, &a.ResolvedAt, &a.Labels); err != nil {
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
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
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
			&a.IsFlapping, &a.FlapCycleCount, &a.FlapWindowStart,
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
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
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
		&a.IsFlapping,
		&a.FlapCycleCount,
		&a.FlapWindowStart,
	)

	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAlertDetailInsensitive - Alert 상세 조회 (대소문자 무시)
func (db *Postgres) GetAlertDetailInsensitive(alertID string) (*model.AlertDetailResponse, error) {
	query := `
		SELECT
			alert_id, incident_id, alarm_title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail,
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
		FROM alerts
		WHERE lower(alert_id) = lower($1)
		LIMIT 1
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
		&a.IsFlapping,
		&a.FlapCycleCount,
		&a.FlapWindowStart,
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

// IsAlertAlreadyResolved - Alert가 이미 resolved 상태인지 확인
func (db *Postgres) IsAlertAlreadyResolved(alertID string, endsAt time.Time) (bool, error) {
	query := `
		SELECT resolved_at FROM alerts
		WHERE alert_id = $1
	`

	var resolvedAt *time.Time
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(&resolvedAt)
	if err != nil {
		return false, err
	}
	if resolvedAt == nil {
		return false, nil
	}
	// 동일/과거 endsAt은 중복으로 간주
	return !endsAt.After(*resolvedAt), nil
}

// RecordStateTransition - Alert 상태 전환 기록
func (db *Postgres) RecordStateTransition(alertID, fromStatus, toStatus string, timestamp time.Time) error {
	query := `
		INSERT INTO alert_state_transitions (alert_id, from_status, to_status, transitioned_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, fromStatus, toStatus, timestamp)
	return err
}

// GetAlertCurrentStatus - Alert의 현재 상태 조회
func (db *Postgres) GetAlertCurrentStatus(alertID string) (string, error) {
	query := `SELECT status FROM alerts WHERE alert_id = $1`

	var status string
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// IsAlertFlapping - Alert가 flapping 상태인지 확인
func (db *Postgres) IsAlertFlapping(alertID string) bool {
	query := `SELECT is_flapping FROM alerts WHERE alert_id = $1`

	var isFlapping bool
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(&isFlapping)
	if err != nil {
		return false
	}
	return isFlapping
}

// CountFlappingCycles - 지정된 시간 윈도우 내 firing→resolved 사이클 수 계산
// Returns: (cycleCount, windowStart, error)
func (db *Postgres) CountFlappingCycles(alertID string, windowMinutes int) (int, time.Time, error) {
	// 현재 flapping 윈도우 정보 조회
	var flapWindowStart *time.Time
	var currentCycleCount int

	query := `SELECT flap_window_start, flap_cycle_count FROM alerts WHERE alert_id = $1`
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(&flapWindowStart, &currentCycleCount)
	if err != nil {
		return 0, time.Time{}, err
	}

	// 윈도우가 시작되지 않은 경우, 새 윈도우 시작
	if flapWindowStart == nil {
		newWindowStart := time.Now()
		return 1, newWindowStart, nil
	}

	// 윈도우 만료 확인
	windowStart := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	if flapWindowStart.Before(windowStart) {
		// 윈도우 만료, 새 윈도우 시작
		newWindowStart := time.Now()
		return 1, newWindowStart, nil
	}

	// 윈도우 내 resolved 전환 횟수 카운트
	countQuery := `
		SELECT COUNT(*)
		FROM alert_state_transitions
		WHERE alert_id = $1
		  AND to_status = 'resolved'
		  AND transitioned_at >= $2
	`

	var count int
	err = db.Pool.QueryRow(context.Background(), countQuery, alertID, flapWindowStart).Scan(&count)
	if err != nil {
		return 0, *flapWindowStart, err
	}

	return count, *flapWindowStart, nil
}

// MarkAlertAsFlapping - Alert flapping 상태 설정/해제
func (db *Postgres) MarkAlertAsFlapping(alertID string, isFlapping bool, cycleCount int, windowStart time.Time) error {
	var query string

	if isFlapping {
		query = `
			UPDATE alerts
			SET is_flapping = $2,
			    flap_cycle_count = $3,
			    flap_window_start = $4,
			    last_flap_notification_at = NOW(),
			    updated_at = NOW()
			WHERE alert_id = $1
		`
		_, err := db.Pool.Exec(context.Background(), query, alertID, isFlapping, cycleCount, windowStart)
		return err
	}

	// Flapping 해제
	query = `
		UPDATE alerts
		SET is_flapping = FALSE,
		    flap_cycle_count = 0,
		    flap_window_start = NULL,
		    updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID)
	return err
}

// UpdateFlappingCycleCount - Flapping cycle 수 업데이트
func (db *Postgres) UpdateFlappingCycleCount(alertID string, cycleCount int) error {
	query := `
		UPDATE alerts
		SET flap_cycle_count = $2, updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, cycleCount)
	return err
}

// HasTransitionsSince - 지정된 시각 이후 상태 전환이 있었는지 확인
func (db *Postgres) HasTransitionsSince(alertID string, since time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM alert_state_transitions
		WHERE alert_id = $1 AND transitioned_at > $2
	`

	var count int
	err := db.Pool.QueryRow(context.Background(), query, alertID, since).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
