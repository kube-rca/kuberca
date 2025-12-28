package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
)

// 헬스체크 엔드포인트
// Ping godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} model.PingResponse
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, model.PingResponse{Message: "pong"})
}

// 루트 엔드포인트
// Root godoc
// @Summary Root endpoint
// @Tags health
// @Produce json
// @Success 200 {object} model.RootResponse
// @Router / [get]
func Root(c *gin.Context) {
	c.JSON(http.StatusOK, model.RootResponse{
		Status:  "ok",
		Message: "Gin basic API server is running",
	})
}
