package templates_test

import (
	"bytes"
	"strings"
	"testing"

	"consolehub/internal/templates"
)

func TestEmbeddedTemplates(t *testing.T) {
	engine, err := templates.New("Local")
	if err != nil {
		t.Fatalf("failed to initialize embedded template engine: %v", err)
	}

	pages := []string{
		"login.html",
		"dashboard.html",
		"tenants.html",
		"hosts.html",
		"apps.html",
		"runs.html",
		"console.html",
		"search.html",
		"users.html",
		"api_keys.html",
		"settings.html",
	}

	sampleData := map[string]any{
		"Title":          "Test Title",
		"RunningCount":   5,
		"OnlineHosts":    3,
		"OfflineHosts":   1,
		"RecentRuns":     12,
		"RecentFailures": 0,
		"Run": map[string]any{
			"ID":  "run-test-123",
			"PID": 9999,
		},
	}

	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			var buf bytes.Buffer
			if err := engine.Render(&buf, page, sampleData); err != nil {
				t.Fatalf("failed to render embedded template '%s': %v", page, err)
			}

			output := buf.String()
			if len(strings.TrimSpace(output)) == 0 {
				t.Fatalf("embedded template '%s' rendered an empty string", page)
			}
		})
	}
}
