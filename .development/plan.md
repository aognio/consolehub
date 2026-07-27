# ConsoleHub Implementation Plan

## 1. Executive Summary & Goals

ConsoleHub is a centralized web console for monitoring long-running command-line applications across multiple remote hosts. It provides real-time streaming, historical log retention, host lifecycle tracking, and multi-tenant isolation in a single self-contained Go binary embedded with PocketBase.

---

## 2. Technology Stack & Operational Model

* **Language**: Go 1.22+ (latest stable)
* **Storage & DB**: Embedded PocketBase (`github.com/pocketbase/pocketbase`) using SQLite backend.
* **Frontend**: Server-rendered HTML templates (`html/template`) enhanced with **HTMX** (dynamic partial updates) and **Alpine.js** (lightweight UI state management).
* **Styling**: Tailwind CSS (compiled or CDN bundle with dark/light mode palette).
* **Ingestion Protocol**: **JSON-RPC 2.0 over WebSockets** (`GET /api/v1/rpc/ws`) as primary real-time ingestion transport, plus HTTP batch API fallback.
* **Configuration**: Loaded via TOML (`--config /path/to/server-config.toml`).
* **Deployment**: Single self-contained binary containing all embedded static assets and templates (`embed.FS`).

---

## 3. Architecture & Package Topology

```
consolehub/
├── cmd/
│   └── consolehub/
│       └── main.go              # Entry point: flags, TOML load, PocketBase boot
├── internal/
│   ├── config/                  # TOML parser, server config structs & defaults
│   ├── models/                  # Domain models (Tenant, Host, App, Run, StreamLine, User)
│   ├── storage/                 # PocketBase DAO repositories & collection setup
│   ├── auth/                    # Independent session auth, token validation, password hashing
│   ├── services/                # Business logic (Tenant, Host, App, Run, Stream, User)
│   ├── stream/                  # Ingestion interfaces, PubSub hub, SSE & WS stream broadcasters
│   ├── api/                     # HTTP & JSON-RPC 2.0 over WebSocket handlers
│   │   ├── jsonrpc/             # JSON-RPC 2.0 transport adapter & procedure routers
│   │   └── rest/                # HTTP REST API handlers for fallback ingestion
│   ├── middleware/              # Auth, RBAC, Multi-tenant scoping, Logging middleware
│   ├── ui/                      # Web UI HTTP handlers rendering HTMX partials & pages
│   └── templates/               # Embed templates, layouts, component fragments
├── docs/                        # Project documentation & specs
│   └── jsonrpc-websocket.md     # Detailed JSON-RPC 2.0 WS protocol specification
└── static/                      # Embedded JS (HTMX, Alpine.js) and CSS assets
```

---

## 4. PocketBase Schema & Domain Models

### Collections Design

1. **`users`**: `id`, `email`, `password_hash`, `name`, `role` (`super_admin`, `admin`, `user`), `active`, `created`, `updated`
2. **`tenants`**: `id`, `name`, `slug`, `active`, `created`, `updated`
3. **`tenant_members`**: `id`, `tenant_id`, `user_id`, `role` (`admin`, `user`), `created`, `updated`
4. **`hosts`**: `id`, `tenant_id`, `hostname`, `display_name`, `platform`, `last_seen`, `online`, `created`, `updated`
5. **`apps`**: `id`, `tenant_id`, `name`, `display_name`, `description`, `created`, `updated`
6. **`runs`**: `id`, `client_run_id`, `tenant_id`, `host_id`, `app_id`, `pid`, `started_at`, `finished_at`, `status`, `version`, `working_directory`, `command_line`, `exit_code`, `last_sequence`, `created`, `updated`
7. **`stream_lines`**: `id`, `run_id`, `tenant_id`, `sequence`, `timestamp`, `stream`, `kind`, `text`, `created`
8. **`groups`**: `id`, `name`, `created`, `updated`

---

## 5. Implementation Phases

1. **Phase 1**: Foundation & Infrastructure (Config loader, PocketBase schema & initialization, CLI scaffolding).
2. **Phase 2**: Auth & Multi-Tenancy Engine (Independent session auth, RBAC & context tenancy middleware).
3. **Phase 3**: Telemetry Ingestion & Real-Time Streaming (JSON-RPC 2.0 over WS adapter, HTTP REST fallback, PubSub stream hub).
4. **Phase 4**: Frontend Framework & UI Component System (Embedded HTML templates, HTMX, Alpine.js, Tailwind CSS, Dark/Light mode cookies).
5. **Phase 5**: Core Web Pages (Dashboard, Tenants, Hosts, Apps, Runs, Search).
6. **Phase 6**: Flagship Console Viewer (SSE live streaming, pause/resume, tailing, auto-scroll, JSON expander).
7. **Phase 7**: Verification & Documentation Suite.
