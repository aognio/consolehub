package writer_test

import (
	"bytes"
	"testing"

	"consolehub/libraries/go/consolehub/queue"
	"consolehub/libraries/go/consolehub/writer"
)

func TestStreamWriter_LineSplitting(t *testing.T) {
	q := queue.New(100)
	var localBuf bytes.Buffer

	sw := writer.New("stdout", q, &localBuf)

	// Write chunked data without trailing newline
	_, _ = sw.Write([]byte("Hello "))
	if q.Len() != 0 {
		t.Errorf("expected 0 events queued for incomplete line, got %d", q.Len())
	}

	// Complete the line
	_, _ = sw.Write([]byte("World!\nSecond line.\n"))

	if q.Len() != 2 {
		t.Errorf("expected 2 events queued, got %d", q.Len())
	}

	if localBuf.String() != "Hello World!\nSecond line.\n" {
		t.Errorf("unexpected local buffer content: %s", localBuf.String())
	}

	// Flush remaining
	sw.Flush()
}
