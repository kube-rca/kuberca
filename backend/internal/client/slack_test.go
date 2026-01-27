package client

import "testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toSlackMarkdown(tt.input); got != tt.want {
				t.Fatalf("toSlackMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}
