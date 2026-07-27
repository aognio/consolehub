package transport

import (
	"context"
	"errors"
)

var (
	ErrDisconnected = errors.New("transport disconnected")
)

// Transport defines the communication interface for JSON-RPC 2.0 frames.
type Transport interface {
	Connect(ctx context.Context, url string) error
	Call(ctx context.Context, method string, params any, result any) error
	Close() error
	IsConnected() bool
}
