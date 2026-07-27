# ConsoleHub Ingestion API

ConsoleHub provides telemetry ingestion endpoints for client agents monitoring process executions.

## Real-Time JSON-RPC 2.0 Endpoint
- **URL**: `GET /api/v1/rpc/ws`
- **Protocol**: JSON-RPC 2.0 over WebSockets
- Full specification details are documented in [docs/jsonrpc-websocket.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/jsonrpc-websocket.md).

### Procedures Summary
- `auth.authenticate`: Client authentication token handshake.
- `process.register`: Registers process execution with `client_run_id` idempotency.
- `stream.append`: Transmits ordered telemetry log line batches with sequence deduplication.
- `process.heartbeat`: Health check heartbeat update.
- `stream.resume`: Queries sequence progress upon reconnection.
- `process.finish`: Reports process completion and exit code.

## Server-Sent Events (SSE) Live Stream Endpoint
- **URL**: `GET /api/v1/runs/live?run_id={id}`
- Streams real-time line telemetry directly to web UI Console Viewers.
