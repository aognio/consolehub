# Performance Targets & Optimization Guidelines

## Benchmarks & Limits
- **Ingestion Throughput**: Capable of ingesting up to 10,000 stream lines/sec across active host processes.
- **Batching Parameters**: Standard batch limits set to 250 lines or 256 KB per frame.
- **In-Memory PubSub Hub**: Ring buffer fan-out for SSE browser viewers to prevent memory allocation spikes.
- **Async Database Writes**: Stream lines are broadcast to active UI subscribers synchronously via memory channels and flushed asynchronously in batches to PocketBase SQLite storage.
