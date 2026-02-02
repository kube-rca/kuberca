// Slack Alert 메시지 관련 메서드 정의

package client

import (
	"fmt"
	"time"

	"github.com/kube-rca/backend/internal/model"
)

// 알림을 Slack으로 전송
//
// firing 알림과 resolved 알림을 다르게 처리:
//   - firing: 새 메시지 전송 후 thread_ts 저장
//   - resolved: 기존 쓰레드에 답글로 전송 후 thread_ts 삭제
func (c *SlackClient) SendAlert(alert model.Alert, status, incidentID string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("slack bot token or channel ID not configured")
	}

	// 1. 메시지 포맷팅
	color := c.getColorByStatus(status, alert.Labels["severity"])
	emoji := c.getEmojiByStatus(status)

	title := fmt.Sprintf("%s [%s] %s",
		emoji,
		alert.Labels["severity"],
		alert.Labels["alertname"],
	)

	fields := []SlackField{
		{Title: "Namespace", Value: alert.Labels["namespace"], Short: true},
		{Title: "Severity", Value: alert.Labels["severity"], Short: true},
		{Title: "Status", Value: status, Short: true},
		{Title: "Started", Value: alert.StartsAt.Format(time.RFC3339), Short: true},
	}

	// Incident 페이지 링크 추가
	if incidentID != "" && c.frontendURL != "" {
		incidentLink := fmt.Sprintf("<%s/incidents/%s|🔍 Incident 대시보드 보러가기>", c.frontendURL, incidentID)
		fields = append(fields, SlackField{Title: "Incident", Value: incidentLink, Short: false})
	}

	msg := SlackMessage{
		Channel: c.channelID,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Text:       alert.Annotations["description"],
				Fields:     fields,
				Footer:     "kube-rca",
				FooterIcon: "https://kubernetes.io/images/favicon.png",
				Ts:         time.Now().Unix(),
			},
		},
	}

	// 2. resolved 알림: 기존 쓰레드로 전송
	// fingerprint로 저장된 thread_ts를 조회하여 해당 쓰레드로 전송
	if status == "resolved" {
		if threadTS, ok := c.GetThreadTS(alert.Fingerprint); ok {
			msg.ThreadTS = threadTS
		}
	}

	// 3. Slack API 호출
	resp, err := c.send(msg)
	if err != nil {
		return err
	}

	// 4. thread_ts 저장
	if status == "firing" && resp.TS != "" {
		c.StoreThreadTS(alert.Fingerprint, resp.TS)
	}
	// resolved: thread_ts 삭제 (메모리 정리)
	if status == "resolved" {
		c.DeleteThreadTS(alert.Fingerprint)
	}
	return nil
}

// Status에 따른 적절한 메시지 색상 반환
func (c *SlackClient) getColorByStatus(status, severity string) string {
	if status == "resolved" {
		return "#36a64f" // green
	}
	switch severity {
	case "critical":
		return "#dc3545" // red
	case "warning":
		return "#ffc107" // yellow
	default:
		return "#17a2b8" // blue
	}
}

// Status에 따른 적절한 메시지 이모지 반환
func (c *SlackClient) getEmojiByStatus(status string) string {
	if status == "resolved" {
		return "✅"
	}
	return "🔥"
}
