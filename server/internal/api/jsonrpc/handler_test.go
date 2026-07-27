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

	host, err := svc.RegisterHost(ctx, "vps-01", "vps-01", "vps-01.example.com", "VPS 01", "linux")
	if err != nil {
		t.Fatalf("failed to register host: %v", err)
	}
	if err := svc.AssociateHostTenant(ctx, host.ID, tenant.ID); err != nil {
		t.Fatalf("failed to associate host with tenant: %v", err)
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

func TestJSONRPC_APIKeyTenantMismatch(t *testing.T) {
	server, svc := setupTestJSONRPCHandler(t)
	defer server.Close()

	ctx := context.Background()
	tenantA, _ := svc.CreateTenant(ctx, "Org A", "org-a")
	tenantB, _ := svc.CreateTenant(ctx, "Org B", "org-b")

	hostB, _ := svc.RegisterHost(ctx, "vps-02", "vps-02", "", "VPS 02", "linux")
	_ = svc.AssociateHostTenant(ctx, hostB.ID, tenantB.ID)

	_, rawKey, err := svc.CreateAPIKey(ctx, tenantA.ID, "Org A Key", "Key for Org A", nil)
	if err != nil {
		t.Fatalf("failed to create API Key: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// Authenticate with Tenant A's key
	_ = ws.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "auth.authenticate",
		"params": map[string]any{"token": rawKey},
	})
	var authResp map[string]any
	_ = ws.ReadJSON(&authResp)

	// Attempt process.register under Tenant B (mismatch)
	regReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "process.register",
		"params": map[string]any{
			"tenant": tenantB.Slug,
			"app":    "app-b",
			"host": map[string]any{"hostname": "vps-02"},
			"process": map[string]any{
				"client_run_id": "client-uuid-222",
				"pid":          200,
				"started_at":   "2026-07-27T00:00:00Z",
			},
		},
	}
	_ = ws.WriteJSON(regReq)

	var regResp map[string]any
	if err := ws.ReadJSON(&regResp); err != nil {
		t.Fatalf("failed to read reg resp: %v", err)
	}

	errObj, ok := regResp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error for tenant mismatch, got %v", regResp)
	}
	if errObj["code"].(float64) != float64(jsonrpc.ErrCodeUnauthorized) {
		t.Fatalf("expected ErrCodeUnauthorized (%d), got %v", jsonrpc.ErrCodeUnauthorized, errObj["code"])
	}
}

func TestJSONRPC_HostSlugNotAssociated(t *testing.T) {
	server, svc := setupTestJSONRPCHandler(t)
	defer server.Close()

	ctx := context.Background()
	tenant, _ := svc.CreateTenant(ctx, "Org C", "org-c")

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// Authenticate
	_ = ws.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "auth.authenticate",
		"params":  map[string]any{"token": "valid-token"},
	})
	var authResp map[string]any
	_ = ws.ReadJSON(&authResp)

	// Attempt process.register with unassociated/non-existent host slug
	regReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "process.register",
		"params": map[string]any{
			"tenant": tenant.Slug,
			"app":    "app-c",
			"host":   map[string]any{"hostname": "unknown-host-slug"},
			"process": map[string]any{
				"client_run_id": "client-uuid-333",
				"pid":          300,
				"started_at":   "2026-07-27T00:00:00Z",
			},
		},
	}
	_ = ws.WriteJSON(regReq)

	var regResp map[string]any
	if err := ws.ReadJSON(&regResp); err != nil {
		t.Fatalf("failed to read reg resp: %v", err)
	}

	errObj, ok := regResp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error for unassociated host, got %v", regResp)
	}
	if errObj["code"].(float64) != float64(jsonrpc.ErrCodeInvalidParams) {
		t.Fatalf("expected ErrCodeInvalidParams (%d), got %v", jsonrpc.ErrCodeInvalidParams, errObj["code"])
	}
}

func TestJSONRPC_HealthzUnauthenticated(t *testing.T) {
	server, _ := setupTestJSONRPCHandler(t)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// Call healthz procedure without prior authentication
	healthReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "healthz",
	}
	if err := ws.WriteJSON(healthReq); err != nil {
		t.Fatalf("failed to write healthz req: %v", err)
	}

	var resp map[string]any
	if err := ws.ReadJSON(&resp); err != nil {
		t.Fatalf("failed to read healthz resp: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok || result["status"] != "ok" {
		t.Fatalf("expected status ok in healthz response, got %v", resp)
	}
}

func TestJSONRPC_TenantInfoAndAppList(t *testing.T) {
	server, svc := setupTestJSONRPCHandler(t)
	defer server.Close()

	ctx := context.Background()
	tenant, err := svc.CreateTenant(ctx, "Acme Corp", "acme-corp")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	_, _ = svc.CreateApp(ctx, tenant.ID, "sync-service", "Sync Service", "Main sync daemon")

	_, rawKey, err := svc.CreateAPIKey(ctx, tenant.ID, "Acme Key", "Acme API Key", nil)
	if err != nil {
		t.Fatalf("failed to create API Key: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// 1. Authenticate
	_ = ws.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "auth.authenticate",
		"params":  map[string]any{"token": rawKey},
	})
	var authResp map[string]any
	_ = ws.ReadJSON(&authResp)

	// 2. Test tenant.info
	_ = ws.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tenant.info",
	})
	var infoResp map[string]any
	if err := ws.ReadJSON(&infoResp); err != nil {
		t.Fatalf("failed to read tenant.info resp: %v", err)
	}
	infoRes, ok := infoResp["result"].(map[string]any)
	if !ok || infoRes["slug"] != "acme-corp" {
		t.Fatalf("expected slug acme-corp in tenant.info, got %v", infoResp)
	}

	// 3. Test tenant.app_list
	_ = ws.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tenant.app_list",
	})
	var listResp map[string]any
	if err := ws.ReadJSON(&listResp); err != nil {
		t.Fatalf("failed to read tenant.app_list resp: %v", err)
	}
	listRes, ok := listResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map in tenant.app_list, got %v", listResp)
	}
	appsList, ok := listRes["apps"].([]any)
	if !ok || len(appsList) == 0 {
		t.Fatalf("expected at least 1 app in tenant.app_list result, got %v", listResp)
	}
}
