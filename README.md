<p align="center">
  <img src="https://raw.githubusercontent.com/aognio/consolehub/main/assets/images/consolehub-logo.png" alt="ConsoleHub Logo" width="480">
</p>

<h1 align="center">ConsoleHub</h1>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.style=flat-square" alt="License"></a>
  <a href="Makefile"><img src="https://img.shields.io/badge/Build-Passing-brightgreen?style=flat-square" alt="Monorepo Build"></a>
</p>

**ConsoleHub** is a centralized web console and real-time telemetry ingestion platform designed specifically for monitoring long-running command-line applications, automation pipelines, background workers, and distributed microservices across remote host machines.

It packages real-time log streaming, historical execution retention, host lifecycle tracking, and multi-tenant isolation into a single self-contained Go binary powered by embedded **PocketBase (SQLite)**, **HTMX**, **Alpine.js**, and **Tailwind CSS**.

---

## 🌟 Key Features & Architecture Highlights

* **Self-Contained Single Binary**: Zero external daemon dependencies. Embedded PocketBase database engine with automatic schema migrations.
* **JSON-RPC 2.0 Ingestion over WebSockets**: Real-time telemetry protocol (`GET /api/v1/rpc/ws`) with client-side run identity, batching, sequence tracking, and heartbeats.
* **Production-Ready Go Client Library (`libraries/go/consolehub`)**:
  * **Mandatory Client-Side ULID**: Process runs use a 26-character sortable Crockford Base32 ULID (`ClientRunID`) generated on the client machine at startup.
  * **Metadata Auto-Detection**: Mandatory system `Hostname` and process `PID` auto-detection with optional `AppVersion` and `OSName` metadata fields.
  * **Non-Blocking Resilience**: Thread-safe bounded in-memory ring-buffer queue, automatic batching (250 lines / 50ms), exponential backoff with jitter, and circuit breaker (10s window after 5 failures).
  * **Standard Library Ergonomics**: Mechanical drop-in replacements for `fmt.Printf`, `log.Logger`, `io.Writer`, progress trackers (`Progress`), and interactive terminal prompts (`Prompt`, `SecretPrompt`, `Confirm`).
* **Structured API Key Authentication**:
  * Format: `sk_<base62(16_random_bytes)>_crc32_<base62(crc32_checksum)>` (e.g. `sk_3q2Z7x9P2R8LmNkJd4Hf6Y_crc32_2AB9XQ`).
  * **Argon2id Hashing**: Raw API keys are displayed **once** upon creation with a prominent warning and stored exclusively as Argon2id hashes (`$argon2id$v=19$m=65536,t=1,p=4$...`).
  * Structural CRC32 checksum verification performed before database querying.
* **Multi-Tenant & Multi-Host Management**:
  * Global Host Registry (`ch_hosts`) with many-to-many Tenant Join associations (`ch_host_tenants`).
  * Full isolation for Tenants, Applications, Process Runs, and Log Line Streams.
  * Web console featuring active tenant switching, live SSE terminal streaming console, application registry, user administration, and API key management.

---

## 📁 Monorepo Layout

```text
consolehub/
├── server/                     # Backend server binary (Go + Embedded PocketBase + HTMX UI)
│   ├── cmd/consolehub/         # Main entrypoint, CLI flag parsing, and server boot
│   ├── internal/               # Core packages
│   │   ├── api/jsonrpc/        # JSON-RPC 2.0 over WebSocket protocol handler
│   │   ├── apikey/             # Structured API key generator & CRC32 validator
│   │   ├── auth/               # Bcrypt password & Argon2id API key hashing
│   │   ├── config/             # TOML configuration loader & environment parsing
│   │   ├── middleware/         # Session authentication & tenant scoping middleware
│   │   ├── models/             # Go domain structs (Tenant, Host, App, Run, APIKey, User, etc.)
│   │   ├── services/           # Core business logic & database services
│   │   ├── storage/            # PocketBase embedded database wrapper & schema migration
│   │   ├── stream/             # Real-time PubSub hub (stream.Hub) for SSE browser broadcast
│   │   ├── templates/          # Embedded HTML templates (html/template)
│   │   └── ui/                 # Web console HTTP page handlers & form processing
│   ├── consolehub.service      # Production systemd unit configuration
│   └── Makefile                # Server build & test targets
├── libraries/                  # Client Ingestion Libraries
│   └── go/consolehub/          # Go client agent library (libraries/go/consolehub)
│       ├── config/             # Environment & host metadata auto-detection
│       ├── events/             # Typed protocol event structures
│       ├── progress/           # Real-time local & remote progress bar trackers
│       ├── prompt/             # Interactive console input prompts
│       ├── protocol/           # JSON-RPC 2.0 frame definition
│       ├── queue/              # Thread-safe bounded ring-buffer queue
│       ├── transport/          # WebSocket & Mock transport interfaces
│       ├── ulid/               # 26-char Crockford Base32 ULID generator
│       ├── worker/             # Asynchronous worker goroutine, batching & circuit breaker
│       ├── writer/             # io.Writer stream wrappers
│       └── README.md           # Go client library documentation
├── demos/                      # Sample Ingestion Demo Applications
│   ├── go-demo/                # Executable Go demo application
│   └── python-demo/            # Executable Python telemetry script
├── docs/                       # Comprehensive Project Documentation Suite
│   ├── jsonrpc-websocket.md    # JSON-RPC 2.0 over WebSocket protocol spec
│   ├── go-client.md            # Go client library overview & architecture
│   ├── go-client-api.md        # Go client library API reference
│   ├── go-client-architecture.md# Threading, queuing & circuit breaker mechanics
│   ├── go-client-migration.md   # Standard library migration guide
│   ├── go-client-testing.md    # Unit testing & mock transport guide
│   └── go-client-roadmap.md    # Future client features roadmap
├── skills/                     # AI Agent Skills
│   ├── consolehub/             # General architecture & JSON-RPC protocol skill
│   └── consolehub-golang/      # Go client library API & usage skill
├── .development/               # Architectural Decision Records (ADRs) & journal
├── AGENTS.md                   # AI Agent guidance & repository conventions
└── Makefile                    # Top-level orchestration Makefile
```

---

## 🚀 Quickstart

### Prerequisites
* **Go**: `1.25` or higher installed.

### 1. Build Monorepo Targets
Build the backend server binary (`bin/consolehub-server`) and the Go demo application:
```bash
make build
```

### 2. Run Monorepo Unit & Integration Tests
```bash
make test
```

### 3. Run Server Locally
```bash
make run-server
```
Navigate to `http://localhost:3787` in your browser.

Default Super Admin Credentials:
* **Email**: `admin@consolehub.local`
* **Password**: `admin123456`

### 4. Install Server Binary to System Path
```bash
sudo make install
```

---

## 🛠️ Go Client Library (`libraries/go/consolehub`)

The official Go client library provides a thread-safe, non-blocking way to instrument any Go application or worker process.

### Quick Code Example

```go
package main

import (
    "time"
    "consolehub/libraries/go/consolehub"
)

func main() {
    // Flush remaining buffer & notify server exit on main return
    defer consolehub.Close()

    // Print text lines to stdout stream
    consolehub.Println("Starting data ingestion worker...")
    consolehub.Printf("Processing batch %d of %d\n", 1, 100)

    // Log structured event
    consolehub.Info("Worker initialized successfully")

    // Progress bar tracking
    p := consolehub.Progress("Downloading payload", 100)
    for i := 0; i <= 100; i += 25 {
        p.Set(int64(i))
        time.Sleep(50 * time.Millisecond)
    }
    p.Done()
}
```

### Standard Library Ergonomics Replacement

| Standard Library API | ConsoleHub Drop-in Replacement | Stream / Channel |
|---|---|---|
| `fmt.Print(...)` | `consolehub.Print(...)` | `"stdout"` |
| `fmt.Printf(...)` | `consolehub.Printf(...)` | `"stdout"` |
| `fmt.Println(...)` | `consolehub.Println(...)` | `"stdout"` |
| `log.Println(...)` | `consolehub.Info(...)` | `"log"` (Level: info) |
| `os.Stdout` | `consolehub.Stdout()` | `io.Writer` |
| `os.Stderr` | `consolehub.Stderr()` | `io.Writer` |

For complete client library details, see [libraries/go/consolehub/README.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/libraries/go/consolehub/README.md).

---

## 🔑 Structured API Key Authentication

ConsoleHub generates self-verifying, high-entropy API keys:

$$\text{Key Format}: \quad \mathtt{sk\_}\langle\text{Base62}(16\text{ random bytes})\rangle\mathtt{\_crc32\_}\langle\text{Base62}(\text{CRC32})\rangle$$

```text
Example Key: sk_3q2Z7x9P2R8LmNkJd4Hf6Y_crc32_2AB9XQ
```

### Key Security Flow
1. **Creation**: Generated via `/api-keys` in the Web Console.
2. **One-Time Display**: Displayed **once** in a copyable text box with a prominent warning banner.
3. **Argon2id Hashing**: Stored strictly as an Argon2id hash (`$argon2id$v=19$m=65536,t=1,p=4$...`).
4. **Fast Validation**: Verification decodes Base62 entropy and checks the CRC32 checksum before validating the Argon2id hash.

---

## 📚 Documentation Index

* **Protocol Specification**: [docs/jsonrpc-websocket.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/jsonrpc-websocket.md)
* **Go Client Overview**: [docs/go-client.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/go-client.md)
* **Go Client API Reference**: [docs/go-client-api.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/go-client-api.md)
* **Go Client Architecture & Resiliency**: [docs/go-client-architecture.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/go-client-architecture.md)
* **Standard Library Migration**: [docs/go-client-migration.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/go-client-migration.md)
* **Testing & Mock Transports**: [docs/go-client-testing.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/docs/go-client-testing.md)
* **Client Library README**: [libraries/go/consolehub/README.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/libraries/go/consolehub/README.md)
* **AI Agent Guidance**: [AGENTS.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/AGENTS.md)
* **Coding Conventions & Deployment Boundaries**: [.development/conventions.md](file:///home/gnrfan/code/experiments/by-language/golang/consolehub/.development/conventions.md)

---

## 📄 License

ConsoleHub is released under the **MIT License**.
