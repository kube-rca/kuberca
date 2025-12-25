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

// 1. GET /api/v1/incidents
func (h *RcaHandler) GetIncidents(c *gin.Context) {
	res, err := h.svc.GetIncidentList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// 2. GET /api/v1/incidents/:id
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

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// 3. PUT /api/v1/incidents/:id
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

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "RCA 정보가 성공적으로 수정되었습니다.",
		"incident_id": id,
	})
}

// Mock 데이터 생성 추후 삭제 에정
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

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Mock 데이터 1개가 DB에 저장되었습니다.",
		"incident_id": newID,
	})
}
