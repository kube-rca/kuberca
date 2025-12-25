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
