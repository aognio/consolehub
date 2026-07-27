package consolehub_test

import (
	"testing"

	"github.com/aognio/consolehub/libraries/go/consolehub"
	"github.com/aognio/consolehub/libraries/go/consolehub/protocol"
	"github.com/aognio/consolehub/libraries/go/consolehub/transport"
)

func TestClient_PublicAPI(t *testing.T) {
	mockTrans := transport.NewMockTransport(func(method string, params any) (any, error) {
		switch method {
		case protocol.MethodAuthAuthenticate:
			return &protocol.AuthResult{Authenticated: true}, nil
		case protocol.MethodProcessRegister:
			return &protocol.ProcessRegisterResult{ProcessID: "run-test-1"}, nil
		case protocol.MethodStreamAppend:
			return &protocol.StreamAppendResult{AcceptedThrough: 10}, nil
		}
		return nil, nil
	})

	client, err := consolehub.New(
		consolehub.WithTenant("acme"),
		consolehub.WithApp("replicator"),
		consolehub.WithTransport(mockTrans),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// 1. Standard Output
	client.Println("Hello ConsoleHub!")
	client.Printf("Item %d processed\n", 42)

	// 2. Structured Logging
	client.Infof("Client initialized successfully")
	client.Warn("High CPU usage")

	// 3. Progress
	p := client.Progress("Processing records", 100)
	p.Set(50)
	p.Done()

	// Graceful close
	if err := client.Close(); err != nil {
		t.Errorf("failed to close client: %v", err)
	}
}

func TestClient_DisabledFlag(t *testing.T) {
	client, err := consolehub.New(consolehub.WithDisabled(true))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if !client.IsDisabled() {
		t.Errorf("expected client to be disabled")
	}

	client.SetDisabled(false)
	if client.IsDisabled() {
		t.Errorf("expected client to be enabled after SetDisabled(false)")
	}

	client.SetDisabled(true)
	if !client.IsDisabled() {
		t.Errorf("expected client to be disabled after SetDisabled(true)")
	}

	// Replacement functions should execute without error when disabled
	client.Println("This message will not be sent to ConsoleHub")
	client.Infof("Local log only")
}

func TestGlobal_DisableEnable(t *testing.T) {
	consolehub.Disable()
	if !consolehub.IsDisabled() {
		t.Errorf("expected global consolehub to be disabled")
	}

	consolehub.Println("Global disabled print test")

	consolehub.Enable()
	if consolehub.IsDisabled() {
		t.Errorf("expected global consolehub to be enabled")
	}
}
