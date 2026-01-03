package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

type RcaHandler struct {
	svc *service.RcaService
}

// 명칭을 NewRcaHandler로 변경
func NewRcaHandler(svc *service.RcaService) *RcaHandler {
	return &RcaHandler{svc: svc}
}

// GetIncidents godoc
// @Summary List incidents
// @Tags incidents
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.IncidentListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/incidents [get]
func (h *RcaHandler) GetIncidents(c *gin.Context) {
	res, err := h.svc.GetIncidentList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetIncidentDetail godoc
// @Summary Get incident detail
// @Tags incidents
// @Produce json
// @Security BearerAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} model.IncidentDetailEnvelope
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/incidents/{id} [get]
func (h *RcaHandler) GetIncidentDetail(c *gin.Context) {
	id := c.Param("id")

	res, err := h.svc.GetIncidentDetail(id)
	if err != nil {
		// 데이터를 못 찾았을 경우 (pgx.ErrNoRows 등 처리 가능하나 일단 500/404 통합)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "데이터를 찾을 수 없거나 DB 오류입니다.",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.IncidentDetailEnvelope{
		Status: "success",
		Data:   res,
	})
}

// UpdateIncident godoc
// @Summary Update incident detail
// @Tags incidents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Incident ID"
// @Param request body model.UpdateIncidentRequest true "Incident update payload"
// @Success 200 {object} model.IncidentUpdateResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/incidents/{id} [put]
func (h *RcaHandler) UpdateIncident(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateIncidentRequest
	// JSON 파싱
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "잘못된 요청 데이터입니다.",
			"error":   err.Error(),
		})
		return
	}

	// 서비스 호출
	err := h.svc.UpdateIncident(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "업데이트 실패",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.IncidentUpdateResponse{
		Status:     "success",
		Message:    "RCA 정보가 성공적으로 수정되었습니다.",
		IncidentID: id,
	})
}

// HideIncident godoc
// @Summary Hide incident
// @Tags incidents
// @Produce json
// @Security BearerAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} model.IncidentUpdateResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/incidents/{id} [patch]
func (h *RcaHandler) HideIncident(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.HideIncident(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Incident hidden successfully",
	})
}

// CreateMockIncident godoc
// @Summary Create mock incident
// @Tags incidents
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.MockIncidentResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/incidents/mock [post]
func (h *RcaHandler) CreateMockIncident(c *gin.Context) {
	newID, err := h.svc.CreateMockIncident()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Mock 데이터 생성 실패",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.MockIncidentResponse{
		Status:     "success",
		Message:    "Mock 데이터 1개가 DB에 저장되었습니다.",
		IncidentID: newID,
	})
}
