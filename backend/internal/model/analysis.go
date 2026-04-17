package model

import (
	"encoding/json"
	"time"
)

type LocalizedText map[string]string

// AlertAnalysis - 상태별 Alert 분석 결과
type AlertAnalysis struct {
	AnalysisID    int64
	AlertID       string
	IncidentID    *string
	Status        string
	Summary       string
	Detail        string
	SummaryI18n   LocalizedText
	DetailI18n    LocalizedText
	Context       json.RawMessage
	AnalysisModel string
	CreatedAt     time.Time
}

// AlertAnalysisItem - Alert Detail API 응답용 분석 DTO
type AlertAnalysisItem struct {
	AnalysisID    int64         `json:"analysis_id"`
	Status        string        `json:"status"`
	Summary       string        `json:"summary"`
	Detail        string        `json:"detail"`
	SummaryI18n   LocalizedText `json:"summary_i18n,omitempty"`
	DetailI18n    LocalizedText `json:"detail_i18n,omitempty"`
	AnalysisModel string        `json:"analysis_model"`
	CreatedAt     time.Time     `json:"created_at"`
}

// AlertAnalysisArtifact - 분석 근거 데이터(메트릭/이벤트/로그/PromQL)
type AlertAnalysisArtifact struct {
	ArtifactID   int64
	AnalysisID   int64
	AlertID      string
	IncidentID   *string
	ArtifactType string
	Query        string
	Result       json.RawMessage
	Summary      string
	CreatedAt    time.Time
}
