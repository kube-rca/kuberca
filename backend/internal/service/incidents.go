package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/sse"
)

type RcaService struct {
	repo              *db.Postgres
	agentService      *AgentService
	embeddingService  *EmbeddingService
	sseHub            *sse.Hub
	mu                sync.Mutex
	inFlightSummaries map[string]time.Time
}

func NewRcaService(repo *db.Postgres, agentService *AgentService, embeddingService *EmbeddingService, sseHub *sse.Hub) *RcaService {
	return &RcaService{
		repo:              repo,
		agentService:      agentService,
		embeddingService:  embeddingService,
		sseHub:            sseHub,
		inFlightSummaries: make(map[string]time.Time),
	}
}

func (s *RcaService) GetIncidentList() ([]model.IncidentListResponse, error) {
	return s.repo.GetIncidentList()
}

func (s *RcaService) GetIncidentDetail(id string) (*model.IncidentDetailResponse, error) {
	// Incident 상세 조회
	incident, err := s.repo.GetIncidentDetail(id)
	if err != nil {
		return nil, err
	}

	// 연결된 Alert 목록 조회
	alerts, err := s.repo.GetAlertsByIncidentID(id)
	if err != nil {
		return nil, err
	}

	incident.Alerts = alerts
	return incident, nil
}

func (s *RcaService) UpdateIncident(id string, req model.UpdateIncidentRequest) error {
	if err := s.repo.UpdateIncident(id, req); err != nil {
		return err
	}
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventIncidentUpdated,
			Data: sse.EventData{IncidentID: id},
		})
	}
	return nil
}

func (s *RcaService) HideIncident(id string) error {
	if err := s.repo.HideIncident(id); err != nil {
		return err
	}
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventIncidentUpdated,
			Data: sse.EventData{IncidentID: id},
		})
	}
	return nil
}

// GetHiddenIncidentList - 숨김 처리된 Incident 목록 조회
func (s *RcaService) GetHiddenIncidentList() ([]model.IncidentListResponse, error) {
	return s.repo.GetHiddenIncidentList()
}

// UnhideIncident - Incident 숨김 해제 (추가됨)
func (s *RcaService) UnhideIncident(id string) error {
	if err := s.repo.UnhideIncident(id); err != nil {
		return err
	}
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventIncidentUpdated,
			Data: sse.EventData{IncidentID: id},
		})
	}
	return nil
}

func (s *RcaService) ResolveIncident(id string, resolvedBy string) error {
	// 1. Incident 상태를 resolved로 변경
	if err := s.repo.ResolveIncident(id, resolvedBy); err != nil {
		return err
	}

	// SSE broadcast: incident resolved
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventIncidentResolved,
			Data: sse.EventData{IncidentID: id},
		})
	}

	// 2. Agent에 최종 분석 요청 (비동기)
	go s.requestIncidentSummary(id)

	return nil
}

// requestIncidentSummary - Incident 최종 분석 요청 (goroutine에서 실행)
func (s *RcaService) requestIncidentSummary(incidentID string) {
	if startedAt, ok := s.beginIncidentSummary(incidentID); !ok {
		log.Printf(
			"Skipping duplicate incident summary request (incident_id=%s, in_flight_ms=%d)",
			incidentID,
			time.Since(startedAt).Milliseconds(),
		)
		return
	}
	defer s.finishIncidentSummary(incidentID)

	// SSE broadcast: 인시던트 분석 시작 이벤트
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventAnalysisStarted,
			Data: sse.EventData{IncidentID: incidentID},
		})
	}

	requestStartedAt := time.Now()

	// Incident 정보 조회
	incident, err := s.repo.GetIncidentDetail(incidentID)
	if err != nil {
		log.Printf("Failed to get incident for summary: %v", err)
		if s.sseHub != nil {
			s.sseHub.Broadcast(sse.Event{
				Type: sse.EventAnalysisFailed,
				Data: sse.EventData{IncidentID: incidentID, Message: err.Error()},
			})
		}
		return
	}

	// Alert 목록 조회 (분석 내용 포함)
	alerts, err := s.repo.GetAlertsWithAnalysisByIncidentID(incidentID)
	if err != nil {
		log.Printf("Failed to get alerts for summary: %v", err)
		if s.sseHub != nil {
			s.sseHub.Broadcast(sse.Event{
				Type: sse.EventAnalysisFailed,
				Data: sse.EventData{IncidentID: incidentID, Message: err.Error()},
			})
		}
		return
	}

	// Agent에 최종 분석 요청
	resp, err := s.agentService.RequestIncidentSummary(incident, alerts)
	if err != nil {
		log.Printf("Failed to request incident summary: %v", err)
		if s.sseHub != nil {
			s.sseHub.Broadcast(sse.Event{
				Type: sse.EventAnalysisFailed,
				Data: sse.EventData{IncidentID: incidentID, Message: err.Error()},
			})
		}
		return
	}

	// DB에 분석 결과 저장 (title 포함)
	if err := s.repo.UpdateIncidentAnalysis(incidentID, resp.Title, resp.Summary, resp.Detail); err != nil {
		log.Printf("Failed to save incident analysis: %v", err)
		return
	}

	log.Printf("Incident summary saved (incident_id=%s)", incidentID)

	// SSE broadcast: incident updated (summary saved)
	if s.sseHub != nil {
		s.sseHub.Broadcast(sse.Event{
			Type: sse.EventIncidentUpdated,
			Data: sse.EventData{IncidentID: incidentID},
		})
	}

	// 임베딩 생성 (유사 인시던트 검색용)
	if s.embeddingService != nil && resp.Summary != "" {
		embeddingID, model, err := s.embeddingService.CreateEmbedding(context.Background(), incidentID, resp.Summary)
		if err != nil {
			log.Printf("Failed to create embedding for incident %s: %v", incidentID, err)
		} else {
			log.Printf("Embedding created (incident_id=%s, embedding_id=%d, model=%s)", incidentID, embeddingID, model)
		}
	}

	log.Printf(
		"Incident summary finished (incident_id=%s, elapsed_ms=%d)",
		incidentID,
		time.Since(requestStartedAt).Milliseconds(),
	)
}

// GetAlertsByIncidentID - Incident에 속한 Alert 목록 조회
func (s *RcaService) GetAlertsByIncidentID(incidentID string) ([]model.AlertListResponse, error) {
	return s.repo.GetAlertsByIncidentID(incidentID)
}

// ============================================================================
// Alert 관련 메서드
// ============================================================================

// GetAlertList - 전체 Alert 목록 조회
func (s *RcaService) GetAlertList() ([]model.AlertListResponse, error) {
	return s.repo.GetAlertList()
}

// GetAlertDetail - Alert 상세 조회
func (s *RcaService) GetAlertDetail(id string) (*model.AlertDetailResponse, error) {
	detail, err := s.repo.GetAlertDetail(id)
	if err != nil {
		return nil, err
	}
	detail.IsAnalyzing = s.agentService.IsAnalyzing(id)
	return detail, nil
}

// UpdateAlertIncidentID - Alert의 Incident ID 변경
func (s *RcaService) UpdateAlertIncidentID(alertID, incidentID string) error {
	return s.repo.UpdateAlertIncidentID(alertID, incidentID)
}

func (s *RcaService) GetFeedback(targetType, targetID string, userID int64) (*model.FeedbackSummary, error) {
	upVotes, downVotes, myVote, err := s.repo.GetVoteSummary(targetType, targetID, userID)
	if err != nil {
		return nil, err
	}

	comments, err := s.repo.GetComments(targetType, targetID, 200)
	if err != nil {
		return nil, err
	}

	return &model.FeedbackSummary{
		TargetType: targetType,
		TargetID:   targetID,
		UpVotes:    upVotes,
		DownVotes:  downVotes,
		MyVote:     myVote,
		Comments:   comments,
	}, nil
}

func (s *RcaService) CreateFeedbackComment(targetType, targetID string, userID int64, loginID, body string) (*model.FeedbackComment, error) {
	return s.repo.CreateComment(targetType, targetID, userID, loginID, body)
}

func (s *RcaService) UpdateFeedbackComment(targetType, targetID string, commentID, userID int64, body string) (*model.FeedbackComment, error) {
	return s.repo.UpdateComment(targetType, targetID, commentID, userID, body)
}

func (s *RcaService) DeleteFeedbackComment(targetType, targetID string, commentID, userID int64) error {
	return s.repo.DeleteComment(targetType, targetID, commentID, userID)
}

func (s *RcaService) VoteFeedback(targetType, targetID string, userID int64, voteType string) error {
	return s.repo.UpsertVote(targetType, targetID, userID, voteType)
}

// TriggerAlertAnalysis - 수동 분석 트리거 (Manual Mode 또는 Auto Mode에서 미분석 alert 대상)
func (s *RcaService) TriggerAlertAnalysis(alertID string) error {
	// 1. Alert 상세 조회
	alertDetail, err := s.repo.GetAlertDetail(alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	// 2. AlertDetailResponse → model.Alert 변환
	alert, err := alertDetailToAlert(alertDetail)
	if err != nil {
		return fmt.Errorf("failed to convert alert: %w", err)
	}

	// 3. threadTS, incidentID 조회
	threadTS, _ := s.repo.GetAlertThreadTS(alertDetail.Fingerprint)
	incidentID := ""
	if alertDetail.IncidentID != nil {
		incidentID = *alertDetail.IncidentID
	}

	// 4. 비동기 분석 요청 (수동 트리거이므로 thread 체크 스킵)
	go s.agentService.RequestAnalysis(alert, alertID, threadTS, incidentID, true)
	log.Printf("Manual analysis triggered (alert_id=%s, incident_id=%s)", alertID, incidentID)

	return nil
}

// TriggerIncidentAnalysis - 수동 Incident 분석 트리거
func (s *RcaService) TriggerIncidentAnalysis(incidentID string) error {
	_, err := s.repo.GetIncidentDetail(incidentID)
	if err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}
	go s.requestIncidentSummary(incidentID)
	log.Printf("Manual incident analysis triggered (incident_id=%s)", incidentID)
	return nil
}

func (s *RcaService) beginIncidentSummary(incidentID string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if startedAt, exists := s.inFlightSummaries[incidentID]; exists {
		if time.Since(startedAt) < inFlightStaleTTL {
			return startedAt, false
		}
		log.Printf("Evicting stale in-flight incident summary entry (incident_id=%s, age_ms=%d)", incidentID, time.Since(startedAt).Milliseconds())
	}

	startedAt := time.Now()
	s.inFlightSummaries[incidentID] = startedAt
	return startedAt, true
}

func (s *RcaService) finishIncidentSummary(incidentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inFlightSummaries, incidentID)
}

// alertDetailToAlert - AlertDetailResponse → model.Alert 변환
func alertDetailToAlert(detail *model.AlertDetailResponse) (model.Alert, error) {
	var labels map[string]string
	if err := json.Unmarshal(detail.Labels, &labels); err != nil {
		return model.Alert{}, fmt.Errorf("failed to unmarshal labels: %w", err)
	}

	var annotations map[string]string
	if detail.Annotations != nil {
		if err := json.Unmarshal(detail.Annotations, &annotations); err != nil {
			return model.Alert{}, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	alert := model.Alert{
		Status:      detail.Status,
		Labels:      labels,
		Annotations: annotations,
		StartsAt:    detail.FiredAt,
		Fingerprint: detail.Fingerprint,
	}
	if detail.ResolvedAt != nil {
		alert.EndsAt = *detail.ResolvedAt
	}
	return alert, nil
}

// Mock 데이터 생성 (테스트용)
func (s *RcaService) CreateMockIncident() (string, error) {
	return s.repo.CreateMockIncident()
}
