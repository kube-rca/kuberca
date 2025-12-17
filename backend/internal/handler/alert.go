// Alertmanager 웹훅 요청을 처리하는 핸들러
//
// 요청 흐름:
//  1. Alertmanager가 POST /webhook/alertmanager로 알림 전송
//  2. JSON 페이로드를 AlertmanagerWebhook 구조체로 파싱
//  3. 각 알림을 로깅하고 service 레이어로 전달

package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

// Alert 핸들러 구조체 정의
type AlertHandler struct {
	alertService *service.AlertService
}

// Alert 핸들러 객체 생성
func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

func (h *AlertHandler) Webhook(c *gin.Context) {
	var webhook model.AlertmanagerWebhook

	// 1. JSON 페이로드 파싱
	// Alertmanager가 보낸 페이로드를 AlertmanagerWebhook 구조체로 변환
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// 2. Raw payload 로깅 (디버깅용)
	rawPayload, _ := json.MarshalIndent(webhook, "", "  ")
	log.Printf("Raw webhook payload:\n%s", string(rawPayload))

	// 3. 웹훅 메타데이터 로깅
	// status: firing(발생) 또는 resolved(해결)
	// alertCount: 웹훅에 포함된 알림 개수
	// receiver: Alertmanager에서 설정한 receiver 이름
	log.Printf("Received alert webhook: status=%s, alertCount=%d, receiver=%s",
		webhook.Status, len(webhook.Alerts), webhook.Receiver)

	// 3. 개별 알림 로깅
	// 여러 알림을 그룹으로 묶어서 전송 가능
	for _, alert := range webhook.Alerts {
		log.Printf("  Alert: name=%s, severity=%s, status=%s, namespace=%s",
			alert.Labels["alertname"], // alertname: 알림 이름 (예: PodCrashLooping)
			alert.Labels["severity"],  // severity: 심각도 (critical, warning, info)
			alert.Status,
			alert.Labels["namespace"], // namespace: 문제 발생 네임스페이스
		)
		log.Printf("    Description: %s", alert.Annotations["description"]) // description: 알림 내용 설명
		log.Printf("    StartsAt: %s", alert.StartsAt.Format(time.RFC3339)) // StartsAt: 알림 발생 시각 (UTC)
		log.Printf("    Fingerprint: %s", alert.Fingerprint)                // Fingerprint: 알림 고유 식별자 (thread_ts 매핑용)
	}

	// 4. 서비스 레이어에서 Slack 전송 처리
	// 비즈니스 로직(필터링)은 service 레이어에 위임
	sent, failed := h.alertService.ProcessWebhook(webhook)

	// 5. 응답 반환
	c.JSON(http.StatusOK, gin.H{
		"status":      "received",          // 수신 상태
		"alertCount":  len(webhook.Alerts), // 수신한 알림 수
		"slackSent":   sent,                // Slack 전송 성공 수
		"slackFailed": failed,              // Slack 전송 실패 수
	})
}
