package logutil

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string is returned unchanged",
			in:   "",
			want: "",
		},
		{
			name: "plain string without control chars is returned unchanged",
			in:   "alert-id-1234",
			want: "alert-id-1234",
		},
		{
			name: "string containing spaces is returned unchanged",
			in:   "incident_id=abc severity=critical",
			want: "incident_id=abc severity=critical",
		},
		{
			name: "newline is replaced with space",
			in:   "abc\ndef",
			want: "abc def",
		},
		{
			name: "carriage return is replaced with space",
			in:   "abc\rdef",
			want: "abc def",
		},
		{
			name: "CRLF is replaced with two spaces",
			in:   "abc\r\ndef",
			want: "abc  def",
		},
		{
			name: "null byte is removed",
			in:   "abc\x00def",
			want: "abcdef",
		},
		{
			name: "mixed control characters are all sanitized",
			in:   "log\nentry\rmix\x00ed",
			want: "log entry mixed",
		},
		{
			name: "forged log entry attempt is neutralised",
			in:   "user-input\nINFO: forged log line claiming admin login",
			want: "user-input INFO: forged log line claiming admin login",
		},
		{
			name: "trailing newline is replaced",
			in:   "abc\n",
			want: "abc ",
		},
		{
			name: "leading newline is replaced",
			in:   "\nabc",
			want: " abc",
		},
		{
			name: "only newlines becomes only spaces",
			in:   "\n\n\n",
			want: "   ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Sanitize(tc.in)
			if got != tc.want {
				t.Fatalf("Sanitize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestSanitizeIdempotent verifies the helper is safe to over-apply: calling
// Sanitize on already-sanitised output returns the same string.
func TestSanitizeIdempotent(t *testing.T) {
	inputs := []string{
		"",
		"plain",
		"abc\ndef",
		"abc\r\ndef\x00",
		"already sanitised entry",
	}
	for _, in := range inputs {
		first := Sanitize(in)
		second := Sanitize(first)
		if first != second {
			t.Fatalf("Sanitize is not idempotent for %q: first=%q second=%q", in, first, second)
		}
	}
}
