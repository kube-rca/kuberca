// Agent 분석 요청 및 응답 처리 비즈니스 로직 정의
//
// 처리 흐름:
//  1. RequestAnalysis: Agent에 분석 요청 (goroutine에서 호출)
//  2. Agent 응답을 DB에 저장 (alerts.analysis_summary, analysis_detail)
//  3. Agent 응답을 Slack 쓰레드에 전송

package service

import (
	"log"
	"sync"
	"time"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/sse"
)

// inFlightStaleTTL is the maximum duration an in-flight entry is considered
// valid. Entries older than this are treated as stale (e.g. leaked by a panic)
// and automatically evicted so that future requests are not permanently blocked.
const inFlightStaleTTL = 5 * time.Minute

// AgentService 구조체 정의
type AgentService struct {
	agentClient *client.AgentClient
	notifier    client.Notifier
	db          *db.Postgres
	sseHub      *sse.Hub
	mu          sync.Mutex
	inFlight    map[string]time.Time
}

// AgentService 객체 생성
func NewAgentService(agentClient *client.AgentClient, notifier client.Notifier, database *db.Postgres, sseHub *sse.Hub) *AgentService {
	return &AgentService{
		agentClient: agentClient,
		notifier:    notifier,
		db:          database,
		sseHub:      sseHub,
		inFlight:    make(map[string]time.Time),
	}
}

func (s *AgentService) RequestAnalysis(alert model.Alert, alertID, threadTS, incidentID string, skipThreadCheck bool) {
	if !skipThreadCheck && threadTS == "" && s.requiresThreadRef() {
		log.Printf("No thread_ref for alert (alert_id=%s, fingerprint=%s), skipping agent request", alertID, alert.Fingerprint)
		return
	}
	if threadTS == "" {
		log.Printf("No thread_ref for alert (alert_id=%s, fingerprint=%s), sending analysis without thread", alertID, alert.Fingerprint)
	}

	key := s.analysisKey(alertID, alert.Fingerprint)
	if key != "" {
		if startedAt, ok := s.beginAnalysis(key); !ok {
			log.Printf(
				"Skipping duplicate alert analysis request (alert_id=%s, fingerprint=%s, incident_id=%s, in_flight_ms=%d)",
				alertID,
				alert.Fingerprint,
				incidentID,
				time.Since(startedAt).Milliseconds(),
			)
			return
		}
		defer s.finishAnalysis(key)
	}

	requestStartedAt := time.Now()
	log.Printf("Requesting agent analysis (alert_id=%s, fingerprint=%s, status=%s, thread_ref=%s)", alertID, alert.Fingerprint, alert.Status, threadTS)

	// Agent에 분석 요청 (동기)
	resp, err := s.agentClient.RequestAnalysis(alert, threadTS, incidentID)
	if err != nil {
		log.Printf("Failed to request agent analysis: %v", err)
		return
	}

	// 분석 결과를 DB에 저장 (alerts.analysis_summary, analysis_detail)
	summary := resp.AnalysisSummary
	detail := resp.AnalysisDetail
	if summary == "" {
		summary = resp.Analysis
	}
	if detail == "" {
		detail = resp.Analysis
	}

	if err := s.db.UpdateAlertAnalysis(alertID, summary, detail); err != nil {
		log.Printf("Failed to save analysis to DB: %v", err)
		// DB 저장 실패해도 Slack 전송은 계속 진행
	} else {
		log.Printf("Saved analysis to DB (alert_id=%s)", alertID)
	}

	analysisID, err := s.db.InsertAlertAnalysis(
		alertID,
		incidentID,
		alert.Status,
		summary,
		detail,
		resp.Context,
	)
	if err != nil {
		log.Printf("Failed to insert alert analysis: %v", err)
	} else if len(resp.Artifacts) > 0 {
		if err := s.db.InsertAlertAnalysisArtifacts(analysisID, alertID, incidentID, resp.Artifacts); err != nil {
			log.Printf("Failed to insert alert analysis artifacts: %v", err)
		}
	}

	// SSE broadcast: 분석 완료 이벤트
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventAnalysisCompleted,
			Data: sse.EventData{AlertID: alertID, IncidentID: incidentID},
		})
	}

	// 분석 결과를 Slack 쓰레드에 전송 (threadTS가 있을 때만)
	if threadTS != "" {
		log.Printf("Sending analysis notification (thread_ref=%s)", threadTS)
		if err := s.notifier.Notify(client.AnalysisResultPostedEvent{
			ThreadRef: threadTS,
			Content:   resp.Analysis,
		}); err != nil {
			log.Printf("Failed to send analysis notification: %v", err)
		}
	} else {
		log.Printf("Skipping analysis notification (no thread_ref, alert_id=%s)", alertID)
	}

	log.Printf(
		"Agent analysis finished (alert_id=%s, fingerprint=%s, incident_id=%s, elapsed_ms=%d)",
		alertID,
		alert.Fingerprint,
		incidentID,
		time.Since(requestStartedAt).Milliseconds(),
	)
}

func (s *AgentService) analysisKey(alertID, fingerprint string) string {
	if alertID != "" {
		return "alert:" + alertID
	}
	if fingerprint != "" {
		return "fingerprint:" + fingerprint
	}
	log.Printf("WARNING: Cannot determine unique analysis key (alert_id=%q, fingerprint=%q), deduplication skipped", alertID, fingerprint)
	return ""
}

func (s *AgentService) beginAnalysis(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if startedAt, exists := s.inFlight[key]; exists {
		if time.Since(startedAt) < inFlightStaleTTL {
			return startedAt, false
		}
		log.Printf("Evicting stale in-flight analysis entry (key=%s, age_ms=%d)", key, time.Since(startedAt).Milliseconds())
	}

	startedAt := time.Now()
	s.inFlight[key] = startedAt
	return startedAt, true
}

func (s *AgentService) finishAnalysis(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inFlight, key)
}

func (s *AgentService) requiresThreadRef() bool {
	if req, ok := s.notifier.(client.ThreadRefRequirement); ok {
		return req.RequiresThreadRef()
	}
	_, ok := s.notifier.(client.ThreadRefStore)
	return ok
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

		var artifacts []client.AlertAnalysisArtifactInput
		if latest, err := s.db.GetLatestAlertAnalysisByAlertID(alert.AlertID); err == nil && latest != nil {
			if latest.Summary != "" {
				summary = latest.Summary
			}
			if latest.Detail != "" {
				detail = latest.Detail
			}

			artifactRows, err := s.db.GetAlertAnalysisArtifactsByAnalysisID(latest.AnalysisID)
			if err != nil {
				log.Printf("Failed to load alert analysis artifacts: %v", err)
			} else {
				for _, artifact := range artifactRows {
					artifacts = append(artifacts, client.AlertAnalysisArtifactInput{
						Type:    artifact.ArtifactType,
						Query:   artifact.Query,
						Result:  artifact.Result,
						Summary: artifact.Summary,
					})
				}
			}
		}

		alertInputs = append(alertInputs, client.AlertSummaryInput{
			Fingerprint:     alert.Fingerprint,
			AlertName:       alert.AlarmTitle,
			Severity:        alert.Severity,
			Status:          alert.Status,
			AnalysisSummary: summary,
			AnalysisDetail:  detail,
			Artifacts:       artifacts,
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
