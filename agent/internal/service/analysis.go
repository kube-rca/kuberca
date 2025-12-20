package service

import "github.com/kube-rca/agent/internal/model"

type AnalysisService struct{}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

// analysisCompleteMessage is a placeholder response until analysis is implemented.
const analysisCompleteMessage = "Analysis Complete!"

func (s *AnalysisService) AnalyzeWebhook(_ model.AlertmanagerWebhook) string {
	return analysisCompleteMessage
}
