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

The project follows clean architecture principles with thin HTTP/WebSocket handlers delegating to domain services.

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

1. **`users`**
   - `id` (text, primary key)
   - `email` (text, unique)
   - `password_hash` (text)
   - `name` (text)
   - `role` (text: `super_admin`, `admin`, `user`)
   - `active` (bool)
   - `created`, `updated`

2. **`tenants`**
   - `id` (text, primary key)
   - `name` (text)
   - `slug` (text, unique)
   - `active` (bool)
   - `created`, `updated`

3. **`tenant_members`**
   - `id` (text, primary key)
   - `tenant_id` (relation -> `tenants`)
   - `user_id` (relation -> `users`)
   - `role` (text: `admin`, `user`)
   - `created`, `updated`

4. **`hosts`**
   - `id` (text, primary key)
   - `tenant_id` (relation -> `tenants`)
   - `hostname` (text)
   - `display_name` (text)
   - `platform` (text)
   - `last_seen` (date/time)
   - `online` (bool)
   - `created`, `updated`

5. **`apps`**
   - `id` (text, primary key)
   - `tenant_id` (relation -> `tenants`)
   - `name` (text)
   - `display_name` (text)
   - `description` (text)
   - `created`, `updated`

6. **`runs`**
   - `id` (UUID text, primary key)
   - `client_run_id` (UUID text, unique index per tenant for deduplication)
   - `tenant_id` (relation -> `tenants`)
   - `host_id` (relation -> `hosts`)
   - `app_id` (relation -> `apps`)
   - `pid` (int)
   - `started_at` (date/time)
   - `finished_at` (date/time, optional)
   - `status` (text: `running`, `exited`, `crashed`, `stopped`)
   - `version` (text)
   - `working_directory` (text)
   - `command_line` (text)
   - `exit_code` (int)
   - `last_sequence` (int)
   - `created`, `updated`

7. **`stream_lines`**
   - `id` (text, primary key)
   - `run_id` (relation -> `runs`)
   - `tenant_id` (relation -> `tenants`)
   - `sequence` (int, monotonic sequence index)
   - `timestamp` (date/time)
   - `stream` (text: `stdout`, `stderr`, `log`)
   - `kind` (text: `text`, `json`)
   - `text` (text)
   - `created`

8. **`groups`** (Reserved for future permissions)
   - `id`, `name`, `created`, `updated`

---

## 5. Configuration (TOML Schema)

Configuration is loaded via `--config /path/to/server-config.toml`.

```toml
[server]
host = "0.0.0.0"
port = 8080
scheme = "http"
base_path = ""
public_url = "http://localhost:8080"

[security]
cookie_secret = "super-secret-random-32-byte-key-here"
secure_cookies = false
same_site = "lax"
session_duration = "24h"

[pocketbase]
data_dir = "./pb_data"

[logging]
level = "info"
retention_days = 30
```

---

## 6. Authentication & Authorization (RBAC & Multi-Tenancy)

### Independent Session Authentication
* Custom session store backed by SQLite/PocketBase, utilizing `security.cookie_secret` for cookie signature and encryption.
* Preserves user login state across binary rebuilds and server restarts.

### RBAC Hierarchy
1. **Super Admin**: Full platform access. Manage tenants, create admins/users, view all tenants.
2. **Admin**: Assigned to specific tenants. Manage tenant hosts/apps/runs, invite tenant users, view all logs inside tenant scope.
3. **User**: Read-only access within assigned tenants. View dashboards, search logs, watch live console streams.

---

## 7. JSON-RPC 2.0 over WebSockets Ingestion & Streaming Architecture

### Endpoint
* `GET /api/v1/rpc/ws`: Primary WebSocket endpoint handling JSON-RPC 2.0 calls.

### Connection & Ingestion Flow
1. WebSocket upgrade (supporting HTTP Bearer Token or `auth.authenticate` first call).
2. Process registration (`process.register`) with `client_run_id` for client-side idempotency.
3. Stream ingestion (`stream.append`) using monotonic sequences & batch deduplication.
4. Heartbeat tracking (`process.heartbeat`).
5. Reconnection & sequence alignment (`stream.resume`).
6. Execution completion (`process.finish`).

### Core JSON-RPC Procedures
- `auth.authenticate`
- `process.register`
- `stream.append`
- `process.heartbeat`
- `stream.resume`
- `process.finish`
- `process.status`
- `connection.ping`

Full details, error codes (`-32001` through `-32010`), and sequence flow diagrams are specified in [docs/jsonrpc-websocket.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/jsonrpc-websocket.md).

### Real-Time UI PubSub Hub
* In-memory event multiplexer (`stream.Hub`) receives line ingestions.
* Fan-outs to connected browser clients via Server-Sent Events (SSE) or WebSockets.
* Asynchronously batches inserts into PocketBase `stream_lines` storage for zero-drop performance under high ingestion throughput.

---

## 8. UI Structure & Flagship Console Viewer

### Navigation & Theme Support
* Responsive sidebar layout (Desktop, Tablet, Mobile).
* Light Mode / Dark Mode toggle saved to browser cookies (accessible prior to authentication).
* Dynamic tenant selector in top navbar for multi-tenant switching.

### Pages Overview
1. **Login**: Session auth entry point.
2. **Dashboard**: Stat cards (Running processes, Offline/Online hosts, Recent runs, Recent failures, Stream activity ticker).
3. **Tenants**: Tenant CRUD (Super Admin).
4. **Hosts**: Host table with online/offline indicators, OS/platform details.
5. **Apps**: Application directory and run metrics.
6. **Running Processes / Historical Runs**: Filterable execution grid (Status, App, Host, PID, Started, Duration, User, Actions).
7. **Flagship Console Viewer (`/runs/:id/console`)**:
   - Monospace terminal window with fast rendering.
   - Controls: Auto-scroll toggle, Pause/Resume stream, Tail mode, Text search within buffer.
   - Copy line to clipboard, Jump to timestamp, Raw download.
   - Formatting support: Plain text view & collapsible JSON lines view.
8. **Global Search**: Multi-facet filter (Text query, Time range, Host, App, Run, Tenant, Stream).
9. **User Administration**: User management & tenant assignment.
10. **Settings**: Configuration overview & data retention controls.

---

## 9. Documentation Roadmap & Deliverables

During implementation, the following documentation files will be created in `docs/`:

- `docs/architecture.md`: System layout, package structure, and transport details.
- `docs/data-model.md`: Schema definition for PocketBase collections and domain structs.
- `docs/api.md`: REST & JSON-RPC API specification for client ingestion.
- `docs/jsonrpc-websocket.md`: Full JSON-RPC 2.0 over WebSockets specification.
- `docs/security.md`: Authentication, session secret handling, RBAC, and multi-tenancy rules.
- `docs/ui.md`: HTMX/Alpine UI design patterns, layout structure, and Console Viewer components.
- `docs/roadmap.md`: Strategic feature timeline and upcoming extensions.
- `docs/backlog.md`: Known enhancements and task tracking.
- `docs/decisions.md`: Architecture Decision Records (ADRs).
- `docs/journal.md`: Work log and implementation progress notes.

---

## 10. Implementation Phases

1. **Phase 1: Foundation & Infrastructure**
   - Config loader (`internal/config`) for TOML parsing.
   - Embedded PocketBase setup & schema migrations (`internal/storage`).
   - Logging and core binary scaffolding (`cmd/consolehub`).

2. **Phase 2: Auth & Multi-Tenancy Engine**
   - Custom session engine (`internal/auth`) with cookie security.
   - Context-based tenant isolation & RBAC middleware (`internal/middleware`).
   - User administration services.

3. **Phase 3: Telemetry Ingestion & Real-Time Streaming**
   - JSON-RPC 2.0 over WebSockets transport adapter (`internal/api/jsonrpc`).
   - Ingestion service & HTTP fallback handlers (`internal/services`, `internal/api/rest`).
   - Streaming Hub & event pub/sub (`internal/stream`).

4. **Phase 4: Frontend Framework & UI Component System**
   - Embed HTML templates, HTMX, Alpine.js, Tailwind CSS bundle.
   - Base responsive layout, theme cookie handler (Light/Dark mode).

5. **Phase 5: Core Web Pages Implementation**
   - Dashboard, Tenants, Hosts, Apps, and Runs views.
   - Search page with multi-dimensional filtering.

6. **Phase 6: Flagship Console Viewer**
   - Real-time SSE connection to streaming hub.
   - Pause, resume, auto-scroll, tailing, copy, timestamp jump, and JSON expandable rendering.

7. **Phase 7: Verification & Documentation Suite**
   - End-to-end integration tests & build verification.
   - Full documentation suite (`docs/*.md`).
