package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
)

// Alertmanager 웹훅 엔드포인트
func AlertmanagerWebhook(c *gin.Context) {
	var webhook model.AlertmanagerWebhook

	// Alertmanager가 보낸 페이로드를 AlertmanagerWebhook 구조체로 변환
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// JSON 메타데이터 로깅
	// status: firing(발생) 또는 resolved(해결)
	// alertCount: 이번 웹훅에 포함된 알림 개수
	// receiver: Alertmanager에서 설정한 receiver 이름
	log.Printf("Received alert webhook: status=%s, alertCount=%d, receiver=%s",
		webhook.Status, len(webhook.Alerts), webhook.Receiver)

	// 그룹핑된 개별 알림 및 로깅
	for _, alert := range webhook.Alerts {
		log.Printf("  Alert: name=%s, severity=%s, status=%s, namespace=%s",
			alert.Labels["alertname"], // alertname: 알림 이름 (예: PodCrashLooping)
			alert.Labels["severity"],  // severity: 심각도 (critical, warning, info)
			alert.Status,
			alert.Labels["namespace"], // namespace: 문제 발생 네임스페이스
		)
		log.Printf("    Description: %s", alert.Annotations["description"]) // description: 알림 내용 설명
		log.Printf("    StartsAt: %s", alert.StartsAt.Format(time.RFC3339)) // StartsAt: 알림 발생 시각 (UTC)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "received",
		"alertCount": len(webhook.Alerts),
	})
}
