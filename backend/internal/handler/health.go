package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

// HealthHandler provides DB-aware health check endpoints.
type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Readyz godoc
// @Summary Readiness check (includes DB connectivity)
// @Tags health
// @Produce json
// @Success 200 {object} model.StatusResponse
// @Failure 503 {object} model.StatusResponse
// @Router /readyz [get]
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, model.StatusResponse{Status: "unavailable"})
		return
	}
	c.JSON(http.StatusOK, model.StatusResponse{Status: "ok"})
}

// Healthz godoc
// @Summary Startup/health check (includes DB connectivity)
// @Tags health
// @Produce json
// @Success 200 {object} model.StatusResponse
// @Failure 503 {object} model.StatusResponse
// @Router /healthz [get]
func (h *HealthHandler) Healthz(c *gin.Context) {
	h.Readyz(c)
}
