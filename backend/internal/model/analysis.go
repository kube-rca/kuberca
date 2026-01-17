package model

import (
	"encoding/json"
	"time"
)

// AlertAnalysis - 상태별 Alert 분석 결과
type AlertAnalysis struct {
	AnalysisID int64
	AlertID    string
	IncidentID *string
	Status     string
	Summary    string
	Detail     string
	Context    json.RawMessage
	CreatedAt  time.Time
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
