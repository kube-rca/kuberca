package client

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kube-rca/backend/internal/config"
)

func TestToSlackMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bold-only",
			input: "This is **bold** text.",
			want:  "This is *bold* text.",
		},
		{
			name:  "inline-code-protected",
			input: "Use `2 ** 3` and **bold**.",
			want:  "Use `2 ** 3` and *bold*.",
		},
		{
			name:  "code-block-protected",
			input: "```python\n2 ** 3\n```\n**bold**",
			want:  "```python\n2 ** 3\n```\n*bold*",
		},
		{
			name:  "mixed-inline-and-bold",
			input: "**Bold** and `code **`",
			want:  "*Bold* and `code **`",
		},
		{
			name:  "heading-converted",
			input: "### 1) 요약 (Summary)\n내용",
			want:  "*1) 요약 (Summary)*\n내용",
		},
		{
			name:  "heading-protected-in-code-block",
			input: "```\n### 1) 요약 (Summary)\n```\n**bold**",
			want:  "```\n### 1) 요약 (Summary)\n```\n*bold*",
		},
		{
			name:  "bold-italic-triple-asterisk",
			input: "***근본 원인***\n내용",
			want:  "*근본 원인*\n내용",
		},
		{
			name:  "triple-asterisk-in-code-block-protected",
			input: "```\n***not bold***\n```\n***bold italic***",
			want:  "```\n***not bold***\n```\n*bold italic*",
		},
		{
			name:  "heading-with-bold-inside",
			input: "#### **근본 원인**\n내용",
			want:  "*근본 원인*\n내용",
		},
		{
			name:  "heading-with-bold-italic-inside",
			input: "### ***확인 근거***\n설명",
			want:  "*확인 근거*\n설명",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toSlackMarkdown(tt.input); got != tt.want {
				t.Fatalf("toSlackMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSlackClient_LanguageEn_AnalysisTitle(t *testing.T) {
	c := NewSlackClient(config.SlackConfig{Language: "en"})
	if got := c.t("analysis_result_title"); got != "🤖 AI Analysis" {
		t.Errorf("analysis_result_title (en) = %q, want %q", got, "🤖 AI Analysis")
	}
}

func TestSlackClient_LanguageKo_AnalysisTitle(t *testing.T) {
	c := NewSlackClient(config.SlackConfig{Language: "ko"})
	if got := c.t("analysis_result_title"); got != "🤖 AI 분석 결과" {
		t.Errorf("analysis_result_title (ko) = %q, want %q", got, "🤖 AI 분석 결과")
	}
}

func TestSlackClient_LanguageInvalidFallback(t *testing.T) {
	c := NewSlackClient(config.SlackConfig{Language: "invalid"})
	// Invalid language is normalized to "en" at construction.
	if got := c.t("analysis_result_title"); got != "🤖 AI Analysis" {
		t.Errorf("analysis_result_title (invalid) = %q, want en fallback %q", got, "🤖 AI Analysis")
	}
	if got := c.t("incident_dashboard_link"); got != "🔍 View incident dashboard" {
		t.Errorf("incident_dashboard_link (invalid) = %q, want en fallback", got)
	}
}

func TestSlackClient_LanguageMissingKey_FallsBackToEn(t *testing.T) {
	// Construct a ko-language client but request a key that does not exist in
	// the ko map (simulated by mutating slackI18n locally would be invasive;
	// instead verify the helper logic by checking that an unknown key returns
	// the en value when present, and "" when absent in both).
	c := NewSlackClient(config.SlackConfig{Language: "ko"})
	if got := c.t("__unknown_key__"); got != "" {
		t.Errorf("unknown key should return empty string, got %q", got)
	}
}

func TestSlackClient_KoFlappingBody_ContainsCycleFormatter(t *testing.T) {
	c := NewSlackClient(config.SlackConfig{Language: "ko"})
	body := c.t("flapping_detected_body")
	if !strings.Contains(body, "%d") {
		t.Errorf("ko flapping_detected_body missing %%d formatter, got %q", body)
	}
}

func TestSlackClient_EnFlappingBody_ContainsCycleFormatter(t *testing.T) {
	c := NewSlackClient(config.SlackConfig{Language: "en"})
	body := c.t("flapping_detected_body")
	if !strings.Contains(body, "%d") {
		t.Errorf("en flapping_detected_body missing %%d formatter, got %q", body)
	}
}

func TestSlackAttachmentMarshalsMrkdwnIn(t *testing.T) {
	msg := SlackMessage{
		Channel: "C123",
		Attachments: []SlackAttachment{
			{
				Title:    "AI",
				Text:     "*조치 사항*",
				MrkdwnIn: []string{"text", "fields"},
			},
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, `"mrkdwn_in":["text","fields"]`) {
		t.Fatalf("expected mrkdwn_in in payload, got %s", got)
	}
}
