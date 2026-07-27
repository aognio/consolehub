package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/aognio/consolehub/libraries/go/consolehub/protocol"

	"github.com/gorilla/websocket"
)

type WebSocketTransport struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	requestID int64
	connected bool
}

func NewWebSocketTransport() *WebSocketTransport {
	return &WebSocketTransport{}
}

func (w *WebSocketTransport) Connect(ctx context.Context, urlStr string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.connected && w.conn != nil {
		return nil
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, urlStr, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	w.conn = conn
	w.connected = true
	return nil
}

func (w *WebSocketTransport) Call(ctx context.Context, method string, params any, result any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected || w.conn == nil {
		return ErrDisconnected
	}

	reqID := atomic.AddInt64(&w.requestID, 1)
	req := protocol.RequestFrame{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	if err := w.conn.WriteJSON(req); err != nil {
		w.closeConnLocked()
		return fmt.Errorf("write websocket json: %w", err)
	}

	_, raw, err := w.conn.ReadMessage()
	if err != nil {
		w.closeConnLocked()
		return fmt.Errorf("read websocket json: %w", err)
	}

	var resp protocol.ResponseFrame
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unmarshal response frame: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("jsonrpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		resBytes, errMarshal := json.Marshal(resp.Result)
		if errMarshal != nil {
			return errMarshal
		}
		if err := json.Unmarshal(resBytes, result); err != nil {
			return fmt.Errorf("unmarshal result payload: %w", err)
		}
	}

	return nil
}

func (w *WebSocketTransport) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeConnLocked()
}

func (w *WebSocketTransport) closeConnLocked() error {
	if w.connected && w.conn != nil {
		w.connected = false
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	w.connected = false
	return nil
}

func (w *WebSocketTransport) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}
