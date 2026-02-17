package service

import (
	"context"
	"log"

	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

type RcaService struct {
	repo             *db.Postgres
	agentService     *AgentService
	embeddingService *EmbeddingService
}

func NewRcaService(repo *db.Postgres, agentService *AgentService, embeddingService *EmbeddingService) *RcaService {
	return &RcaService{repo: repo, agentService: agentService, embeddingService: embeddingService}
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
	return s.repo.UpdateIncident(id, req)
}

func (s *RcaService) HideIncident(id string) error {
	return s.repo.HideIncident(id)
}

// GetHiddenIncidentList - 숨김 처리된 Incident 목록 조회
func (s *RcaService) GetHiddenIncidentList() ([]model.IncidentListResponse, error) {
	return s.repo.GetHiddenIncidentList()
}

// UnhideIncident - Incident 숨김 해제 (추가됨)
func (s *RcaService) UnhideIncident(id string) error {
	return s.repo.UnhideIncident(id)
}

func (s *RcaService) ResolveIncident(id string, resolvedBy string) error {
	// 1. Incident 상태를 resolved로 변경
	if err := s.repo.ResolveIncident(id, resolvedBy); err != nil {
		return err
	}

	// 2. Agent에 최종 분석 요청 (비동기)
	go s.requestIncidentSummary(id)

	return nil
}

// requestIncidentSummary - Incident 최종 분석 요청 (goroutine에서 실행)
func (s *RcaService) requestIncidentSummary(incidentID string) {
	// Incident 정보 조회
	incident, err := s.repo.GetIncidentDetail(incidentID)
	if err != nil {
		log.Printf("Failed to get incident for summary: %v", err)
		return
	}

	// Alert 목록 조회 (분석 내용 포함)
	alerts, err := s.repo.GetAlertsWithAnalysisByIncidentID(incidentID)
	if err != nil {
		log.Printf("Failed to get alerts for summary: %v", err)
		return
	}

	// Agent에 최종 분석 요청
	resp, err := s.agentService.RequestIncidentSummary(incident, alerts)
	if err != nil {
		log.Printf("Failed to request incident summary: %v", err)
		return
	}

	// DB에 분석 결과 저장 (title 포함)
	if err := s.repo.UpdateIncidentAnalysis(incidentID, resp.Title, resp.Summary, resp.Detail); err != nil {
		log.Printf("Failed to save incident analysis: %v", err)
		return
	}

	log.Printf("Incident summary saved (incident_id=%s)", incidentID)

	// 임베딩 생성 (유사 인시던트 검색용)
	if s.embeddingService != nil && resp.Summary != "" {
		embeddingID, model, err := s.embeddingService.CreateEmbedding(context.Background(), incidentID, resp.Summary)
		if err != nil {
			log.Printf("Failed to create embedding for incident %s: %v", incidentID, err)
		} else {
			log.Printf("Embedding created (incident_id=%s, embedding_id=%d, model=%s)", incidentID, embeddingID, model)
		}
	}
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
	return s.repo.GetAlertDetail(id)
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

func (s *RcaService) VoteFeedback(targetType, targetID string, userID int64, voteType string) error {
	return s.repo.UpsertVote(targetType, targetID, userID, voteType)
}

// Mock 데이터 생성 (테스트용)
func (s *RcaService) CreateMockIncident() (string, error) {
	return s.repo.CreateMockIncident()
}
