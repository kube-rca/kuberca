package db

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db/dbtest"
	"github.com/kube-rca/backend/internal/model"
)

func setupAnalysisDB(t *testing.T) *Postgres {
	t.Helper()
	pool := dbtest.StartPostgres(t)
	pg := &Postgres{Pool: pool}
	if err := pg.EnsureAlertSchema(); err != nil {
		t.Fatalf("EnsureAlertSchema: %v", err)
	}
	if err := pg.EnsureAlertAnalysisSchema(); err != nil {
		t.Fatalf("EnsureAlertAnalysisSchema: %v", err)
	}
	return pg
}

// seedAlert inserts a firing alert and returns its alertID.
func seedAlert(t *testing.T, pg *Postgres, fingerprint string) string {
	t.Helper()
	alertID, err := pg.SaveAlert(model.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "Seed"},
		Annotations: map[string]string{},
		StartsAt:    time.Now().Add(-time.Minute),
		Fingerprint: fingerprint,
	}, "")
	if err != nil {
		t.Fatalf("seedAlert(%s): %v", fingerprint, err)
	}
	return alertID
}

func TestInsertAlertAnalysis_RoundTrip(t *testing.T) {
	t.Parallel()
	pg := setupAnalysisDB(t)
	alertID := seedAlert(t, pg, "fp-analysis-001")

	analysisID, err := pg.InsertAlertAnalysis(
		alertID, "", "firing",
		"Root cause found", "Pod OOMKilled due to memory limit",
		model.LocalizedText{"ko": "근본 원인 발견"},
		model.LocalizedText{"ko": "메모리 제한으로 OOMKilled"},
		json.RawMessage(`{"pods":["app-123"]}`),
	)
	if err != nil {
		t.Fatalf("InsertAlertAnalysis: %v", err)
	}
	if analysisID <= 0 {
		t.Fatalf("analysisID = %d; want >0", analysisID)
	}

	got, err := pg.GetLatestAlertAnalysisByAlertID(alertID)
	if err != nil {
		t.Fatalf("GetLatestAlertAnalysisByAlertID: %v", err)
	}
	if got == nil {
		t.Fatal("GetLatestAlertAnalysisByAlertID returned nil")
	}
	if got.Summary != "Root cause found" {
		t.Errorf("Summary = %q; want 'Root cause found'", got.Summary)
	}
	if got.Status != "firing" {
		t.Errorf("Status = %q; want firing", got.Status)
	}
	if got.AlertID != alertID {
		t.Errorf("AlertID = %q; want %q", got.AlertID, alertID)
	}
}

func TestInsertAlertAnalysis_NoRows_ReturnsNil(t *testing.T) {
	t.Parallel()
	pg := setupAnalysisDB(t)

	got, err := pg.GetLatestAlertAnalysisByAlertID("nonexistent-alert")
	if err != nil {
		t.Fatalf("GetLatestAlertAnalysisByAlertID: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing alertID, got %+v", got)
	}
}

func TestInsertAlertAnalysisArtifacts(t *testing.T) {
	t.Parallel()
	pg := setupAnalysisDB(t)
	alertID := seedAlert(t, pg, "fp-artifact-001")

	analysisID, err := pg.InsertAlertAnalysis(
		alertID, "", "firing",
		"Summary", "Detail",
		model.LocalizedText{}, model.LocalizedText{},
		nil,
	)
	if err != nil {
		t.Fatalf("InsertAlertAnalysis: %v", err)
	}

	artifacts := []client.AlertAnalysisArtifactInput{
		{Type: "prometheus", Query: `rate(http_requests_total[5m])`, Result: json.RawMessage(`{"value":42}`), Summary: "High rate"},
		{Type: "k8s_event", Query: "get pods", Result: json.RawMessage(`{"events":[]}`), Summary: "No events"},
	}
	if err := pg.InsertAlertAnalysisArtifacts(analysisID, alertID, "", artifacts); err != nil {
		t.Fatalf("InsertAlertAnalysisArtifacts: %v", err)
	}

	rows, err := pg.GetAlertAnalysisArtifactsByAnalysisID(analysisID)
	if err != nil {
		t.Fatalf("GetAlertAnalysisArtifactsByAnalysisID: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("artifact count = %d; want 2", len(rows))
	}
	if rows[0].ArtifactType != "prometheus" {
		t.Errorf("first artifact type = %q; want prometheus", rows[0].ArtifactType)
	}
}

func TestGetLatestAnalysesByAlertID_DistinctOnStatus(t *testing.T) {
	t.Parallel()
	pg := setupAnalysisDB(t)
	alertID := seedAlert(t, pg, "fp-distinct-001")

	// Insert two firing analyses for the same alert.
	for _, summary := range []string{"first", "second"} {
		if _, err := pg.InsertAlertAnalysis(alertID, "", "firing", summary, "", nil, nil, nil); err != nil {
			t.Fatalf("InsertAlertAnalysis(%s): %v", summary, err)
		}
	}

	analyses, err := pg.GetLatestAnalysesByAlertID(alertID)
	if err != nil {
		t.Fatalf("GetLatestAnalysesByAlertID: %v", err)
	}
	// DISTINCT ON (status) → only 1 firing row expected.
	if len(analyses) != 1 {
		t.Errorf("len(analyses) = %d; want 1 (DISTINCT ON status)", len(analyses))
	}
	// Must be the most recent one.
	if analyses[0].Summary != "second" {
		t.Errorf("Summary = %q; want 'second'", analyses[0].Summary)
	}
}

func TestGetLatestFiringAnalysisByFingerprint(t *testing.T) {
	t.Parallel()
	pg := setupAnalysisDB(t)

	fp := "fp-firing-by-fp-001"
	alertID := seedAlert(t, pg, fp)

	if _, err := pg.InsertAlertAnalysis(alertID, "", "firing", "Firing summary", "", nil, nil, nil); err != nil {
		t.Fatalf("InsertAlertAnalysis: %v", err)
	}

	got, err := pg.GetLatestFiringAnalysisByFingerprint(fp)
	if err != nil {
		t.Fatalf("GetLatestFiringAnalysisByFingerprint: %v", err)
	}
	if got == nil {
		t.Fatal("expected analysis, got nil")
	}
	if got.Summary != "Firing summary" {
		t.Errorf("Summary = %q; want 'Firing summary'", got.Summary)
	}
}
