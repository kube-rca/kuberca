package model

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type PingResponse struct {
	Message string `json:"message"`
}

type RootResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AuthLogoutResponse struct {
	Status string `json:"status"`
}

type AuthMeResponse struct {
	UserID  int64  `json:"userId"`
	LoginID string `json:"loginId"`
}

type AlertWebhookResponse struct {
	Status      string `json:"status"`
	AlertCount  int    `json:"alertCount"`
	SlackSent   int    `json:"slackSent"`
	SlackFailed int    `json:"slackFailed"`
}
