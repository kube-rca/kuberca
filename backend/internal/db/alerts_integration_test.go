package db

import (
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/db/dbtest"
	"github.com/kube-rca/backend/internal/model"
)

// setupAlertDB creates a Postgres backed by a real container and applies
// the alert schema (including alert_state_transitions).
func setupAlertDB(t *testing.T) *Postgres {
	t.Helper()
	pool := dbtest.StartPostgres(t)
	pg := &Postgres{Pool: pool}
	if err := pg.EnsureAlertSchema(); err != nil {
		t.Fatalf("EnsureAlertSchema: %v", err)
	}
	return pg
}

func TestSaveAlert_InsertAndRead(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "HighCPU", "severity": "warning"},
		Annotations: map[string]string{"summary": "CPU too high"},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: "fp-insert-read-001",
	}

	alertID, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}
	if alertID == "" {
		t.Fatal("SaveAlert returned empty alertID")
	}

	got, err := pg.GetAlertDetail(alertID)
	if err != nil {
		t.Fatalf("GetAlertDetail: %v", err)
	}
	if got.AlarmTitle != "HighCPU" {
		t.Errorf("AlarmTitle = %q; want %q", got.AlarmTitle, "HighCPU")
	}
	if got.Status != "firing" {
		t.Errorf("Status = %q; want firing", got.Status)
	}
	if got.Fingerprint != "fp-insert-read-001" {
		t.Errorf("Fingerprint = %q; want fp-insert-read-001", got.Fingerprint)
	}
}

func TestSaveAlert_IdempotentOnSameFingerprint(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	fp := "fp-idem-001"
	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "DiskFull", "severity": "critical"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-2 * time.Minute),
		Fingerprint: fp,
	}

	id1, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("first SaveAlert: %v", err)
	}

	// Second call with same fingerprint + firing → must return same alertID
	id2, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("second SaveAlert: %v", err)
	}
	if id1 != id2 {
		t.Errorf("idempotency broken: id1=%q id2=%q", id1, id2)
	}
}

func TestManualResolveAlert(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "PodCrash"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: "fp-resolve-001",
	}
	alertID, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}

	if err := pg.ManualResolveAlert(alertID); err != nil {
		t.Fatalf("ManualResolveAlert: %v", err)
	}

	got, err := pg.GetAlertDetail(alertID)
	if err != nil {
		t.Fatalf("GetAlertDetail after resolve: %v", err)
	}
	if got.Status != "resolved" {
		t.Errorf("Status = %q; want resolved", got.Status)
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt is nil after ManualResolveAlert")
	}
}

func TestManualResolveAlert_AlreadyResolved(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "AlreadyDone"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: "fp-already-resolved-001",
	}
	alertID, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}

	if err := pg.ManualResolveAlert(alertID); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	// Second resolve of an already-resolved alert should return an error.
	if err := pg.ManualResolveAlert(alertID); err == nil {
		t.Error("expected error on double resolve, got nil")
	}
}

func TestGetAlertList_ReturnsEnabledAlerts(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	for i, fp := range []string{"fp-list-a", "fp-list-b"} {
		_, err := pg.SaveAlert(model.Alert{
			Status:      "firing",
			Labels:      map[string]string{"alertname": "Alert", "severity": "info"},
			Annotations: map[string]string{},
			StartsAt:    time.Now().Add(-time.Duration(i+1) * time.Minute),
			Fingerprint: fp,
		}, "")
		if err != nil {
			t.Fatalf("SaveAlert %d: %v", i, err)
		}
	}

	list, err := pg.GetAlertList()
	if err != nil {
		t.Fatalf("GetAlertList: %v", err)
	}
	if len(list) < 2 {
		t.Errorf("GetAlertList returned %d items; want ≥2", len(list))
	}
}

func TestUpdateAlertAnalysis(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "MemLeak"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: "fp-analysis-update-001",
	}
	alertID, err := pg.SaveAlert(alert, "")
	if err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}

	err = pg.UpdateAlertAnalysis(alertID, "Memory leak detected", "OOM in container foo",
		model.LocalizedText{"ko": "메모리 누수 감지"},
		model.LocalizedText{"ko": "컨테이너 foo에서 OOM 발생"},
	)
	if err != nil {
		t.Fatalf("UpdateAlertAnalysis: %v", err)
	}

	got, err := pg.GetAlertDetail(alertID)
	if err != nil {
		t.Fatalf("GetAlertDetail: %v", err)
	}
	if got.AnalysisSummary == nil || *got.AnalysisSummary != "Memory leak detected" {
		t.Errorf("AnalysisSummary = %v; want 'Memory leak detected'", got.AnalysisSummary)
	}
}

func TestRecordStateTransition_AndHasTransitionsSince(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	fp := "fp-transition-001"
	before := time.Now().Add(-time.Second)

	if err := pg.RecordStateTransition(fp, "firing", "resolved", time.Now()); err != nil {
		t.Fatalf("RecordStateTransition: %v", err)
	}

	has, err := pg.HasTransitionsSince(fp, before)
	if err != nil {
		t.Fatalf("HasTransitionsSince: %v", err)
	}
	if !has {
		t.Error("HasTransitionsSince returned false; want true")
	}
}

func TestUpdateAlertThreadTS(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	fp := "fp-thread-001"
	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "SlackThread"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: fp,
	}
	if _, err := pg.SaveAlert(alert, ""); err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}

	const wantTS = "1234567890.123456"
	if err := pg.UpdateAlertThreadTS(fp, wantTS); err != nil {
		t.Fatalf("UpdateAlertThreadTS: %v", err)
	}

	gotTS, ok := pg.GetAlertThreadTS(fp)
	if !ok {
		t.Fatal("GetAlertThreadTS: not found")
	}
	if gotTS != wantTS {
		t.Errorf("thread_ts = %q; want %q", gotTS, wantTS)
	}
}

func TestGetAlertCurrentStatus(t *testing.T) {
	t.Parallel()
	pg := setupAlertDB(t)

	fp := "fp-status-001"
	alert := model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "StatusCheck"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: fp,
	}
	if _, err := pg.SaveAlert(alert, ""); err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}

	status, err := pg.GetAlertCurrentStatus(fp)
	if err != nil {
		t.Fatalf("GetAlertCurrentStatus: %v", err)
	}
	if status != "firing" {
		t.Errorf("status = %q; want firing", status)
	}
}
