package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

// 허용된 설정 키
var allowedKeys = map[string]bool{
	"flapping": true,
	"slack":    true,
	"ai":      true,
}

// appSettingsRepo - DB 인터페이스
type appSettingsRepo interface {
	GetAppSetting(ctx context.Context, key string) (*model.AppSetting, error)
	UpsertAppSetting(ctx context.Context, key string, value json.RawMessage) error
	GetAllAppSettings(ctx context.Context) ([]model.AppSetting, error)
}

// AppSettingsService - 앱 설정 비즈니스 로직
type AppSettingsService struct {
	db         appSettingsRepo
	envFlapping config.FlappingConfig
	envSlack    config.SlackConfig
}

func NewAppSettingsService(db appSettingsRepo, envFlapping config.FlappingConfig, envSlack config.SlackConfig) *AppSettingsService {
	return &AppSettingsService{
		db:         db,
		envFlapping: envFlapping,
		envSlack:    envSlack,
	}
}

// IsAllowedKey - 허용된 키인지 확인
func IsAllowedKey(key string) bool {
	return allowedKeys[key]
}

// GetSetting - 단건 조회 (DB 값 없으면 nil)
func (s *AppSettingsService) GetSetting(ctx context.Context, key string) (*model.AppSetting, error) {
	return s.db.GetAppSetting(ctx, key)
}

// UpdateSetting - 설정 저장 (유효성 검증 후)
func (s *AppSettingsService) UpdateSetting(ctx context.Context, key string, value json.RawMessage) error {
	if !IsAllowedKey(key) {
		return fmt.Errorf("unknown setting key: %s", key)
	}

	// 키별 JSON 구조 검증
	switch key {
	case "flapping":
		var v model.FlappingSettings
		if err := json.Unmarshal(value, &v); err != nil {
			return fmt.Errorf("invalid flapping settings: %w", err)
		}
	case "slack":
		var v model.SlackSettings
		if err := json.Unmarshal(value, &v); err != nil {
			return fmt.Errorf("invalid slack settings: %w", err)
		}
	case "ai":
		var v model.AISettings
		if err := json.Unmarshal(value, &v); err != nil {
			return fmt.Errorf("invalid ai settings: %w", err)
		}
	}

	return s.db.UpsertAppSetting(ctx, key, value)
}

// GetAllSettings - 전체 설정 조회
func (s *AppSettingsService) GetAllSettings(ctx context.Context) ([]model.AppSetting, error) {
	return s.db.GetAllAppSettings(ctx)
}

// GetFlappingSettings - DB 조회 + ENV fallback 병합
func (s *AppSettingsService) GetFlappingSettings() *model.FlappingSettings {
	ctx := context.Background()
	setting, err := s.db.GetAppSetting(ctx, "flapping")
	if err != nil {
		log.Printf("Failed to get flapping settings from DB: %v", err)
		return nil
	}
	if setting == nil {
		return nil
	}

	var fs model.FlappingSettings
	if err := json.Unmarshal(setting.Value, &fs); err != nil {
		log.Printf("Failed to unmarshal flapping settings: %v", err)
		return nil
	}
	return &fs
}

// GetEffectiveFlappingConfig - DB 값 우선, 없으면 ENV fallback
func (s *AppSettingsService) GetEffectiveFlappingConfig() config.FlappingConfig {
	if dbSettings := s.GetFlappingSettings(); dbSettings != nil {
		return config.FlappingConfig{
			Enabled:                dbSettings.Enabled,
			DetectionWindowMinutes: dbSettings.DetectionWindowMinutes,
			CycleThreshold:         dbSettings.CycleThreshold,
			ClearanceWindowMinutes: dbSettings.ClearanceWindowMinutes,
		}
	}
	return s.envFlapping
}

// GetSlackSettings - DB 조회
func (s *AppSettingsService) GetSlackSettings() *model.SlackSettings {
	ctx := context.Background()
	setting, err := s.db.GetAppSetting(ctx, "slack")
	if err != nil {
		log.Printf("Failed to get slack settings from DB: %v", err)
		return nil
	}
	if setting == nil {
		return nil
	}

	var ss model.SlackSettings
	if err := json.Unmarshal(setting.Value, &ss); err != nil {
		log.Printf("Failed to unmarshal slack settings: %v", err)
		return nil
	}
	return &ss
}

// IsSlackFallbackEnabled - Slack fallback 활성화 여부 (DB 설정 우선)
func (s *AppSettingsService) IsSlackFallbackEnabled() bool {
	if dbSettings := s.GetSlackSettings(); dbSettings != nil {
		return dbSettings.Enabled
	}
	// DB에 설정 없으면 ENV 기반 Slack이 있으면 enabled
	return s.envSlack.BotToken != "" && s.envSlack.ChannelID != ""
}

// GetAISettings - DB 조회
func (s *AppSettingsService) GetAISettings() *model.AISettings {
	ctx := context.Background()
	setting, err := s.db.GetAppSetting(ctx, "ai")
	if err != nil {
		log.Printf("Failed to get AI settings from DB: %v", err)
		return nil
	}
	if setting == nil {
		return nil
	}

	var as model.AISettings
	if err := json.Unmarshal(setting.Value, &as); err != nil {
		log.Printf("Failed to unmarshal AI settings: %v", err)
		return nil
	}
	return &as
}

// GetSettingWithFallback - 개별 키 조회 + ENV fallback 포함 응답 생성
func (s *AppSettingsService) GetSettingWithFallback(ctx context.Context, key string) (*model.AppSetting, error) {
	setting, err := s.db.GetAppSetting(ctx, key)
	if err != nil {
		return nil, err
	}
	if setting != nil {
		return setting, nil
	}

	// DB에 없으면 ENV 기반 기본값 생성
	var fallbackValue interface{}
	switch key {
	case "flapping":
		fallbackValue = model.FlappingSettings{
			Enabled:                s.envFlapping.Enabled,
			DetectionWindowMinutes: s.envFlapping.DetectionWindowMinutes,
			CycleThreshold:         s.envFlapping.CycleThreshold,
			ClearanceWindowMinutes: s.envFlapping.ClearanceWindowMinutes,
		}
	case "slack":
		fallbackValue = model.SlackSettings{
			Enabled:   s.envSlack.BotToken != "" && s.envSlack.ChannelID != "",
			ChannelId: s.envSlack.ChannelID,
		}
	case "ai":
		fallbackValue = model.AISettings{
			Provider: "gemini",
			ModelId:  "",
		}
	default:
		return nil, fmt.Errorf("unknown setting key: %s", key)
	}

	data, err := json.Marshal(fallbackValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fallback value: %w", err)
	}

	return &model.AppSetting{
		Key:   key,
		Value: data,
	}, nil
}
