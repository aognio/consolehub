---
name: consolehub-golang
description: Go-specific implementation guidance, public API ergonomics, mandatory client-side ULID (ClientRunID), mandatory metadata (Hostname, PID) & optional metadata (AppVersion, OSName), stdout/stderr writers, progress bars, interactive prompts, structured logging, and migration patterns for ConsoleHub Go Client.
---

# ConsoleHub Go Client Implementation & Usage Guide

This skill provides Go-specific guidance for integrating and using the `consolehub` Go client library in applications.

---

## 1. Quick Start & Mandatory / Optional Metadata

Import `consolehub` package:
```go
import "github.com/aognio/consolehub/libraries/go/consolehub"
```

### Metadata Transmission & ULID Rules:
* **Mandatory Client-Generated ULID**:
  * `ClientRunID`: A 26-character canonical Crockford Base32 ULID generated on the client machine at startup (auto-generated via `ulid.Make()`, or overridden via `consolehub.WithClientRunID(...)`).
* **Mandatory Metadata**:
  * `Hostname`: System hostname (auto-detected via `os.Hostname()` if omitted, or overridden via `consolehub.WithHostname(...)`).
  * `PID`: Process ID (auto-detected via `os.Getpid()` if omitted, or overridden via `consolehub.WithPID(...)`).
* **Optional Metadata**:
  * `AppVersion`: Application version string (set via `consolehub.WithAppVersion("1.4.2")`).
  * `OSName`: Operating System name (auto-detected via `runtime.GOOS`, or overridden via `consolehub.WithOSName("linux")`).

### Option A: Standard Global Default Client
Auto-configures from environment variables (`CONSOLEHUB_ENDPOINT`, `CONSOLEHUB_TOKEN`, `CONSOLEHUB_TENANT`, `CONSOLEHUB_APP`, `CONSOLEHUB_DISABLED`).

```go
func main() {
    defer consolehub.Close()

    // Optionally disable telemetry transmission globally
    // consolehub.Disable()

    consolehub.Println("Application starting up...")
}
```

### Globally Disabling Telemetry Transmission
You can disable sending telemetry/log messages to ConsoleHub while retaining standard local stdout/stderr terminal printing:

* **Environment Variable**: Set `CONSOLEHUB_DISABLED=true` (or `1`, `yes`).
* **Global Helpers**:
  ```go
  consolehub.Disable()                 // Disables forwarding replacement functions to ConsoleHub
  consolehub.Enable()                  // Re-enables forwarding replacement functions to ConsoleHub
  consolehub.SetDisabled(true)         // Set global disabled state
  disabled := consolehub.IsDisabled() // Check global disabled state
  ```
* **Client Instance Methods**:
  ```go
  client.SetDisabled(true)            // Toggle telemetry transmission dynamically on client instance
  disabled := client.IsDisabled()
  ```

### Option B: Custom Explicit Client
```go
client, err := consolehub.New(
    consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
    consolehub.WithTenant("engineering"),
    consolehub.WithApp("telegram-replicator"),
    consolehub.WithToken("client-secret-token"),
    // Mandatory client-side ULID (auto-generated if omitted)
    consolehub.WithClientRunID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
    // Mandatory overrides (optional if auto-detection is desired)
    consolehub.WithHostname("vps-prod-01"),
    consolehub.WithPID(42817),
    // Optional metadata
    consolehub.WithAppVersion("1.4.2"),
    consolehub.WithOSName("linux"),
)
if err != nil {
    log.Fatalf("failed to initialize consolehub: %v", err)
}
defer client.Close()
```

---

## 2. Public API Surface

### 2.1 Standard Output Replacements (`fmt` style)
Replaces `fmt.Print*` calls mechanically for real-time telemetry streaming:

```go
consolehub.Print("Starting worker...")
consolehub.Printf("Processing item %d of %d\n", i, total)
consolehub.Println("Done.")
```

### 2.2 Direct `io.Writer` Access
Access standard streams or wrap custom loggers/writers:

```go
fmt.Fprintln(consolehub.Stdout(), "Writing directly to stdout stream")
fmt.Fprintln(consolehub.Stderr(), "Writing directly to stderr stream")

// Wrap standard Go log package
logger := log.New(consolehub.Stdout(), "[APP] ", log.LstdFlags)
logger.Println("Log message routed through ConsoleHub")

// Multi-writer (local terminal + ConsoleHub)
mw := io.MultiWriter(os.Stdout, consolehub.Stdout())
```

### 2.3 Structured Logging (`log` style)
```go
consolehub.Debug("Initializing storage...")
consolehub.Infof("Server listening on port %d", 8080)
consolehub.Warn("High CPU usage detected")
consolehub.Errorf("Database query failed: %v", err)
```

### 2.4 Progress Trackers
First-class progress rendering both locally and via protocol stream:

```go
p := consolehub.Progress("Downloading media file", 100)
for i := 0; i <= 100; i += 20 {
    p.Set(int64(i))
    time.Sleep(100 * time.Millisecond)
}
p.Done()

// Upload tracker
up := consolehub.UploadProgress("Uploading payload.tar.gz", 10485760)
up.Add(102400)
up.Finish()
```

### 2.5 Interactive Console Prompts
```go
// Text Prompt
channel := consolehub.Prompt("Enter target Telegram channel", "@mychannel")

// Masked Secret Prompt
apiKey := consolehub.SecretPrompt("Enter API Key")

// Confirmation Prompt
if consolehub.Confirm("Proceed with database migration?", true) {
    // perform migration
}

// Choice Selection
env := consolehub.Choice("Select environment", []string{"Development", "Staging", "Production"}, "Development")
```

---

## 3. Low-Level Protocol Package & JSON-RPC Procedure Constants

Import low-level protocol definitions:
```go
import "github.com/aognio/consolehub/libraries/go/consolehub/protocol"
```

### Complete JSON-RPC Procedure Constants:
| Procedure Constant | JSON-RPC Method Name | Description | Required Auth |
| :--- | :--- | :--- | :--- |
| `protocol.MethodHealthz` | `"healthz"` | System status check (`HealthzResult`) | Unauthenticated |
| `protocol.MethodAuthAuthenticate` | `"auth.authenticate"` | Session authentication (`AuthParams`, `AuthResult`) | Unauthenticated |
| `protocol.MethodTenantInfo` | `"tenant.info"` | Tenant metadata (`TenantInfoParams`, `TenantInfoResult`) | Authenticated |
| `protocol.MethodTenantAppList` | `"tenant.app_list"` | Application listing (`TenantAppListParams`, `TenantAppListResult`) | Authenticated |
| `protocol.MethodProcessRegister` | `"process.register"` | Run registration (`ProcessRegisterParams`, `ProcessRegisterResult`) | Authenticated |
| `protocol.MethodStreamAppend` | `"stream.append"` | Batch stream ingestion (`StreamAppendParams`, `StreamAppendResult`) | Authenticated |
| `protocol.MethodStreamResume` | `"stream.resume"` | Reconnection replay (`StreamResumeParams`, `StreamResumeResult`) | Authenticated |
| `protocol.MethodProcessFinish` | `"process.finish"` | Execution completion (`ProcessFinishParams`) | Authenticated |
| `protocol.MethodProcessHeartbeat` | `"process.heartbeat"` | Keepalive ping (`connection.ping` / `process.heartbeat`) | Authenticated |
