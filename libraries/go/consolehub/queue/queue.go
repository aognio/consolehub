package queue

import (
	"sync"
	"sync/atomic"

	"consolehub/libraries/go/consolehub/events"
)

type BoundedQueue struct {
	mu           sync.Mutex
	capacity     int
	items        []events.Event
	nextSequence int64
	droppedCount int64
}

func New(capacity int) *BoundedQueue {
	if capacity <= 0 {
		capacity = 10000
	}
	return &BoundedQueue{
		capacity: capacity,
		items:    make([]events.Event, 0, 1024),
	}
}

// Push enqueues an event, assigning a monotonically increasing sequence number.
func (q *BoundedQueue) Push(evt events.Event) bool {
	if evt == nil {
		return false
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Assign monotonic sequence number
	seq := atomic.AddInt64(&q.nextSequence, 1)
	evt.SetSequence(seq)

	// Ring buffer overflow drop strategy (drops oldest item if capacity is reached)
	if len(q.items) >= q.capacity {
		q.items = q.items[1:]
		atomic.AddInt64(&q.droppedCount, 1)
	}

	q.items = append(q.items, evt)
	return true
}

// PopBatch retrieves up to maxCount events from the front of the queue.
func (q *BoundedQueue) PopBatch(maxCount int) []events.Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	count := maxCount
	if count > len(q.items) {
		count = len(q.items)
	}

	batch := make([]events.Event, count)
	copy(batch, q.items[:count])
	q.items = q.items[count:]
	return batch
}

func (q *BoundedQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *BoundedQueue) DroppedCount() int64 {
	return atomic.LoadInt64(&q.droppedCount)
}

func (q *BoundedQueue) NextSequence() int64 {
	return atomic.LoadInt64(&q.nextSequence)
}

func (q *BoundedQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = make([]events.Event, 0, 1024)
}
