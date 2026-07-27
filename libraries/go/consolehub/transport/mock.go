package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/aognio/consolehub/libraries/go/consolehub/protocol"
)

type MockHandler func(method string, params any) (any, error)

type MockTransport struct {
	mu            sync.Mutex
	connected     bool
	handler       MockHandler
	calls         []string
	paramsHistory []any
}

func NewMockTransport(handler MockHandler) *MockTransport {
	return &MockTransport{
		handler: handler,
	}
}

func (m *MockTransport) Connect(ctx context.Context, urlStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	return nil
}

func (m *MockTransport) Call(ctx context.Context, method string, params any, result any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return ErrDisconnected
	}

	m.calls = append(m.calls, method)
	m.paramsHistory = append(m.paramsHistory, params)

	if m.handler != nil {
		res, err := m.handler(method, params)
		if err != nil {
			return err
		}
		if result != nil && res != nil {
			switch target := result.(type) {
			case *protocol.AuthResult:
				if r, ok := res.(*protocol.AuthResult); ok {
					*target = *r
				}
			case *protocol.ProcessRegisterResult:
				if r, ok := res.(*protocol.ProcessRegisterResult); ok {
					*target = *r
				}
			case *protocol.StreamAppendResult:
				if r, ok := res.(*protocol.StreamAppendResult); ok {
					*target = *r
				}
			case *protocol.StreamResumeResult:
				if r, ok := res.(*protocol.StreamResumeResult); ok {
					*target = *r
				}
			}
		}
		return nil
	}

	return fmt.Errorf("no mock handler configured for method: %s", method)
}

func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

func (m *MockTransport) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *MockTransport) GetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.calls))
	copy(cp, m.calls)
	return cp
}
