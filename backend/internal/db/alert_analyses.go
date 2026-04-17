package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/model"
)

// EnsureAlertAnalysisSchema - alert_analyses / alert_analysis_artifacts 테이블 생성
func (db *Postgres) EnsureAlertAnalysisSchema() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS alert_analyses (
			analysis_id BIGSERIAL PRIMARY KEY,
			alert_id TEXT NOT NULL,
			incident_id TEXT,
			status TEXT NOT NULL DEFAULT 'firing',
			summary TEXT NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '',
			summary_i18n JSONB NOT NULL DEFAULT '{}',
			detail_i18n JSONB NOT NULL DEFAULT '{}',
			context JSONB NOT NULL DEFAULT '{}',
			analysis_model TEXT NOT NULL DEFAULT '',
			analysis_version TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS alert_analyses_alert_id_idx ON alert_analyses(alert_id)`,
		`CREATE INDEX IF NOT EXISTS alert_analyses_incident_id_idx ON alert_analyses(incident_id)`,
		`CREATE INDEX IF NOT EXISTS alert_analyses_status_idx ON alert_analyses(status)`,
		`CREATE INDEX IF NOT EXISTS alert_analyses_created_at_idx ON alert_analyses(created_at DESC)`,
		`ALTER TABLE alert_analyses ADD COLUMN IF NOT EXISTS summary_i18n JSONB NOT NULL DEFAULT '{}'`,
		`ALTER TABLE alert_analyses ADD COLUMN IF NOT EXISTS detail_i18n JSONB NOT NULL DEFAULT '{}'`,
		`
		CREATE TABLE IF NOT EXISTS alert_analysis_artifacts (
			artifact_id BIGSERIAL PRIMARY KEY,
			analysis_id BIGINT NOT NULL REFERENCES alert_analyses(analysis_id) ON DELETE CASCADE,
			alert_id TEXT NOT NULL,
			incident_id TEXT,
			artifact_type TEXT NOT NULL,
			query TEXT NOT NULL DEFAULT '',
			result JSONB NOT NULL DEFAULT '{}',
			summary TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
		`,
		`CREATE INDEX IF NOT EXISTS alert_analysis_artifacts_analysis_id_idx ON alert_analysis_artifacts(analysis_id)`,
		`CREATE INDEX IF NOT EXISTS alert_analysis_artifacts_alert_id_idx ON alert_analysis_artifacts(alert_id)`,
		`CREATE INDEX IF NOT EXISTS alert_analysis_artifacts_incident_id_idx ON alert_analysis_artifacts(incident_id)`,
		`CREATE INDEX IF NOT EXISTS alert_analysis_artifacts_type_idx ON alert_analysis_artifacts(artifact_type)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

// InsertAlertAnalysis - Alert 분석 결과 저장
func (db *Postgres) InsertAlertAnalysis(
	alertID string,
	incidentID string,
	status string,
	summary string,
	detail string,
	summaryI18n model.LocalizedText,
	detailI18n model.LocalizedText,
	contextJSON json.RawMessage,
) (int64, error) {
	var incidentIDPtr *string
	if incidentID != "" {
		incidentIDPtr = &incidentID
	}

	if len(contextJSON) == 0 {
		contextJSON = json.RawMessage("{}")
	}

	query := `
		INSERT INTO alert_analyses (
			alert_id, incident_id, status, summary, detail, summary_i18n, detail_i18n, context, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8::jsonb, NOW())
		RETURNING analysis_id
	`

	var analysisID int64
	err := db.Pool.QueryRow(context.Background(), query,
		alertID,
		incidentIDPtr,
		status,
		summary,
		detail,
		localizedJSON(summaryI18n, summary),
		localizedJSON(detailI18n, detail),
		[]byte(contextJSON),
	).Scan(&analysisID)
	if err != nil {
		return 0, err
	}

	return analysisID, nil
}

// InsertAlertAnalysisArtifacts - 분석 근거 데이터 저장
func (db *Postgres) InsertAlertAnalysisArtifacts(
	analysisID int64,
	alertID string,
	incidentID string,
	artifacts []client.AlertAnalysisArtifactInput,
) error {
	if len(artifacts) == 0 {
		return nil
	}

	var incidentIDPtr *string
	if incidentID != "" {
		incidentIDPtr = &incidentID
	}

	query := `
		INSERT INTO alert_analysis_artifacts (
			analysis_id, alert_id, incident_id, artifact_type,
			query, result, summary, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, NOW())
	`

	for _, artifact := range artifacts {
		resultJSON := artifact.Result
		if len(resultJSON) == 0 {
			resultJSON = json.RawMessage("{}")
		}
		if _, err := db.Pool.Exec(context.Background(), query,
			analysisID,
			alertID,
			incidentIDPtr,
			artifact.Type,
			artifact.Query,
			[]byte(resultJSON),
			artifact.Summary,
		); err != nil {
			return err
		}
	}
	return nil
}

// GetLatestAnalysesByAlertID - alert_id 기준 status별 최신 분석 1건씩 조회 (firing, resolved)
func (db *Postgres) GetLatestAnalysesByAlertID(alertID string) ([]model.AlertAnalysis, error) {
	// PostgreSQL DISTINCT ON: status별 최신 1건만 반환
	query := `
		SELECT analysis_id, alert_id, incident_id, status, summary, detail, summary_i18n, detail_i18n,
		       analysis_model, created_at
		FROM (
			SELECT DISTINCT ON (status)
				analysis_id, alert_id, incident_id, status, summary, detail, summary_i18n, detail_i18n,
				analysis_model, created_at
			FROM alert_analyses
			WHERE alert_id = $1
			ORDER BY status, created_at DESC
		) sub
		ORDER BY created_at ASC
	`

	rows, err := db.Pool.Query(context.Background(), query, alertID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AlertAnalysis
	for rows.Next() {
		var a model.AlertAnalysis
		var summaryI18n, detailI18n []byte
		if err := rows.Scan(
			&a.AnalysisID,
			&a.AlertID,
			&a.IncidentID,
			&a.Status,
			&a.Summary,
			&a.Detail,
			&summaryI18n,
			&detailI18n,
			&a.AnalysisModel,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		a.SummaryI18n = toLocalizedText(summaryI18n, a.Summary)
		a.DetailI18n = toLocalizedText(detailI18n, a.Detail)
		list = append(list, a)
	}

	if list == nil {
		list = []model.AlertAnalysis{}
	}
	return list, nil
}

// GetLatestAlertAnalysisByAlertID - alert_id 기준 최신 분석 1건 조회
func (db *Postgres) GetLatestAlertAnalysisByAlertID(alertID string) (*model.AlertAnalysis, error) {
	query := `
		SELECT analysis_id, alert_id, incident_id, status, summary, detail, summary_i18n, detail_i18n, context, created_at
		FROM alert_analyses
		WHERE alert_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var analysis model.AlertAnalysis
	var summaryI18n, detailI18n []byte
	err := db.Pool.QueryRow(context.Background(), query, alertID).Scan(
		&analysis.AnalysisID,
		&analysis.AlertID,
		&analysis.IncidentID,
		&analysis.Status,
		&analysis.Summary,
		&analysis.Detail,
		&summaryI18n,
		&detailI18n,
		&analysis.Context,
		&analysis.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	analysis.SummaryI18n = toLocalizedText(summaryI18n, analysis.Summary)
	analysis.DetailI18n = toLocalizedText(detailI18n, analysis.Detail)
	return &analysis, nil
}

// GetLatestFiringAnalysisByFingerprint - fingerprint 기준 최신 firing 분석 조회
func (db *Postgres) GetLatestFiringAnalysisByFingerprint(fingerprint string) (*model.AlertAnalysis, error) {
	query := `
		SELECT aa.analysis_id, aa.alert_id, aa.incident_id, aa.status,
		       aa.summary, aa.detail, aa.summary_i18n, aa.detail_i18n, aa.context, aa.created_at
		FROM alert_analyses aa
		JOIN alerts a ON aa.alert_id = a.alert_id
		WHERE a.fingerprint = $1 AND aa.status = 'firing'
		ORDER BY aa.created_at DESC
		LIMIT 1
	`

	var analysis model.AlertAnalysis
	var summaryI18n, detailI18n []byte
	err := db.Pool.QueryRow(context.Background(), query, fingerprint).Scan(
		&analysis.AnalysisID,
		&analysis.AlertID,
		&analysis.IncidentID,
		&analysis.Status,
		&analysis.Summary,
		&analysis.Detail,
		&summaryI18n,
		&detailI18n,
		&analysis.Context,
		&analysis.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	analysis.SummaryI18n = toLocalizedText(summaryI18n, analysis.Summary)
	analysis.DetailI18n = toLocalizedText(detailI18n, analysis.Detail)
	return &analysis, nil
}

// GetAlertAnalysisArtifactsByAnalysisID - analysis_id 기준 근거 데이터 조회
func (db *Postgres) GetAlertAnalysisArtifactsByAnalysisID(analysisID int64) ([]model.AlertAnalysisArtifact, error) {
	query := `
		SELECT artifact_id, analysis_id, alert_id, incident_id, artifact_type, query, result, summary, created_at
		FROM alert_analysis_artifacts
		WHERE analysis_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.Pool.Query(context.Background(), query, analysisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AlertAnalysisArtifact
	for rows.Next() {
		var artifact model.AlertAnalysisArtifact
		if err := rows.Scan(
			&artifact.ArtifactID,
			&artifact.AnalysisID,
			&artifact.AlertID,
			&artifact.IncidentID,
			&artifact.ArtifactType,
			&artifact.Query,
			&artifact.Result,
			&artifact.Summary,
			&artifact.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, artifact)
	}

	if list == nil {
		list = []model.AlertAnalysisArtifact{}
	}
	return list, nil
}
