package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kube-rca/backend/internal/model"
)

// Postgres 구조체의 필드 타입을 NewPostgresPool의 리턴 타입과 맞춥니다.
type Postgres struct {
	Pool *pgxpool.Pool
}

func (db *Postgres) GetIncidentList() ([]model.IncidentListResponse, error) {
	query := `
		SELECT incident_id, alarm_title, severity, resolved_at 
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
