# Changelog

All notable changes to the ConsoleHub monorepo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.2] - 2026-07-27

### Added
- **Unauthenticated `healthz` JSON-RPC Procedure**:
  - Added an unauthenticated `healthz` (and `system.healthz`) JSON-RPC 2.0 procedure returning server status (`"ok"`), server version (`"v0.1.2"`), and timestamp.
- **Tenant Management JSON-RPC Procedures**:
  - Added authenticated `tenant.info` procedure returning full metadata for the authenticated tenant.
  - Added authenticated `tenant.app_list` procedure returning all applications associated with the tenant.
- **Go Client Library Protocol Extensions**:
  - Added `MethodHealthz`, `MethodTenantInfo`, `MethodTenantAppList` constants and `HealthzResult`, `TenantInfoResult`, `TenantAppListResult` structs to `libraries/go/consolehub/protocol/protocol.go`.
- **Global Telemetry Disable Control**:
  - Added global helper functions `consolehub.Disable()`, `consolehub.Enable()`, `consolehub.SetDisabled(bool)`, and `consolehub.IsDisabled()` to globally toggle sending telemetry messages while preserving local standard terminal output.
  - Added dynamic instance methods `client.SetDisabled(bool)` and `client.IsDisabled()` on `*consolehub.Client`.
  - Expanded `CONSOLEHUB_DISABLED` environment variable recognition to accept `"true"`, `"1"`, or `"yes"`.
- **Automatic Host Machine Registration & Tenant Association**:
  - Telemetry clients connecting from new host machines now automatically register the host machine (`slug = hostname`) and associate it with the registering tenant, allowing clients to ingest telemetry out-of-the-box without requiring manual host pre-creation.
- **Web Console Process Runs Filtering & ViewModel Resolution**:
  - Filtered `/runs` page by active tenant cookie selection so runs are displayed per tenant.
  - Resolved `App` and `Host` models in `RunViewModel` so application names and hostnames are cleanly rendered on `/runs`.
- **Structured JSONL Server Logging**:
  - Added structured JSON Lines (`.jsonl`) logging package (`server/internal/logger`).
  - Added `[logging]` TOML configuration options (`log_path = "/var/log/consolehub/consolehub.jsonl"` / `log_file`, `log_level = "debug"`).
  - Logs HTTP requests, WebSocket connection events, JSON-RPC procedures (`process.register`, `stream.append`, `auth.authenticate`), and errors to the configured `.jsonl` log file (with automatic fallback to `./consolehub.log` if target directory permissions fail).
- **Documentation & Agent Skill Updates**:
  - Updated `docs/jsonrpc-websocket.md` and `skills/consolehub/SKILL.md` with procedure references and payload schemas.

---

## [0.1.1] - 2026-07-27

### Added
- **Many-to-Many Host-Tenant Associations**:
  - Hosts can now be associated with multiple tenants simultaneously.
  - Telemetry agents automatically associate registered host machines with active tenant IDs during `process.register` JSON-RPC ingestion.
- **Interactive Multi-Select & Auto-Complete UI**:
  - Added an interactive auto-complete search textbox to filter system tenants dynamically on the `/hosts` Web Console.
  - Added removable tag pills with **"✕"** buttons to easily remove associated tenants.
  - Added checkbox multi-select dropdown for assigning host tenants in real-time.
  - Added **Associate Tenants** multi-select field to the *Register Host Machine* modal.
- **Strict API Key Tenant Scoping**:
  - Enforced API key tenant isolation during `process.register` JSON-RPC ingestion (`ErrCodeUnauthorized` returned if an API key attempts to register apps or runs under a different tenant).
  - Pre-selected active tenant in Web UI when generating API keys and displayed tenant names on the `/api-keys` management page.
- **Host Slug Ingestion Validation**:
  - Enforced that the hostname/slug sent during `process.register` JSON-RPC ingestion must match an existing host slug associated with the target tenant.
- **Root Repository Changelog**: Introduced `CHANGELOG.md` tracking all release milestones and patch details.

### Fixed
- **Host Visibility Filter**: Fixed an issue where newly created hosts with zero initial tenant associations were hidden when filtering by active tenant.
- **Alpine.js Template Safety**: Converted inline `@click` template bindings in `hosts.html` to HTML dataset attributes (`$el.dataset`), preventing JavaScript syntax errors on hostnames with special characters.
- **Fallback Host Slugs**: Added auto-fallback to default `slug = hostname` when registering hosts with empty slugs.

---

## [0.1.0] - 2026-07-27

### Added
- **ConsoleHub Core Monorepo**: Initial release of the ConsoleHub telemetry backend server, web console, and Go ingestion client library (`github.com/aognio/consolehub/libraries/go/consolehub`).
- **Telemetry Ingestion Engine**:
  - JSON-RPC 2.0 over WebSocket endpoint (`/api/v1/rpc/ws`) for host and process telemetry ingestion.
  - Real-time SSE PubSub stream hub (`/api/v1/stream`) broadcasting log lines and metrics to browser clients.
- **Argon2id API Key Security**:
  - Secure API key generation formatted as `sk_<random_bytes>_crc32_<checksum>`.
  - One-time display of raw keys with Argon2id hash storage.
- **Embedded Web Console**:
  - HTMX and Alpine.js UI for managing Tenants, Hosts, Apps, Process Runs, API Keys, and Users.
  - Binary versioning (`v0.1.0`) with CLI `-version` / `--version` flags and `-ldflags` build injection.
