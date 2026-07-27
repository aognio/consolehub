package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"consolehub/internal/config"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 3787 {
		t.Errorf("expected default port 3787, got %d", cfg.Server.Port)
	}
	if cfg.Server.Scheme != "https" {
		t.Errorf("expected default scheme 'https', got '%s'", cfg.Server.Scheme)
	}
}

func TestLoadConfig_FromTOMLFile(t *testing.T) {
	content := `
[server]
host = "127.0.0.1"
port = 3787
scheme = "https"
base_path = "/console"
public_url = "https://consolehub.example.com:3787"

[security]
cookie_secret = "test-secret-key-32-bytes-length-x"
secure_cookies = true
same_site = "strict"
session_duration = "12h"

[pocketbase]
data_dir = "/var/lib/consolehub/data"

[logging]
log_path = "/var/log/consolehub/consolehub.jsonl"
log_level = "debug"
retention_days = 60
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "server-config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load TOML config: %v", err)
	}

	if cfg.Server.Port != 3787 {
		t.Errorf("expected port 3787, got %d", cfg.Server.Port)
	}
	if cfg.Server.PublicURL != "https://consolehub.example.com:3787" {
		t.Errorf("expected public_url match, got '%s'", cfg.Server.PublicURL)
	}
	if cfg.Logging.ResolvedLogPath() != "/var/log/consolehub/consolehub.jsonl" {
		t.Errorf("expected resolved log path '/var/log/consolehub/consolehub.jsonl', got '%s'", cfg.Logging.ResolvedLogPath())
	}
}
