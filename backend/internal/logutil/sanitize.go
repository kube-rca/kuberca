// Package logutil provides small helpers for safely emitting log records
// when user-controlled values are interpolated into log messages.
//
// The standard library log package writes one record per call, but the
// underlying writer is line-oriented: a caller-controlled value containing
// CR / LF (or NUL) characters can forge additional log entries that look
// like they originated from the application itself ("log injection",
// CodeQL rule go/log-injection).
//
// Sanitize replaces those control characters so the value can be safely
// passed to log.Printf or any other line-oriented logger. It is a no-op
// for strings that contain no control characters, so over-applying it is
// cheap and side-effect free.
package logutil

import "strings"

// Sanitize returns s with line-terminating and NUL control characters
// stripped (CR, LF) or removed (NUL) so it cannot forge new log entries
// when written through log.Printf or similar line-oriented loggers.
//
// CR and LF are replaced with a single space rather than removed so that
// the surrounding log message remains visually delimited; NUL is removed
// entirely because it has no useful display form.
func Sanitize(s string) string {
	if s == "" {
		return s
	}
	if !strings.ContainsAny(s, "\n\r\x00") {
		return s
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}
