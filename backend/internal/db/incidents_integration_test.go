package db

import (
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/db/dbtest"
	"github.com/kube-rca/backend/internal/model"
)

func setupIncidentDB(t *testing.T) *Postgres {
	t.Helper()
	pool := dbtest.StartPostgres(t)
	pg := &Postgres{Pool: pool}
	if err := pg.EnsureIncidentSchema(); err != nil {
		t.Fatalf("EnsureIncidentSchema: %v", err)
	}
	// alerts table is needed for GetIncidentList (LEFT JOIN).
	if err := pg.EnsureAlertSchema(); err != nil {
		t.Fatalf("EnsureAlertSchema: %v", err)
	}
	return pg
}

func TestCreateIncident_RoundTrip(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	incidentID, err := pg.CreateIncident("Pod crash loop in prod", "critical", time.Now())
	if err != nil {
		t.Fatalf("CreateIncident: %v", err)
	}
	if incidentID == "" {
		t.Fatal("CreateIncident returned empty ID")
	}

	got, err := pg.GetIncidentDetail(incidentID)
	if err != nil {
		t.Fatalf("GetIncidentDetail: %v", err)
	}
	if got.Title != "Pod crash loop in prod" {
		t.Errorf("Title = %q; want 'Pod crash loop in prod'", got.Title)
	}
	if got.Status != "firing" {
		t.Errorf("Status = %q; want firing", got.Status)
	}
	if got.Severity != "critical" {
		t.Errorf("Severity = %q; want critical", got.Severity)
	}
}

func TestResolveIncident(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	// The firing unique index allows only one firing incident at a time.
	// Use GetOrCreateFiringIncident so the schema constraint is satisfied.
	incidentID, created, err := pg.GetOrCreateFiringIncident("Test incident", "warning", time.Now())
	if err != nil {
		t.Fatalf("GetOrCreateFiringIncident: %v", err)
	}
	if !created {
		t.Skip("a firing incident already existed — skipping isolation test")
	}

	if err := pg.ResolveIncident(incidentID, "test-user"); err != nil {
		t.Fatalf("ResolveIncident: %v", err)
	}

	got, err := pg.GetIncidentDetail(incidentID)
	if err != nil {
		t.Fatalf("GetIncidentDetail: %v", err)
	}
	if got.Status != "resolved" {
		t.Errorf("Status = %q; want resolved", got.Status)
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt is nil after ResolveIncident")
	}
	if got.ResolvedBy == nil || *got.ResolvedBy != "test-user" {
		t.Errorf("ResolvedBy = %v; want 'test-user'", got.ResolvedBy)
	}
}

func TestResolveIncident_AlreadyResolved(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	incidentID, created, err := pg.GetOrCreateFiringIncident("Another incident", "info", time.Now())
	if err != nil {
		t.Fatalf("GetOrCreateFiringIncident: %v", err)
	}
	if !created {
		t.Skip("pre-existing firing incident")
	}

	if err := pg.ResolveIncident(incidentID, "user1"); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	// Second resolve must fail (no firing incident).
	if err := pg.ResolveIncident(incidentID, "user2"); err == nil {
		t.Error("expected error on double resolve, got nil")
	}
}

func TestUpdateIncident(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	incidentID, err := pg.CreateIncident("Original title", "info", time.Now())
	if err != nil {
		t.Fatalf("CreateIncident: %v", err)
	}

	// Resolve first so we can update without firing unique index issues.
	_ = pg.ResolveIncident(incidentID, "sys")

	req := model.UpdateIncidentRequest{
		Title:           "Updated title",
		Severity:        "warning",
		AnalysisSummary: "RCA done",
		AnalysisDetail:  "Details here",
	}
	if err := pg.UpdateIncident(incidentID, req); err != nil {
		t.Fatalf("UpdateIncident: %v", err)
	}

	got, err := pg.GetIncidentDetail(incidentID)
	if err != nil {
		t.Fatalf("GetIncidentDetail: %v", err)
	}
	if got.Title != "Updated title" {
		t.Errorf("Title = %q; want 'Updated title'", got.Title)
	}
	if got.AnalysisSummary == nil || *got.AnalysisSummary != "RCA done" {
		t.Errorf("AnalysisSummary = %v; want 'RCA done'", got.AnalysisSummary)
	}
}

func TestHideAndUnhideIncident(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	incidentID, err := pg.CreateIncident("Hide me", "info", time.Now())
	if err != nil {
		t.Fatalf("CreateIncident: %v", err)
	}
	// Resolve so we can create another firing incident later without conflicts.
	_ = pg.ResolveIncident(incidentID, "sys")

	if err := pg.HideIncident(incidentID); err != nil {
		t.Fatalf("HideIncident: %v", err)
	}

	// Hidden incidents should appear in the hidden list.
	hidden, err := pg.GetHiddenIncidentList()
	if err != nil {
		t.Fatalf("GetHiddenIncidentList: %v", err)
	}
	found := false
	for _, inc := range hidden {
		if inc.IncidentID == incidentID {
			found = true
		}
	}
	if !found {
		t.Errorf("incident %s not in hidden list after HideIncident", incidentID)
	}

	if err := pg.UnhideIncident(incidentID); err != nil {
		t.Fatalf("UnhideIncident: %v", err)
	}

	list, err := pg.GetIncidentList()
	if err != nil {
		t.Fatalf("GetIncidentList: %v", err)
	}
	found = false
	for _, inc := range list {
		if inc.IncidentID == incidentID {
			found = true
		}
	}
	if !found {
		t.Errorf("incident %s not visible after UnhideIncident", incidentID)
	}
}

func TestGetOrCreateFiringIncident_Idempotent(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	id1, created1, err := pg.GetOrCreateFiringIncident("Firing 1", "warning", time.Now())
	if err != nil {
		t.Fatalf("first GetOrCreateFiringIncident: %v", err)
	}
	if !created1 {
		t.Skip("pre-existing firing incident")
	}

	id2, created2, err := pg.GetOrCreateFiringIncident("Firing 2", "critical", time.Now())
	if err != nil {
		t.Fatalf("second GetOrCreateFiringIncident: %v", err)
	}

	if created2 {
		t.Error("expected created2=false (existing firing incident)")
	}
	if id1 != id2 {
		t.Errorf("id mismatch: id1=%q id2=%q", id1, id2)
	}
}

func TestUpdateIncidentSeverity(t *testing.T) {
	t.Parallel()
	pg := setupIncidentDB(t)

	incidentID, err := pg.CreateIncident("Sev test", "info", time.Now())
	if err != nil {
		t.Fatalf("CreateIncident: %v", err)
	}
	// Resolve before updating to avoid unique index conflict.
	_ = pg.ResolveIncident(incidentID, "sys")

	// info → warning: should succeed (warning > info).
	if err := pg.UpdateIncidentSeverity(incidentID, "warning"); err != nil {
		t.Fatalf("UpdateIncidentSeverity(warning): %v", err)
	}

	got, err := pg.GetIncidentDetail(incidentID)
	if err != nil {
		t.Fatalf("GetIncidentDetail: %v", err)
	}
	if got.Severity != "warning" {
		t.Errorf("Severity = %q; want warning", got.Severity)
	}
}
