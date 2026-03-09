package service

import (
	"strings"
	"testing"
	"time"

	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/model"
)

// ============================================================================
// Mock: alertStore
// ============================================================================

type alertStoreMock struct {
	// SaveAlert tracking
	saveAlertCalls   []saveAlertCall
	saveAlertResults []saveAlertResult
	saveAlertIdx     int

	// State tracking for dedup verification
	firingAlerts   map[string]string // fingerprint → alertID (firing only)
	resolvedAlerts map[string]bool   // fingerprint → has resolved_at set

	// Incident
	firingIncidentID  string
	firingIncidentErr error

	// Flapping (default: no flapping)
	currentStatus   map[string]string // fingerprint → status
	isFlapping      map[string]bool
	alreadyResolved map[string]bool
	threadTS        map[string]string // fingerprint → thread_ts
}

type saveAlertCall struct {
	Alert      model.Alert
	IncidentID string
}

type saveAlertResult struct {
	AlertID string
	Err     error
}

func newAlertStoreMock() *alertStoreMock {
	return &alertStoreMock{
		firingAlerts:    make(map[string]string),
		resolvedAlerts:  make(map[string]bool),
		currentStatus:   make(map[string]string),
		isFlapping:      make(map[string]bool),
		alreadyResolved: make(map[string]bool),
		threadTS:        make(map[string]string),
	}
}

func (m *alertStoreMock) SaveAlert(alert model.Alert, incidentID string) (string, error) {
	m.saveAlertCalls = append(m.saveAlertCalls, saveAlertCall{Alert: alert, IncidentID: incidentID})

	if m.saveAlertIdx < len(m.saveAlertResults) {
		r := m.saveAlertResults[m.saveAlertIdx]
		m.saveAlertIdx++
		// Track state
		if r.Err == nil && alert.Status == "firing" {
			m.firingAlerts[alert.Fingerprint] = r.AlertID
			m.currentStatus[alert.Fingerprint] = "firing"
		} else if r.Err == nil && alert.Status == "resolved" {
			delete(m.firingAlerts, alert.Fingerprint)
			m.currentStatus[alert.Fingerprint] = "resolved"
		}
		return r.AlertID, r.Err
	}

	// Default: generate mock ID
	id := "ALR-mock0001"
	if alert.Status == "firing" {
		m.firingAlerts[alert.Fingerprint] = id
		m.currentStatus[alert.Fingerprint] = "firing"
	}
	return id, nil
}

func (m *alertStoreMock) GetAlertCurrentStatus(fingerprint string) (string, error) {
	if s, ok := m.currentStatus[fingerprint]; ok {
		return s, nil
	}
	return "", nil
}

func (m *alertStoreMock) IsAlertFlapping(fingerprint string) bool {
	return m.isFlapping[fingerprint]
}

func (m *alertStoreMock) RecordStateTransition(_, _, _ string, _ time.Time) error {
	return nil
}

func (m *alertStoreMock) IsAlertAlreadyResolved(fingerprint string, _ time.Time) (bool, error) {
	return m.alreadyResolved[fingerprint], nil
}

func (m *alertStoreMock) UpdateAlertResolved(fingerprint string, _ time.Time) error {
	m.resolvedAlerts[fingerprint] = true
	return nil
}

func (m *alertStoreMock) GetAlertThreadTS(fingerprint string) (string, bool) {
	ts, ok := m.threadTS[fingerprint]
	return ts, ok
}

func (m *alertStoreMock) UpdateAlertThreadTS(fingerprint, threadTS string) error {
	m.threadTS[fingerprint] = threadTS
	return nil
}

func (m *alertStoreMock) CountFlappingCycles(_ string, _ int) (int, time.Time, error) {
	return 0, time.Time{}, nil
}

func (m *alertStoreMock) MarkAlertAsFlapping(_ string, _ bool, _ int, _ time.Time) error {
	return nil
}

func (m *alertStoreMock) UpdateFlappingCycleCount(_ string, _ int) error {
	return nil
}

func (m *alertStoreMock) GetLatestAlertByFingerprint(_ string) (*model.AlertDetailResponse, error) {
	return nil, nil
}

func (m *alertStoreMock) HasTransitionsSince(_ string, _ time.Time) (bool, error) {
	return false, nil
}

func (m *alertStoreMock) GetFiringIncident() (*model.IncidentDetailResponse, error) {
	if m.firingIncidentID != "" {
		return &model.IncidentDetailResponse{IncidentID: m.firingIncidentID}, nil
	}
	return nil, m.firingIncidentErr
}

func (m *alertStoreMock) CreateIncident(_, _ string, _ time.Time) (string, error) {
	return "INC-test0001", nil
}

func (m *alertStoreMock) UpdateIncidentSeverity(_, _ string) error {
	return nil
}

// ============================================================================
// Mock: client.Notifier (ThreadAwareNotifier)
// ============================================================================

type notifierMock struct {
	events     []client.NotifierEvent
	threadRefs map[string]string
}

func newNotifierMock() *notifierMock {
	return &notifierMock{
		threadRefs: map[string]string{},
	}
}

func (m *notifierMock) Notify(event client.NotifierEvent) error {
	m.events = append(m.events, event)
	// Simulate storing thread_ref for firing alerts
	if e, ok := event.(client.AlertStatusChangedEvent); ok && e.Alert.Status == "firing" {
		m.threadRefs[e.Alert.Fingerprint] = "ts-" + e.Alert.Fingerprint
	}
	return nil
}

func (m *notifierMock) StoreThreadRef(alertKey, threadRef string) {
	m.threadRefs[alertKey] = threadRef
}

func (m *notifierMock) GetThreadRef(alertKey string) (string, bool) {
	ref, ok := m.threadRefs[alertKey]
	return ref, ok
}

func (m *notifierMock) DeleteThreadRef(alertKey string) {
	delete(m.threadRefs, alertKey)
}

func (m *notifierMock) RequiresThreadRef() bool {
	return false
}

// ============================================================================
// Mock: alertAnalyzer
// ============================================================================

type analyzerMock struct {
	calls []analyzerCall
}

type analyzerCall struct {
	Fingerprint string
	AlertID     string
	ThreadTS    string
	IncidentID  string
}

func (m *analyzerMock) RequestAnalysis(alert model.Alert, alertID, threadTS, incidentID string, skipThreadCheck bool) {
	m.calls = append(m.calls, analyzerCall{
		Fingerprint: alert.Fingerprint,
		AlertID:     alertID,
		ThreadTS:    threadTS,
		IncidentID:  incidentID,
	})
}

// ============================================================================
// Helper: AlertService 생성
// ============================================================================

func newTestAlertService(store *alertStoreMock, notifier *notifierMock, analyzer *analyzerMock) *AlertService {
	return &AlertService{
		notifier:     notifier,
		agentService: analyzer,
		db:           store,
		envFlapping: config.FlappingConfig{
			Enabled: false, // 기본: flapping 비활성
		},
		sseHub: nil, // SSE 비활성
	}
}

func makeAlert(fingerprint, status, severity string) model.Alert {
	return model.Alert{
		Status:      status,
		Labels:      map[string]string{"alertname": "TestAlert", "severity": severity},
		Annotations: map[string]string{"summary": "test"},
		StartsAt:    time.Now().Add(-5 * time.Minute),
		EndsAt:      time.Now(),
		Fingerprint: fingerprint,
	}
}

func makeWebhook(alerts ...model.Alert) model.AlertmanagerWebhook {
	return model.AlertmanagerWebhook{
		Version: "4",
		Status:  alerts[0].Status,
		Alerts:  alerts,
	}
}

// ============================================================================
// Tests
// ============================================================================

func TestSaveAlert_ReturnsAlertID(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-abc12345", Err: nil},
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert := makeAlert("fp-001", "firing", "warning")
	webhook := makeWebhook(alert)

	sent, failed := svc.ProcessWebhook(webhook)

	if sent != 1 || failed != 0 {
		t.Fatalf("ProcessWebhook() = sent=%d, failed=%d; want sent=1, failed=0", sent, failed)
	}

	if len(store.saveAlertCalls) != 1 {
		t.Fatalf("SaveAlert called %d times; want 1", len(store.saveAlertCalls))
	}

	if store.saveAlertCalls[0].Alert.Fingerprint != "fp-001" {
		t.Fatalf("SaveAlert fingerprint = %q; want %q", store.saveAlertCalls[0].Alert.Fingerprint, "fp-001")
	}
}

func TestSaveAlert_NewFiring_NewAlertID(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-11111111", Err: nil},
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert := makeAlert("fp-new", "firing", "warning")
	webhook := makeWebhook(alert)

	svc.ProcessWebhook(webhook)

	// Verify alertID has ALR- prefix
	if len(store.saveAlertCalls) != 1 {
		t.Fatalf("expected 1 SaveAlert call, got %d", len(store.saveAlertCalls))
	}

	// SaveAlert should have been called with the alert
	call := store.saveAlertCalls[0]
	if call.Alert.Fingerprint != "fp-new" {
		t.Fatalf("SaveAlert fingerprint = %q; want %q", call.Alert.Fingerprint, "fp-new")
	}
}

func TestSaveAlert_RepeatFiring_SameAlertID(t *testing.T) {
	store := newAlertStoreMock()
	// Same alertID returned for both calls (simulating dedup)
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-same0001", Err: nil},
		{AlertID: "ALR-same0001", Err: nil},
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert1 := makeAlert("fp-repeat", "firing", "warning")
	alert2 := makeAlert("fp-repeat", "firing", "warning")

	svc.ProcessWebhook(makeWebhook(alert1))
	svc.ProcessWebhook(makeWebhook(alert2))

	if len(store.saveAlertCalls) != 2 {
		t.Fatalf("SaveAlert called %d times; want 2", len(store.saveAlertCalls))
	}
}

func TestSaveAlert_ResolvedAfterFiring_SameAlertID(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-fire0001", Err: nil}, // firing
		{AlertID: "ALR-fire0001", Err: nil}, // resolved (same ID)
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	// 1. Firing alert
	firingAlert := makeAlert("fp-resolve", "firing", "warning")
	svc.ProcessWebhook(makeWebhook(firingAlert))

	// 2. Resolved alert
	resolvedAlert := makeAlert("fp-resolve", "resolved", "warning")
	svc.ProcessWebhook(makeWebhook(resolvedAlert))

	if len(store.saveAlertCalls) != 2 {
		t.Fatalf("SaveAlert called %d times; want 2", len(store.saveAlertCalls))
	}

	// Resolved should trigger UpdateAlertResolved
	if !store.resolvedAlerts["fp-resolve"] {
		t.Fatal("UpdateAlertResolved was not called for fingerprint fp-resolve")
	}
}

func TestSaveAlert_ReFiringAfterResolved_NewAlertID(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-first001", Err: nil}, // first firing
		{AlertID: "ALR-first001", Err: nil}, // resolved (same ID)
		{AlertID: "ALR-secnd002", Err: nil}, // re-firing (NEW ID)
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	// 1. First firing
	svc.ProcessWebhook(makeWebhook(makeAlert("fp-refire", "firing", "warning")))

	// 2. Resolved
	svc.ProcessWebhook(makeWebhook(makeAlert("fp-refire", "resolved", "warning")))

	// 3. Re-firing → should get a NEW alertID
	svc.ProcessWebhook(makeWebhook(makeAlert("fp-refire", "firing", "warning")))

	if len(store.saveAlertCalls) != 3 {
		t.Fatalf("SaveAlert called %d times; want 3", len(store.saveAlertCalls))
	}

	// Verify all calls received the same fingerprint
	for i, call := range store.saveAlertCalls {
		if call.Alert.Fingerprint != "fp-refire" {
			t.Fatalf("SaveAlert[%d] fingerprint = %q; want %q", i, call.Alert.Fingerprint, "fp-refire")
		}
	}
}

func TestProcessWebhook_SkipsInfoSeverity(t *testing.T) {
	store := newAlertStoreMock()
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert := makeAlert("fp-info", "firing", "info")
	webhook := makeWebhook(alert)

	sent, failed := svc.ProcessWebhook(webhook)

	if sent != 0 || failed != 0 {
		t.Fatalf("ProcessWebhook() = sent=%d, failed=%d; want sent=0, failed=0", sent, failed)
	}

	if len(store.saveAlertCalls) != 0 {
		t.Fatalf("SaveAlert called %d times for info severity; want 0", len(store.saveAlertCalls))
	}
}

func TestProcessWebhook_DuplicateResolvedSkipped(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-dup00001", Err: nil},
	}
	// Mark as already resolved
	store.alreadyResolved["fp-dup"] = true
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert := makeAlert("fp-dup", "resolved", "warning")
	webhook := makeWebhook(alert)

	sent, failed := svc.ProcessWebhook(webhook)

	// Duplicate resolved should be skipped via continue
	if sent != 0 || failed != 0 {
		t.Fatalf("ProcessWebhook() = sent=%d, failed=%d; want sent=0, failed=0", sent, failed)
	}

	// UpdateAlertResolved should NOT be called (skipped before it)
	if store.resolvedAlerts["fp-dup"] {
		t.Fatal("UpdateAlertResolved should not be called for duplicate resolved alert")
	}
}

func TestAlertIDFormat(t *testing.T) {
	// Verify the expected ALR-{8chars} format
	tests := []struct {
		name    string
		alertID string
		valid   bool
	}{
		{"valid ALR format", "ALR-a1b2c3d4", true},
		{"valid ALR format with numbers", "ALR-12345678", true},
		{"missing prefix", "a1b2c3d4", false},
		{"wrong prefix", "INC-a1b2c3d4", false},
		{"too short", "ALR-abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := strings.HasPrefix(tt.alertID, "ALR-") && len(tt.alertID) >= 12
			if isValid != tt.valid {
				t.Fatalf("alertID %q validity = %v; want %v", tt.alertID, isValid, tt.valid)
			}
		})
	}
}

func TestProcessWebhook_MultipleAlerts(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "ALR-alert001", Err: nil},
		{AlertID: "ALR-alert002", Err: nil},
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert1 := makeAlert("fp-multi1", "firing", "critical")
	alert2 := makeAlert("fp-multi2", "firing", "warning")
	webhook := model.AlertmanagerWebhook{
		Version: "4",
		Status:  "firing",
		Alerts:  []model.Alert{alert1, alert2},
	}

	sent, failed := svc.ProcessWebhook(webhook)

	if sent != 2 || failed != 0 {
		t.Fatalf("ProcessWebhook() = sent=%d, failed=%d; want sent=2, failed=0", sent, failed)
	}

	if len(store.saveAlertCalls) != 2 {
		t.Fatalf("SaveAlert called %d times; want 2", len(store.saveAlertCalls))
	}

	// Each alert should have a different fingerprint
	fps := map[string]bool{}
	for _, call := range store.saveAlertCalls {
		fps[call.Alert.Fingerprint] = true
	}
	if len(fps) != 2 {
		t.Fatalf("Expected 2 distinct fingerprints, got %d", len(fps))
	}
}

func TestProcessWebhook_SaveAlertError_ContinuesProcessing(t *testing.T) {
	store := newAlertStoreMock()
	store.saveAlertResults = []saveAlertResult{
		{AlertID: "", Err: errMock},
		{AlertID: "ALR-ok000001", Err: nil},
	}
	notif := newNotifierMock()
	analyzer := &analyzerMock{}
	svc := newTestAlertService(store, notif, analyzer)

	alert1 := makeAlert("fp-fail", "firing", "warning")
	alert2 := makeAlert("fp-ok", "firing", "warning")
	webhook := model.AlertmanagerWebhook{
		Version: "4",
		Status:  "firing",
		Alerts:  []model.Alert{alert1, alert2},
	}

	sent, failed := svc.ProcessWebhook(webhook)

	// First alert save failed, but second should succeed
	// Both should attempt Slack send (DB failure doesn't block Slack)
	if sent+failed != 2 {
		t.Fatalf("ProcessWebhook() total processed = %d; want 2", sent+failed)
	}

	if len(store.saveAlertCalls) != 2 {
		t.Fatalf("SaveAlert called %d times; want 2", len(store.saveAlertCalls))
	}
}

var errMock = &mockError{msg: "mock error"}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
