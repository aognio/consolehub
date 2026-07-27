package stream_test

import (
	"testing"
	"time"

	"consolehub/internal/models"
	"consolehub/internal/stream"
)

func TestStreamHub_PubSub(t *testing.T) {
	hub := stream.NewHub()
	runID := "run-100"

	ch := hub.Subscribe(runID)
	defer hub.Unsubscribe(runID, ch)

	line := models.StreamLine{
		RunID:     runID,
		Sequence:  1,
		Timestamp: time.Now(),
		Stream:    models.StreamStdout,
		Kind:      models.KindText,
		Text:      "Hello world from live stream",
	}

	hub.Publish(line)

	select {
	case received := <-ch:
		if received.Text != line.Text {
			t.Errorf("expected text '%s', got '%s'", line.Text, received.Text)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for pubsub stream line")
	}
}
