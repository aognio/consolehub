# Agentic Coding Guidelines for ConsoleHub (Monorepo)

Welcome! This document provides guidelines, architectural layout, and commands for AI agents operating on the ConsoleHub repository.

---

## 1. Monorepo Structure

* `server/`: Centralized web console server (Go binary + embedded PocketBase + HTMX UI).
  * `server/cmd/consolehub/main.go`: Entry point, flag parsing (`--config`), server boot.
  * `server/internal/config/`: TOML configuration loader.
  * `server/internal/models/`: Go domain structs (`Tenant`, `Host`, `App`, `Run`, `StreamLine`, `User`, `TenantMember`).
  * `server/internal/storage/`: PocketBase embedded database wrapper & schema initialization.
  * `server/internal/auth/`: Password hashing (bcrypt) and session token signing (HMAC-SHA256).
  * `server/internal/services/`: Core business logic rules (Tenant, Host, App, Run, Stream, User).
  * `server/internal/stream/`: Real-time PubSub hub (`stream.Hub`) for SSE broadcast to browser UI.
  * `server/internal/api/jsonrpc/`: Real-time JSON-RPC 2.0 over WebSocket transport adapter (`GET /api/v1/rpc/ws`).
  * `server/internal/middleware/`: Authentication and context scoping middleware.
  * `server/internal/templates/`: Embedded HTML templates (`html/template`).
  * `server/internal/ui/`: Web UI HTTP page handlers.
* `libraries/`: Ingestion agent client libraries (`libraries/go/consolehub`).
* `demos/`: Sample applications showcasing live telemetry ingestion (`demos/go-demo`, `demos/python-demo`).
* `docs/`: Comprehensive project documentation suite (`docs/jsonrpc-websocket.md`, `docs/plan.md`, etc.).
* `.development/`: Development tracking, ADR decision logs, and milestone notes.

---

## 2. Essential Commands

```bash
# Build monorepo targets (server & demo agent)
make build

# Run unit and integration tests across monorepo
make test

# Install server binary to /usr/local/bin
sudo make install

# Run server locally
make run-server

# Run Go demo agent
make run-demo-go

# Clean build artifacts
make clean
```

---

## 3. Development & Workflow Rules

1. **Git Discipline**: All active development takes place on the `wip` working branch. Commit frequently with small, logical changes.
2. **Development Workspace**: Keep `.development/` synchronized with all architectural changes, journal entries, decisions, and backlog updates.
3. **Test-Driven Development**: Maintain unit and integration tests under `*_test.go`. Ensure `make test` passes before declaring tasks completed.
4. **Thin Handlers**: Business logic belongs inside `server/internal/services`, not directly inside HTTP or WebSocket handlers.
5. **No Hardcoded Config**: Ports, hosts, data paths, and secrets must come from `server/internal/config`.
