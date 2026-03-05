package model

import (
	"encoding/json"
	"time"
)

// AppSetting - app_settings 테이블 구조체 (JSONB key-value)
type AppSetting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// FlappingSettings - Flapping Detection 설정
type FlappingSettings struct {
	Enabled                bool `json:"enabled"`
	DetectionWindowMinutes int  `json:"detectionWindowMinutes"`
	CycleThreshold         int  `json:"cycleThreshold"`
	ClearanceWindowMinutes int  `json:"clearanceWindowMinutes"`
}

// NotificationSettings - 알림 파이프라인 활성화/비활성화 설정
type NotificationSettings struct {
	Enabled bool `json:"enabled"`
}

// AISettings - AI Provider/Model 설정
type AISettings struct {
	Provider string `json:"provider"`
	ModelId  string `json:"modelId"`
}

// AppSettingResponse - 단건 조회 응답
type AppSettingResponse struct {
	Status string     `json:"status"`
	Data   AppSetting `json:"data"`
}

// AppSettingListResponse - 전체 조회 응답
type AppSettingListResponse struct {
	Status string       `json:"status"`
	Data   []AppSetting `json:"data"`
}
