# ConsoleHub Go Client Architecture

This document details the internal architecture, threading model, queueing strategy, circuit breaker, and lifecycle of the `consolehub` Go client library.

---

## 1. Threading & Concurrency Model

Application goroutines calling `consolehub.Printf`, `consolehub.Info`, or `writer.Write` operate completely concurrently without lock contention or network blocking.

```text
+-------------------------------------------------------------+
| Application Threads (Fprint, Printf, Info, Progress)       |
+-------------------------------------------------------------+
                              | (Non-blocking queue push)
                              v
+-------------------------------------------------------------+
| Bounded Queue (capacity: 10,000 events)                    |
+-------------------------------------------------------------+
                              | (Batched pop)
                              v
+-------------------------------------------------------------+
| Background Worker Goroutine                                 |
| - Manages WebSocket Connection Lifecycle                    |
| - Performs Exponential Backoff + Circuit Breaker            |
| - Sends JSON-RPC 2.0 Batches (Max 250 items / 50ms)         |
+-------------------------------------------------------------+
```

---

## 2. Queueing & Overflow Strategy

- **Bounded Queue**: Default size is 10,000 items.
- **Ring Buffer Eviction**: If queue capacity is reached due to prolonged server outage, the oldest item is evicted to make room for new events.
- **Dropped Counter**: `q.DroppedCount()` tracks cumulative dropped events.

---

## 3. Circuit Breaker & Reconnection

- **Threshold**: 5 consecutive connection/write failures trigger the circuit breaker.
- **State**: Circuit opens for 10 seconds, during which background network attempts are suspended and client operates in local-only mode.
- **Backoff & Jitter**: Reconnections utilize randomized exponential backoff: $t = \min(t_{max}, t_{initial} \times 2^n) \pm \text{jitter}$.
