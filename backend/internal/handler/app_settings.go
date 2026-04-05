package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

// appSettingsService - 서비스 인터페이스
type appSettingsService interface {
	GetSetting(ctx context.Context, key string) (*model.AppSetting, error)
	UpdateSetting(ctx context.Context, key string, value json.RawMessage) error
	GetAllSettings(ctx context.Context) ([]model.AppSetting, error)
	GetSettingWithFallback(ctx context.Context, key string) (*model.AppSetting, error)
	GetAISettings() *model.AISettings
}

// agentConfigUpdater - Agent AI 설정 업데이트 인터페이스
type agentConfigUpdater interface {
	UpdateAIConfig(provider, modelId string) error
}

// AppSettingsHandler - 앱 설정 핸들러
type AppSettingsHandler struct {
	svc         appSettingsService
	agentClient agentConfigUpdater
}

func NewAppSettingsHandler(svc appSettingsService, agentClient agentConfigUpdater) *AppSettingsHandler {
	return &AppSettingsHandler{svc: svc, agentClient: agentClient}
}

// ListAppSettings godoc
// @Summary List all app settings
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.AppSettingListResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/settings/app [get]
func (h *AppSettingsHandler) ListAppSettings(c *gin.Context) {
	settings, err := h.svc.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.AppSettingListResponse{Status: "success", Data: settings})
}

// GetAppSetting godoc
// @Summary Get an app setting by key (with ENV fallback)
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Param key path string true "Setting key (flapping, slack, ai)"
// @Success 200 {object} model.AppSettingResponse
// @Failure 400,404,500 {object} model.ErrorResponse
// @Router /api/v1/settings/app/{key} [get]
func (h *AppSettingsHandler) GetAppSetting(c *gin.Context) {
	key := c.Param("key")
	if !service.IsAllowedKey(key) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid setting key: " + key})
		return
	}

	setting, err := h.svc.GetSettingWithFallback(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.AppSettingResponse{Status: "success", Data: *setting})
}

// UpdateAppSetting godoc
// @Summary Update an app setting
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Setting key (flapping, slack, ai)"
// @Success 200 {object} model.AppSettingResponse
// @Failure 400,500 {object} model.ErrorResponse
// @Router /api/v1/settings/app/{key} [put]
func (h *AppSettingsHandler) UpdateAppSetting(c *gin.Context) {
	key := c.Param("key")
	if !service.IsAllowedKey(key) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid setting key: " + key})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "failed to read request body"})
		return
	}

	// JSON 유효성 확인
	if !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid JSON"})
		return
	}

	if err := h.svc.UpdateSetting(c.Request.Context(), key, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// AI 설정 변경 시 Agent에 알림
	if key == "ai" && h.agentClient != nil {
		if aiSettings := h.svc.GetAISettings(); aiSettings != nil {
			if err := h.agentClient.UpdateAIConfig(aiSettings.Provider, aiSettings.ModelId); err != nil {
				// Agent 알림 실패해도 DB 저장은 성공한 상태이므로 warning만 남김
				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "설정이 저장되었습니다. Agent 설정 반영에 실패했습니다: " + err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "설정이 저장되었습니다."})
}
