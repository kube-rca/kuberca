package model

type AnalysisResult struct {
	AlertCount    int            `json:"alertCount"`
	SeverityCount map[string]int `json:"severityCount"`
	StatusCount   map[string]int `json:"statusCount"`
}
