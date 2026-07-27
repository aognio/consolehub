package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
)

// Environment holds auto-detected system and process metadata.
type Environment struct {
	Hostname         string
	PID              int
	CommandLine      string
	WorkingDirectory string
	Platform         string
	Architecture     string
	Username         string
}

// AutoDetect resolves environment metadata from standard OS calls.
func AutoDetect() Environment {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}

	workDir, _ := os.Getwd()
	if workDir == "" {
		workDir = "."
	}

	cmdLine := ""
	if len(os.Args) > 0 {
		cmdLine = filepath.Base(os.Args[0])
		if len(os.Args) > 1 {
			for _, arg := range os.Args[1:] {
				cmdLine += " " + arg
			}
		}
	}

	username := ""
	if currentUser, err := user.Current(); err == nil {
		username = currentUser.Username
	}

	return Environment{
		Hostname:         hostname,
		PID:              os.Getpid(),
		CommandLine:      cmdLine,
		WorkingDirectory: workDir,
		Platform:         runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Username:         username,
	}
}

// GetEnvOrDefault fetches environment variable or returns fallback.
func GetEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// GetEnvIntOrDefault fetches integer environment variable or returns fallback.
func GetEnvIntOrDefault(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
