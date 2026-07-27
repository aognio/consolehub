# Architecture Decision Records (ADRs)

## ADR 001: Embedded PocketBase Database
- **Decision**: Embed PocketBase (`github.com/pocketbase/pocketbase`) using SQLite.
- **Rationale**: Keeps ConsoleHub as a single self-contained binary while providing collection schema management and migrations.

## ADR 002: JSON-RPC 2.0 over WebSockets Ingestion
- **Decision**: Use JSON-RPC 2.0 over WebSockets (`GET /api/v1/rpc/ws`) as primary telemetry transport.
- **Rationale**: Well-specified, bi-directional, lightweight, with explicit sequence deduplication and error codes.

## ADR 003: HTMX + Alpine.js + Server Templates UI
- **Decision**: Go `html/template` with HTMX and Alpine.js.
- **Rationale**: Minimal bundle footprint, fast server rendering, and zero SPA build pipeline overhead.
