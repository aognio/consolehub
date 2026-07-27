package version

import "fmt"

var (
	// Version holds the current release version of ConsoleHub.
	Version = "v0.1.1"

	// GitCommit holds the short git commit hash set via -ldflags during build.
	GitCommit = "none"

	// BuildDate holds the build timestamp set via -ldflags during build.
	BuildDate = "unknown"
)

// String returns a formatted version, git commit, and build date string.
func String() string {
	return fmt.Sprintf("ConsoleHub Server %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
