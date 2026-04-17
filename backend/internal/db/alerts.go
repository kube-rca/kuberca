package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
			analysis_summary_i18n JSONB NOT NULL DEFAULT '{}',
			analysis_detail_i18n JSONB NOT NULL DEFAULT '{}',
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
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS analysis_summary_i18n JSONB NOT NULL DEFAULT '{}'`,
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS analysis_detail_i18n JSONB NOT NULL DEFAULT '{}'`,
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
		// Partial unique index: 동일 fingerprint의 firing alert는 1건만 허용
		`CREATE UNIQUE INDEX IF NOT EXISTS alerts_fingerprint_firing_uniq ON alerts(fingerprint) WHERE status = 'firing'`,
		// alert_state_transitions에 fingerprint 컬럼 추가 (flapping은 fingerprint 단위)
		`ALTER TABLE alert_state_transitions ADD COLUMN IF NOT EXISTS fingerprint TEXT NOT NULL DEFAULT ''`,
		// 기존 데이터 백필 (현재 alert_id == fingerprint)
		`UPDATE alert_state_transitions SET fingerprint = alert_id WHERE fingerprint = ''`,
		// fingerprint 인덱스
		`CREATE INDEX IF NOT EXISTS alert_state_transitions_fingerprint_idx ON alert_state_transitions(fingerprint, transitioned_at DESC)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

// SaveAlert - Alertmanager 알림을 alerts 테이블에 저장
// 동일 fingerprint + firing 중인 alert가 있으면 UPDATE, 없으면 새 UUID로 INSERT
// 원자적 COALESCE 서브쿼리 + RETURNING으로 TOCTOU race condition 최소화
// 반환: 생성/업데이트된 alertID
func (db *Postgres) SaveAlert(alert model.Alert, incidentID string) (string, error) {
	alertID, err := db.saveAlertInner(alert, incidentID)
	if err != nil {
		// Retry once: concurrent insert로 partial unique index 위반 시
		// 재시도하면 COALESCE가 방금 생성된 firing row를 찾아서 UPDATE
		alertID, err = db.saveAlertInner(alert, incidentID)
	}
	return alertID, err
}

func (db *Postgres) saveAlertInner(alert model.Alert, incidentID string) (string, error) {
	alertName := alert.Labels["alertname"]
	severity := alert.Labels["severity"]
	if severity == "" {
		severity = "warning"
	}

	var incidentIDPtr *string
	if incidentID != "" {
		incidentIDPtr = &incidentID
	}

	newUUID := "ALR-" + uuid.New().String()[:8]

	// 원자적 COALESCE: 동일 fingerprint + firing alert가 있으면 그 ID 재사용, 없으면 새 UUID
	query := `
		INSERT INTO alerts (
			alert_id, incident_id, alarm_title, severity, status, fired_at,
			fingerprint, labels, annotations, created_at, updated_at
		)
		VALUES (
			COALESCE(
				(SELECT alert_id FROM alerts WHERE fingerprint = $7 AND status = 'firing' LIMIT 1),
				$1
			),
			$2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
		ON CONFLICT (alert_id) DO UPDATE SET
			incident_id = COALESCE(EXCLUDED.incident_id, alerts.incident_id),
			alarm_title = EXCLUDED.alarm_title,
			severity = EXCLUDED.severity,
			status = EXCLUDED.status,
			labels = EXCLUDED.labels,
			annotations = EXCLUDED.annotations,
			updated_at = NOW()
		RETURNING alert_id
	`

	var alertID string
	err := db.Pool.QueryRow(context.Background(), query,
		newUUID,           // $1 (fallback UUID)
		incidentIDPtr,     // $2
		alertName,         // $3
		severity,          // $4
		alert.Status,      // $5
		alert.StartsAt,    // $6
		alert.Fingerprint, // $7
		alert.Labels,      // $8
		alert.Annotations, // $9
	).Scan(&alertID)
	return alertID, err
}

// GetFiringAlertByFingerprint - 동일 fingerprint의 firing 중인 alert_id 조회
func (db *Postgres) GetFiringAlertByFingerprint(fingerprint string) (string, error) {
	query := `SELECT alert_id FROM alerts WHERE fingerprint = $1 AND status = 'firing' LIMIT 1`

	var alertID string
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&alertID)
	if err != nil {
		return "", err
	}
	return alertID, nil
}

// GetLatestAlertByFingerprint - fingerprint 기준 최신 alert 조회
func (db *Postgres) GetLatestAlertByFingerprint(fingerprint string) (*model.AlertDetailResponse, error) {
	query := `
		SELECT
			alert_id, incident_id, alarm_title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail, analysis_summary_i18n, analysis_detail_i18n,
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
		FROM alerts
		WHERE fingerprint = $1
		ORDER BY fired_at DESC
		LIMIT 1
	`

	var a model.AlertDetailResponse
	var summaryI18n, detailI18n []byte
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(
		&a.AlertID,
		&a.IncidentID,
		&a.AlarmTitle,
		&a.Severity,
		&a.Status,
		&a.FiredAt,
		&a.ResolvedAt,
		&a.AnalysisSummary,
		&a.AnalysisDetail,
		&summaryI18n,
		&detailI18n,
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
	a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
	a.AnalysisDetailI18n = toLocalizedText(detailI18n, strPtrValue(a.AnalysisDetail))
	return &a, nil
}

// GetAlertList - Alert 목록 조회
func (db *Postgres) GetAlertList() ([]model.AlertListResponse, error) {
	query := `
		SELECT alert_id, incident_id, alarm_title, labels->>'namespace' as namespace, severity, status, fired_at, resolved_at, analysis_summary, analysis_summary_i18n, labels
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
		var summaryI18n []byte
		if err := rows.Scan(&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Namespace, &a.Severity, &a.Status, &a.FiredAt, &a.ResolvedAt, &a.AnalysisSummary, &summaryI18n, &a.Labels); err != nil {
			return nil, err
		}
		a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
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
		SELECT alert_id, incident_id, alarm_title, severity, status, fired_at, resolved_at, analysis_summary, analysis_summary_i18n
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
		var summaryI18n []byte
		if err := rows.Scan(&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Severity, &a.Status, &a.FiredAt, &a.ResolvedAt, &a.AnalysisSummary, &summaryI18n); err != nil {
			return nil, err
		}
		a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
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
			fired_at, resolved_at, analysis_summary, analysis_detail, analysis_summary_i18n, analysis_detail_i18n,
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
		var summaryI18n, detailI18n []byte
		if err := rows.Scan(
			&a.AlertID, &a.IncidentID, &a.AlarmTitle, &a.Severity, &a.Status,
			&a.FiredAt, &a.ResolvedAt, &a.AnalysisSummary, &a.AnalysisDetail, &summaryI18n, &detailI18n,
			&a.Fingerprint, &a.ThreadTS, &a.Labels, &a.Annotations,
			&a.IsFlapping, &a.FlapCycleCount, &a.FlapWindowStart,
		); err != nil {
			return nil, err
		}
		a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
		a.AnalysisDetailI18n = toLocalizedText(detailI18n, strPtrValue(a.AnalysisDetail))
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
			fired_at, resolved_at, analysis_summary, analysis_detail, analysis_summary_i18n, analysis_detail_i18n,
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
		FROM alerts
		WHERE alert_id = $1
	`

	var a model.AlertDetailResponse
	var summaryI18n, detailI18n []byte
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
		&summaryI18n,
		&detailI18n,
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
	a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
	a.AnalysisDetailI18n = toLocalizedText(detailI18n, strPtrValue(a.AnalysisDetail))
	return &a, nil
}

// GetAlertDetailInsensitive - Alert 상세 조회 (대소문자 무시)
func (db *Postgres) GetAlertDetailInsensitive(alertID string) (*model.AlertDetailResponse, error) {
	query := `
		SELECT
			alert_id, incident_id, alarm_title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail, analysis_summary_i18n, analysis_detail_i18n,
			fingerprint, thread_ts, labels, annotations,
			is_flapping, flap_cycle_count, flap_window_start
		FROM alerts
		WHERE lower(alert_id) = lower($1)
		LIMIT 1
	`

	var a model.AlertDetailResponse
	var summaryI18n, detailI18n []byte
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
		&summaryI18n,
		&detailI18n,
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
	a.AnalysisSummaryI18n = toLocalizedText(summaryI18n, strPtrValue(a.AnalysisSummary))
	a.AnalysisDetailI18n = toLocalizedText(detailI18n, strPtrValue(a.AnalysisDetail))
	return &a, nil
}

// UpdateAlertThreadTS - fingerprint 기준 firing alert에 Slack thread_ts 저장
func (db *Postgres) UpdateAlertThreadTS(fingerprint, threadTS string) error {
	query := `
		UPDATE alerts
		SET thread_ts = $2, updated_at = NOW()
		WHERE fingerprint = $1 AND status = 'firing'
	`
	_, err := db.Pool.Exec(context.Background(), query, fingerprint, threadTS)
	return err
}

// GetAlertThreadTS - fingerprint 기준 최신 alert의 thread_ts 조회
func (db *Postgres) GetAlertThreadTS(fingerprint string) (string, bool) {
	query := `
		SELECT thread_ts FROM alerts
		WHERE fingerprint = $1 AND thread_ts != ''
		ORDER BY fired_at DESC LIMIT 1
	`

	var threadTS string
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&threadTS)
	if err != nil || threadTS == "" {
		return "", false
	}
	return threadTS, true
}

// UpdateAlertResolved - fingerprint 기준 alert에 resolved_at 설정
// SaveAlert이 이미 status='resolved'로 변경했으므로, resolved_at IS NULL인 row를 찾아 갱신
func (db *Postgres) UpdateAlertResolved(fingerprint string, resolvedAt time.Time) error {
	query := `
		UPDATE alerts
		SET status = 'resolved', resolved_at = $2, updated_at = NOW()
		WHERE fingerprint = $1 AND status = 'resolved' AND resolved_at IS NULL
	`
	_, err := db.Pool.Exec(context.Background(), query, fingerprint, resolvedAt)
	return err
}

// UpdateAlertAnalysis - Alert 분석 결과 저장
func (db *Postgres) UpdateAlertAnalysis(alertID, summary, detail string, summaryI18n, detailI18n model.LocalizedText) error {
	query := `
		UPDATE alerts
		SET analysis_summary = $2,
		    analysis_detail = $3,
		    analysis_summary_i18n = $4::jsonb,
		    analysis_detail_i18n = $5::jsonb,
		    updated_at = NOW()
		WHERE alert_id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, alertID, summary, detail, localizedJSON(summaryI18n, summary), localizedJSON(detailI18n, detail))
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

// IsAlertAlreadyResolved - fingerprint 기준 최신 alert가 이미 resolved 상태인지 확인
func (db *Postgres) IsAlertAlreadyResolved(fingerprint string, endsAt time.Time) (bool, error) {
	query := `
		SELECT resolved_at FROM alerts
		WHERE fingerprint = $1
		ORDER BY fired_at DESC LIMIT 1
	`

	var resolvedAt *time.Time
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&resolvedAt)
	if err != nil {
		return false, err
	}
	if resolvedAt == nil {
		return false, nil
	}
	// 동일/과거 endsAt은 중복으로 간주
	return !endsAt.After(*resolvedAt), nil
}

// RecordStateTransition - Alert 상태 전환 기록 (fingerprint 기반)
func (db *Postgres) RecordStateTransition(fingerprint, fromStatus, toStatus string, timestamp time.Time) error {
	query := `
		INSERT INTO alert_state_transitions (alert_id, fingerprint, from_status, to_status, transitioned_at)
		VALUES ($1, $1, $2, $3, $4)
	`
	_, err := db.Pool.Exec(context.Background(), query, fingerprint, fromStatus, toStatus, timestamp)
	return err
}

// GetAlertCurrentStatus - fingerprint 기준 최신 alert의 현재 상태 조회
func (db *Postgres) GetAlertCurrentStatus(fingerprint string) (string, error) {
	query := `SELECT status FROM alerts WHERE fingerprint = $1 ORDER BY fired_at DESC LIMIT 1`

	var status string
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// IsAlertFlapping - fingerprint 기준 최신 alert가 flapping 상태인지 확인
func (db *Postgres) IsAlertFlapping(fingerprint string) bool {
	query := `SELECT is_flapping FROM alerts WHERE fingerprint = $1 ORDER BY fired_at DESC LIMIT 1`

	var isFlapping bool
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&isFlapping)
	if err != nil {
		return false
	}
	return isFlapping
}

// CountFlappingCycles - 지정된 시간 윈도우 내 firing→resolved 사이클 수 계산
// Returns: (cycleCount, windowStart, error)
func (db *Postgres) CountFlappingCycles(fingerprint string, windowMinutes int) (int, time.Time, error) {
	// 현재 flapping 윈도우 정보 조회 (fingerprint 기준 최신)
	var flapWindowStart *time.Time
	var currentCycleCount int

	query := `SELECT flap_window_start, flap_cycle_count FROM alerts WHERE fingerprint = $1 ORDER BY fired_at DESC LIMIT 1`
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(&flapWindowStart, &currentCycleCount)
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

	// 윈도우 내 resolved 전환 횟수 카운트 (fingerprint 기반)
	countQuery := `
		SELECT COUNT(*)
		FROM alert_state_transitions
		WHERE fingerprint = $1
		  AND to_status = 'resolved'
		  AND transitioned_at >= $2
	`

	var count int
	err = db.Pool.QueryRow(context.Background(), countQuery, fingerprint, flapWindowStart).Scan(&count)
	if err != nil {
		return 0, *flapWindowStart, err
	}

	return count, *flapWindowStart, nil
}

// MarkAlertAsFlapping - fingerprint 기준 firing alert의 flapping 상태 설정/해제
func (db *Postgres) MarkAlertAsFlapping(fingerprint string, isFlapping bool, cycleCount int, windowStart time.Time) error {
	var query string

	if isFlapping {
		query = `
			UPDATE alerts
			SET is_flapping = $2,
			    flap_cycle_count = $3,
			    flap_window_start = $4,
			    last_flap_notification_at = NOW(),
			    updated_at = NOW()
			WHERE fingerprint = $1 AND status = 'firing'
		`
		_, err := db.Pool.Exec(context.Background(), query, fingerprint, isFlapping, cycleCount, windowStart)
		return err
	}

	// Flapping 해제 (최신 alert 대상)
	query = `
		UPDATE alerts
		SET is_flapping = FALSE,
		    flap_cycle_count = 0,
		    flap_window_start = NULL,
		    updated_at = NOW()
		WHERE fingerprint = $1 AND is_flapping = TRUE
	`
	_, err := db.Pool.Exec(context.Background(), query, fingerprint)
	return err
}

// UpdateFlappingCycleCount - fingerprint 기준 firing alert의 Flapping cycle 수 업데이트
func (db *Postgres) UpdateFlappingCycleCount(fingerprint string, cycleCount int) error {
	query := `
		UPDATE alerts
		SET flap_cycle_count = $2, updated_at = NOW()
		WHERE fingerprint = $1 AND status = 'firing'
	`
	_, err := db.Pool.Exec(context.Background(), query, fingerprint, cycleCount)
	return err
}

// HasTransitionsSince - 지정된 시각 이후 상태 전환이 있었는지 확인 (fingerprint 기반)
func (db *Postgres) HasTransitionsSince(fingerprint string, since time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM alert_state_transitions
		WHERE fingerprint = $1 AND transitioned_at > $2
	`

	var count int
	err := db.Pool.QueryRow(context.Background(), query, fingerprint, since).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ManualResolveAlert - alert_id 기준으로 수동 resolve (status='firing' → 'resolved')
func (db *Postgres) ManualResolveAlert(alertID string) error {
	query := `
		UPDATE alerts
		SET status = 'resolved', resolved_at = NOW(), updated_at = NOW()
		WHERE alert_id = $1 AND status = 'firing'
	`
	result, err := db.Pool.Exec(context.Background(), query, alertID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found or already resolved: %s", alertID)
	}
	return nil
}
