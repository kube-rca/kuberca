// Agent 서비스와 HTTP 통신하는 클라이언트 정의
//
// 환경변수:
//   - AGENT_URL: Agent 서비스 URL (예: http://kube-rca-agent.kube-rca.svc:8082)
//
// Agent에 전달하는 데이터:
//   - alert: 개별 알림 정보
//   - thread_ts: Slack 메시지 timestamp (응답을 해당 쓰레드에 전송하기 위함)
//   - callback_url: Agent가 분석 완료 후 호출할 Backend URL

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kube-rca/backend/internal/model"
)

// AgentClient 구조체 정의
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
}

// AgentAnalysisRequest - Agent로 전송하는 요청 구조체
type AgentAnalysisRequest struct {
	Alert       model.Alert `json:"alert"`
	ThreadTS    string      `json:"thread_ts"`
	CallbackURL string      `json:"callback_url"`
}

// AgentClient 객체 생성
func NewAgentClient() *AgentClient {
	baseURL := os.Getenv("AGENT_URL")
	if baseURL == "" {
		baseURL = "http://kube-rca-agent.kube-rca.svc:8082"
	}

	return &AgentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Agent 설정 여부 체크
func (c *AgentClient) IsConfigured() bool {
	return c.baseURL != ""
}

// Agent에 분석 요청 전송
// POST /analyze/alertmanager 엔드포인트로 요청
func (c *AgentClient) RequestAnalysis(alert model.Alert, threadTS, callbackURL string) error {
	req := AgentAnalysisRequest{
		Alert:       alert,
		ThreadTS:    threadTS,
		CallbackURL: callbackURL,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal agent request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/analyze/alertmanager", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("agent returned status: %d", resp.StatusCode)
	}

	return nil
}
