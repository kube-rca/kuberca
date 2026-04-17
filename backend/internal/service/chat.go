package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
)

var ErrInvalidChatRequest = errors.New("invalid chat request")

type ChatService struct {
	repo        *db.Postgres
	agentClient *client.AgentClient
}

func NewChatService(repo *db.Postgres, agentClient *client.AgentClient) *ChatService {
	return &ChatService{
		repo:        repo,
		agentClient: agentClient,
	}
}

func (s *ChatService) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	req.Message = strings.TrimSpace(req.Message)
	req.Language = normalizePreferredLanguage(req.Language)
	req.Page = strings.TrimSpace(req.Page)
	req.IncidentID = strings.TrimSpace(req.IncidentID)
	req.AlertID = strings.TrimSpace(req.AlertID)
	req.IncidentTitle = strings.TrimSpace(req.IncidentTitle)
	req.IncidentContent = strings.TrimSpace(req.IncidentContent)
	req.AlertTitle = strings.TrimSpace(req.AlertTitle)
	req.AlertContent = strings.TrimSpace(req.AlertContent)

	if req.Message == "" && !req.Auto {
		return nil, fmt.Errorf("%w: message is required when auto=false", ErrInvalidChatRequest)
	}

	contextUsed := &model.ChatContextUsed{}
	agentContext := map[string]any{}

	if req.Page != "" {
		agentContext["page"] = req.Page
	}
	agentContext["language"] = req.Language

	if req.IncidentID != "" {
		incident, err := s.repo.GetIncidentDetailInsensitive(req.IncidentID)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to load incident_id=%s", ErrInvalidChatRequest, req.IncidentID)
		}
		summary := localizedValue(incident.AnalysisSummaryI18n, strValue(incident.AnalysisSummary), req.Language)
		detail := localizedValue(incident.AnalysisDetailI18n, strValue(incident.AnalysisDetail), req.Language)
		contextUsed.Incident = &model.IncidentChatContext{
			IncidentID: incident.IncidentID,
			Title:      incident.Title,
			Severity:   incident.Severity,
			Status:     incident.Status,
			Summary:    summary,
		}
		agentContext["incident"] = map[string]any{
			"incident_id":      incident.IncidentID,
			"title":            incident.Title,
			"severity":         incident.Severity,
			"status":           incident.Status,
			"analysis_summary": summary,
			"analysis_detail":  detail,
		}
	}

	if req.AlertID != "" {
		alert, err := s.repo.GetAlertDetailInsensitive(req.AlertID)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to load alert_id=%s", ErrInvalidChatRequest, req.AlertID)
		}
		summary := localizedValue(alert.AnalysisSummaryI18n, strValue(alert.AnalysisSummary), req.Language)
		detail := localizedValue(alert.AnalysisDetailI18n, strValue(alert.AnalysisDetail), req.Language)
		incidentID := strValue(alert.IncidentID)
		contextUsed.Alert = &model.AlertChatContext{
			AlertID:    alert.AlertID,
			IncidentID: incidentID,
			AlarmTitle: alert.AlarmTitle,
			Severity:   alert.Severity,
			Status:     alert.Status,
			Summary:    summary,
		}
		agentContext["alert"] = map[string]any{
			"alert_id":         alert.AlertID,
			"incident_id":      incidentID,
			"alarm_title":      alert.AlarmTitle,
			"severity":         alert.Severity,
			"status":           alert.Status,
			"analysis_summary": summary,
			"analysis_detail":  detail,
			"labels":           alert.Labels,
			"annotations":      alert.Annotations,
		}
	}

	if req.IncidentTitle != "" || req.IncidentContent != "" || req.AlertTitle != "" || req.AlertContent != "" {
		agentContext["user_input_context"] = map[string]any{
			"incident_title":   req.IncidentTitle,
			"incident_content": req.IncidentContent,
			"alert_title":      req.AlertTitle,
			"alert_content":    req.AlertContent,
		}
	}

	userMessage := req.Message
	if userMessage == "" && req.Auto {
		userMessage = s.defaultAutoMessage(req)
	}

	agentResp, err := s.agentClient.RequestChat(ctx, client.AgentChatRequest{
		Message:        userMessage,
		ConversationID: req.ConversationID,
		Language:       req.Language,
		Context:        agentContext,
		Metadata: map[string]any{
			"source":   "frontend-chat",
			"language": req.Language,
		},
	})
	if err != nil {
		return nil, err
	}

	answer := firstNonEmpty(agentResp.Answer, agentResp.Message, agentResp.Response, agentResp.OutputText, agentResp.Analysis)
	if answer == "" {
		return nil, fmt.Errorf("agent returned empty answer")
	}

	resp := &model.ChatResponse{
		Status:         "success",
		Answer:         answer,
		ConversationID: firstNonEmpty(agentResp.ConversationID, req.ConversationID),
	}
	if contextUsed.Incident != nil || contextUsed.Alert != nil {
		resp.ContextUsed = contextUsed
	}
	return resp, nil
}

func (s *ChatService) defaultAutoMessage(req model.ChatRequest) string {
	if req.Language == "en" {
		switch {
		case req.IncidentID != "":
			return "Summarize this incident's current status, likely cause, and next actions."
		case req.AlertID != "":
			return "Explain this alert's current status, likely cause, and what should be checked next."
		default:
			return "Summarize the key situation based on the current context."
		}
	}

	switch {
	case req.IncidentID != "":
		return "이 incident의 현재 상태, 원인 추정, 다음 액션을 요약해줘."
	case req.AlertID != "":
		return "이 alert의 현재 상태와 원인 추정, 확인해야 할 항목을 알려줘."
	default:
		return "현재 제공된 컨텍스트를 기준으로 핵심 상황을 요약해줘."
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func strValue[T ~string | *string](value T) string {
	switch v := any(value).(type) {
	case string:
		return strings.TrimSpace(v)
	case *string:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(*v)
	default:
		return ""
	}
}

func localizedValue(values model.LocalizedText, fallback, language string) string {
	language = normalizePreferredLanguage(language)
	if values != nil {
		if value := strings.TrimSpace(values[language]); value != "" {
			return value
		}
		if value := strings.TrimSpace(values["ko"]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(fallback)
}
