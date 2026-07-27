# ConsoleHub Go Client Testing Guide

This document details unit testing strategies and mock transport patterns for testing applications integrated with the `consolehub` Go client.

---

## 1. Using `MockTransport` in Unit Tests

The `transport.MockTransport` implementation allows testing client calls without connecting to a real WebSocket server.

```go
package myapp_test

import (
    "testing"

    "consolehub/libraries/go/consolehub"
    "consolehub/libraries/go/consolehub/protocol"
    "consolehub/libraries/go/consolehub/transport"
)

func TestMyApplication(t *testing.T) {
    mockTrans := transport.NewMockTransport(func(method string, params any) (any, error) {
        switch method {
        case protocol.MethodAuthAuthenticate:
            return &protocol.AuthResult{Authenticated: true}, nil
        case protocol.MethodProcessRegister:
            return &protocol.ProcessRegisterResult{ProcessID: "test-run-1"}, nil
        case protocol.MethodStreamAppend:
            return &protocol.StreamAppendResult{AcceptedThrough: 100}, nil
        }
        return nil, nil
    })

    client, err := consolehub.New(
        consolehub.WithTenant("test-tenant"),
        consolehub.WithApp("test-app"),
        consolehub.WithTransport(mockTrans),
    )
    if err != nil {
        t.Fatalf("failed to initialize client: %v", err)
    }
    defer client.Close()

    client.Println("Executing test routine...")
}
```

---

## 2. Local-Only Testing with `WithDisabled(true)`

For isolated unit tests where no network calls or telemetry streaming should occur, pass `WithDisabled(true)`:

```go
client, _ := consolehub.New(consolehub.WithDisabled(true))
defer client.Close()
```
