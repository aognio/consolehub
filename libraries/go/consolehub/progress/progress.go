package progress

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/aognio/consolehub/libraries/go/consolehub/events"
	"github.com/aognio/consolehub/libraries/go/consolehub/queue"
)

type Tracker struct {
	id       string
	label    string
	total    int64
	current  int64
	finished int32
	queue    *queue.BoundedQueue
	mu       sync.Mutex
}

func New(id, label string, total int64, q *queue.BoundedQueue) *Tracker {
	t := &Tracker{
		id:    id,
		label: label,
		total: total,
		queue: q,
	}
	t.emit(0, false)
	return t
}

func (t *Tracker) Set(current int64) {
	if atomic.LoadInt32(&t.finished) == 1 {
		return
	}
	atomic.StoreInt64(&t.current, current)
	t.emit(current, false)
}

func (t *Tracker) Add(delta int64) {
	if atomic.LoadInt32(&t.finished) == 1 {
		return
	}
	curr := atomic.AddInt64(&t.current, delta)
	t.emit(curr, false)
}

func (t *Tracker) Done() {
	t.Finish()
}

func (t *Tracker) Finish() {
	if atomic.CompareAndSwapInt32(&t.finished, 0, 1) {
		curr := atomic.LoadInt64(&t.current)
		if t.total > 0 && curr < t.total {
			curr = t.total
			atomic.StoreInt64(&t.current, curr)
		}
		t.emit(curr, true)
		fmt.Println()
	}
}

func (t *Tracker) emit(current int64, finished bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	percent := float64(0)
	if t.total > 0 {
		percent = (float64(current) / float64(t.total)) * 100
		if percent > 100 {
			percent = 100
		}
	}

	// Render progress locally to stdout
	fmt.Printf("\r\033[K[%-20s] %3.0f%% (%d/%d) %s", renderBar(percent), percent, current, t.total, t.label)

	// Emit structured protocol event to queue
	if t.queue != nil {
		t.queue.Push(events.NewProgressEvent(t.id, t.label, current, t.total, finished))
	}
}

func renderBar(percent float64) string {
	filled := int(percent / 5.0)
	bar := ""
	for i := 0; i < 20; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	return bar
}
