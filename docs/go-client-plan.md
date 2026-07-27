# ConsoleHub Production Go Client Library Implementation Plan

This document outlines the architectural plan, API ergonomics, transport design, resilience mechanisms, testing strategy, and migration guide for the official ConsoleHub Go Client Library (`libraries/go/consolehub`).

---

## 1. Overview & Philosophy

The ConsoleHub Go client library provides real-time telemetry ingestion, structured log streaming, interactive progress tracking, and console interaction for Go applications running across distributed hosts.

### Core Philosophy & Guarantees
1. **Extension of Standard Library**: The public API mirrors Go's standard `fmt`, `log`, and `io.Writer` packages (`consolehub.Printf`, `consolehub.Stdout()`, `consolehub.Info`).
2. **Zero-Block Application Code**: Application goroutines emitting output never perform synchronous network I/O. Messages are enqueued into a lock-free/bounded lock buffer and processed asynchronously by a background worker.
3. **Graceful Degradation**: If the ConsoleHub server is unreachable, undergoing maintenance, or drops connection, the client degrades silently. Output renders locally, and application execution is never blocked or terminated due to ConsoleHub server failures.
4. **Mechanical Migration**: Standard `fmt.Printf("Downloading %s\n", file)` calls can be converted to `consolehub.Printf("Downloading %s\n", file)` with minimal code refactoring.
5. **Idempotency & Sequence Order**: Every line emitted receives a monotonically increasing 64-bit sequence number. Batches are deduplicated server-side via `batch_id` and process executions are registered idempotently using `client_run_id`.

---

## 2. Package Architecture & Layout

```text
libraries/go/consolehub/
├── client.go                 # Main Client instance & public entry points
├── options.go                # ClientOptions & functional option builders
├── std.go                    # Top-level standard helpers (Print, Printf, Info, Progress, Prompt)
├── writer.go                 # io.Writer wrappers for Stdout, Stderr, Custom Streams
├── config/                   # Auto-detection of hostname, PID, command line, environment
│   └── config.go
├── events/                   # Typed structured event definitions (TextLine, LogEvent, Progress, Prompt)
│   └── events.go
├── protocol/                 # JSON-RPC 2.0 frame structures & method constants
│   └── protocol.go
├── transport/                # Transport interface, WebSocket transport, and Mock transport
│   ├── transport.go
│   ├── websocket.go
│   └── mock.go
├── queue/                    # Thread-safe bounded buffer queue with drop policy
│   └── queue.go
├── worker/                   # Async background worker, batching, heartbeat, reconnect, backoff, circuit breaker
│   └── worker.go
├── progress/                 # First-class progress bar trackers
│   └── progress.go
├── prompt/                   # Interactive console prompts (Prompt, Secret, Confirm, Choice)
│   └── prompt.go
└── examples/                 # Comprehensive example applications
    ├── basic/
    ├── progress/
    ├── logger/
    ├── multiwriter/
    ├── prompt/
    └── telegram-style/
```

---

## 3. Public API Ergonomics

### 3.1 Lifecycle & Global Access
```go
// Initialize custom client
client, err := consolehub.New(
    consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
    consolehub.WithTenant("engineering"),
    consolehub.WithApp("telegram-replicator"),
    consolehub.WithToken("my-secret-token"),
)
defer client.Close()

// Or use global default client auto-configured from environment variables
defer consolehub.Close()
```

### 3.2 Standard Output Replacements (`fmt` style)
```go
consolehub.Print("Starting worker...")
consolehub.Printf("Connected to host %s on port %d\n", host, port)
consolehub.Println("Ready.")

// Direct io.Writer access
fmt.Fprintln(consolehub.Stdout(), "Writing to stdout stream")
fmt.Fprintln(consolehub.Stderr(), "Writing to stderr stream")
mw := io.MultiWriter(os.Stdout, consolehub.Stdout())
```

### 3.3 Structured Logging (`log` style)
```go
consolehub.Debug("Initializing components...")
consolehub.Infof("Processing batch of %d items", batchSize)
consolehub.Warn("High memory consumption detected")
consolehub.Errorf("Failed to connect to DB: %v", err)

// Log typed events
consolehub.Log(events.LogEvent{
    Level: "info",
    Message: "Download complete",
    Fields: map[string]any{"bytes": 1048576, "duration_ms": 420},
})
```

### 3.4 Progress Trackers
```go
p := consolehub.Progress("Downloading media", 100)
for i := 0; i <= 100; i += 10 {
    p.Set(int64(i))
    time.Sleep(100 * time.Millisecond)
}
p.Done()

// Upload progress
up := consolehub.UploadProgress("Uploading artifact.tar.gz", 10485760)
up.Add(102400)
up.Finish()
```

### 3.5 Interactive Console Prompts
```go
// Text Prompt
answer := consolehub.Prompt("Enter target channel username", "@mychannel")

// Masked Secret Prompt
apiKey := consolehub.SecretPrompt("Enter API secret key")

// Confirmation Prompt
if consolehub.Confirm("Proceed with database migration?", true) {
    // perform migration
}

// Choice Selection
selected := consolehub.Choice("Select environment", []string{"Development", "Staging", "Production"}, "Development")
```

---

## 4. Transport, Buffering & Circuit Breaker Architecture

```text
+-----------------------------------------------------------------------------------+
| Application Goroutines (Fprint, Printf, Info, Progress, Prompt)                   |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| Bounded Queue (Fixed Capacity: 10,000 events)                                     |
| - Monotonic Sequence Generation (Sequence #1, #2, #3...)                          |
| - Non-blocking push (drops lowest priority or oldest on overflow)                 |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| Asynchronous Background Worker                                                    |
| - Batch Aggregator (Max 250 items or 50ms flush tick)                             |
| - Circuit Breaker (Half-Open/Closed/Open states)                                  |
| - Reconnection Loop with Exponential Backoff + Random Jitter                      |
| - Automatic Heartbeat Ticker (30s interval)                                       |
| - Sequence Acknowledgement & Resume Replay                                        |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| Transport Adapter (JSON-RPC 2.0 over WebSocket / Mock Transport)                  |
+-----------------------------------------------------------------------------------+
```

### Key Transport & Resilience Features:
1. **Bounded Buffer Queue**: Memory footprint is strictly bounded (default 10,000 messages max). If queue fills due to prolonged server outage, oldest non-critical logs drop gracefully without blocking application threads.
2. **Circuit Breaker**: When connection failures exceed consecutive threshold (default 5 failures), circuit breaker opens for a backoff duration (10s), skipping network attempts while continuing local output.
3. **Exponential Backoff with Jitter**: Reconnection attempts use $t_{backoff} = \min(t_{max}, t_{initial} \times 2^n) \pm \text{jitter}$.
4. **Sequence Tracking & Resumption**: Client tracks `accepted_through_sequence`. Upon reconnect, client invokes `stream.resume` procedure to fetch server sequence state and replay unacknowledged messages seamlessly.
5. **Idempotent Client Run ID**: Auto-generates a UUID `client_run_id` per process instance so repeated registrations return the existing `process_id` without creating duplicate run entries.

---

## 5. Documentation & Skill Deliverables

- **Client Documentation Suite**:
  - `docs/go-client-plan.md` (this plan)
  - `docs/go-client.md`
  - `docs/go-client-api.md`
  - `docs/go-client-architecture.md`
  - `docs/go-client-migration.md`
  - `docs/go-client-testing.md`
  - `docs/go-client-roadmap.md`

- **Agent Reusable Skills**:
  - `skills/consolehub/SKILL.md` (General concepts, architecture, JSON-RPC 2.0 protocol, batching, sequence numbers, streaming philosophy)
  - `skills/consolehub-golang/SKILL.md` (Go-specific public API ergonomics, io.Writer, progress, prompts, structured logging, testing, migration examples)

- **Example Applications**:
  - `libraries/go/consolehub/examples/basic/main.go`
  - `libraries/go/consolehub/examples/progress/main.go`
  - `libraries/go/consolehub/examples/logger/main.go`
  - `libraries/go/consolehub/examples/multiwriter/main.go`
  - `libraries/go/consolehub/examples/prompt/main.go`
  - `libraries/go/consolehub/examples/telegram-style/main.go`

---

## 6. Implementation Workflow

1. Create `docs/go-client-plan.md` (completed).
2. Update `.development/journal.md`, `.development/decisions.md`, and `.development/backlog.md`.
3. Create agent skills `skills/consolehub/SKILL.md` and `skills/consolehub-golang/SKILL.md`.
4. Implement subpackages under `libraries/go/consolehub/`: `config`, `events`, `protocol`, `transport`, `queue`, `worker`, `writer`, `progress`, `prompt`.
5. Implement main `consolehub` public API package and std helpers.
6. Write unit and integration test suite across all subpackages (`go test -v ./...`).
7. Write complete documentation suite under `docs/go-client*.md`.
8. Create 6 example applications under `libraries/go/consolehub/examples/`.
