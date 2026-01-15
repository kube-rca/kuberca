// Agent 분석 요청 및 응답 처리 비즈니스 로직 정의
//
// 처리 흐름:
//  1. RequestAnalysis: Agent에 분석 요청 (goroutine에서 호출)
//  2. Agent 응답을 DB에 저장 (alerts.analysis_summary, analysis_detail)
//  3. Agent 응답을 Slack 쓰레드에 전송

package service

import (
	"log"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

// AgentService 구조체 정의
type AgentService struct {
	agentClient *client.AgentClient
	slackClient *client.SlackClient
	db          *db.Postgres
}

// AgentService 객체 생성
func NewAgentService(agentClient *client.AgentClient, slackClient *client.SlackClient, database *db.Postgres) *AgentService {
	return &AgentService{
		agentClient: agentClient,
		slackClient: slackClient,
		db:          database,
	}
}

func (s *AgentService) RequestAnalysis(alert model.Alert, threadTS string) {
	if threadTS == "" {
		log.Printf("No thread_ts for alert (alert_id=%s), skipping agent request", alert.Fingerprint)
		return
	}

	log.Printf("Requesting agent analysis (alert_id=%s, status=%s, thread_ts=%s)", alert.Fingerprint, alert.Status, threadTS)

	// Agent에 분석 요청 (동기)
	resp, err := s.agentClient.RequestAnalysis(alert, threadTS)
	if err != nil {
		log.Printf("Failed to request agent analysis: %v", err)
		return
	}

	// 분석 결과를 DB에 저장 (alerts.analysis_summary, analysis_detail)
	if err := s.db.UpdateAlertAnalysis(alert.Fingerprint, resp.Analysis, resp.Analysis); err != nil {
		log.Printf("Failed to save analysis to DB: %v", err)
		// DB 저장 실패해도 Slack 전송은 계속 진행
	} else {
		log.Printf("Saved analysis to DB (alert_id=%s)", alert.Fingerprint)
	}

	// 분석 결과를 Slack 쓰레드에 전송
	log.Printf("Sending analysis to Slack thread (thread_ts=%s)", threadTS)
	if err := s.slackClient.SendToThread(threadTS, resp.Analysis); err != nil {
		log.Printf("Failed to send analysis to Slack: %v", err)
	}
}

// RequestIncidentSummary - Incident 종료 시 전체 Alert 분석을 기반으로 최종 요약 요청
func (s *AgentService) RequestIncidentSummary(incident *model.IncidentDetailResponse, alerts []model.AlertDetailResponse) (*client.IncidentSummaryResponse, error) {
	// Alert 분석 내용을 Agent 요청 포맷으로 변환
	alertInputs := make([]client.AlertSummaryInput, 0, len(alerts))
	for _, alert := range alerts {
		// *string -> string 변환 (nil이면 빈 문자열)
		summary := ""
		if alert.AnalysisSummary != nil {
			summary = *alert.AnalysisSummary
		}
		detail := ""
		if alert.AnalysisDetail != nil {
			detail = *alert.AnalysisDetail
		}

		alertInputs = append(alertInputs, client.AlertSummaryInput{
			Fingerprint:     alert.Fingerprint,
			AlertName:       alert.AlarmTitle,
			Severity:        alert.Severity,
			Status:          alert.Status,
			AnalysisSummary: summary,
			AnalysisDetail:  detail,
		})
	}

	// resolved_at 포맷팅
	resolvedAt := ""
	if incident.ResolvedAt != nil {
		resolvedAt = incident.ResolvedAt.Format("2006-01-02T15:04:05Z")
	}

	req := client.IncidentSummaryRequest{
		IncidentID: incident.IncidentID,
		Title:      incident.Title,
		Severity:   incident.Severity,
		FiredAt:    incident.FiredAt.Format("2006-01-02T15:04:05Z"),
		ResolvedAt: resolvedAt,
		Alerts:     alertInputs,
	}

	log.Printf("Requesting incident summary (incident_id=%s, alert_count=%d)", incident.IncidentID, len(alerts))

	resp, err := s.agentClient.RequestIncidentSummary(req)
	if err != nil {
		log.Printf("Failed to request incident summary: %v", err)
		return nil, err
	}

	log.Printf("Received incident summary (incident_id=%s)", incident.IncidentID)
	return resp, nil
}
