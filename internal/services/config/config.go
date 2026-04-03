package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	APIKey         string         `json:"api_key,omitempty"`
	Model          string         `json:"model,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	PermissionMode string         `json:"permission_mode,omitempty"`
	Settings       map[string]any `json:"settings,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Model:          "claude-sonnet-4-20250514",
		MaxTokens:      8192,
		PermissionMode: "default",
		Settings:       make(map[string]any),
	}
}

// Load loads configuration from files and environment
func Load(cwd string) (*Config, error) {
	cfg := DefaultConfig()

	// Load user settings
	userPath := filepath.Join(homeDir(), ".claude", "settings.json")
	if data, err := os.ReadFile(userPath); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Load project settings
	projectPath := filepath.Join(cwd, "CLAUDE.md")
	if _, err := os.Stat(projectPath); err == nil {
		// CLAUDE.md exists, project-level config may be present
	}

	// Environment variables override file config
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if model := os.Getenv("CLAUDE_CODE_MODEL"); model != "" {
		cfg.Model = model
	}

	return cfg, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
