# Architecture Overview

## System Layout

ConsoleHub is designed as a single self-contained binary built with Go, embedding PocketBase (SQLite), HTML templates, HTMX, Alpine.js, and Tailwind CSS.

```
[ Console Client / Agent ]
         |
         | JSON-RPC 2.0 / WebSockets (GET /api/v1/rpc/ws)
         v
+--------------------------------------------------------+
|                      ConsoleHub                        |
|                                                        |
|  +--------------------------------------------------+  |
|  | Ingestion Transport Adapter                      |  |
|  | (internal/api/jsonrpc)                           |  |
|  +------------------------+-------------------------+  |
|                           |                            |
|                           v                            |
|  +--------------------------------------------------+  |
|  | Domain Services Layer                            |  |
|  | (RunService, HostService, StreamService, etc.)   |  |
|  +------------+-------------------------+-----------+  |
|               |                         |              |
|               v                         v              |
|  +------------------------+  +----------------------+  |
|  | In-Memory Stream Hub   |  | PocketBase Storage   |  |
|  | (Pub/Sub for SSE/WS)   |  | (SQLite DAO)         |  |
|  +-----------+------------+  +----------------------+  |
|              |                                         |
|              v                                         |
|  +--------------------------------------------------+  |
|  | Web UI Layer (HTMX + Alpine.js + html/template)  |  |
|  +--------------------------------------------------+  |
+--------------------------------------------------------+
```

## Package Responsibilities

* `cmd/consolehub`: CLI entrypoint, flag parsing, TOML config loading, PocketBase initialization.
* `internal/config`: TOML configuration parsing and validation.
* `internal/models`: Pure Go structs for domain entities.
* `internal/storage`: PocketBase repository wrappers & DAO migrations.
* `internal/auth`: Session auth, password hashing, and token validation.
* `internal/services`: Core business logic rules.
* `internal/stream`: Transport interfaces & real-time PubSub hub.
* `internal/api/jsonrpc`: JSON-RPC 2.0 over WebSocket transport adapter.
* `internal/api/rest`: HTTP REST fallback endpoints.
* `internal/middleware`: Session auth, RBAC authorization, tenancy scoping.
* `internal/ui`: Web UI HTTP handlers.
* `internal/templates`: Embedded HTML templates and HTMX partials.
