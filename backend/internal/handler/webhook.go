package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
)

// webhookService - 서비스 인터페이스
type webhookService interface {
	ListWebhookConfigs(ctx context.Context) ([]model.WebhookConfig, error)
	GetWebhookConfig(ctx context.Context, id int) (*model.WebhookConfig, error)
	CreateWebhookConfig(ctx context.Context, req model.WebhookConfigRequest) (int, error)
	UpdateWebhookConfig(ctx context.Context, id int, req model.WebhookConfigRequest) error
	DeleteWebhookConfig(ctx context.Context, id int) error
}

// WebhookSettingsHandler - 웹훅 설정 관련 핸들러
type WebhookSettingsHandler struct {
	svc webhookService
}

func NewWebhookSettingsHandler(svc webhookService) *WebhookSettingsHandler {
	return &WebhookSettingsHandler{svc: svc}
}

// ListWebhookConfigs godoc
// @Summary List webhook configs
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.WebhookConfigListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/settings/webhooks [get]
func (h *WebhookSettingsHandler) ListWebhookConfigs(c *gin.Context) {
	configs, err := h.svc.ListWebhookConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.WebhookConfigListResponse{Status: "success", Data: configs})
}

// GetWebhookConfig godoc
// @Summary Get a webhook config by ID
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Param id path int true "Webhook Config ID"
// @Success 200 {object} model.WebhookConfigResponse
// @Failure 400,404,500 {object} model.ErrorResponse
// @Router /api/v1/settings/webhooks/{id} [get]
func (h *WebhookSettingsHandler) GetWebhookConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid id"})
		return
	}
	cfg, err := h.svc.GetWebhookConfig(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.WebhookConfigResponse{Status: "success", Data: cfg})
}

// CreateWebhookConfig godoc
// @Summary Create a webhook config
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.WebhookConfigRequest true "Webhook config"
// @Success 201 {object} model.WebhookConfigMutationResponse
// @Failure 400,500 {object} model.ErrorResponse
// @Router /api/v1/settings/webhooks [post]
func (h *WebhookSettingsHandler) CreateWebhookConfig(c *gin.Context) {
	var req model.WebhookConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}
	id, err := h.svc.CreateWebhookConfig(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model.WebhookConfigMutationResponse{
		Status:  "success",
		Message: "웹훅 설정이 생성되었습니다.",
		ID:      id,
	})
}

// UpdateWebhookConfig godoc
// @Summary Update a webhook config
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Webhook Config ID"
// @Param request body model.WebhookConfigRequest true "Webhook config"
// @Success 200 {object} model.WebhookConfigMutationResponse
// @Failure 400,404,500 {object} model.ErrorResponse
// @Router /api/v1/settings/webhooks/{id} [put]
func (h *WebhookSettingsHandler) UpdateWebhookConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid id"})
		return
	}
	var req model.WebhookConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if err := h.svc.UpdateWebhookConfig(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.WebhookConfigMutationResponse{
		Status:  "success",
		Message: "웹훅 설정이 수정되었습니다.",
		ID:      id,
	})
}

// DeleteWebhookConfig godoc
// @Summary Delete a webhook config
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Param id path int true "Webhook Config ID"
// @Success 200 {object} model.WebhookConfigMutationResponse
// @Failure 400,404,500 {object} model.ErrorResponse
// @Router /api/v1/settings/webhooks/{id} [delete]
func (h *WebhookSettingsHandler) DeleteWebhookConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid id"})
		return
	}
	if err := h.svc.DeleteWebhookConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.WebhookConfigMutationResponse{
		Status:  "success",
		Message: "웹훅 설정이 삭제되었습니다.",
		ID:      id,
	})
}
