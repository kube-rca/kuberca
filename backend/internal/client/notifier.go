package client

import "github.com/kube-rca/backend/internal/model"

const (
	NotifierEventAlertStatusChanged   = "alert.status_changed"
	NotifierEventFlappingDetected     = "alert.flapping_detected"
	NotifierEventFlappingCleared      = "alert.flapping_cleared"
	NotifierEventAnalysisResultPosted = "analysis.result_posted"
)

// NotifierEvent는 알림 채널(Slack, Teams 등) 전송 이벤트를 표현한다.
type NotifierEvent interface {
	EventType() string
}

// AlertStatusChangedEvent는 firing/resolved 일반 알림 이벤트다.
type AlertStatusChangedEvent struct {
	Alert      model.Alert
	IncidentID string
}

func (AlertStatusChangedEvent) EventType() string {
	return NotifierEventAlertStatusChanged
}

// FlappingDetectedEvent는 flapping 감지 알림 이벤트다.
type FlappingDetectedEvent struct {
	Alert      model.Alert
	IncidentID string
	CycleCount int
}

func (FlappingDetectedEvent) EventType() string {
	return NotifierEventFlappingDetected
}

// FlappingClearedEvent는 flapping 해제 알림 이벤트다.
type FlappingClearedEvent struct {
	Fingerprint string
	ThreadRef   string
}

func (FlappingClearedEvent) EventType() string {
	return NotifierEventFlappingCleared
}

// AnalysisResultPostedEvent는 분석 결과 스레드 메시지 이벤트다.
type AnalysisResultPostedEvent struct {
	ThreadRef string
	Content   string
}

func (AnalysisResultPostedEvent) EventType() string {
	return NotifierEventAnalysisResultPosted
}

// Notifier는 플랫폼별 알림 전송 구현의 공통 인터페이스다.
type Notifier interface {
	Notify(event NotifierEvent) error
}

// ThreadRefStore는 thread/reply 개념이 있는 플랫폼에서만 구현한다.
// thread 개념이 없는 플랫폼은 이 인터페이스를 구현하지 않아도 된다.
type ThreadRefStore interface {
	StoreThreadRef(alertKey, threadRef string)
	GetThreadRef(alertKey string) (string, bool)
	DeleteThreadRef(alertKey string)
}

// ThreadRefRequirement는 thread_ref가 필수인지 동적으로 판단할 수 있는 capability다.
// 구현하지 않으면 ThreadRefStore 구현 여부를 기준으로 판단한다.
type ThreadRefRequirement interface {
	RequiresThreadRef() bool
}

// ThreadAwareNotifier는 thread capability를 함께 제공하는 notifier다.
type ThreadAwareNotifier interface {
	Notifier
	ThreadRefStore
	ThreadRefRequirement
}
