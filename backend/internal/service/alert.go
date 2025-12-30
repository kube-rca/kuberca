// Alert 처리 비즈니스 로직 정의
// handler에서 받은 알림을 필터링하고 client를 통해 Slack으로 전송
//
// 처리 흐름:
//  1. 알림을 DB에 저장 (incidents 테이블)
//  2. resolved 상태면 resolved_at 업데이트
//  3. shouldSendToSlack으로 필터링
//  4. resolved 알림: DB에서 thread_ts 조회하여 메모리에 복원
//  5. SlackClient.SendAlert로 Slack 전송
//  6. firing 알림: thread_ts를 DB에 저장
//  7. Agent에 비동기 분석 요청 (firing, resolved)
//  8. 전송 성공/실패 카운트 반환

package service

import (
	"context"
	"log"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

// AlertService 구조체 정의
type AlertService struct {
	slackClient  *client.SlackClient
	agentService *AgentService
	db           *db.Postgres
}

// AlertService 객체 생성
func NewAlertService(slackClient *client.SlackClient, agentService *AgentService, database *db.Postgres) *AlertService {
	return &AlertService{
		slackClient:  slackClient,
		agentService: agentService,
		db:           database,
	}
}

func (s *AlertService) ProcessWebhook(webhook model.AlertmanagerWebhook) (sent, failed int) {
	ctx := context.Background()

	for _, alert := range webhook.Alerts {
		// 1. DB에 알림 저장 (incidents 테이블)
		if err := s.db.SaveAlertAsIncident(ctx, alert); err != nil {
			log.Printf("Failed to save alert to DB: %v", err)
			// DB 저장 실패해도 Slack 전송은 계속 진행
		}

		// 2. resolved 상태면 resolved_at 업데이트
		if alert.Status == "resolved" {
			if err := s.db.UpdateIncidentResolved(ctx, alert.Fingerprint, alert.EndsAt); err != nil {
				log.Printf("Failed to update incident resolved status: %v", err)
			}
		}

		// 3. 필터링: Slack으로 전송할 알림인지 확인
		if !s.shouldSendToSlack(alert) {
			continue
		}

		// 4. resolved 알림: DB에서 thread_ts 조회하여 메모리에 복원
		// (백엔드 재시작 시 메모리가 초기화되므로 DB에서 복원 필요)
		if alert.Status == "resolved" {
			if threadTS, ok := s.db.GetThreadTS(ctx, alert.Fingerprint); ok {
				s.slackClient.StoreThreadTS(alert.Fingerprint, threadTS)
			}
		}

		// 5. Client 레이어로 해당 알림을 전달
		err := s.slackClient.SendAlert(alert, alert.Status)
		if err != nil {
			log.Printf("Failed to send alert to Slack: %v", err)
			failed++
			continue
		}

		log.Printf("Sent alert to Slack (fingerprint=%s, status=%s)", alert.Fingerprint, alert.Status)
		sent++

		// 5. thread_ts를 DB에 저장 (firing 알림일 때)
		if alert.Status == "firing" {
			if threadTS, ok := s.slackClient.GetThreadTS(alert.Fingerprint); ok {
				if err := s.db.UpdateThreadTS(ctx, alert.Fingerprint, threadTS); err != nil {
					log.Printf("Failed to save thread_ts to DB: %v", err)
				}
			}
		}

		// 6. Agent에 비동기 분석 요청 (firing, resolved)
		// DB에서 thread_ts 조회 (메모리 대신 DB 사용)
		threadTS, _ := s.db.GetThreadTS(ctx, alert.Fingerprint)
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
