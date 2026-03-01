// 외부 Slack API와 통신하는 클라이언트 정의
// Client 레이어에서만 사용하는 구조체 및 Slack 공통 메서드 정의
//
// 환경변수:
//   - SLACK_BOT_TOKEN: Slack Bot Token (xoxb-...)
//   - SLACK_CHANNEL_ID: Slack 채널 ID (C...)
//
// Webhook 대신 Bot Token을 사용하는 이유:
//   - thread_ts 반환: 메시지 전송 후 timestamp를 받아 쓰레드 관리 가능
//   - 스레드 답글: resolved 알림을 firing과 같은 스레드로 전송 가능
//   - 추후 AI 분석 결과도 같은 스레드로 전송 가능

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kube-rca/backend/internal/config"
)

// SlackClient(메시지 메타데이터) 구조체 정의
type SlackClient struct {
	botToken    string
	channelID   string
	frontendURL string
	httpClient  *http.Client

	// threadMap: fingerprint -> thread_ts 매핑
	//   - resolved 알림을 firing과 같은 스레드로 보내기 위함
	//   - 추후 AI 분석 결과를 같은 스레드로 보내기 위함
	// sync.Map 사용 이유: 동시성 안전 (여러 알림이 동시에 처리될 수 있음)
	threadMap sync.Map
}

var _ Notifier = (*SlackClient)(nil)
var _ ThreadRefStore = (*SlackClient)(nil)
var _ ThreadRefRequirement = (*SlackClient)(nil)

// SlackMessage(메시지 내용)) 구조체 정의
type SlackMessage struct {
	Channel     string            `json:"channel"`               // 메시지를 보낼 채널 ID
	Text        string            `json:"text,omitempty"`        // 메시지 본문
	Attachments []SlackAttachment `json:"attachments,omitempty"` // 색상, 필드
	ThreadTS    string            `json:"thread_ts,omitempty"`   // 쓰레드 메시지의 timestamp
}

// SlackAttachment(메시지 포맷) 구조체 정의
type SlackAttachment struct {
	// - critical: #dc3545 (빨강)
	// - warning: #ffc107 (노랑)
	// - resolved: #36a64f (초록)
	Color      string       `json:"color"`
	Title      string       `json:"title"`
	Text       string       `json:"text"`
	Footer     string       `json:"footer,omitempty"`
	FooterIcon string       `json:"footer_icon,omitempty"`
	Ts         int64        `json:"ts,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
}

// SlackField(메시지 포맷 필드) 구조체 정의
type SlackField struct {
	Title string `json:"title"` // 필드 제목 (예: "Namespace")
	Value string `json:"value"` // 필드 값 (예: "default")
	Short bool   `json:"short"` // true면 좁은 너비 (한 줄에 2개)
}

// SlackResponse(메시지 응답) 구조체 정의
type SlackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	TS    string `json:"ts,omitempty"`
}

// SlackClient 객체 생성
func NewSlackClient(cfg config.SlackConfig) *SlackClient {
	return &SlackClient{
		botToken:    cfg.BotToken,
		channelID:   cfg.ChannelID,
		frontendURL: strings.TrimRight(cfg.FrontendURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackClient에 Bot Token과 Channel ID가 모두 설정되어 있는지 체크
func (c *SlackClient) IsConfigured() bool {
	return c.botToken != "" && c.channelID != ""
}

// Notify는 공통 Notifier 이벤트를 Slack 메시지로 변환해 전송한다.
func (c *SlackClient) Notify(event NotifierEvent) error {
	switch e := event.(type) {
	case AlertStatusChangedEvent:
		return c.SendAlert(e.Alert, e.Alert.Status, e.IncidentID)
	case *AlertStatusChangedEvent:
		return c.SendAlert(e.Alert, e.Alert.Status, e.IncidentID)
	case FlappingDetectedEvent:
		return c.SendFlappingDetection(e.Alert, e.IncidentID, e.CycleCount)
	case *FlappingDetectedEvent:
		return c.SendFlappingDetection(e.Alert, e.IncidentID, e.CycleCount)
	case FlappingClearedEvent:
		return c.SendFlappingCleared(e.Fingerprint, e.ThreadRef)
	case *FlappingClearedEvent:
		return c.SendFlappingCleared(e.Fingerprint, e.ThreadRef)
	case AnalysisResultPostedEvent:
		return c.SendToThread(e.ThreadRef, e.Content)
	case *AnalysisResultPostedEvent:
		return c.SendToThread(e.ThreadRef, e.Content)
	case nil:
		return fmt.Errorf("unsupported notifier event: <nil>")
	default:
		return fmt.Errorf("unsupported notifier event: %T (%s)", event, event.EventType())
	}
}

// Slack API 호출
func (c *SlackClient) send(msg SlackMessage) (*SlackResponse, error) {
	// JSON 직렬화
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.botToken)

	// 요청 전송
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// JSON 파싱
	var slackResp SlackResponse
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 에러 확인
	if !slackResp.OK {
		return nil, fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return &slackResp, nil
}

// firing 알림 전송 후 thread_ts를 저장
func (c *SlackClient) StoreThreadRef(alertKey, threadRef string) {
	c.threadMap.Store(alertKey, threadRef)
}

// resolved 알림 전송 전 thread_ts를 조회
func (c *SlackClient) GetThreadRef(alertKey string) (string, bool) {
	val, ok := c.threadMap.Load(alertKey)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// resolved 알림 전송 후 thread_ts를 제거
func (c *SlackClient) DeleteThreadRef(alertKey string) {
	c.threadMap.Delete(alertKey)
}

// StoreThreadTS는 기존 호출부 호환을 위해 유지한다.
func (c *SlackClient) StoreThreadTS(fingerprint, threadTS string) {
	c.StoreThreadRef(fingerprint, threadTS)
}

// GetThreadTS는 기존 호출부 호환을 위해 유지한다.
func (c *SlackClient) GetThreadTS(fingerprint string) (string, bool) {
	return c.GetThreadRef(fingerprint)
}

// DeleteThreadTS는 기존 호출부 호환을 위해 유지한다.
func (c *SlackClient) DeleteThreadTS(fingerprint string) {
	c.DeleteThreadRef(fingerprint)
}

func (c *SlackClient) RequiresThreadRef() bool {
	return true
}

// 특정 쓰레드에 메시지 전송 (Agent 분석 결과 전송용, 일단 분리하지않음)
func (c *SlackClient) SendToThread(threadTS, text string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("slack bot token or channel ID not configured")
	}

	msg := SlackMessage{
		Channel:  c.channelID,
		ThreadTS: threadTS,
		Attachments: []SlackAttachment{
			{
				Color: "#6f42c1", // purple for AI analysis
				Title: "🤖 AI 분석 결과",
				Text:  toSlackMarkdown(text),
			},
		},
	}

	_, err := c.send(msg)
	return err
}

func toSlackMarkdown(text string) string {
	if text == "" {
		return text
	}

	var builder strings.Builder
	builder.Grow(len(text))

	inCodeBlock := false
	inInlineCode := false

	for i := 0; i < len(text); {
		if !inInlineCode && strings.HasPrefix(text[i:], "```") {
			inCodeBlock = !inCodeBlock
			builder.WriteString("```")
			i += 3
			continue
		}
		if !inCodeBlock && text[i] == '`' {
			inInlineCode = !inInlineCode
			builder.WriteByte('`')
			i++
			continue
		}
		if !inCodeBlock && !inInlineCode && (i == 0 || text[i-1] == '\n') && text[i] == '#' {
			j := i
			for j < len(text) && text[j] == '#' {
				j++
			}
			if j < len(text) && (text[j] == ' ' || text[j] == '\t') {
				k := j
				for k < len(text) && (text[k] == ' ' || text[k] == '\t') {
					k++
				}
				end := k
				for end < len(text) && text[end] != '\n' {
					end++
				}
				heading := strings.TrimSpace(text[k:end])
				if heading != "" {
					builder.WriteByte('*')
					builder.WriteString(heading)
					builder.WriteByte('*')
					if end < len(text) && text[end] == '\n' {
						builder.WriteByte('\n')
						end++
					}
					i = end
					continue
				}
			}
		}
		if !inCodeBlock && !inInlineCode && strings.HasPrefix(text[i:], "**") {
			builder.WriteByte('*')
			i += 2
			continue
		}

		builder.WriteByte(text[i])
		i++
	}

	return builder.String()
}
