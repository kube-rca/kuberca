package model

import "time"

type AnalyticsCountItem struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type AnalyticsSummary struct {
	TotalIncidents       int     `json:"total_incidents"`
	FiringIncidents      int     `json:"firing_incidents"`
	ResolvedIncidents    int     `json:"resolved_incidents"`
	TotalAlerts          int     `json:"total_alerts"`
	FiringAlerts         int     `json:"firing_alerts"`
	ResolvedAlerts       int     `json:"resolved_alerts"`
	AvgMTTRMinutes       float64 `json:"avg_mttr_minutes"`
	AvgAlertsPerIncident float64 `json:"avg_alerts_per_incident"`
}

type AnalyticsBreakdown struct {
	IncidentSeverity []AnalyticsCountItem `json:"incident_severity"`
	AlertSeverity    []AnalyticsCountItem `json:"alert_severity"`
	TopNamespaces    []AnalyticsCountItem `json:"top_namespaces"`
}

type AnalyticsDailyPoint struct {
	Date      string `json:"date"`
	Incidents int    `json:"incidents"`
	Alerts    int    `json:"alerts"`
}

type AnalyticsSeries struct {
	Daily []AnalyticsDailyPoint `json:"daily"`
}

type AnalyticsDashboardResponse struct {
	Window      string             `json:"window"`
	GeneratedAt time.Time          `json:"generated_at"`
	Summary     AnalyticsSummary   `json:"summary"`
	Breakdown   AnalyticsBreakdown `json:"breakdown"`
	Series      AnalyticsSeries    `json:"series"`
}
