// Alert 처리 비즈니스 로직 정의
// handler에서 받은 알림을 필터링하고 client를 통해 Slack으로 전송
//
// 처리 흐름:
//  1. 현재 firing 상태인 Incident가 있는지 확인
//     - 없으면: 새 Incident 생성
//  2. Alert를 DB에 저장 (alerts 테이블) + Incident 연결
//  3. resolved 상태면 resolved_at 업데이트
//  4. shouldSendToSlack으로 필터링
//  5. resolved 알림: DB에서 thread_ts 조회하여 메모리에 복원
//  6. SlackClient.SendAlert로 Slack 전송
//  7. firing 알림: thread_ts를 DB에 저장
//  8. Agent에 비동기 분석 요청 (firing, resolved)
//  9. 전송 성공/실패 카운트 반환

package service

import (
	"log"

	"github.com/jackc/pgx/v5"
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
	for _, alert := range webhook.Alerts {
		// 0. severity 필터링 (info, none 등은 DB 저장도 하지 않음)
		if !s.shouldProcess(alert) {
			log.Printf("Skipping alert with severity=%s (alert_id=%s)", alert.Labels["severity"], alert.Fingerprint)
			continue
		}

		// 1. 현재 firing 상태인 Incident 확인/생성
		incidentID, err := s.getOrCreateIncident(alert)
		if err != nil {
			log.Printf("Failed to get or create incident: %v", err)
			// Incident 처리 실패해도 Alert 저장 및 Slack 전송은 계속 진행
			incidentID = ""
		}

		// 2. Alert를 DB에 저장 (alerts 테이블)
		if err := s.db.SaveAlert(alert, incidentID); err != nil {
			log.Printf("Failed to save alert to DB: %v", err)
			// DB 저장 실패해도 Slack 전송은 계속 진행
		}

		// 3. resolved 상태면 중복 체크 후 resolved_at 업데이트
		if alert.Status == "resolved" {
			// 이미 resolved된 알림인지 확인 (중복 웹훅 방지)
			if alreadyResolved, _ := s.db.IsAlertAlreadyResolved(alert.Fingerprint); alreadyResolved {
				log.Printf("Skipping duplicate resolved alert (alert_id=%s)", alert.Fingerprint)
				continue
			}
			if err := s.db.UpdateAlertResolved(alert.Fingerprint, alert.EndsAt); err != nil {
				log.Printf("Failed to update alert resolved status: %v", err)
			}
		}

		// 4. 필터링: Slack으로 전송할 알림인지 확인
		if !s.shouldSendToSlack(alert) {
			continue
		}

		// 5. resolved 알림: DB에서 thread_ts 조회하여 메모리에 복원
		// (백엔드 재시작 시 메모리가 초기화되므로 DB에서 복원 필요)
		if alert.Status == "resolved" {
			if threadTS, ok := s.db.GetAlertThreadTS(alert.Fingerprint); ok {
				s.slackClient.StoreThreadTS(alert.Fingerprint, threadTS)
			}
		}

		// 6. Client 레이어로 해당 알림을 전달
		err = s.slackClient.SendAlert(alert, alert.Status)
		if err != nil {
			log.Printf("Failed to send alert to Slack: %v", err)
			failed++
			continue
		}

		log.Printf("Sent alert to Slack (alert_id=%s, status=%s, incident_id=%s)", alert.Fingerprint, alert.Status, incidentID)
		sent++

		// 7. thread_ts를 DB에 저장 (firing 알림일 때)
		if alert.Status == "firing" {
			if threadTS, ok := s.slackClient.GetThreadTS(alert.Fingerprint); ok {
				if err := s.db.UpdateAlertThreadTS(alert.Fingerprint, threadTS); err != nil {
					log.Printf("Failed to save thread_ts to DB: %v", err)
				}
			}
		}

		// 8. Agent에 비동기 분석 요청 (firing, resolved)
		// DB에서 thread_ts 조회 (메모리 대신 DB 사용)
		threadTS, _ := s.db.GetAlertThreadTS(alert.Fingerprint)
		go s.agentService.RequestAnalysis(alert, threadTS, incidentID)
	}
	return sent, failed
}

// getOrCreateIncident - 현재 firing 상태인 Incident를 조회하거나 새로 생성
func (s *AlertService) getOrCreateIncident(alert model.Alert) (string, error) {
	// firing 상태인 Incident 조회
	incident, err := s.db.GetFiringIncident()
	if err == nil && incident != nil {
		// 기존 Incident에 연결 + severity 업데이트
		severity := alert.Labels["severity"]
		if severity != "" {
			_ = s.db.UpdateIncidentSeverity(incident.IncidentID, severity)
		}
		return incident.IncidentID, nil
	}

	// firing Incident가 없으면 새로 생성 (pgx.ErrNoRows인 경우)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	// severity := alert.Labels["severity"]
	// if severity == "" {
	// 	severity = "warning"
	// }

	severity := "TBD"

	// 초기 title은 Ongoing으로 설정 (에이전트 분석 후 title 업데이트)
	incidentID, err := s.db.CreateIncident("Ongoing", severity, alert.StartsAt)
	if err != nil {
		return "", err
	}

	log.Printf("Created new incident: %s (triggered by alert: %s)", incidentID, alert.Fingerprint)
	return incidentID, nil
}

// shouldProcess - DB 저장 및 처리 여부 결정 (info, none 등은 완전 무시)
func (s *AlertService) shouldProcess(alert model.Alert) bool {
	severity := alert.Labels["severity"]
	return severity == "warning" || severity == "critical"
}

// 필터링 로직 예시:
//   - severity가 warning 이상만 전송
//   - 특정 namespace 제외 (예: kube-system)
//   - 특정 alertname만 전송
//
// Returns:
//   - bool: true면 Slack으로 전송, false면 무시
func (s *AlertService) shouldSendToSlack(alert model.Alert) bool {
	// warning, critical만 전송 (info, none 등 필터링)
	severity := alert.Labels["severity"]
	if severity == "warning" || severity == "critical" {
		return true
	}
	return false
}
