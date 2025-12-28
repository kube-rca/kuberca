package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

type EmbeddingHandler struct {
	svc *service.EmbeddingService
}

func NewEmbeddingHandler(svc *service.EmbeddingService) *EmbeddingHandler {
	return &EmbeddingHandler{svc: svc}
}

// CreateEmbedding godoc
// @Summary Create incident embedding
// @Tags embeddings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.EmbeddingRequest true "Incident embedding payload"
// @Success 200 {object} model.EmbeddingResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/embeddings [post]
func (h *EmbeddingHandler) CreateEmbedding(c *gin.Context) {
	var req model.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IncidentID == "" || req.IncidentSummary == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incident_id and incident_summary are required"})
		return
	}
	id, modelName, err := h.svc.CreateEmbedding(c.Request.Context(), req.IncidentID, req.IncidentSummary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.EmbeddingResponse{Status: "success", EmbeddingID: id, Model: modelName})
}
