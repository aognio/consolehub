package jsonrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"consolehub/internal/api/jsonrpc"
	"consolehub/internal/config"
	"consolehub/internal/services"
	"consolehub/internal/storage"
	"consolehub/internal/stream"

	"github.com/gorilla/websocket"
)

func setupTestJSONRPCHandler(t *testing.T) (*httptest.Server, *services.Services) {
	tmpDir := t.TempDir()
	cfg := config.Default()
	cfg.PocketBase.DataDir = filepath.Join(tmpDir, "pb_data")

	store, err := storage.New(cfg)
	if err != nil {
		t.Fatalf("failed to setup storage: %v", err)
	}

	svc := services.New(store)
	hub := stream.NewHub()
	handler := jsonrpc.NewHandler(cfg, svc, hub)

	server := httptest.NewServer(http.HandlerFunc(handler.ServeHTTP))
	return server, svc
}

func TestJSONRPC_AuthenticateAndRegister(t *testing.T) {
	server, svc := setupTestJSONRPCHandler(t)
	defer server.Close()

	ctx := context.Background()
	tenant, err := svc.CreateTenant(ctx, "Test Org", "test-org")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// 1. Authenticate
	authReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "auth.authenticate",
		"params": map[string]any{
			"token": "valid-token",
		},
	}
	if err := ws.WriteJSON(authReq); err != nil {
		t.Fatalf("failed to write auth req: %v", err)
	}

	var authResp map[string]any
	if err := ws.ReadJSON(&authResp); err != nil {
		t.Fatalf("failed to read auth resp: %v", err)
	}

	result, ok := authResp["result"].(map[string]any)
	if !ok || result["authenticated"] != true {
		t.Fatalf("expected authenticated result, got %v", authResp)
	}

	// 2. Process Register
	regReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "process.register",
		"params": map[string]any{
			"tenant": tenant.Slug,
			"app":    "my-app",
			"host": map[string]any{
				"hostname":     "vps-01",
				"display_name": "VPS 01",
				"platform":     "linux",
			},
			"process": map[string]any{
				"client_run_id":     "client-uuid-111",
				"pid":               100,
				"started_at":        "2026-07-26T17:30:00Z",
				"version":           "1.0.0",
				"command_line":      "./app",
				"working_directory": "/app",
			},
		},
	}

	if err := ws.WriteJSON(regReq); err != nil {
		t.Fatalf("failed to write reg req: %v", err)
	}

	var regResp map[string]any
	if err := ws.ReadJSON(&regResp); err != nil {
		t.Fatalf("failed to read reg resp: %v", err)
	}

	regResult, ok := regResp["result"].(map[string]any)
	if !ok || regResult["process_id"] == nil {
		t.Fatalf("expected process_id in result, got %v", regResp)
	}
}
