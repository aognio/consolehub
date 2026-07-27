package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/aognio/consolehub/libraries/go/consolehub/config"
	"github.com/aognio/consolehub/libraries/go/consolehub/events"
	"github.com/aognio/consolehub/libraries/go/consolehub/protocol"
	"github.com/aognio/consolehub/libraries/go/consolehub/queue"
	"github.com/aognio/consolehub/libraries/go/consolehub/transport"
	"github.com/aognio/consolehub/libraries/go/consolehub/worker"
)

func TestWorker_LifecycleAndMockTransport(t *testing.T) {
	q := queue.New(100)
	q.Push(events.NewTextLine("stdout", "Test line 1"))
	q.Push(events.NewTextLine("stdout", "Test line 2"))

	mockTrans := transport.NewMockTransport(func(method string, params any) (any, error) {
		switch method {
		case protocol.MethodAuthAuthenticate:
			return &protocol.AuthResult{Authenticated: true, TenantID: "t-1"}, nil
		case protocol.MethodProcessRegister:
			return &protocol.ProcessRegisterResult{ProcessID: "run-999"}, nil
		case protocol.MethodStreamAppend:
			return &protocol.StreamAppendResult{AcceptedThrough: 2}, nil
		case protocol.MethodProcessFinish:
			return nil, nil
		}
		return nil, nil
	})

	opts := worker.WorkerOptions{
		Endpoint:      "ws://mock/api/v1/rpc/ws",
		Token:         "token-123",
		Tenant:        "test-tenant",
		App:           "test-app",
		ClientRunID:   "run-id-123",
		Environment:   config.AutoDetect(),
		MaxBatchSize:  50,
		FlushInterval: 10 * time.Millisecond,
	}

	w := worker.New(opts, mockTrans, q)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	w.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	calls := mockTrans.GetCalls()
	if len(calls) < 3 {
		t.Errorf("expected at least 3 RPC calls (auth, register, append), got %v", calls)
	}
}
