package service

import (
	"context"

	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

type RcaService struct {
	repo *db.Postgres
}

func NewRcaService(repo *db.Postgres) *RcaService {
	return &RcaService{repo: repo}
}

func (s *RcaService) GetIncidentList() ([]model.IncidentListResponse, error) {
	return s.repo.GetIncidentList()
}

func (s *RcaService) GetIncidentDetail(id string) (*model.IncidentDetailResponse, error) {
	ctx := context.Background()

	// Incident 상세 조회
	incident, err := s.repo.GetIncidentDetail(id)
	if err != nil {
		return nil, err
	}

	// 연결된 Alert 목록 조회
	alerts, err := s.repo.GetAlertsByIncidentID(ctx, id)
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
	ctx := context.Background()
	return s.repo.ResolveIncident(ctx, id, resolvedBy)
}

// GetAlertsByIncidentID - Incident에 속한 Alert 목록 조회
func (s *RcaService) GetAlertsByIncidentID(incidentID string) ([]model.AlertListResponse, error) {
	ctx := context.Background()
	return s.repo.GetAlertsByIncidentID(ctx, incidentID)
}

// ============================================================================
// Alert 관련 메서드
// ============================================================================

// GetAlertList - 전체 Alert 목록 조회
func (s *RcaService) GetAlertList() ([]model.AlertListResponse, error) {
	ctx := context.Background()
	return s.repo.GetAlertList(ctx)
}

// GetAlertDetail - Alert 상세 조회
func (s *RcaService) GetAlertDetail(id string) (*model.AlertDetailResponse, error) {
	ctx := context.Background()
	return s.repo.GetAlertDetail(ctx, id)
}

// Mock 데이터 생성 (테스트용)
func (s *RcaService) CreateMockIncident() (string, error) {
	return s.repo.CreateMockIncident()
}
