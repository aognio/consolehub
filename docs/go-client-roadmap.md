# ConsoleHub Go Client Roadmap & Future Extensibility

This document outlines upcoming capabilities and future extension points planned for the ConsoleHub Go client library.

---

## 1. Upcoming Features

- **HTTP Fallback Ingestion Transport**: Support HTTP POST batching fallback when WebSockets are blocked by corporate proxies.
- **Remote Control Receivers**: Implementation of server control notifications (`client.set_log_level`, `process.stop`, `process.pause`).
- **OpenTelemetry Bridge**: Adapter package converting OpenTelemetry log records into ConsoleHub `events.LogEvent` instances.
- **Compression**: Gzip/Snappy compression support for high-throughput batching frames.
