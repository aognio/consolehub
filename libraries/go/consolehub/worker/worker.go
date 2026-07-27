package worker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"consolehub/libraries/go/consolehub/config"
	"consolehub/libraries/go/consolehub/protocol"
	"consolehub/libraries/go/consolehub/queue"
	"consolehub/libraries/go/consolehub/transport"
)

type WorkerState int32

const (
	StateDisconnected WorkerState = iota
	StateConnecting
	StateConnected
	StateDegraded
)

type WorkerOptions struct {
	Endpoint         string
	Token            string
	Tenant           string
	App              string
	ClientRunID      string
	Hostname         string // Mandatory
	PID              int    // Mandatory
	AppVersion       string // Optional
	OSName           string // Optional
	Environment      config.Environment
	MaxBatchSize     int
	FlushInterval    time.Duration
	HeartbeatSeconds int
}

type Worker struct {
	opts             WorkerOptions
	trans            transport.Transport
	q                *queue.BoundedQueue
	state            int32
	processID        string
	acceptedSeq      int64
	consecutiveErrs  int32
	circuitOpenUntil time.Time
	stopChan         chan struct{}
	stopOnce         sync.Once
	wg               sync.WaitGroup
	mu               sync.Mutex
}

func New(opts WorkerOptions, trans transport.Transport, q *queue.BoundedQueue) *Worker {
	if opts.MaxBatchSize <= 0 {
		opts.MaxBatchSize = 250
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = 50 * time.Millisecond
	}
	if opts.HeartbeatSeconds <= 0 {
		opts.HeartbeatSeconds = 30
	}
	if trans == nil {
		trans = transport.NewWebSocketTransport()
	}

	return &Worker{
		opts:     opts,
		trans:    trans,
		q:        q,
		stopChan: make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.runLoop(ctx)
}

func (w *Worker) runLoop(ctx context.Context) {
	defer w.wg.Done()

	flushTicker := time.NewTicker(w.opts.FlushInterval)
	defer flushTicker.Stop()

	heartbeatTicker := time.NewTicker(time.Duration(w.opts.HeartbeatSeconds) * time.Second)
	defer heartbeatTicker.Stop()

	// Initial connection attempt
	_ = w.ensureConnected(ctx)

	for {
		select {
		case <-w.stopChan:
			w.flushAndFinish(ctx)
			return
		case <-ctx.Done():
			w.flushAndFinish(context.Background())
			return
		case <-flushTicker.C:
			w.processBatch(ctx)
		case <-heartbeatTicker.C:
			w.sendHeartbeat(ctx)
		}
	}
}

func (w *Worker) ensureConnected(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if time.Now().Before(w.circuitOpenUntil) {
		return fmt.Errorf("circuit breaker open")
	}

	if w.trans.IsConnected() && w.processID != "" {
		return nil
	}

	atomic.StoreInt32(&w.state, int32(StateConnecting))

	// Connect transport
	if err := w.trans.Connect(ctx, w.opts.Endpoint); err != nil {
		w.recordFailureLocked()
		return err
	}

	// Authenticate
	authParams := protocol.AuthParams{
		Token: w.opts.Token,
		Protocol: map[string]string{
			"name":    "ConsoleHub-Go-Client",
			"version": "1.0.0",
		},
	}
	var authRes protocol.AuthResult
	if err := w.trans.Call(ctx, protocol.MethodAuthAuthenticate, authParams, &authRes); err != nil {
		w.recordFailureLocked()
		return err
	}

	// Resolve Mandatory & Optional metadata
	hostname := w.opts.Hostname
	if hostname == "" {
		hostname = w.opts.Environment.Hostname
	}

	pid := w.opts.PID
	if pid <= 0 {
		pid = w.opts.Environment.PID
	}

	osName := w.opts.OSName
	if osName == "" {
		osName = w.opts.Environment.Platform
	}

	version := w.opts.AppVersion
	if version == "" {
		version = "1.0.0"
	}

	// Register Process
	regParams := protocol.ProcessRegisterParams{
		Tenant: w.opts.Tenant,
		App:    w.opts.App,
		Host: protocol.HostParams{
			Hostname:    hostname,
			DisplayName: hostname,
			Platform:    osName,
		},
		Process: protocol.ProcessDetails{
			ClientRunID:      w.opts.ClientRunID,
			PID:              pid,
			StartedAt:        time.Now().Format(time.RFC3339),
			Version:          version,
			CommandLine:      w.opts.Environment.CommandLine,
			WorkingDirectory: w.opts.Environment.WorkingDirectory,
		},
	}

	var regRes protocol.ProcessRegisterResult
	if err := w.trans.Call(ctx, protocol.MethodProcessRegister, regParams, &regRes); err != nil {
		w.recordFailureLocked()
		return err
	}

	w.processID = regRes.ProcessID
	atomic.StoreInt32(&w.state, int32(StateConnected))
	atomic.StoreInt32(&w.consecutiveErrs, 0)
	return nil
}

func (w *Worker) processBatch(ctx context.Context) {
	if w.q.Len() == 0 {
		return
	}

	if err := w.ensureConnected(ctx); err != nil {
		// Degraded mode: keep events in buffer for next flush attempt
		return
	}

	batch := w.q.PopBatch(w.opts.MaxBatchSize)
	if len(batch) == 0 {
		return
	}

	linesMap := make([]map[string]any, len(batch))
	for i, item := range batch {
		linesMap[i] = item.ToMap()
	}

	batchID := fmt.Sprintf("b-%d-%d", time.Now().UnixNano(), rand.Int63())
	appendParams := protocol.StreamAppendParams{
		ProcessID:     w.processID,
		BatchID:       batchID,
		FirstSequence: batch[0].GetSequence(),
		Lines:         linesMap,
	}

	var res protocol.StreamAppendResult
	if err := w.trans.Call(ctx, protocol.MethodStreamAppend, appendParams, &res); err != nil {
		w.mu.Lock()
		w.recordFailureLocked()
		w.mu.Unlock()
		return
	}

	atomic.StoreInt64(&w.acceptedSeq, res.AcceptedThrough)
}

func (w *Worker) sendHeartbeat(ctx context.Context) {
	if atomic.LoadInt32(&w.state) != int32(StateConnected) {
		return
	}

	params := map[string]any{
		"process_id":         w.processID,
		"timestamp":          time.Now().Format(time.RFC3339),
		"last_sent_sequence": atomic.LoadInt64(&w.acceptedSeq),
	}

	_ = w.trans.Call(ctx, protocol.MethodProcessHeartbeat, params, nil)
}

func (w *Worker) flushAndFinish(ctx context.Context) {
	// Flush remaining items
	for w.q.Len() > 0 {
		w.processBatch(ctx)
	}

	if atomic.LoadInt32(&w.state) == int32(StateConnected) && w.processID != "" {
		finishParams := protocol.ProcessFinishParams{
			ProcessID:    w.processID,
			FinishedAt:   time.Now().Format(time.RFC3339),
			Status:       "exited",
			ExitCode:     0,
			LastSequence: atomic.LoadInt64(&w.acceptedSeq),
		}
		_ = w.trans.Call(ctx, protocol.MethodProcessFinish, finishParams, nil)
	}

	_ = w.trans.Close()
	atomic.StoreInt32(&w.state, int32(StateDisconnected))
}

func (w *Worker) recordFailureLocked() {
	errs := atomic.AddInt32(&w.consecutiveErrs, 1)
	atomic.StoreInt32(&w.state, int32(StateDegraded))

	if errs >= 5 {
		// Circuit breaker opens for 10 seconds
		w.circuitOpenUntil = time.Now().Add(10 * time.Second)
	}
}

func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
	})
	w.wg.Wait()
}

func (w *Worker) IsConnected() bool {
	return atomic.LoadInt32(&w.state) == int32(StateConnected)
}
