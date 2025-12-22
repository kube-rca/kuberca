// Alert 처리 비즈니스 로직 정의
// handler에서 받은 알림을 필터링하고 client를 통해 Slack으로 전송
//
// 처리 흐름:
//  1. 각 알림에 대해 shouldSendToSlack으로 필터링
//  2. 필터 통과한 알림을 SlackClient.SendAlert로 전송
//  3. 모든 알림(firing, resolved)에 대해 Agent에 비동기 분석 요청
//  4. 전송 성공/실패 카운트 반환

package service

import (
	"log"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/model"
)

// AlertService 구조체 정의
type AlertService struct {
	slackClient  *client.SlackClient
	agentService *AgentService
}

// AlertService 객체 생성
func NewAlertService(slackClient *client.SlackClient, agentService *AgentService) *AlertService {
	return &AlertService{
		slackClient:  slackClient,
		agentService: agentService,
	}
}

func (s *AlertService) ProcessWebhook(webhook model.AlertmanagerWebhook) (sent, failed int) {
	for _, alert := range webhook.Alerts {
		// 필터링: Slack으로 전송할 알림인지 확인
		if !s.shouldSendToSlack(alert) {
			continue
		}

		// Client 레이어로 해당 알림을 전달
		err := s.slackClient.SendAlert(alert, alert.Status)
		if err != nil {
			log.Printf("Failed to send alert to Slack: %v", err)
			failed++
			continue
		}

		log.Printf("Sent alert to Slack (fingerprint=%s, status=%s)", alert.Fingerprint, alert.Status)
		sent++

		// Agent에 비동기 분석 요청 (firing, resolved)
		threadTS, _ := s.slackClient.GetThreadTS(alert.Fingerprint)
		go s.agentService.RequestAnalysis(alert, threadTS)
	}
	return sent, failed
}

// 필터링 로직 예시:
//   - severity가 warning 이상만 전송
//   - 특정 namespace 제외 (예: kube-system)
//   - 특정 alertname만 전송
//
// Returns:
//   - bool: true면 Slack으로 전송, false면 무시
func (s *AlertService) shouldSendToSlack(alert model.Alert) bool {
	// 현재는 모든 알림 전송
	return true
}
