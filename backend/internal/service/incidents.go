package service

import (
	"fmt"
	"time"

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
	return s.repo.GetIncidentDetail(id)
}

func (s *RcaService) UpdateIncident(id string, req model.UpdateIncidentRequest) error {
	return s.repo.UpdateIncident(id, req)
}

func (s *RcaService) HideIncident(id string) error {
	return s.repo.HideIncident(id)
}

// Mock 데이터 생성 추후 삭제 에정
func (s *RcaService) CreateMockIncident() (string, error) {
	// 1. 고유 ID 생성 (예: INC-1703501234)
	timestamp := time.Now().Unix()
	id := fmt.Sprintf("INC-%d", timestamp)

	// 2. 테스트 데이터 생성
	title := fmt.Sprintf("테스트용 알람 (생성시간: %d)", timestamp)
	severity := "Warning"
	status := "Firing"

	// 3. DB 저장 요청
	err := s.repo.CreateIncident(id, title, severity, status)
	if err != nil {
		return "", err
	}

	return id, nil
}
