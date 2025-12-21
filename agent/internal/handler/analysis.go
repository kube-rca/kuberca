package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/agent/internal/model"
	"github.com/kube-rca/agent/internal/service"
)

type AnalysisHandler struct {
	analysisService *service.AnalysisService
}

func NewAnalysisHandler(analysisService *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService}
}

func (h *AnalysisHandler) AnalyzeAlertRequest(c *gin.Context) {
	var request model.AlertAnalysisRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Failed to parse alert request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result := h.analysisService.AnalyzeAlertRequest(request)
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"thread_ts": request.ThreadTS,
		"analysis":  result,
	})
}
