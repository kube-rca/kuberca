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
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

// AlertService 구조체 정의
type AlertService struct {
	slackClient    *client.SlackClient
	agentService   *AgentService
	db             *db.Postgres
	flappingConfig config.FlappingConfig
}

// AlertService 객체 생성
func NewAlertService(slackClient *client.SlackClient, agentService *AgentService, database *db.Postgres, flappingConfig config.FlappingConfig) *AlertService {
	return &AlertService{
		slackClient:    slackClient,
		agentService:   agentService,
		db:             database,
		flappingConfig: flappingConfig,
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

		// 2.5. Flapping 감지 (resolved 상태 업데이트 전에 수행)
		isFlapping, isNewFlapping := s.detectFlapping(alert)

		// 3. resolved 상태면 중복 체크 후 resolved_at 업데이트
		if alert.Status == "resolved" {
			// 이미 resolved된 알림인지 확인 (중복 웹훅 방지)
			if alreadyResolved, _ := s.db.IsAlertAlreadyResolved(alert.Fingerprint, alert.EndsAt); alreadyResolved {
				log.Printf("Skipping duplicate resolved alert (alert_id=%s)", alert.Fingerprint)
				continue
			}
			if err := s.db.UpdateAlertResolved(alert.Fingerprint, alert.EndsAt); err != nil {
				log.Printf("Failed to update alert resolved status: %v", err)
			}

			// Flapping 상태라면 clearance 체크 스케줄링
			if isFlapping {
				go s.scheduleFlappingClearanceCheck(alert.Fingerprint, alert.EndsAt)
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

		// 6. Slack 전송 - Flapping 상태에 따라 다르게 처리
		if isNewFlapping {
			// 새로 Flapping 감지됨 - 오렌지 경고 메시지 전송
			err = s.slackClient.SendFlappingDetection(alert, incidentID, s.getFlappingCycleCount(alert.Fingerprint))
		} else if isFlapping {
			// 이미 Flapping 중 - Slack 알림 스킵
			log.Printf("Skipping Slack notification for flapping alert (alert_id=%s)", alert.Fingerprint)
			sent++
			goto skipSlack
		} else {
			// 정상 Alert 전송
			err = s.slackClient.SendAlert(alert, alert.Status, incidentID)
		}

		if err != nil {
			log.Printf("Failed to send alert to Slack: %v", err)
			failed++
			continue
		}

		log.Printf("Sent alert to Slack (alert_id=%s, status=%s, incident_id=%s, flapping=%v)", alert.Fingerprint, alert.Status, incidentID, isFlapping)
		sent++

	skipSlack:
		// 7. thread_ts를 DB에 저장 (firing 알림일 때)
		if alert.Status == "firing" {
			if threadTS, ok := s.slackClient.GetThreadTS(alert.Fingerprint); ok {
				if err := s.db.UpdateAlertThreadTS(alert.Fingerprint, threadTS); err != nil {
					log.Printf("Failed to save thread_ts to DB: %v", err)
				}
			}
		}

		// 8. Agent에 비동기 분석 요청 - Flapping 중이면 스킵
		if !isFlapping {
			threadTS, _ := s.db.GetAlertThreadTS(alert.Fingerprint)
			go s.agentService.RequestAnalysis(alert, threadTS, incidentID)
		} else {
			log.Printf("Skipping Agent analysis for flapping alert (alert_id=%s)", alert.Fingerprint)
		}
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

// detectFlapping - Alert flapping 감지
// Returns: (isFlapping, isNewFlapping)
func (s *AlertService) detectFlapping(alert model.Alert) (bool, bool) {
	// Flapping 기능이 비활성화된 경우 스킵
	if !s.flappingConfig.Enabled {
		return false, false
	}

	// 현재 Alert 상태 조회
	currentStatus, err := s.db.GetAlertCurrentStatus(alert.Fingerprint)
	if err != nil {
		// Alert가 DB에 없는 경우 (첫 전환)
		return false, false
	}

	// 상태 변경이 없으면 전환 아님
	if currentStatus == alert.Status {
		return s.db.IsAlertFlapping(alert.Fingerprint), false
	}

	// 상태 전환 기록
	timestamp := alert.StartsAt
	if alert.Status == "resolved" {
		timestamp = alert.EndsAt
	}

	if err := s.db.RecordStateTransition(alert.Fingerprint, currentStatus, alert.Status, timestamp); err != nil {
		log.Printf("Failed to record state transition: %v", err)
		return false, false
	}

	// resolved 전환(사이클 완료)에만 flapping 감지
	if alert.Status != "resolved" {
		return s.db.IsAlertFlapping(alert.Fingerprint), false
	}

	// Flapping 윈도우 및 threshold 조회
	windowMinutes := s.getFlappingWindowMinutes()
	threshold := s.getFlappingThreshold()

	// 윈도우 내 사이클 수 계산
	cycleCount, windowStart, err := s.db.CountFlappingCycles(alert.Fingerprint, windowMinutes)
	if err != nil {
		log.Printf("Failed to count flapping cycles: %v", err)
		return false, false
	}

	// Flapping 상태 확인
	wasFlapping := s.db.IsAlertFlapping(alert.Fingerprint)
	isNowFlapping := cycleCount >= threshold

	if isNowFlapping && !wasFlapping {
		// 새로 Flapping 감지됨
		if err := s.db.MarkAlertAsFlapping(alert.Fingerprint, true, cycleCount, windowStart); err != nil {
			log.Printf("Failed to mark alert as flapping: %v", err)
		}
		log.Printf("Flapping detected for alert %s (cycles=%d, threshold=%d)", alert.Fingerprint, cycleCount, threshold)
		return true, true
	} else if isNowFlapping {
		// 이미 Flapping 중, cycle count 업데이트
		s.db.UpdateFlappingCycleCount(alert.Fingerprint, cycleCount)
		return true, false
	}

	return false, false
}

// scheduleFlappingClearanceCheck - Flapping 해제 체크 스케줄링
func (s *AlertService) scheduleFlappingClearanceCheck(fingerprint string, resolvedAt time.Time) {
	clearanceMinutes := s.getFlappingClearanceMinutes()
	checkTime := resolvedAt.Add(time.Duration(clearanceMinutes) * time.Minute)

	// Clearance 시간까지 대기
	time.Sleep(time.Until(checkTime))

	// Alert 상태 재확인
	alert, err := s.db.GetAlertDetail(fingerprint)
	if err != nil {
		log.Printf("Failed to get alert for flapping clearance check: %v", err)
		return
	}

	// 이미 Flapping이 아니면 종료
	if !alert.IsFlapping {
		return
	}

	// Alert가 다시 firing되었거나 resolved 시각이 변경되었으면 clearance 취소
	if alert.Status == "firing" || (alert.ResolvedAt != nil && alert.ResolvedAt.Before(resolvedAt)) {
		log.Printf("Alert re-fired, not clearing flapping status (alert_id=%s)", fingerprint)
		return
	}

	// resolvedAt 이후 새로운 전환이 있었는지 확인
	hasNewTransitions, err := s.db.HasTransitionsSince(fingerprint, resolvedAt)
	if err != nil || hasNewTransitions {
		log.Printf("New transitions detected, not clearing flapping status (alert_id=%s)", fingerprint)
		return
	}

	// Flapping 해제
	log.Printf("Clearing flapping status after %d min stability (alert_id=%s)", clearanceMinutes, fingerprint)
	if err := s.db.MarkAlertAsFlapping(fingerprint, false, 0, time.Time{}); err != nil {
		log.Printf("Failed to clear flapping status: %v", err)
		return
	}

	// Slack에 Flapping 해제 메시지 전송
	if threadTS, ok := s.db.GetAlertThreadTS(fingerprint); ok {
		s.slackClient.SendFlappingCleared(fingerprint, threadTS)
	}
}

// Configuration helper methods
func (s *AlertService) getFlappingWindowMinutes() int {
	return s.flappingConfig.DetectionWindowMinutes
}

func (s *AlertService) getFlappingThreshold() int {
	return s.flappingConfig.CycleThreshold
}

func (s *AlertService) getFlappingClearanceMinutes() int {
	return s.flappingConfig.ClearanceWindowMinutes
}

func (s *AlertService) getFlappingCycleCount(fingerprint string) int {
	count, _, _ := s.db.CountFlappingCycles(fingerprint, s.getFlappingWindowMinutes())
	return count
}
