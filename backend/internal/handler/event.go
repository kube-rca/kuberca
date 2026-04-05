// SSE stream endpoint handler.
//
// GET /api/v1/events — long-lived HTTP connection that pushes server-sent
// events (alerts, incidents, analysis results, heartbeats) to the browser.

package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kube-rca/backend/internal/sse"
)

// EventHandler serves the SSE stream.
type EventHandler struct {
	hub *sse.Hub
}

// NewEventHandler creates an EventHandler bound to the given Hub.
func NewEventHandler(hub *sse.Hub) *EventHandler {
	return &EventHandler{hub: hub}
}

// Stream handles the SSE connection lifecycle:
//  1. Set required SSE headers.
//  2. Register the client with the hub.
//  3. Send an initial "connected" event.
//  4. Forward hub events until the client disconnects.
func (h *EventHandler) Stream(c *gin.Context) {
	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	clientID := uuid.New().String()
	client := h.hub.Register(clientID)
	defer h.hub.Unregister(clientID)

	// Send initial "connected" event
	connectedEvent := sse.Event{
		Type:      "connected",
		Timestamp: time.Now(),
		Data: sse.EventData{
			Message: fmt.Sprintf("connected (client_id=%s)", clientID),
		},
	}
	connData, err := json.Marshal(connectedEvent)
	if err != nil {
		log.Printf("Failed to marshal connected event: %v", err)
		c.AbortWithStatus(500)
		return
	}
	fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(connData))
	c.Writer.Flush()

	// Event loop
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE client disconnected (client_id=%s)", clientID)
			return
		case <-client.Done:
			return
		case eventData, ok := <-client.Events:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(eventData))
			c.Writer.Flush()
		}
	}
}
