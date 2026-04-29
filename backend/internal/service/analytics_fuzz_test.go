package service

import (
	"testing"
	"time"
)

func FuzzParseAnalyticsWindow(f *testing.F) {
	seeds := []string{
		"",
		"1h",
		"24h",
		"1d",
		"30d",
		"365d",
		"0d",
		"-1d",
		"366d",
		"abc",
		"1.5d",
		"1 d",
		"  7d  ",
		"1H",
		"1y",
		"99999999999999d",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		duration, normalized, err := parseAnalyticsWindow(raw)
		if err != nil {
			if duration != 0 || normalized != "" {
				t.Fatalf("error path returned non-zero values: duration=%v normalized=%q err=%v", duration, normalized, err)
			}
			return
		}

		if duration <= 0 {
			t.Fatalf("non-error path returned non-positive duration %v for input %q", duration, raw)
		}
		if duration > 365*24*time.Hour {
			t.Fatalf("non-error path exceeded 365d cap: duration=%v for input %q", duration, raw)
		}
		if normalized == "" {
			t.Fatalf("non-error path returned empty normalized window for input %q", raw)
		}
	})
}
