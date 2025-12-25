package model

import "time"

// 프론트엔드 리스트 출력용 구조체
type IncidentListResponse struct {
	IncidentID string     `json:"incident_id"`
	AlarmTitle string     `json:"alarm_title"`
	Severity   string     `json:"severity"`
	ResolvedAt *time.Time `json:"resolved_at"`
}
