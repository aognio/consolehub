package queue_test

import (
	"testing"

	"consolehub/libraries/go/consolehub/events"
	"consolehub/libraries/go/consolehub/queue"
)

func TestQueue_SequenceAndOverflow(t *testing.T) {
	q := queue.New(5)

	// Push 10 items into capacity 5 queue
	for i := 1; i <= 10; i++ {
		q.Push(events.NewTextLine("stdout", "line"))
	}

	if q.Len() != 5 {
		t.Errorf("expected queue length 5, got %d", q.Len())
	}

	if q.DroppedCount() != 5 {
		t.Errorf("expected 5 dropped items, got %d", q.DroppedCount())
	}

	if q.NextSequence() != 10 {
		t.Errorf("expected next sequence 10, got %d", q.NextSequence())
	}

	// Pop batch of 3
	batch := q.PopBatch(3)
	if len(batch) != 3 {
		t.Errorf("expected batch of 3, got %d", len(batch))
	}

	// Sequences in remaining batch should start from 6
	if batch[0].GetSequence() != 6 {
		t.Errorf("expected sequence 6 for first popped item, got %d", batch[0].GetSequence())
	}
}
