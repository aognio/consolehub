# ConsoleHub Go Client Library

The official Go client library (`libraries/go/consolehub`) provides lightweight, zero-blocking real-time telemetry streaming, progress tracking, and structured logging for Go applications.

---

## 1. Installation & Quick Start

```bash
go get consolehub/libraries/go/consolehub
```

```go
package main

import (
    "consolehub/libraries/go/consolehub"
)

func main() {
    defer consolehub.Close()

    consolehub.Println("Application starting up...")
    consolehub.Printf("Worker pool ready with %d threads\n", 4)
}
```

---

## 2. Key Features

- **Standard Library Parity**: Drop-in replacements for `fmt.Print*`, `log.Output`, and `io.Writer`.
- **Zero Blocking**: Application goroutines push messages to a bounded in-memory buffer in nanoseconds without performing synchronous network I/O.
- **Graceful Degradation**: If the ConsoleHub server is offline, client applications continue executing normally and rendering local output.
- **Circuit Breaker & Backoff**: Automatically recovers from network disconnections using exponential backoff with random jitter.
- **Monotonic Sequences**: Messages receive sequence numbers for deduplication and reconnection playback.
- **First-Class Progress & Prompts**: Track progress bars (`consolehub.Progress`) and prompt users interactively (`consolehub.Prompt`, `consolehub.Confirm`).
