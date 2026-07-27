# Architectural & Operational Assumptions

1. **Single Binary Target**: ConsoleHub will be distributed as a single executable containing all static assets, SQLite driver, PocketBase, and Web templates.
2. **Environment & Host Deployment**: Remote host applications running agents have outbound network connectivity to the ConsoleHub WebSocket endpoint (`GET /api/v1/rpc/ws`).
3. **Database Concurrency**: PocketBase handles internal SQLite connection pooling and WAL mode for high concurrent read/write throughput during ingestion.
4. **Log Line Structure**: Log output lines are UTF-8 text strings or valid JSON payloads with maximum line byte limits capped at 256 KB.
5. **Time Synchronization**: Remote host clocks and ConsoleHub server clocks are synchronized via NTP (ISO 8601 UTC timestamps).
