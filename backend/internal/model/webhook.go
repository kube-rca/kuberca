package model

import "time"

// WebhookConfig - DB에 저장되는 웹훅 설정 구조체
type WebhookConfig struct {
	ID         int       `json:"id"`
	URL        string    `json:"url"`
	Type       string    `json:"type"`
	Token      string    `json:"token,omitempty"`
	Channel    string    `json:"channel,omitempty"`
	Severities []string  `json:"severities"` // 빈 배열 = 모든 severity 수신
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebhookConfigRequest - 웹훅 설정 생성/수정 요청 구조체
type WebhookConfigRequest struct {
	URL        string   `json:"url"`
	Type       string   `json:"type"`
	Token      string   `json:"token,omitempty"`
	Channel    string   `json:"channel,omitempty"`
	Severities []string `json:"severities"` // 빈 배열 = 모든 severity 수신
}

// WebhookConfigResponse - 단건 조회 응답
type WebhookConfigResponse struct {
	Status string         `json:"status"`
	Data   *WebhookConfig `json:"data"`
}

// WebhookConfigListResponse - 목록 조회 응답
type WebhookConfigListResponse struct {
	Status string          `json:"status"`
	Data   []WebhookConfig `json:"data"`
}

// WebhookConfigMutationResponse - 생성/수정/삭제 응답
type WebhookConfigMutationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      int    `json:"id,omitempty"`
}
