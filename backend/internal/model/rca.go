package model

import (
	"encoding/json"
	"time"
)

// 프론트엔드 리스트 출력용 구조체
type IncidentListResponse struct {
	IncidentID string     `json:"incident_id"`
	AlarmTitle string     `json:"alarm_title"`
	Severity   string     `json:"severity"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

type IncidentDetailResponse struct {
	IncidentID      string     `json:"incident_id"`
	AlarmTitle      string     `json:"alarm_title"`
	Severity        string     `json:"severity"`
	Status          string     `json:"status"`
	FiredAt         time.Time  `json:"fired_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`      // null 가능
	AnalysisSummary *string    `json:"analysis_summary"` // null 가능
	AnalysisDetail  *string    `json:"analysis_detail"`  // null 가능

	// DB의 JSONB 컬럼을 그대로 바이트로 받아서 전달 (구조를 몰라도 됨)
	SimilarIncidents json.RawMessage `json:"similar_incidents" swaggertype:"object"`
}

type UpdateIncidentRequest struct {
	AlarmTitle      string `json:"alarm_title"`
	Severity        string `json:"severity"`
	AnalysisSummary string `json:"analysis_summary"`
	AnalysisDetail  string `json:"analysis_detail"`
}
