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

func (h *AnalysisHandler) AnalyzeAlertmanager(c *gin.Context) {
	var webhook model.AlertmanagerWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result := h.analysisService.AnalyzeWebhook(webhook)
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"receiver": webhook.Receiver,
		"analysis": result,
	})
}
