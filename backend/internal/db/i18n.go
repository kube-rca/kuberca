package db

import (
	"encoding/json"
	"strings"

	"github.com/kube-rca/backend/internal/model"
)

func toLocalizedText(raw []byte, fallback string) model.LocalizedText {
	text := model.LocalizedText{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &text)
	}
	if strings.TrimSpace(fallback) != "" {
		if strings.TrimSpace(text["ko"]) == "" {
			text["ko"] = fallback
		}
	}
	if len(text) == 0 {
		return nil
	}
	return text
}

func localizedJSON(text model.LocalizedText, fallback string) []byte {
	if len(text) == 0 && strings.TrimSpace(fallback) != "" {
		text = model.LocalizedText{"ko": fallback}
	}
	if len(text) == 0 {
		return []byte("{}")
	}
	payload, err := json.Marshal(text)
	if err != nil {
		return []byte("{}")
	}
	return payload
}
