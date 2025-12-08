// Alertmanager 웹훅 페이로드 및 개별 알림 구조체를 정의
// handler, service, client 레이어에서 공통으로 사용하기 때문에 model 레이어에 별도로 정의

package model

import "time"

// AlertmanagerWebhook - Alertmanager 웹훅 페이로드
// 여러 개의 알림이 그룹으로 묶여서 전송 가능
type AlertmanagerWebhook struct {
	Version string `json:"version"`

	// 동일한 GroupKey를 가진 알림들은 함께 그룹핑됨
	GroupKey string `json:"groupKey"`

	// max_alerts 설정으로 인해 생략된 알림이 있을 경우 그 개수
	TruncatedAlerts int    `json:"truncatedAlerts"`
	Status          string `json:"status"`
	Receiver        string `json:"receiver"`

	// route.group_by 설정에 따라 결정되는 그룹핑에 사용된 라벨
	GroupLabels map[string]string `json:"groupLabels"`

	// 그룹 내 모든 알림에 공통으로 존재하는 라벨
	CommonLabels map[string]string `json:"commonLabels"`

	// 그룹 내 모든 알림에 공통으로 존재하는 어노테이션
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`

	// 개별 알림 리스트
	Alerts []Alert `json:"alerts"`
}

// Alert - 개별 알림
// 각 Alert은 고유한 Fingerprint를 가지며, 이를 통해 동일한 알림을 식별
type Alert struct {
	Status string `json:"status"`

	// - alertname: 알림 이름 (예: "PodCrashLooping", "HighMemoryUsage")
	// - severity: 심각도 (예: "critical", "warning", "info")
	// - namespace: 문제 발생 네임스페이스
	// - pod: 문제 발생 파드 이름
	Labels map[string]string `json:"labels"`

	// - summary: 알림 요약
	// - description: 알림 상세 설명
	// - runbook_url: 대응 매뉴얼 URL
	Annotations map[string]string `json:"annotations"`

	// StartsAt: 알림 발생 시각 (UTC)
	StartsAt time.Time `json:"startsAt"`

	// EndsAt: 알림 종료 시각 (UTC)
	// resolved 상태일 때만 유효한 값 설정
	// firing 상태일 때는 "0001-01-01T00:00:00Z"
	EndsAt time.Time `json:"endsAt"`

	// GeneratorURL: 알림을 생성한 Prometheus 쿼리 URL
	GeneratorURL string `json:"generatorURL"`

	// Fingerprint: 알림 고유 식별자 (Labels의 조합으로 생성되는 해시값)
	// thread_ts 매핑에 사용하여 같은 스레드로 메시지 전송이 가능하게함
	Fingerprint string `json:"fingerprint"`
}
