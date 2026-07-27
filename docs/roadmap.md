# ConsoleHub Roadmap

## Version 1.0 (Current Release)
- Self-contained binary compilation with embedded PocketBase storage.
- TOML configuration loader (`--config /path/to/server-config.toml`).
- Independent session authentication surviving binary rebuilds.
- Multi-tenancy isolation and RBAC (`Super Admin`, `Admin`, `User`).
- JSON-RPC 2.0 over WebSocket telemetry ingestion (`GET /api/v1/rpc/ws`).
- Flagship Console Viewer with real-time SSE streaming, auto-scroll, tailing, pause/resume, and expandable JSON rendering.
- Light and Dark mode UI theme switcher with cookie persistence.

## Future Versions
- Client agent CLI (`consolehub-agent`).
- Remote process control (`process.stop`, `process.pause`, `process.resume`).
- Anomaly alerts and webhook integrations.
