package model

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Incident 모델 (장애 단위)
// ============================================================================

// IncidentListResponse - Incident 목록 조회용 구조체
type IncidentListResponse struct {
	IncidentID string     `json:"incident_id"`
	Title      string     `json:"title"`
	Severity   string     `json:"severity"`
	Status     string     `json:"status"` // firing, resolved
	FiredAt    time.Time  `json:"fired_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	AlertCount int        `json:"alert_count"` // 연결된 Alert 개수
}

// IncidentDetailResponse - Incident 상세 조회용 구조체
type IncidentDetailResponse struct {
	IncidentID      string     `json:"incident_id"`
	Title           string     `json:"title"`
	Severity        string     `json:"severity"`
	Status          string     `json:"status"` // firing, resolved
	FiredAt         time.Time  `json:"fired_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	AnalysisSummary *string    `json:"analysis_summary"`
	AnalysisDetail  *string    `json:"analysis_detail"`
	CreatedBy       *string    `json:"created_by"`
	ResolvedBy      *string    `json:"resolved_by"`

	// DB의 JSONB 컬럼을 그대로 바이트로 받아서 전달
	SimilarIncidents json.RawMessage `json:"similar_incidents" swaggertype:"object"`

	// 연결된 Alert 목록 (상세 조회 시 포함)
	Alerts []AlertListResponse `json:"alerts,omitempty"`
}

// UpdateIncidentRequest - Incident 수정 요청 구조체
type UpdateIncidentRequest struct {
	Title           string `json:"title"`
	Severity        string `json:"severity"`
	AnalysisSummary string `json:"analysis_summary"`
	AnalysisDetail  string `json:"analysis_detail"`
}

// ResolveIncidentRequest - Incident 종료 요청 구조체
type ResolveIncidentRequest struct {
	ResolvedBy string `json:"resolved_by"`
}

// ============================================================================
// Alert 모델 (개별 알람 단위)
// ============================================================================

// AlertListResponse - Alert 목록 조회용 구조체
type AlertListResponse struct {
	AlertID    string     `json:"alert_id"`
	IncidentID *string    `json:"incident_id"` // null 가능 (아직 Incident에 연결되지 않은 경우)
	AlarmTitle string     `json:"alarm_title"`
	Namespace  *string    `json:"namespace"`
	Severity   string     `json:"severity"`
	Status     string     `json:"status"` // firing, resolved
	FiredAt    time.Time  `json:"fired_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

// AlertDetailResponse - Alert 상세 조회용 구조체
type AlertDetailResponse struct {
	AlertID         string          `json:"alert_id"`
	IncidentID      *string         `json:"incident_id"`
	AlarmTitle      string          `json:"alarm_title"`
	Severity        string          `json:"severity"`
	Status          string          `json:"status"`
	FiredAt         time.Time       `json:"fired_at"`
	ResolvedAt      *time.Time      `json:"resolved_at"`
	AnalysisSummary *string         `json:"analysis_summary"`
	AnalysisDetail  *string         `json:"analysis_detail"`
	Fingerprint     string          `json:"fingerprint"`
	ThreadTS        string          `json:"thread_ts"`
	Labels          json.RawMessage `json:"labels" swaggertype:"object"`
	Annotations     json.RawMessage `json:"annotations" swaggertype:"object"`
}

// ============================================================================
// API Response Envelope
// ============================================================================

// IncidentDetailEnvelope - Incident 상세 API 응답 구조체
type IncidentDetailEnvelope struct {
	Status string                  `json:"status"`
	Data   *IncidentDetailResponse `json:"data"`
}

// IncidentUpdateResponse - Incident 수정 API 응답 구조체
type IncidentUpdateResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	IncidentID string `json:"incident_id"`
}

// MockIncidentResponse - Mock Incident 생성 API 응답 구조체
type MockIncidentResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	IncidentID string `json:"incident_id"`
}

// UpdateAlertIncidentRequest - Alert의 Incident ID 변경 요청 구조체
type UpdateAlertIncidentRequest struct {
	IncidentID string `json:"incident_id" binding:"required"`
}

// AlertUpdateResponse - Alert 수정 API 응답 구조체
type AlertUpdateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	AlertID string `json:"alert_id"`
}
