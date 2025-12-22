// Agent 분석 요청 및 응답 처리 비즈니스 로직 정의
//
// 처리 흐름:
//  1. RequestAnalysis: Agent에 비동기 분석 요청 (goroutine에서 호출)
//  2. HandleCallback: Agent 응답을 받아 Slack 쓰레드에 전송

package service

import (
	"log"
	"os"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/model"
)

// AgentService 구조체 정의
type AgentService struct {
	agentClient *client.AgentClient
	slackClient *client.SlackClient
	callbackURL string
}

// AgentCallbackRequest 구조체 정의
type AgentCallbackRequest struct {
	ThreadTS string `json:"thread_ts"`         // Slack 쓰레드 ID
	Analysis string `json:"analysis"`          // AI 분석 결과
	Error    string `json:"error,omitempty"`   // 분석 실패 에러 메시지
}

// AgentService 객체 생성
func NewAgentService(agentClient *client.AgentClient, slackClient *client.SlackClient) *AgentService {
	callbackURL := os.Getenv("BACKEND_CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "http://kube-rca-backend.kube-rca.svc:8080"
	}

	return &AgentService{
		agentClient: agentClient,
		slackClient: slackClient,
		callbackURL: callbackURL + "/callback/agent",
	}
}

// Agent에 분석 요청 (비동기)
// firing, resolved 알림 모두 Slack 전송 후 goroutine에서 호출
func (s *AgentService) RequestAnalysis(alert model.Alert, threadTS string) {
	if threadTS == "" {
		log.Printf("No thread_ts for alert (fingerprint=%s), skipping agent request", alert.Fingerprint)
		return
	}

	err := s.agentClient.RequestAnalysis(alert, threadTS, s.callbackURL)
	if err != nil {
		log.Printf("Failed to request agent analysis: %v", err)
	} else {
		log.Printf("Requested agent analysis (fingerprint=%s, status=%s, thread_ts=%s)", alert.Fingerprint, alert.Status, threadTS)
	}
}

// Agent 분석 완료 후 /callback/agent  호출
func (s *AgentService) HandleCallback(req AgentCallbackRequest) error {
	if req.Error != "" {
		log.Printf("Agent returned error: %s", req.Error)
		return s.slackClient.SendToThread(req.ThreadTS, "분석 중 오류가 발생했습니다: "+req.Error)
	}

	log.Printf("Sending analysis to Slack thread (thread_ts=%s)", req.ThreadTS)
	return s.slackClient.SendToThread(req.ThreadTS, req.Analysis)
}
