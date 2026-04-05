package client

import (
	"encoding/json"
	"strings"
	"testing"
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
