package version_test

import (
	"strings"
	"testing"

	"consolehub/internal/version"
)

func TestVersion_String(t *testing.T) {
	str := version.String()
	if !strings.HasPrefix(str, "ConsoleHub Server v0.1.0") {
		t.Errorf("expected string to start with 'ConsoleHub Server v0.1.0', got %s", str)
	}
}
