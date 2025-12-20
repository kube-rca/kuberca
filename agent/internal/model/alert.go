package model

import "time"

type AlertmanagerWebhook struct {
	Version         string            `json:"version"`
	GroupKey        string            `json:"groupKey"`
	TruncatedAlerts int               `json:"truncatedAlerts"`
	Status          string            `json:"status"`
	Receiver        string            `json:"receiver"`
	GroupLabels     map[string]string `json:"groupLabels"`
	CommonLabels    map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL     string            `json:"externalURL"`
	Alerts          []Alert           `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}
