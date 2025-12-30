// Agent 분석 요청 및 응답 처리 비즈니스 로직 정의
//
// 처리 흐름:
//  1. RequestAnalysis: Agent에 분석 요청 (goroutine에서 호출)
//  2. Agent 응답을 DB에 저장 (incidents.analysis_summary, analysis_detail)
//  3. Agent 응답을 Slack 쓰레드에 전송

package service

import (
	"context"
	"log"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

// AgentService 구조체 정의
type AgentService struct {
	agentClient *client.AgentClient
	slackClient *client.SlackClient
	db          *db.Postgres
}

// AgentService 객체 생성
func NewAgentService(agentClient *client.AgentClient, slackClient *client.SlackClient, database *db.Postgres) *AgentService {
	return &AgentService{
		agentClient: agentClient,
		slackClient: slackClient,
		db:          database,
	}
}

func (s *AgentService) RequestAnalysis(alert model.Alert, threadTS string) {
	if threadTS == "" {
		log.Printf("No thread_ts for alert (fingerprint=%s), skipping agent request", alert.Fingerprint)
		return
	}

	log.Printf("Requesting agent analysis (fingerprint=%s, status=%s, thread_ts=%s)", alert.Fingerprint, alert.Status, threadTS)

	// Agent에 분석 요청 (동기)
	resp, err := s.agentClient.RequestAnalysis(alert, threadTS)
	if err != nil {
		log.Printf("Failed to request agent analysis: %v", err)
		return
	}

	// 분석 결과를 DB에 저장 (incidents.analysis_summary, analysis_detail)
	ctx := context.Background()
	if err := s.db.UpdateAnalysis(ctx, alert.Fingerprint, resp.Analysis, resp.Analysis); err != nil {
		log.Printf("Failed to save analysis to DB: %v", err)
		// DB 저장 실패해도 Slack 전송은 계속 진행
	} else {
		log.Printf("Saved analysis to DB (fingerprint=%s)", alert.Fingerprint)
	}

	// 분석 결과를 Slack 쓰레드에 전송
	log.Printf("Sending analysis to Slack thread (thread_ts=%s)", threadTS)
	if err := s.slackClient.SendToThread(threadTS, resp.Analysis); err != nil {
		log.Printf("Failed to send analysis to Slack: %v", err)
	}
}
