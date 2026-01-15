// Agent 서비스와 HTTP 통신하는 클라이언트 정의
//
// 환경변수:
//   - AGENT_URL: Agent 서비스 URL (예: http://kube-rca-agent.kube-rca.svc:8000)
//
// Agent에 전달하는 데이터:
//   - alert: 개별 알림 정보
//   - thread_ts: Slack 쓰레드 ID

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

// AgentClient 구조체 정의
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
}

// AgentAnalysisRequest 구조체 정의
type AgentAnalysisRequest struct {
	Alert      model.Alert `json:"alert"`
	ThreadTS   string      `json:"thread_ts"`
	IncidentID string      `json:"incident_id,omitempty"`
}

// AgentAnalysisResponse 구조체 정의
type AgentAnalysisResponse struct {
	Status   string `json:"status"`
	ThreadTS string `json:"thread_ts"`
	Analysis string `json:"analysis"`
}

// IncidentSummaryRequest - Incident 최종 분석 요청
type IncidentSummaryRequest struct {
	IncidentID string              `json:"incident_id"`
	Title      string              `json:"title"`
	Severity   string              `json:"severity"`
	FiredAt    string              `json:"fired_at"`
	ResolvedAt string              `json:"resolved_at"`
	Alerts     []AlertSummaryInput `json:"alerts"`
}

// AlertSummaryInput - 개별 Alert 분석 내용 (Agent에 전달)
type AlertSummaryInput struct {
	Fingerprint     string `json:"fingerprint"`
	AlertName       string `json:"alert_name"`
	Severity        string `json:"severity"`
	Status          string `json:"status"`
	AnalysisSummary string `json:"analysis_summary"`
	AnalysisDetail  string `json:"analysis_detail"`
}

// IncidentSummaryResponse - Incident 최종 분석 응답
type IncidentSummaryResponse struct {
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Detail  string `json:"detail"`
}

// AgentClient 객체 생성
func NewAgentClient(cfg config.AgentConfig) *AgentClient {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://kube-rca-agent.kube-rca.svc:8000"
	}

	return &AgentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // AI 분석 시간 고려
		},
	}
}

// Agent 설정 여부 체크
func (c *AgentClient) IsConfigured() bool {
	return c.baseURL != ""
}

// POST /analyze 분석 요청하고 분석 결과 반환 (동기)
func (c *AgentClient) RequestAnalysis(alert model.Alert, threadTS string) (*AgentAnalysisResponse, error) {
	req := AgentAnalysisRequest{
		Alert:      alert,
		ThreadTS:   threadTS,
		IncidentID: alert.Fingerprint,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/analyze", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var analysisResp AgentAnalysisResponse
	if err := json.Unmarshal(body, &analysisResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &analysisResp, nil
}

// POST /summarize-incident - Incident 최종 분석 요청
func (c *AgentClient) RequestIncidentSummary(req IncidentSummaryRequest) (*IncidentSummaryResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incident summary request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/summarize-incident", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var summaryResp IncidentSummaryResponse
	if err := json.Unmarshal(body, &summaryResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &summaryResp, nil
}
