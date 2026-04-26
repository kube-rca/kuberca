package sse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub_StartsWithNoClients(t *testing.T) {
	t.Parallel()
	h := NewHub()
	if h.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", h.ClientCount())
	}
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	t.Parallel()
	h := NewHub()

	c := h.Register("client-1")
	if c == nil {
		t.Fatal("Register returned nil client")
	}
	if c.ID != "client-1" {
		t.Errorf("client ID = %q, want %q", c.ID, "client-1")
	}
	if h.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", h.ClientCount())
	}

	h.Unregister("client-1")
	if h.ClientCount() != 0 {
		t.Errorf("expected 0 clients after unregister, got %d", h.ClientCount())
	}
}

func TestHub_UnregisterUnknownID_IsNoop(t *testing.T) {
	t.Parallel()
	h := NewHub()
	// Should not panic for unknown ID.
	h.Unregister("does-not-exist")
	if h.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", h.ClientCount())
	}
}

func TestHub_Broadcast_DeliversToClient(t *testing.T) {
	t.Parallel()
	h := NewHub()
	c := h.Register("client-2")

	event := Event{
		Type: EventAlertCreated,
		Data: EventData{AlertID: "a1", IncidentID: "inc-1"},
	}
	h.Broadcast(event)

	select {
	case raw := <-c.Events:
		var got Event
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Type != EventAlertCreated {
			t.Errorf("type = %q, want %q", got.Type, EventAlertCreated)
		}
		if got.Data.AlertID != "a1" {
			t.Errorf("alertID = %q, want %q", got.Data.AlertID, "a1")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast event")
	}

	h.Unregister("client-2")
}

func TestHub_Broadcast_ZeroTimestamp_IsFilledIn(t *testing.T) {
	t.Parallel()
	h := NewHub()
	c := h.Register("client-ts")

	before := time.Now()
	h.Broadcast(Event{Type: EventIncidentCreated})

	select {
	case raw := <-c.Events:
		var got Event
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Timestamp.IsZero() {
			t.Error("timestamp should not be zero")
		}
		if got.Timestamp.Before(before) {
			t.Errorf("timestamp %v is before broadcast start %v", got.Timestamp, before)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast event")
	}

	h.Unregister("client-ts")
}

func TestHub_Broadcast_MultipleClients(t *testing.T) {
	t.Parallel()
	h := NewHub()
	c1 := h.Register("multi-1")
	c2 := h.Register("multi-2")

	h.Broadcast(Event{Type: EventAlertResolved, Data: EventData{AlertID: "a2"}})

	for _, ch := range []chan []byte{c1.Events, c2.Events} {
		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting for broadcast event on a client")
		}
	}

	h.Unregister("multi-1")
	h.Unregister("multi-2")
}

func TestHub_Broadcast_NoClients_DoesNotPanic(t *testing.T) {
	t.Parallel()
	h := NewHub()
	// Should not panic when no clients are registered.
	h.Broadcast(Event{Type: EventHeartbeat})
}

func TestHub_ClientCount_Concurrent(t *testing.T) {
	t.Parallel()
	h := NewHub()
	const n = 20

	for i := range n {
		id := string(rune('A' + i))
		h.Register(id)
	}
	if h.ClientCount() != n {
		t.Errorf("expected %d clients, got %d", n, h.ClientCount())
	}

	for i := range n {
		id := string(rune('A' + i))
		h.Unregister(id)
	}
	if h.ClientCount() != 0 {
		t.Errorf("expected 0 clients after cleanup, got %d", h.ClientCount())
	}
}
