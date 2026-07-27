# ConsoleHub System Architecture

ConsoleHub is a centralized web console for monitoring long-running command-line applications across remote hosts. It is delivered as a single self-contained Go binary with embedded PocketBase storage and web UI assets.

```
                  +-----------------------------------+
                  | Console Client / Host Agent       |
                  +-----------------+-----------------+
                                    |
                                    | JSON-RPC 2.0 / WS (GET /api/v1/rpc/ws)
                                    v
+-------------------------------------------------------------------------+
| ConsoleHub Binary                                                       |
|                                                                         |
|  +---------------------------+       +-------------------------------+  |
|  | JSON-RPC WS Adapter       |       | HTML / HTMX Web UI Handlers   |  |
|  | (internal/api/jsonrpc)    |       | (internal/ui)                 |  |
|  +-------------+-------------+       +---------------+---------------+  |
|                |                                     |                  |
|                v                                     v                  |
|  +-------------------------------------------------------------------+  |
|  | Domain Services Layer (internal/services)                         |  |
|  +-------------+------------------------------------+----------------+  |
|                |                                    |                   |
|                v                                    v                   |
|  +---------------------------+    +----------------------------------+  |
|  | Real-Time PubSub Hub      |    | PocketBase SQLite Storage        |  |
|  | (internal/stream)         |    | (internal/storage)               |  |
|  +---------------------------+    +----------------------------------+  |
+-------------------------------------------------------------------------+
```

## Package Responsibilities

* `cmd/consolehub`: CLI entrypoint, flag parsing (`--config /path/to/server-config.toml`), server startup.
* `internal/config`: TOML configuration loader and defaults.
* `internal/models`: Go struct definitions for domain entities.
* `internal/storage`: PocketBase embedded database DAO and schema initialization.
* `internal/auth`: Session HMAC token signing, validation, and bcrypt password hashing.
* `internal/services`: Core domain business logic (Tenant, Host, App, Run, StreamLine, User).
* `internal/stream`: Real-time PubSub hub for live SSE browser streaming.
* `internal/api/jsonrpc`: JSON-RPC 2.0 over WebSocket telemetry ingestion transport handler.
* `internal/middleware`: Session authentication and context injection.
* `internal/templates`: Embedded HTML templates (`html/template`).
* `internal/ui`: Web UI page handlers.
