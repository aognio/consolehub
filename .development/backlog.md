# ConsoleHub Backlog

## Critical
- [ ] TOML configuration loader supporting server, security, PocketBase, and logging parameters.
- [ ] PocketBase embedded integration and collection migrations.
- [ ] Session authentication surviving server restarts.
- [ ] JSON-RPC 2.0 over WebSocket ingestion endpoint (`GET /api/v1/rpc/ws`).

## High
- [ ] Tenant context isolation middleware.
- [ ] Real-time PubSub hub & SSE stream endpoint (`GET /api/v1/runs/live`).
- [ ] Flagship Console Viewer page with auto-scroll, tailing, pause/resume, and JSON rendering.
- [ ] Global stream search across multi-tenant scopes.

## Medium
- [ ] Light/Dark mode UI theme switcher (cookie-based).
- [ ] Dashboard metrics cards (Running processes, Host online/offline status, Failures).
- [ ] HTTP REST ingestion API fallback endpoints (`POST /api/v1/runs/...`).

## Low
- [ ] Stack trace grouping in Console Viewer.
- [ ] Raw stream log export/download feature.

## Nice to Have
- [ ] Future client daemon binary (`consolehub-agent`).
- [ ] Multi-region host distribution map.

## Completed
- [x] Initial project setup & plan generation (`docs/plan.md`, `docs/jsonrpc-websocket.md`).
- [x] Development workspace creation (`.development/`).
