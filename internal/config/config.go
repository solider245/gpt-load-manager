// Package config provides configuration loading for the panel.
package config

import (
	"os"
	"path/filepath"
)

// Config holds all panel configuration.
type Config struct {
	Port   string
	DBPath string
	WebDir string
}

// Load reads configuration from environment variables.
func Load() *Config {
	cfg := &Config{
		Port:   "3002",
		DBPath: filepath.Join(".", "data", "panel.db"),
		WebDir: "",
	}
	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}
	if p := os.Getenv("DB_PATH"); p != "" {
		cfg.DBPath = p
	}
	if p := os.Getenv("WEB_DIR"); p != "" {
		cfg.WebDir = p
	}
	return cfg
}
