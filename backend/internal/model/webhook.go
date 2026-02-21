package model

import "time"

// WebhookHeader - 헤더 키-값 쌍
type WebhookHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// WebhookConfig - DB에 저장되는 웹훅 설정 구조체
type WebhookConfig struct {
	ID        int             `json:"id"`
	URL       string          `json:"url"`
	Method    string          `json:"method"`
	Headers   []WebhookHeader `json:"headers"`
	Body      string          `json:"body"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// WebhookConfigRequest - 웹훅 설정 생성/수정 요청 구조체
type WebhookConfigRequest struct {
	URL     string          `json:"url"`
	Method  string          `json:"method"`
	Headers []WebhookHeader `json:"headers"`
	Body    string          `json:"body"`
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
