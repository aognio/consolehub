# ConsoleHub Go Client API Specification

This document provides a reference for all public functions, methods, and options in the `consolehub` Go package.

---

## 1. Client Initialization & Configuration

### `New(opts ...Option) (*Client, error)`
Constructs a new ConsoleHub client instance.

```go
client, err := consolehub.New(
    consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
    consolehub.WithTenant("engineering"),
    consolehub.WithApp("my-service"),
    consolehub.WithToken("secret-token"),
)
```

### Functional Options
- `WithEndpoint(url string)`: Overrides the JSON-RPC WebSocket URL.
- `WithTenant(tenant string)`: Sets the target organization tenant ID or slug.
- `WithApp(app string)`: Sets the application identifier name.
- `WithToken(token string)`: Sets client authentication token.
- `WithHostname(hostname string)`: Overrides auto-detected hostname.
- `WithQueueCapacity(capacity int)`: Overrides buffer queue capacity (default 10,000).
- `WithDisabled(disabled bool)`: Runs in local-only mode without network transmission.
- `WithTransport(trans transport.Transport)`: Injects custom or mock transport.

---

## 2. Standard Output Functions (`fmt` style)

- `consolehub.Print(v ...any)`
- `consolehub.Printf(format string, v ...any)`
- `consolehub.Println(v ...any)`
- `consolehub.Fprint(w io.Writer, v ...any)`
- `consolehub.Fprintf(w io.Writer, format string, v ...any)`
- `consolehub.Fprintln(w io.Writer, v ...any)`
- `consolehub.Stdout() io.Writer`
- `consolehub.Stderr() io.Writer`
- `consolehub.Writer(streamName string) io.Writer`

---

## 3. Structured Logging Functions

- `consolehub.Debug(msg string)`
- `consolehub.Info(msg string)`
- `consolehub.Warn(msg string)`
- `consolehub.Error(msg string)`
- `consolehub.Debugf(format string, v ...any)`
- `consolehub.Infof(format string, v ...any)`
- `consolehub.Warnf(format string, v ...any)`
- `consolehub.Errorf(format string, v ...any)`
- `consolehub.Log(evt events.Event)`
- `consolehub.Report(evt events.Event)`

---

## 4. Progress Trackers

### `consolehub.Progress(label string, total int64) *progress.Tracker`
Creates a progress tracker rendering both locally and via protocol stream.

```go
p := consolehub.Progress("Processing tasks", 100)
p.Set(50)
p.Add(10)
p.Done()
```

---

## 5. Interactive Console Prompts

- `consolehub.Prompt(promptText, defaultVal string) string`
- `consolehub.SecretPrompt(promptText string) string`
- `consolehub.Confirm(promptText string, defaultVal bool) bool`
- `consolehub.Choice(promptText string, options []string, defaultChoice string) string`
