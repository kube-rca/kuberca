package model

type ChatRequest struct {
	Message         string `json:"message"`
	ConversationID  string `json:"conversation_id"`
	Page            string `json:"page"`
	Auto            bool   `json:"auto"`
	IncidentID      string `json:"incident_id"`
	AlertID         string `json:"alert_id"`
	IncidentTitle   string `json:"incident_title"`
	IncidentContent string `json:"incident_content"`
	AlertTitle      string `json:"alert_title"`
	AlertContent    string `json:"alert_content"`
}

type ChatContextUsed struct {
	Incident *IncidentChatContext `json:"incident,omitempty"`
	Alert    *AlertChatContext    `json:"alert,omitempty"`
}

type IncidentChatContext struct {
	IncidentID string `json:"incident_id"`
	Title      string `json:"title"`
	Severity   string `json:"severity"`
	Status     string `json:"status"`
	Summary    string `json:"summary,omitempty"`
}

type AlertChatContext struct {
	AlertID    string `json:"alert_id"`
	IncidentID string `json:"incident_id,omitempty"`
	AlarmTitle string `json:"alarm_title"`
	Severity   string `json:"severity"`
	Status     string `json:"status"`
	Summary    string `json:"summary,omitempty"`
}

type ChatResponse struct {
	Status         string           `json:"status"`
	Answer         string           `json:"answer"`
	ConversationID string           `json:"conversation_id,omitempty"`
	ContextUsed    *ChatContextUsed `json:"context_used,omitempty"`
}
