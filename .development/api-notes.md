# API Design Notes

## Real-Time JSON-RPC 2.0 Endpoint
- **URL**: `GET /api/v1/rpc/ws`
- Full spec detailed in [docs/jsonrpc-websocket.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/jsonrpc-websocket.md).

## REST Ingestion Endpoints (Fallback)
- `POST /api/v1/runs/register`
- `POST /api/v1/runs/{id}/heartbeat`
- `POST /api/v1/runs/{id}/stream`
- `POST /api/v1/runs/{id}/finish`
- `GET /api/v1/runs/live` (SSE stream for live UI)
