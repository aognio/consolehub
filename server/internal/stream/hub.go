package stream

import (
	"sync"

	"consolehub/internal/models"
)

// Hub manages real-time subscribers for live log streams.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan models.StreamLine
}

// NewHub initializes an in-memory PubSub hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string][]chan models.StreamLine),
	}
}

// Subscribe returns a channel listening to stream lines for a specific run ID.
func (h *Hub) Subscribe(runID string) chan models.StreamLine {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan models.StreamLine, 256)
	h.subscribers[runID] = append(h.subscribers[runID], ch)
	return ch
}

// Unsubscribe removes a channel subscriber for a run ID.
func (h *Hub) Unsubscribe(runID string, ch chan models.StreamLine) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs := h.subscribers[runID]
	for i, sub := range subs {
		if sub == ch {
			h.subscribers[runID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(h.subscribers[runID]) == 0 {
		delete(h.subscribers, runID)
	}
}

// Publish broadcasts a stream line to all active subscribers of a run ID.
func (h *Hub) Publish(line models.StreamLine) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs := h.subscribers[line.RunID]
	for _, ch := range subs {
		select {
		case ch <- line:
		default:
			// Non-blocking drop if consumer buffer is full to avoid slowing down ingestion
		}
	}
}
