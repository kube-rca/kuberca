// Alertmanager 웹훅 요청을 처리하는 핸들러
//
// 요청 흐름:
//  1. Alertmanager가 POST /webhook/alertmanager로 알림 전송
//  2. JSON 페이로드를 AlertmanagerWebhook 구조체로 파싱
//  3. 각 알림을 로깅하고 service 레이어로 전달

package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/logutil"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
	"log"
	"net/http"
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

// Webhook godoc
// @Summary Receive Alertmanager webhook
// @Tags webhook
// @Accept json
// @Produce json
// @Param payload body model.AlertmanagerWebhook true "Alertmanager webhook payload"
// @Success 200 {object} model.AlertWebhookResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /webhook/alertmanager [post]
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
	log.Printf("Raw webhook payload:\n%s", logutil.Sanitize(string(rawPayload)))

	// 3. 웹훅 메타데이터 로깅
	// status: firing(발생) 또는 resolved(해결)
	// alertCount: 웹훅에 포함된 알림 개수
	// receiver: Alertmanager에서 설정한 receiver 이름
	log.Printf("Received alert webhook: status=%s, alertCount=%d, receiver=%s",
		logutil.Sanitize(webhook.Status), len(webhook.Alerts), logutil.Sanitize(webhook.Receiver))

	// 4. 서비스 레이어에서 Slack 전송 처리
	// 비즈니스 로직(필터링)은 service 레이어에 위임
	sent, failed := h.alertService.ProcessWebhook(webhook)

	// 5. 응답 반환
	c.JSON(http.StatusOK, model.AlertWebhookResponse{
		Status:      "received",          // 수신 상태
		AlertCount:  len(webhook.Alerts), // 수신한 알림 수
		SlackSent:   sent,                // Slack 전송 성공 수
		SlackFailed: failed,              // Slack 전송 실패 수
	})
}
