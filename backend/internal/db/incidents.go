package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kube-rca/backend/internal/model"
)

// Postgres 구조체의 필드 타입을 NewPostgresPool의 리턴 타입과 맞춥니다.
type Postgres struct {
	Pool *pgxpool.Pool
}

func (db *Postgres) GetIncidentList() ([]model.IncidentListResponse, error) {
	query := `
		SELECT incident_id, alarm_title, severity, fired_at, resolved_at 
		FROM incidents 
		ORDER BY created_at DESC`

	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IncidentListResponse
	for rows.Next() {
		var i model.IncidentListResponse
		if err := rows.Scan(&i.IncidentID, &i.AlarmTitle, &i.Severity, &i.ResolvedAt); err != nil {
			return nil, err
		}
		list = append(list, i)
	}

	if list == nil {
		list = []model.IncidentListResponse{}
	}
	return list, nil
}

func (db *Postgres) GetIncidentDetail(id string) (*model.IncidentDetailResponse, error) {
	query := `
		SELECT 
			incident_id, alarm_title, severity, status, 
			fired_at, resolved_at, analysis_summary, analysis_detail, similar_incidents
		FROM incidents
		WHERE incident_id = $1
	`

	var i model.IncidentDetailResponse

	err := db.Pool.QueryRow(context.Background(), query, id).Scan(
		&i.IncidentID,
		&i.AlarmTitle,
		&i.Severity,
		&i.Status,
		&i.FiredAt,
		&i.ResolvedAt,
		&i.AnalysisSummary,
		&i.AnalysisDetail,
		&i.SimilarIncidents,
	)

	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (db *Postgres) UpdateIncident(id string, req model.UpdateIncidentRequest) error {
	query := `
		UPDATE incidents 
		SET 
			alarm_title = $1, 
			severity = $2, 
			analysis_summary = $3, 
			analysis_detail = $4,
			updated_at = NOW()
		WHERE incident_id = $5
	`

	commandTag, err := db.Pool.Exec(context.Background(), query,
		req.AlarmTitle,
		req.Severity,
		req.AnalysisSummary,
		req.AnalysisDetail,
		id,
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no incident found with id: %s", id)
	}

	return nil
}

// Mock 데이터 생성 추후 삭제 에정
func (db *Postgres) CreateIncident(id, title, severity, status string) error {
	query := `
		INSERT INTO incidents (incident_id, alarm_title, severity, status, fired_at, created_at, similar_incidents)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), '{}')
	`

	_, err := db.Pool.Exec(context.Background(), query, id, title, severity, status)
	return err
}

// ============================================================================
// Alertmanager/Agent 통합용 메서드
// ============================================================================

// EnsureIncidentSchema - incidents 테이블에 새 컬럼 추가 (fingerprint, thread_ts, labels, annotations)
// 기존 테이블이 있으면 컬럼만 추가, 없으면 전체 생성
func (db *Postgres) EnsureIncidentSchema(ctx context.Context) error {
	queries := []string{
		// 기존 테이블이 없으면 생성
		`
		CREATE TABLE IF NOT EXISTS incidents (
			incident_id TEXT PRIMARY KEY,
			alarm_title TEXT NOT NULL DEFAULT '',
			severity TEXT NOT NULL DEFAULT 'warning',
			status TEXT NOT NULL DEFAULT 'firing',
			fired_at TIMESTAMPTZ,
			resolved_at TIMESTAMPTZ,
			analysis_summary TEXT NOT NULL DEFAULT '',
			analysis_detail TEXT NOT NULL DEFAULT '',
			similar_incidents JSONB NOT NULL DEFAULT '[]',
			fingerprint TEXT NOT NULL DEFAULT '',
			thread_ts TEXT NOT NULL DEFAULT '',
			labels JSONB NOT NULL DEFAULT '{}',
			annotations JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		// 기존 테이블에 새 컬럼 추가 (이미 있으면 무시)
		`ALTER TABLE incidents ADD COLUMN IF NOT EXISTS fingerprint TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE incidents ADD COLUMN IF NOT EXISTS thread_ts TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE incidents ADD COLUMN IF NOT EXISTS labels JSONB NOT NULL DEFAULT '{}'`,
		`ALTER TABLE incidents ADD COLUMN IF NOT EXISTS annotations JSONB NOT NULL DEFAULT '{}'`,
		// 인덱스 생성
		`CREATE INDEX IF NOT EXISTS incidents_fingerprint_idx ON incidents(fingerprint) WHERE fingerprint != ''`,
		`CREATE INDEX IF NOT EXISTS incidents_thread_ts_idx ON incidents(thread_ts) WHERE thread_ts != ''`,
		`CREATE INDEX IF NOT EXISTS incidents_status_idx ON incidents(status)`,
		`CREATE INDEX IF NOT EXISTS incidents_fired_at_idx ON incidents(fired_at DESC)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

// Alertmanager 알림을 incidents 테이블에 저장
// fingerprint를 incident_id로 사용
func (db *Postgres) SaveAlertAsIncident(ctx context.Context, alert model.Alert) error {
	alertName := alert.Labels["alertname"]
	severity := alert.Labels["severity"]
	if severity == "" {
		severity = "warning"
	}

	query := `
		INSERT INTO incidents (
			incident_id, alarm_title, severity, status, fired_at,
			fingerprint, labels, annotations, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (incident_id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = NOW()
	`

	_, err := db.Pool.Exec(ctx, query,
		alert.Fingerprint, // incident_id == fingerprint
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

// incident에 Slack thread_ts 저장
func (db *Postgres) UpdateThreadTS(ctx context.Context, fingerprint, threadTS string) error {
	query := `
		UPDATE incidents
		SET thread_ts = $2, updated_at = NOW()
		WHERE incident_id = $1
	`
	_, err := db.Pool.Exec(ctx, query, fingerprint, threadTS)
	return err
}

// fingerprint로 thread_ts 조회
func (db *Postgres) GetThreadTS(ctx context.Context, fingerprint string) (string, bool) {
	query := `
		SELECT thread_ts FROM incidents
		WHERE incident_id = $1 AND thread_ts != ''
	`

	var threadTS string
	err := db.Pool.QueryRow(ctx, query, fingerprint).Scan(&threadTS)
	if err != nil || threadTS == "" {
		return "", false
	}
	return threadTS, true
}

// resolved 상태로 업데이트
func (db *Postgres) UpdateIncidentResolved(ctx context.Context, fingerprint string, resolvedAt time.Time) error {
	query := `
		UPDATE incidents
		SET status = 'resolved', resolved_at = $2, updated_at = NOW()
		WHERE incident_id = $1
	`
	_, err := db.Pool.Exec(ctx, query, fingerprint, resolvedAt)
	return err
}

// Agent 분석 결과 저장
func (db *Postgres) UpdateAnalysis(ctx context.Context, fingerprint, summary, detail string) error {
	query := `
		UPDATE incidents
		SET analysis_summary = $2, analysis_detail = $3, updated_at = NOW()
		WHERE incident_id = $1
	`
	_, err := db.Pool.Exec(ctx, query, fingerprint, summary, detail)
	return err
}

// fingerprint로 incident 조회
func (db *Postgres) GetIncidentByFingerprint(ctx context.Context, fingerprint string) (*model.IncidentDetailResponse, error) {
	return db.GetIncidentDetail(fingerprint)
}
