package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kube-rca/backend/internal/model"
)

// Postgres 구조체의 필드 타입을 NewPostgresPool의 리턴 타입과 맞춥니다.
type Postgres struct {
	Pool *pgxpool.Pool
}

// EnsureIncidentSchema - incidents 테이블 생성 (장애 단위)
func (db *Postgres) EnsureIncidentSchema(ctx context.Context) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS incidents (
			incident_id TEXT PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			severity TEXT NOT NULL DEFAULT 'warning',
			status TEXT NOT NULL DEFAULT 'firing',
			fired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			analysis_summary TEXT NOT NULL DEFAULT '',
			analysis_detail TEXT NOT NULL DEFAULT '',
			similar_incidents JSONB NOT NULL DEFAULT '[]',
			created_by TEXT NOT NULL DEFAULT 'system',
			resolved_by TEXT,
			is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
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

// GetIncidentList - Incident 목록 조회 (Alert 개수 포함)
func (db *Postgres) GetIncidentList() ([]model.IncidentListResponse, error) {
	query := `
		SELECT
			i.incident_id,
			i.title,
			i.severity,
			i.status,
			i.fired_at,
			i.resolved_at,
			COUNT(a.alert_id) as alert_count
		FROM incidents i
		LEFT JOIN alerts a ON i.incident_id = a.incident_id AND a.is_enabled = TRUE
		WHERE i.is_enabled = TRUE
		GROUP BY i.incident_id, i.title, i.severity, i.status, i.fired_at, i.resolved_at
		ORDER BY i.fired_at DESC`

	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IncidentListResponse
	for rows.Next() {
		var i model.IncidentListResponse
		if err := rows.Scan(&i.IncidentID, &i.Title, &i.Severity, &i.Status, &i.FiredAt, &i.ResolvedAt, &i.AlertCount); err != nil {
			return nil, err
		}
		list = append(list, i)
	}

	if list == nil {
		list = []model.IncidentListResponse{}
	}
	return list, nil
}

// GetIncidentDetail - Incident 상세 조회
func (db *Postgres) GetIncidentDetail(id string) (*model.IncidentDetailResponse, error) {
	query := `
		SELECT
			incident_id, title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail, similar_incidents,
			created_by, resolved_by
		FROM incidents
		WHERE incident_id = $1
	`

	var i model.IncidentDetailResponse
	err := db.Pool.QueryRow(context.Background(), query, id).Scan(
		&i.IncidentID,
		&i.Title,
		&i.Severity,
		&i.Status,
		&i.FiredAt,
		&i.ResolvedAt,
		&i.AnalysisSummary,
		&i.AnalysisDetail,
		&i.SimilarIncidents,
		&i.CreatedBy,
		&i.ResolvedBy,
	)

	if err != nil {
		return nil, err
	}

	return &i, nil
}

// UpdateIncident - Incident 수정
func (db *Postgres) UpdateIncident(id string, req model.UpdateIncidentRequest) error {
	query := `
		UPDATE incidents
		SET
			title = $1,
			severity = $2,
			analysis_summary = $3,
			analysis_detail = $4,
			updated_at = NOW()
		WHERE incident_id = $5
	`

	commandTag, err := db.Pool.Exec(context.Background(), query,
		req.Title,
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

// HideIncident - Incident 숨기기
func (db *Postgres) HideIncident(id string) error {
	query := `
		UPDATE incidents
		SET is_enabled = false, updated_at = NOW()
		WHERE incident_id = $1
	`
	commandTag, err := db.Pool.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no incident found with id: %s", id)
	}

	return nil
}

// ResolveIncident - Incident 종료 (사용자가 수동으로 종료)
func (db *Postgres) ResolveIncident(ctx context.Context, id string, resolvedBy string) error {
	query := `
		UPDATE incidents
		SET status = 'resolved', resolved_at = NOW(), resolved_by = $2, updated_at = NOW()
		WHERE incident_id = $1 AND status = 'firing'
	`
	commandTag, err := db.Pool.Exec(ctx, query, id, resolvedBy)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no firing incident found with id: %s", id)
	}

	return nil
}

// GetFiringIncident - 현재 firing 상태인 Incident 조회
func (db *Postgres) GetFiringIncident(ctx context.Context) (*model.IncidentDetailResponse, error) {
	query := `
		SELECT
			incident_id, title, severity, status,
			fired_at, resolved_at, analysis_summary, analysis_detail, similar_incidents,
			created_by, resolved_by
		FROM incidents
		WHERE status = 'firing' AND is_enabled = TRUE
		ORDER BY fired_at DESC
		LIMIT 1
	`

	var i model.IncidentDetailResponse
	err := db.Pool.QueryRow(ctx, query).Scan(
		&i.IncidentID,
		&i.Title,
		&i.Severity,
		&i.Status,
		&i.FiredAt,
		&i.ResolvedAt,
		&i.AnalysisSummary,
		&i.AnalysisDetail,
		&i.SimilarIncidents,
		&i.CreatedBy,
		&i.ResolvedBy,
	)

	if err != nil {
		return nil, err
	}

	return &i, nil
}

// CreateIncident - 새 Incident 생성
func (db *Postgres) CreateIncident(ctx context.Context, title, severity string, firedAt time.Time) (string, error) {
	incidentID := "INC-" + uuid.New().String()[:8]

	query := `
		INSERT INTO incidents (incident_id, title, severity, status, fired_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'firing', $4, NOW(), NOW())
	`

	_, err := db.Pool.Exec(ctx, query, incidentID, title, severity, firedAt)
	if err != nil {
		return "", err
	}

	return incidentID, nil
}

// UpdateIncidentSeverity - Incident severity 업데이트 (가장 높은 severity로)
func (db *Postgres) UpdateIncidentSeverity(ctx context.Context, incidentID, severity string) error {
	// critical > warning > info 순으로 높은 severity만 업데이트
	query := `
		UPDATE incidents
		SET severity = $2, updated_at = NOW()
		WHERE incident_id = $1
		AND (
			(severity = 'info') OR
			(severity = 'warning' AND $2 = 'critical')
		)
	`
	_, err := db.Pool.Exec(ctx, query, incidentID, severity)
	return err
}

// UpdateIncidentAnalysis - Incident 분석 결과 저장
func (db *Postgres) UpdateIncidentAnalysis(ctx context.Context, incidentID, summary, detail string) error {
	query := `
		UPDATE incidents
		SET analysis_summary = $2, analysis_detail = $3, updated_at = NOW()
		WHERE incident_id = $1
	`
	_, err := db.Pool.Exec(ctx, query, incidentID, summary, detail)
	return err
}

// CreateMockIncident - Mock 데이터 생성 (테스트용)
func (db *Postgres) CreateMockIncident() (string, error) {
	timestamp := time.Now().Unix()
	id := fmt.Sprintf("INC-%d", timestamp)
	title := fmt.Sprintf("테스트용 장애 (생성시간: %d)", timestamp)

	query := `
		INSERT INTO incidents (incident_id, title, severity, status, fired_at, created_at, updated_at)
		VALUES ($1, $2, 'warning', 'firing', NOW(), NOW(), NOW())
	`

	_, err := db.Pool.Exec(context.Background(), query, id, title)
	if err != nil {
		return "", err
	}

	return id, nil
}
