package service

import (
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
