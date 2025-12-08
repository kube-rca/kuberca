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
	"os"
	"sync"
	"time"
)

// SlackClient(메시지 메타데이터) 구조체 정의
type SlackClient struct {
	botToken   string
	channelID  string
	httpClient *http.Client

	// threadMap: fingerprint -> thread_ts 매핑
	//   - resolved 알림을 firing과 같은 스레드로 보내기 위함
	//   - 추후 AI 분석 결과를 같은 스레드로 보내기 위함
	// sync.Map 사용 이유: 동시성 안전 (여러 알림이 동시에 처리될 수 있음)
	threadMap sync.Map
}

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
func NewSlackClient() *SlackClient {
	return &SlackClient{
		botToken:  os.Getenv("SLACK_BOT_TOKEN"),
		channelID: os.Getenv("SLACK_CHANNEL_ID"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackClient에 Bot Token과 Channel ID가 모두 설정되어 있는지 체크
func (c *SlackClient) IsConfigured() bool {
	return c.botToken != "" && c.channelID != ""
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
func (c *SlackClient) StoreThreadTS(fingerprint, threadTS string) {
	c.threadMap.Store(fingerprint, threadTS)
}

// resolved 알림 전송 전 thread_ts를 조회
func (c *SlackClient) GetThreadTS(fingerprint string) (string, bool) {
	val, ok := c.threadMap.Load(fingerprint)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// resolved 알림 전송 후 thread_ts를 제거
func (c *SlackClient) DeleteThreadTS(fingerprint string) {
	c.threadMap.Delete(fingerprint)
}
