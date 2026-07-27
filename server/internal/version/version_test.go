package version_test

import (
	"strings"
	"testing"

	"consolehub/internal/version"
)

func TestVersion_String(t *testing.T) {
	str := version.String()
	expectedPrefix := "ConsoleHub Server " + version.Version
	if !strings.HasPrefix(str, expectedPrefix) {
		t.Errorf("expected string to start with '%s', got %s", expectedPrefix, str)
	}
}
