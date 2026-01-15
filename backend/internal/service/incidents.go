package service

import (
	"log"

	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

type RcaService struct {
	repo         *db.Postgres
	agentService *AgentService
}

func NewRcaService(repo *db.Postgres, agentService *AgentService) *RcaService {
	return &RcaService{repo: repo, agentService: agentService}
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

	// DB에 분석 결과 저장
	if err := s.repo.UpdateIncidentAnalysis(incidentID, resp.Summary, resp.Detail); err != nil {
		log.Printf("Failed to save incident analysis: %v", err)
		return
	}

	log.Printf("Incident summary saved (incident_id=%s)", incidentID)
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

// Mock 데이터 생성 (테스트용)
func (s *RcaService) CreateMockIncident() (string, error) {
	return s.repo.CreateMockIncident()
}
