# ConsoleHub Go Client Library (`libraries/go/consolehub`)

[![Go Reference](https://pkg.go.dev/badge/consolehub.svg)](https://pkg.go.dev/consolehub)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

The official Go client library for **ConsoleHub**—a remote application console, real-time log stream aggregator, and process telemetry platform.

Designed as an extension of the Go standard library (`fmt`, `log`, `io.Writer`), `consolehub` allows Go applications to stream structured logs, stdout/stderr, progress bars, and interactive prompts without blocking application execution or introducing network latency.

---

## Key Features & Guarantees

* **Standard Library Parity**: Drop-in replacements for standard `fmt.Print*` and `log.Output` functions.
* **Zero-Blocking Concurrency**: Application threads push events into a bounded in-memory buffer in nanoseconds without performing synchronous network I/O.
* **Graceful Degradation**: If the ConsoleHub server is offline, maintenance-bound, or unreachable, client applications continue executing normally and rendering local output.
* **Circuit Breaker & Backoff**: Features a built-in circuit breaker (opens after 5 consecutive network errors for 10s) and randomized exponential backoff ($t = \min(t_{max}, t_{initial} \times 2^n) \pm \text{jitter}$).
* **Idempotency & Monotonic Sequences**: Every emitted line receives a 64-bit sequence number. Server deduplicates batches via `batch_id` and process runs via `client_run_id`.
* **First-Class Progress & Prompts**: Track progress bars (`consolehub.Progress`) and prompt users interactively (`consolehub.Prompt`, `consolehub.SecretPrompt`, `consolehub.Confirm`).

---

## Installation

```bash
go get consolehub/libraries/go/consolehub
```

Import into your Go application:

```go
import "consolehub/libraries/go/consolehub"
```

---

## Quick Start

### 1. Minimal Global Client (Zero Config)
Auto-configures from environment variables (`CONSOLEHUB_ENDPOINT`, `CONSOLEHUB_TOKEN`, `CONSOLEHUB_TENANT`, `CONSOLEHUB_APP`).

```go
package main

import (
    "time"

    "consolehub/libraries/go/consolehub"
)

func main() {
    defer consolehub.Close()

    consolehub.Println("Application starting up...")
    consolehub.Infof("Processing batch of %d tasks...", 10)
    
    time.Sleep(100 * time.Millisecond)
    consolehub.Println("Work completed.")
}
```

### 2. Custom Explicit Client
```go
package main

import (
    "log"

    "consolehub/libraries/go/consolehub"
)

func main() {
    client, err := consolehub.New(
        consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
        consolehub.WithTenant("engineering"),
        consolehub.WithApp("telegram-replicator"),
        consolehub.WithToken("my-secret-token"),
    )
    if err != nil {
        log.Fatalf("Failed to initialize consolehub: %v", err)
    }
    defer client.Close()

    client.Println("Connected to ConsoleHub telemetry server.")
}
```

---

## API Reference

### 1. Client Initialization & Configuration

| Function / Option | Type | Description |
|---|---|---|
| `consolehub.New(opts ...Option)` | Constructor | Creates a new isolated `*Client` instance |
| `consolehub.Default()` | Constructor | Returns or initializes the global default `*Client` |
| `consolehub.Close()` | Lifecycle | Flushes remaining buffers and closes the background worker |
| `WithEndpoint(url)` | Config | Overrides JSON-RPC WebSocket URL (default: `ws://localhost:3787/api/v1/rpc/ws`) |
| `WithTenant(tenant)` | Config | Sets organization tenant ID or slug (default: `default`) |
| `WithApp(app)` | Config | Sets application identifier name |
| `WithToken(token)` | Config | Sets client authentication token |
| `WithClientRunID(ulid)` | **Mandatory** | Sets client-generated 26-char Crockford Base32 ULID (auto-generated) |
| `WithHostname(hostname)` | **Mandatory** | Sets host machine hostname (auto-detected fallback) |
| `WithPID(pid)` | **Mandatory** | Sets process ID integer (auto-detected fallback) |
| `WithAppVersion(version)` | *Optional* | Sets application version string (e.g. `"1.4.2"`) |
| `WithOSName(osName)` | *Optional* | Sets operating system name (e.g. `"linux"`, `"darwin"`) |
| `WithQueueCapacity(cap)` | Config | Sets bounded queue memory capacity (default: 10,000 items) |
| `WithDisabled(disabled)` | Config | Disables network transmission (runs in local-only mode) |
| `WithTransport(trans)` | Config | Injects custom or mock transport implementation |

---

### 2. Printing API (`fmt` Replacements)

Replace standard `fmt.Print*` calls to stream output in real-time while printing locally to stdout:

```go
// Direct printing
consolehub.Print("Starting worker...")
consolehub.Printf("Processing item %d of %d\n", current, total)
consolehub.Println("Operation complete.")

// io.Writer printing
consolehub.Fprint(w, "Writing to writer...")
consolehub.Fprintf(w, "Formatted %s\n", val)
consolehub.Fprintln(w, "Done.")

// Stream Writers
stdoutWriter := consolehub.Stdout()
stderrWriter := consolehub.Stderr()
customWriter := consolehub.Writer("custom_stream")
```

---

### 3. Structured Logging API

Emit leveled structured logs directly to ConsoleHub:

```go
consolehub.Debug("Initializing cache storage")
consolehub.Info("Server listening on port 8080")
consolehub.Warn("High memory utilization detected")
consolehub.Error("Failed to connect to database")

consolehub.Debugf("Cache size: %d entries", count)
consolehub.Infof("Processed %d events in %v", total, elapsed)
consolehub.Warnf("Queue depth reached %d%%", depth)
consolehub.Errorf("Query error: %v", err)
```

---

### 4. Progress Tracking API

Create terminal progress trackers that render locally while transmitting structured `events.ProgressEvent` payloads:

```go
// Task progress
p := consolehub.Progress("Downloading dataset archive", 100)
for i := 0; i <= 100; i += 10 {
    p.Set(int64(i))
    time.Sleep(50 * time.Millisecond)
}
p.Done()

// Upload progress
up := consolehub.UploadProgress("Uploading artifact.tar.gz", 10485760)
up.Add(102400)
up.Finish()
```

---

### 5. Interactive Console Prompts API

Prompt terminal users interactively while logging responses:

```go
// Text input prompt
channel := consolehub.Prompt("Enter target channel username", "@mychannel")

// Masked secret input prompt
apiKey := consolehub.SecretPrompt("Enter API secret key")

// Yes/No confirmation prompt
if consolehub.Confirm("Proceed with deployment?", true) {
    consolehub.Println("Deployment started.")
}

// Choice selection prompt
env := consolehub.Choice("Select environment", []string{"Development", "Staging", "Production"}, "Development")
```

---

## Standard Library Migration Guide

Migrating existing Go code to ConsoleHub requires minimal code refactoring:

| Existing Code | ConsoleHub Replacement | Description |
|---|---|---|
| `fmt.Print(...)` | `consolehub.Print(...)` | Output stream |
| `fmt.Printf(...)` | `consolehub.Printf(...)` | Formatted output stream |
| `fmt.Println(...)` | `consolehub.Println(...)` | Line output stream |
| `log.Println(...)` | `consolehub.Infof(...)` | Structured info log |
| `os.Stdout` | `consolehub.Stdout()` | `io.Writer` stdout wrapper |
| `os.Stderr` | `consolehub.Stderr()` | `io.Writer` stderr wrapper |

> **IMPORTANT**: Do **NOT** replace standard Go error handling or formatting primitives:
> - Keep `fmt.Sprintf(...)`
> - Keep `fmt.Errorf(...)`
> - Keep `errors.Is(...)` / `errors.As(...)`

---

## Architecture & Resilience

```text
+-------------------------------------------------------------------+
| Application Threads (Print, Printf, Info, Progress, Prompt)       |
+-------------------------------------------------------------------+
                                  |
                                  v (Non-blocking enqueue)
+-------------------------------------------------------------------+
| Bounded Buffer Queue (Capacity: 10,000 events)                    |
| - Monotonic sequence numbers (#1, #2, #3...)                      |
| - Ring buffer eviction strategy on overflow                       |
+-------------------------------------------------------------------+
                                  |
                                  v (Asynchronous batch pop)
+-------------------------------------------------------------------+
| Background Worker Goroutine                                       |
| - WebSocket connection & authentication                           |
| - Automatic process registration (`client_run_id`)                 |
| - Batch aggregation (Max 250 items or 50ms interval)              |
| - Circuit Breaker (Opens after 5 errors for 10s window)           |
| - Exponential backoff with random jitter                          |
+-------------------------------------------------------------------+
```

---

## Unit Testing & Mock Transport

Test application routines without connecting to a live ConsoleHub server using `transport.MockTransport`:

```go
package myapp_test

import (
    "testing"

    "consolehub/libraries/go/consolehub"
    "consolehub/libraries/go/consolehub/protocol"
    "consolehub/libraries/go/consolehub/transport"
)

func TestMyService(t *testing.T) {
    mockTrans := transport.NewMockTransport(func(method string, params any) (any, error) {
        switch method {
        case protocol.MethodAuthAuthenticate:
            return &protocol.AuthResult{Authenticated: true}, nil
        case protocol.MethodProcessRegister:
            return &protocol.ProcessRegisterResult{ProcessID: "test-run-1"}, nil
        case protocol.MethodStreamAppend:
            return &protocol.StreamAppendResult{AcceptedThrough: 100}, nil
        }
        return nil, nil
    })

    client, err := consolehub.New(
        consolehub.WithTenant("test"),
        consolehub.WithApp("test-service"),
        consolehub.WithTransport(mockTrans),
    )
    if err != nil {
        t.Fatalf("Failed to initialize client: %v", err)
    }
    defer client.Close()

    client.Println("Testing service output...")
}
```

Or disable networking entirely during local unit tests:

```go
client, _ := consolehub.New(consolehub.WithDisabled(true))
defer client.Close()
```

---

## Examples Directory

Explore fully functional example applications in `examples/`:

- [`examples/basic/`](examples/basic/main.go): Basic standard print streaming
- [`examples/progress/`](examples/progress/main.go): Progress bar tracking
- [`examples/logger/`](examples/logger/main.go): Leveled structured logging
- [`examples/multiwriter/`](examples/multiwriter/main.go): Multi-writer target integration
- [`examples/prompt/`](examples/prompt/main.go): Interactive terminal prompts
- [`examples/telegram-style/`](examples/telegram-style/main.go): Telegram replicator migration target

---

## License

ConsoleHub Go Client Library is open-source software licensed under the [MIT License](LICENSE).
