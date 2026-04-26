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

// logInjectionReplacer strips the control characters CodeQL's go/log-injection
// query (and CWE-117 in general) cares about: CR, LF, and NUL. CR and LF are
// replaced with a single space so the surrounding log message remains visually
// delimited; NUL is removed entirely because it has no useful display form.
//
// We intentionally use strings.NewReplacer here (instead of a hand-rolled scan
// or repeated strings.ReplaceAll calls) because CodeQL's Go data-flow library
// recognizes Replacer.Replace and Replacer.WriteString as built-in log-injection
// barriers when all CR / LF code points are covered, so call sites wrapped with
// Sanitize get treated as sanitised by the bundled go/log-injection query
// without needing a custom model-as-data extension. See:
//   - https://codeql.github.com/codeql-query-help/go/go-log-injection/
//   - github/codeql-go#731 (string replacement sanitizers for log-injection)
var logInjectionReplacer = strings.NewReplacer(
	"\n", " ",
	"\r", " ",
	"\x00", "",
)

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
	return logInjectionReplacer.Replace(s)
}
