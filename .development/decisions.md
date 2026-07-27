# Decision Log (Architecture Decision Records)

## ADR-001: Embedded PocketBase over External Database
* **Status**: Accepted
* **Context**: ConsoleHub requires an embedded, zero-dependency database to deliver a single self-contained binary.
* **Alternatives Considered**:
  - Embedded SQLite (`mattn/go-sqlite3` or `modernc.org/sqlite` directly).
  - External PostgreSQL / MySQL.
* **Decision**: Embed PocketBase (`github.com/pocketbase/pocketbase`).
* **Rationale**: PocketBase wraps SQLite with built-in migrations, JSON queries, admin utilities, and Go collection APIs, drastically simplifying local data management while keeping deployment to a single binary.

## ADR-002: JSON-RPC 2.0 over WebSockets for Telemetry Ingestion
* **Status**: Accepted
* **Context**: Real-time log ingestion requires low latency, bi-directional heartbeats, sequence acknowledgements, and explicit error feedback.
* **Alternatives Considered**:
  - Pure gRPC / HTTP2.
  - Raw custom TCP binary protocol.
  - Standard HTTP POST polling.
* **Decision**: JSON-RPC 2.0 over WebSockets (`GET /api/v1/rpc/ws`).
* **Rationale**: JSON-RPC 2.0 is lightweight, well-specified, human-readable, easily debuggable, and provides clear request/response/notification semantics over WebSockets without requiring heavy code generators like protoc.

## ADR-003: HTMX + Alpine.js + Server-Rendered HTML over SPA
* **Status**: Accepted
* **Context**: Web application needs modern interactivity without single-page application (SPA) complexity.
* **Alternatives Considered**:
  - React / Next.js SPA.
  - Vue.js / Inertia.js.
* **Decision**: Server-rendered Go `html/template` with HTMX for dynamic partial replacement and Alpine.js for lightweight local UI state.
* **Rationale**: Reduces bundle footprint, simplifies server-client state synchronization, fits neatly into a single Go binary, and eliminates complex frontend build pipelines.

## ADR-004: Asynchronous Non-Blocking Client Worker & Bounded Ring Buffer
* **Status**: Accepted
* **Context**: Client applications must never block or crash if ConsoleHub telemetry server is unavailable or degraded.
* **Alternatives Considered**:
  - Synchronous WebSocket writes on `Print`/`Printf`/`Info` calls.
  - Unlimited memory buffer queue.
* **Decision**: Lock-free/bounded ring-buffer queue paired with an asynchronous background worker goroutine, circuit breaker, and exponential backoff with jitter.
* **Rationale**: Guarantees that `consolehub.Printf` or `log.Output` calls operate at in-memory speed (~nanoseconds), enforcing graceful degradation and zero application blocking if network is severed.
