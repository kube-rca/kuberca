package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

type AnalyticsService struct {
	repo *db.Postgres
}

func NewAnalyticsService(repo *db.Postgres) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) BuildDashboard(windowRaw string) (*model.AnalyticsDashboardResponse, error) {
	window, normalizedWindow, err := parseAnalyticsWindow(windowRaw)
	if err != nil {
		return nil, err
	}

	incidents, err := s.repo.GetIncidentList()
	if err != nil {
		return nil, err
	}
	alerts, err := s.repo.GetAlertList()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	cutoff := now.Add(-window)
	dayCount := int(window.Hours()/24) + 1
	if dayCount < 1 {
		dayCount = 1
	}

	incidentSeverity := make(map[string]int)
	alertSeverity := make(map[string]int)
	namespaceCount := make(map[string]int)
	trend := make(map[string]*model.AnalyticsDailyPoint)

	for i := dayCount - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		trend[date] = &model.AnalyticsDailyPoint{Date: date}
	}

	summary := model.AnalyticsSummary{}
	var totalMTTRMinutes float64
	var resolvedForMTTR int

	for _, incident := range incidents {
		if incident.FiredAt.Before(cutoff) {
			continue
		}

		summary.TotalIncidents++
		incidentSeverity[normalizeKey(incident.Severity)]++

		status := strings.ToLower(incident.Status)
		if status == "resolved" || incident.ResolvedAt != nil {
			summary.ResolvedIncidents++
		} else {
			summary.FiringIncidents++
		}

		if incident.ResolvedAt != nil && incident.ResolvedAt.After(incident.FiredAt) {
			totalMTTRMinutes += incident.ResolvedAt.Sub(incident.FiredAt).Minutes()
			resolvedForMTTR++
		}

		bucket := incident.FiredAt.UTC().Format("2006-01-02")
		if point, ok := trend[bucket]; ok {
			point.Incidents++
		}
	}

	for _, alert := range alerts {
		if alert.FiredAt.Before(cutoff) {
			continue
		}

		summary.TotalAlerts++
		alertSeverity[normalizeKey(alert.Severity)]++

		status := strings.ToLower(alert.Status)
		if status == "resolved" {
			summary.ResolvedAlerts++
		} else {
			summary.FiringAlerts++
		}

		namespace := "unknown"
		if alert.Namespace != nil && strings.TrimSpace(*alert.Namespace) != "" {
			namespace = strings.TrimSpace(*alert.Namespace)
		}
		namespaceCount[namespace]++

		bucket := alert.FiredAt.UTC().Format("2006-01-02")
		if point, ok := trend[bucket]; ok {
			point.Alerts++
		}
	}

	if resolvedForMTTR > 0 {
		summary.AvgMTTRMinutes = totalMTTRMinutes / float64(resolvedForMTTR)
	}
	if summary.TotalIncidents > 0 {
		summary.AvgAlertsPerIncident = float64(summary.TotalAlerts) / float64(summary.TotalIncidents)
	}

	daily := make([]model.AnalyticsDailyPoint, 0, len(trend))
	keys := make([]string, 0, len(trend))
	for date := range trend {
		keys = append(keys, date)
	}
	sort.Strings(keys)
	for _, date := range keys {
		daily = append(daily, *trend[date])
	}

	return &model.AnalyticsDashboardResponse{
		Window:      normalizedWindow,
		GeneratedAt: now,
		Summary:     summary,
		Breakdown: model.AnalyticsBreakdown{
			IncidentSeverity: mapToSortedItems(incidentSeverity, 10),
			AlertSeverity:    mapToSortedItems(alertSeverity, 10),
			TopNamespaces:    mapToSortedItems(namespaceCount, 7),
		},
		Series: model.AnalyticsSeries{Daily: daily},
	}, nil
}

func normalizeKey(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "unknown"
	}
	return strings.ToLower(trimmed)
}

func parseAnalyticsWindow(raw string) (time.Duration, string, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		v = "30d"
	}

	if strings.HasSuffix(v, "h") {
		duration, err := time.ParseDuration(v)
		if err != nil || duration <= 0 {
			return 0, "", fmt.Errorf("invalid window: %s", raw)
		}
		return duration, v, nil
	}

	if strings.HasSuffix(v, "d") {
		number := strings.TrimSuffix(v, "d")
		days, err := strconv.Atoi(number)
		if err != nil || days <= 0 || days > 365 {
			return 0, "", fmt.Errorf("invalid window: %s", raw)
		}
		return time.Duration(days) * 24 * time.Hour, fmt.Sprintf("%dd", days), nil
	}

	return 0, "", fmt.Errorf("invalid window: %s", raw)
}

func mapToSortedItems(values map[string]int, limit int) []model.AnalyticsCountItem {
	items := make([]model.AnalyticsCountItem, 0, len(values))
	for key, count := range values {
		items = append(items, model.AnalyticsCountItem{Key: key, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Key < items[j].Key
		}
		return items[i].Count > items[j].Count
	})

	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}
