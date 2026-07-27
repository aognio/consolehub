package writer

import (
	"bytes"
	"io"
	"os"
	"sync"

	"consolehub/libraries/go/consolehub/events"
	"consolehub/libraries/go/consolehub/queue"
)

type StreamWriter struct {
	mu          sync.Mutex
	stream      string
	queue       *queue.BoundedQueue
	localTarget io.Writer
	buf         bytes.Buffer
}

func New(stream string, q *queue.BoundedQueue, localTarget io.Writer) *StreamWriter {
	return &StreamWriter{
		stream:      stream,
		queue:       q,
		localTarget: localTarget,
	}
}

func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	// Write to local target (e.g. os.Stdout / os.Stderr) if provided
	if sw.localTarget != nil {
		_, _ = sw.localTarget.Write(p)
	}

	// Buffer line splits
	sw.buf.Write(p)
	for {
		line, errRead := sw.buf.ReadBytes('\n')
		if errRead != nil {
			// Incomplete line remains in buffer
			sw.buf.Write(line)
			break
		}

		lineStr := string(bytes.TrimRight(line, "\r\n"))
		if sw.queue != nil {
			sw.queue.Push(events.NewTextLine(sw.stream, lineStr))
		}
	}

	return len(p), nil
}

func (sw *StreamWriter) Flush() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if sw.buf.Len() > 0 {
		lineStr := sw.buf.String()
		sw.buf.Reset()
		if sw.queue != nil {
			sw.queue.Push(events.NewTextLine(sw.stream, lineStr))
		}
	}
}

func Stdout(q *queue.BoundedQueue) *StreamWriter {
	return New(events.StreamStdout, q, os.Stdout)
}

func Stderr(q *queue.BoundedQueue) *StreamWriter {
	return New(events.StreamStderr, q, os.Stderr)
}
