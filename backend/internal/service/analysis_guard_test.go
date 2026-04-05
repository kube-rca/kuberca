package service

import (
	"testing"
	"time"
)

func TestAgentServiceBeginAnalysisDeduplicatesInFlightRequests(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}

	if _, ok := svc.beginAnalysis("alert:ALR-1"); !ok {
		t.Fatal("first beginAnalysis call should succeed")
	}
	if _, ok := svc.beginAnalysis("alert:ALR-1"); ok {
		t.Fatal("duplicate beginAnalysis call should be rejected while in flight")
	}

	svc.finishAnalysis("alert:ALR-1")

	if _, ok := svc.beginAnalysis("alert:ALR-1"); !ok {
		t.Fatal("beginAnalysis should succeed again after finishAnalysis")
	}
}

func TestAgentServiceBeginAnalysisEvictsStaleEntry(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}

	// Simulate a stale entry that exceeded the TTL.
	svc.inFlight["alert:ALR-2"] = time.Now().Add(-(inFlightStaleTTL + time.Second))

	if _, ok := svc.beginAnalysis("alert:ALR-2"); !ok {
		t.Fatal("beginAnalysis should evict stale entry and succeed")
	}
}

func TestRcaServiceBeginIncidentSummaryDeduplicatesInFlightRequests(t *testing.T) {
	svc := &RcaService{inFlightSummaries: make(map[string]time.Time)}

	if _, ok := svc.beginIncidentSummary("INC-1"); !ok {
		t.Fatal("first beginIncidentSummary call should succeed")
	}
	if _, ok := svc.beginIncidentSummary("INC-1"); ok {
		t.Fatal("duplicate beginIncidentSummary call should be rejected while in flight")
	}

	svc.finishIncidentSummary("INC-1")

	if _, ok := svc.beginIncidentSummary("INC-1"); !ok {
		t.Fatal("beginIncidentSummary should succeed again after finishIncidentSummary")
	}
}

func TestRcaServiceBeginIncidentSummaryEvictsStaleEntry(t *testing.T) {
	svc := &RcaService{inFlightSummaries: make(map[string]time.Time)}

	// Simulate a stale entry that exceeded the TTL.
	svc.inFlightSummaries["INC-2"] = time.Now().Add(-(inFlightStaleTTL + time.Second))

	if _, ok := svc.beginIncidentSummary("INC-2"); !ok {
		t.Fatal("beginIncidentSummary should evict stale entry and succeed")
	}
}

func TestAnalysisKeyReturnsAlertIDWhenPresent(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}
	key := svc.analysisKey("ALR-1", "fp-abc")
	if key != "alert:ALR-1" {
		t.Fatalf("expected alert:ALR-1, got %s", key)
	}
}

func TestAnalysisKeyFallsBackToFingerprint(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}
	key := svc.analysisKey("", "fp-abc")
	if key != "fingerprint:fp-abc" {
		t.Fatalf("expected fingerprint:fp-abc, got %s", key)
	}
}

func TestAnalysisKeyReturnsEmptyWhenBothMissing(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}
	key := svc.analysisKey("", "")
	if key != "" {
		t.Fatalf("expected empty string, got %s", key)
	}
}

func TestWaitForAnalysisCompletionReturnsAfterFinish(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}

	// Firing 분석 시작
	if _, ok := svc.beginAnalysis("alert:ALR-1"); !ok {
		t.Fatal("firing beginAnalysis should succeed")
	}

	// Resolved 분석 시도 → 차단됨
	if _, ok := svc.beginAnalysis("alert:ALR-1"); ok {
		t.Fatal("resolved beginAnalysis should be blocked while firing is in-flight")
	}

	// Firing 완료 시뮬레이션 (별도 goroutine)
	go func() {
		time.Sleep(100 * time.Millisecond)
		svc.finishAnalysis("alert:ALR-1")
	}()

	// waitForAnalysisCompletion: firing 완료 후 반환
	completed := svc.waitForAnalysisCompletion("alert:ALR-1", 5*time.Second)
	if !completed {
		t.Fatal("should complete after firing finishes")
	}

	// Resolved 분석 시작 가능
	if _, ok := svc.beginAnalysis("alert:ALR-1"); !ok {
		t.Fatal("resolved beginAnalysis should succeed after firing completes")
	}
}

func TestWaitForAnalysisCompletionTimesOut(t *testing.T) {
	svc := &AgentService{inFlight: make(map[string]time.Time)}
	svc.inFlight["alert:ALR-1"] = time.Now()

	// Ticker 간격(2s)보다 짧은 timeout으로 deadline 경로를 빠르게 검증
	completed := svc.waitForAnalysisCompletion("alert:ALR-1", 100*time.Millisecond)
	if completed {
		t.Fatal("should timeout when analysis never completes")
	}
}
