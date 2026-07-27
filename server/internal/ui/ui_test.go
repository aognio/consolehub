package ui_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"consolehub/internal/config"
	"consolehub/internal/services"
	"consolehub/internal/storage"
	"consolehub/internal/templates"
	"consolehub/internal/ui"
)

func TestUI_LoginPage(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.Default()
	cfg.PocketBase.DataDir = filepath.Join(tmpDir, "pb_data")

	store, err := storage.New(cfg)
	if err != nil {
		t.Fatalf("failed to setup storage: %v", err)
	}

	svc := services.New(store)
	tmpl, err := templates.New("Local")
	if err != nil {
		t.Fatalf("failed to setup template engine: %v", err)
	}

	handler := ui.NewHandler(cfg, svc, tmpl)

	req := httptest.NewRequest("GET", "/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected login page status OK 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty body for login page, got empty string")
	}
	if !strings.Contains(body, "ConsoleHub") || !strings.Contains(body, "Sign In") {
		t.Errorf("expected body to contain 'ConsoleHub' and 'Sign In', got:\n%s", body)
	}
}
