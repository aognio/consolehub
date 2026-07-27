# Testing Strategy

## Strategy & Principles
- Follow Test-Driven Development (TDD) for business logic (`Red -> Green -> Refactor`).
- Unit tests cover configuration parsing, authentication token generation, password hashing, sequence deduplication, and JSON-RPC dispatching.
- Integration tests verify PocketBase storage repositories, HTTP API handlers, and WebSocket ingestion sessions using `httptest.Server` and WebSocket test clients.

## Test Commands
```bash
# Run unit tests
go test -v -race ./...

# Run integration tests
go test -v -tags=integration ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```
