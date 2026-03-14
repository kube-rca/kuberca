// SSE Hub: in-memory broadcast hub for Server-Sent Events
//
// - Hub manages connected SSE clients and broadcasts events to all of them.
// - Each client has a buffered channel (size 64) for events.
// - A heartbeat goroutine sends ping events every 30 seconds to keep
//   connections alive through proxies (e.g. Nginx 60s timeout).
// - Broadcast is non-blocking: if a client's buffer is full, the event is
//   dropped and a warning is logged.

package sse

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// EventType represents the type of SSE event.
type EventType string

const (
	EventAlertCreated      EventType = "alert_created"
	EventAlertResolved     EventType = "alert_resolved"
	EventAnalysisStarted   EventType = "analysis_started"
	EventAnalysisCompleted EventType = "analysis_completed"
	EventAnalysisFailed    EventType = "analysis_failed"
	EventIncidentCreated   EventType = "incident_created"
	EventIncidentUpdated   EventType = "incident_updated"
	EventIncidentResolved  EventType = "incident_resolved"
	EventHeartbeat         EventType = "heartbeat"
)

// Event is the payload broadcast to all SSE clients.
type Event struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      EventData `json:"data"`
}

// EventData carries optional identifiers and a human-readable message.
type EventData struct {
	AlertID    string `json:"alert_id,omitempty"`
	IncidentID string `json:"incident_id,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Client represents a single connected SSE consumer.
type Client struct {
	ID     string
	Events chan []byte
	Done   chan struct{}
}

// Hub holds all active SSE clients and provides broadcast capabilities.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewHub creates a Hub and starts the background heartbeat goroutine.
func NewHub() *Hub {
	h := &Hub{
		clients: make(map[string]*Client),
	}
	go h.heartbeatLoop()
	return h
}

// Register creates a new Client with a buffered event channel (size 64)
// and adds it to the hub.
func (h *Hub) Register(clientID string) *Client {
	c := &Client{
		ID:     clientID,
		Events: make(chan []byte, 64),
		Done:   make(chan struct{}),
	}
	h.mu.Lock()
	h.clients[clientID] = c
	h.mu.Unlock()
	log.Printf("SSE client registered (client_id=%s, total=%d)", clientID, h.ClientCount())
	return c
}

// Unregister removes the client from the hub and closes its channels.
func (h *Hub) Unregister(clientID string) {
	h.mu.Lock()
	if c, ok := h.clients[clientID]; ok {
		close(c.Done)
		close(c.Events)
		delete(h.clients, clientID)
	}
	h.mu.Unlock()
	log.Printf("SSE client unregistered (client_id=%s, total=%d)", clientID, h.ClientCount())
}

// Broadcast JSON-marshals the event and sends it to every connected client.
// If a client's buffer is full the event is dropped (non-blocking send).
func (h *Hub) Broadcast(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("SSE: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, c := range h.clients {
		select {
		case c.Events <- data:
		default:
			log.Printf("SSE: dropped event for client %s (buffer full)", id)
		}
	}
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// heartbeatLoop sends a heartbeat event every 30 seconds.
func (h *Hub) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.Broadcast(Event{
			Type:      EventHeartbeat,
			Timestamp: time.Now(),
			Data:      EventData{Message: "ping"},
		})
	}
}
