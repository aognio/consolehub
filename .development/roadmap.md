# ConsoleHub Roadmap

## v0.1 - Foundation & Infrastructure
- [ ] Directory layout & configuration module (`internal/config`).
- [ ] Embedded PocketBase initialization & collection schema migrations (`internal/storage`).
- [ ] Custom session-based authentication engine & cookie management (`internal/auth`).
- [ ] RBAC & Multi-tenancy isolation middleware (`internal/middleware`).

## v0.2 - Real-Time Ingestion & Streaming
- [ ] JSON-RPC 2.0 over WebSocket ingestion transport (`internal/api/jsonrpc`).
- [ ] HTTP REST batch ingestion fallback handlers (`internal/api/rest`).
- [ ] Real-time PubSub stream hub & SSE broadcaster (`internal/stream`).
- [ ] Monotonic sequence tracking & deduplication engine.

## v0.3 - UI Engine & Core Pages
- [ ] Embed template system (`internal/templates`, HTMX, Alpine.js, Tailwind CSS).
- [ ] Theme switcher (Light/Dark mode stored in browser cookies).
- [ ] Dashboard view (Metrics cards: Running processes, Offline/Online hosts, Recent runs, Failures).
- [ ] Tenant management, Host management, App directory, and Run history views.

## v0.4 - Flagship Console Viewer & Global Search
- [ ] Monospace real-time Console Viewer with SSE integration.
- [ ] Auto-scroll, pause/resume, tail mode, text search, line copy, timestamp jump, raw download.
- [ ] Plain text & collapsible JSON lines rendering.
- [ ] Global search engine across stream lines.

## v1.0 - Production Readiness & Documentation
- [ ] User administration & permission management.
- [ ] Full end-to-end integration test suite.
- [ ] Complete documentation suite (`docs/*.md`).

## Future Ideas & Beyond
- Remote process control (`process.stop`, `process.pause`, `process.resume`).
- Dynamic log level control & alerting webhooks.
