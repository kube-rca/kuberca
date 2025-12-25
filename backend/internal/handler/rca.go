package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/service"
)

type RcaHandler struct {
	svc *service.RcaService
}

// 명칭을 NewRcaHandler로 변경
func NewRcaHandler(svc *service.RcaService) *RcaHandler {
	return &RcaHandler{svc: svc}
}

func (h *RcaHandler) GetIncidents(c *gin.Context) {
	res, err := h.svc.GetIncidentList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

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
