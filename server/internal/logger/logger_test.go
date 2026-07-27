package logger_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"consolehub/internal/config"
	"consolehub/internal/logger"
)

func TestJSONL_Logger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	l, err := logger.Init(config.LoggingConfig{
		LogFile:  logFile,
		LogLevel: "debug",
	})
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	defer l.Close()

	l.Info("test-component", "server started", map[string]any{"port": 3787})
	l.Error("test-component", "handler error", map[string]any{"error": "something failed"})

	f, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		lines++
		var entry logger.Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse JSONL line: %v", err)
		}
		if entry.Timestamp == "" {
			t.Errorf("expected non-empty timestamp")
		}
	}

	if lines != 2 {
		t.Errorf("expected 2 log lines, got %d", lines)
	}
}
