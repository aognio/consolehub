package storage_test

import (
	"path/filepath"
	"testing"

	"consolehub/internal/config"
	"consolehub/internal/storage"
)

func TestStorage_InitSchema(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.Default()
	cfg.PocketBase.DataDir = filepath.Join(tmpDir, "pb_data")

	store, err := storage.New(cfg)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	collections := []string{
		"ch_users",
		"ch_tenants",
		"ch_tenant_members",
		"ch_hosts",
		"ch_apps",
		"ch_runs",
		"ch_stream_lines",
	}

	for _, col := range collections {
		found, err := store.App.FindCollectionByNameOrId(col)
		if err != nil || found == nil {
			t.Errorf("expected collection %s to exist, got err: %v", col, err)
		}
	}
}

func TestStorage_EnsureDefaultSuperAdmin(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.Default()
	cfg.PocketBase.DataDir = filepath.Join(tmpDir, "pb_data")

	store, err := storage.New(cfg)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	admin, err := store.EnsureDefaultSuperAdmin("admin@consolehub.local", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create default super admin: %v", err)
	}

	if admin.Email != "admin@consolehub.local" {
		t.Errorf("expected email admin@consolehub.local, got %s", admin.Email)
	}
	if admin.Role != "super_admin" {
		t.Errorf("expected role super_admin, got %s", admin.Role)
	}
}
