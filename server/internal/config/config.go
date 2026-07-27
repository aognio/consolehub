package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server     ServerConfig     `toml:"server"`
	Security   SecurityConfig   `toml:"security"`
	PocketBase PocketBaseConfig `toml:"pocketbase"`
	Logging    LoggingConfig    `toml:"logging"`
}

type LoggingConfig struct {
	LogFile  string `toml:"log_file"`
	LogPath  string `toml:"log_path"`
	LogLevel string `toml:"log_level"`
}

func (l LoggingConfig) ResolvedLogPath() string {
	if l.LogPath != "" {
		return l.LogPath
	}
	if l.LogFile != "" {
		return l.LogFile
	}
	return "/var/log/consolehub/consolehub.log"
}

type ServerConfig struct {
	Host      string `toml:"host"`
	Port      int    `toml:"port"`
	Scheme    string `toml:"scheme"`
	BasePath  string `toml:"base_path"`
	PublicURL string `toml:"public_url"`
	Timezone  string `toml:"timezone"`
}

type SecurityConfig struct {
	CookieSecret    string `toml:"cookie_secret"`
	SecureCookies   bool   `toml:"secure_cookies"`
	SessionDuration string `toml:"session_duration"`
}

func (s SecurityConfig) Duration() time.Duration {
	d, err := time.ParseDuration(s.SessionDuration)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

type PocketBaseConfig struct {
	DataDir string `toml:"data_dir"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:      "0.0.0.0",
			Port:      3787,
			Scheme:    "https",
			BasePath:  "/",
			PublicURL: "https://consolehub.example.com:3787",
			Timezone:  "Local",
		},
		Security: SecurityConfig{
			CookieSecret:    "consolehub-dev-secret-key-32bytesmin!!",
			SecureCookies:   false,
			SessionDuration: "86400s",
		},
		PocketBase: PocketBaseConfig{
			DataDir: "./pb_data",
		},
		Logging: LoggingConfig{
			LogFile:  "/var/log/consolehub/consolehub.log",
			LogLevel: "debug",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	searchPaths := []string{}
	if path != "" {
		searchPaths = append(searchPaths, path)
	}
	searchPaths = append(searchPaths, "./server-config.toml", "/etc/consolehub/server-config.toml")

	var foundPath string
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(foundPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", foundPath, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config toml %s: %w", foundPath, err)
	}

	if cfg.PocketBase.DataDir != "" {
		cfg.PocketBase.DataDir = filepath.Clean(cfg.PocketBase.DataDir)
	}

	return cfg, nil
}
