package services_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"consolehub/internal/config"
	"consolehub/internal/models"
	"consolehub/internal/services"
	"consolehub/internal/storage"
)

func setupTestStorage(t *testing.T) *storage.Storage {
	tmpDir := t.TempDir()
	cfg := config.Default()
	cfg.PocketBase.DataDir = filepath.Join(tmpDir, "pb_data")

	store, err := storage.New(cfg)
	if err != nil {
		t.Fatalf("failed to setup test storage: %v", err)
	}
	return store
}

func TestFullCRUDOperations(t *testing.T) {
	store := setupTestStorage(t)
	svc := services.New(store)
	ctx := context.Background()

	// 1. Tenant CRUD
	tenant, err := svc.CreateTenant(ctx, "Test Org", "test-org")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	updatedTenant, err := svc.UpdateTenant(ctx, tenant.ID, "Test Org Updated", "test-org-updated", true)
	if err != nil || updatedTenant.Name != "Test Org Updated" {
		t.Fatalf("failed to update tenant: %v", err)
	}

	// 2. Host CRUD
	host, err := svc.RegisterHost(ctx, "vps-01-slug", "vps-01", "vps-01.example.com", "Main VPS", "linux")
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}

	updatedHost, err := svc.UpdateHost(ctx, host.ID, "vps-01-slug", "vps-01", "vps-01.example.com", "Updated VPS Name", "linux", true)
	if err != nil || updatedHost.DisplayName != "Updated VPS Name" {
		t.Fatalf("failed to update host: %v", err)
	}

	hosts, err := svc.ListHosts(ctx)
	if err != nil || len(hosts) != 1 {
		t.Fatalf("failed to list hosts, count: %d", len(hosts))
	}

	// Test Host-Tenant Many-to-Many Association
	tenant2, err := svc.CreateTenant(ctx, "Engineering 2", "engineering-2")
	if err != nil {
		t.Fatalf("failed to create tenant2: %v", err)
	}

	if err := svc.AssociateHostTenant(ctx, host.ID, tenant.ID); err != nil {
		t.Fatalf("failed to associate host with tenant: %v", err)
	}
	if err := svc.AssociateHostTenant(ctx, host.ID, tenant2.ID); err != nil {
		t.Fatalf("failed to associate host with tenant2: %v", err)
	}

	hTenants, err := svc.ListTenantsByHost(ctx, host.ID)
	if err != nil || len(hTenants) != 2 {
		t.Fatalf("expected 2 associated tenants for host, got %d (err: %v)", len(hTenants), err)
	}

	tHosts, err := svc.ListHostsByTenant(ctx, tenant.ID)
	if err != nil || len(tHosts) != 1 {
		t.Fatalf("expected 1 host for tenant, got %d (err: %v)", len(tHosts), err)
	}

	if err := svc.DissociateHostTenant(ctx, host.ID, tenant2.ID); err != nil {
		t.Fatalf("failed to dissociate host from tenant2: %v", err)
	}

	hTenantsAfter, err := svc.ListTenantsByHost(ctx, host.ID)
	if err != nil || len(hTenantsAfter) != 1 {
		t.Fatalf("expected 1 associated tenant for host after dissociation, got %d", len(hTenantsAfter))
	}

	// 3. App CRUD
	app, err := svc.CreateApp(ctx, tenant.ID, "replicator", "Replicator Service", "Main Sync Service")
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	updatedApp, err := svc.UpdateApp(ctx, app.ID, "replicator", "Replicator Service V2", "Updated Desc")
	if err != nil || updatedApp.DisplayName != "Replicator Service V2" {
		t.Fatalf("failed to update app: %v", err)
	}

	apps, err := svc.ListApps(ctx)
	if err != nil || len(apps) != 1 {
		t.Fatalf("failed to list apps, count: %d", len(apps))
	}

	// 4. Run CRUD
	run, err := svc.RegisterProcessRun(ctx, services.RegisterRunParams{
		TenantID:         tenant.ID,
		HostID:           host.ID,
		AppID:            app.ID,
		ClientRunID:      "run-uuid-001",
		PID:              1234,
		StartedAt:        time.Now(),
		Version:          "1.0.0",
		WorkingDirectory: "/srv/app",
		CommandLine:      "./app",
	})
	if err != nil {
		t.Fatalf("failed to register run: %v", err)
	}

	_, err = svc.UpdateRunStatus(ctx, run.ID, models.RunStatusExited, 0)
	if err != nil {
		t.Fatalf("failed to update run status: %v", err)
	}

	runs, err := svc.ListRuns(ctx)
	if err != nil || len(runs) != 1 {
		t.Fatalf("failed to list runs, count: %d", len(runs))
	}

	// Delete run
	err = svc.DeleteRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to delete run: %v", err)
	}

	// 5. API Key CRUD & Validation
	apiKey, rawKey, err := svc.CreateAPIKey(ctx, tenant.ID, "Test Key", "Integration test key", nil)
	if err != nil || rawKey == "" {
		t.Fatalf("failed to create API key: %v", err)
	}

	valKey, err := svc.ValidateAPIKey(ctx, rawKey)
	if err != nil || valKey.ID != apiKey.ID {
		t.Fatalf("failed to validate API key: %v", err)
	}

	err = svc.UpdateAPIKey(ctx, apiKey.ID, "Updated Key Title", "Updated Description", true, nil)
	if err != nil {
		t.Fatalf("failed to update API key: %v", err)
	}

	err = svc.DeleteAPIKey(ctx, apiKey.ID)
	if err != nil {
		t.Fatalf("failed to delete API key: %v", err)
	}

	// Delete app, host, tenant
	_ = svc.DeleteApp(ctx, app.ID)
	_ = svc.DeleteHost(ctx, host.ID)
	_ = svc.DeleteTenant(ctx, tenant.ID)
}
