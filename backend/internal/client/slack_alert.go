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
func (c *SlackClient) SendAlert(alert model.Alert, status, incidentID string, isManual bool) error {
	if !c.IsConfigured() {
		return fmt.Errorf("slack bot token or channel ID not configured")
	}

	// 1. 메시지 포맷팅
	color := c.getColorByStatus(status, alert.Labels["severity"])
	emoji := c.getEmojiByStatus(status)

	var title string
	if isManual {
		title = fmt.Sprintf("🔧 [Manually Resolved] [%s] %s",
			alert.Labels["severity"],
			alert.Labels["alertname"],
		)
	} else {
		title = fmt.Sprintf("%s [%s] %s",
			emoji,
			alert.Labels["severity"],
			alert.Labels["alertname"],
		)
	}

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

// SendFlappingDetection - Flapping 감지 시 오렌지 경고 메시지 전송
func (c *SlackClient) SendFlappingDetection(alert model.Alert, incidentID string, cycleCount int) error {
	if !c.IsConfigured() {
		return fmt.Errorf("slack bot token or channel ID not configured")
	}

	// 기존 thread_ts 조회 (같은 스레드에 계속 전송)
	var threadTS string
	if ts, ok := c.GetThreadTS(alert.Fingerprint); ok {
		threadTS = ts
	}

	emoji := "⚠️"
	title := fmt.Sprintf("%s [FLAPPING DETECTED] %s", emoji, alert.Labels["alertname"])

	description := fmt.Sprintf(
		"이 알림이 30분 내에 %d회 반복(firing→resolved)되어 Flapping으로 감지되었습니다.\n"+
			"안정화될 때까지 추가 알림 및 AI 분석이 일시 중지됩니다.\n"+
			"상태는 계속 추적되며, 30분간 안정 시 자동으로 해제됩니다.",
		cycleCount,
	)

	fields := []SlackField{
		{Title: "Alert Name", Value: alert.Labels["alertname"], Short: true},
		{Title: "Namespace", Value: alert.Labels["namespace"], Short: true},
		{Title: "Severity", Value: alert.Labels["severity"], Short: true},
		{Title: "Cycle Count", Value: fmt.Sprintf("%d cycles", cycleCount), Short: true},
		{Title: "Current Status", Value: alert.Status, Short: true},
	}

	// Incident 링크 추가
	if incidentID != "" && c.frontendURL != "" {
		incidentLink := fmt.Sprintf("<%s/incidents/%s|🔍 Incident 대시보드>", c.frontendURL, incidentID)
		fields = append(fields, SlackField{Title: "Incident", Value: incidentLink, Short: false})
	}

	msg := SlackMessage{
		Channel:  c.channelID,
		ThreadTS: threadTS, // 기존 스레드에 계속 전송
		Attachments: []SlackAttachment{
			{
				Color:      "#ff9800", // Orange for flapping warning
				Title:      title,
				Text:       description,
				Fields:     fields,
				Footer:     "kube-rca",
				FooterIcon: "https://kubernetes.io/images/favicon.png",
				Ts:         time.Now().Unix(),
			},
		},
	}

	resp, err := c.send(msg)
	if err != nil {
		return err
	}

	// 새 메시지인 경우 thread_ts 저장
	if threadTS == "" && resp.TS != "" {
		c.StoreThreadTS(alert.Fingerprint, resp.TS)
	}

	return nil
}

// SendFlappingCleared - Flapping 해제 시 녹색 성공 메시지 전송
func (c *SlackClient) SendFlappingCleared(fingerprint, threadTS string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("slack bot token or channel ID not configured")
	}

	emoji := "✅"
	title := fmt.Sprintf("%s Flapping Cleared", emoji)
	description := "이 알림이 30분 이상 안정 상태를 유지하여 Flapping 상태가 해제되었습니다.\n정상 알림 모니터링이 재개됩니다."

	msg := SlackMessage{
		Channel:  c.channelID,
		ThreadTS: threadTS,
		Attachments: []SlackAttachment{
			{
				Color:      "#36a64f", // Green for cleared
				Title:      title,
				Text:       description,
				Footer:     "kube-rca",
				FooterIcon: "https://kubernetes.io/images/favicon.png",
				Ts:         time.Now().Unix(),
			},
		},
	}

	_, err := c.send(msg)
	return err
}
