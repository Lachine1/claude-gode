package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	rawSettings map[string]any
	rawEnv      map[string]string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		rawSettings: make(map[string]any),
		rawEnv:      make(map[string]string),
	}
}

// Load loads configuration from files and environment
// Priority (lowest to highest):
// 1. ~/.claude/settings.json (user settings)
// 2. <cwd>/.claude/settings.json (project settings, overrides user)
// 3. ~/.claude/.env (user env)
// 4. <cwd>/.claude/.env (project env, overrides user)
// 5. Environment variables (override all)
func Load(cwd string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Load user settings
	userSettingsPath := filepath.Join(homeDir(), ".claude", "settings.json")
	if data, err := os.ReadFile(userSettingsPath); err == nil {
		_ = json.Unmarshal(data, &cfg.rawSettings)
	}

	// 2. Load project settings (overrides user)
	projectSettingsPath := filepath.Join(cwd, ".claude", "settings.json")
	if data, err := os.ReadFile(projectSettingsPath); err == nil {
		projectSettings := make(map[string]any)
		if err := json.Unmarshal(data, &projectSettings); err == nil {
			for k, v := range projectSettings {
				cfg.rawSettings[k] = v
			}
		}
	}

	// 3. Load user env
	userEnvPath := filepath.Join(homeDir(), ".claude", ".env")
	if data, err := os.ReadFile(userEnvPath); err == nil {
		cfg.rawEnv = parseEnvFile(string(data))
	}

	// 4. Load project env (overrides user)
	projectEnvPath := filepath.Join(cwd, ".claude", ".env")
	if data, err := os.ReadFile(projectEnvPath); err == nil {
		projectEnv := parseEnvFile(string(data))
		for k, v := range projectEnv {
			cfg.rawEnv[k] = v
		}
	}

	// 5. Environment variables override all
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.rawEnv["ANTHROPIC_API_KEY"] = key
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		cfg.rawSettings["model"] = model
	}
	if model := os.Getenv("CLAUDE_CODE_MODEL"); model != "" {
		cfg.rawSettings["model"] = model
	}
	if maxTokens := os.Getenv("CLAUDE_CODE_MAX_TOKENS"); maxTokens != "" {
		cfg.rawSettings["max_tokens"] = maxTokens
	}
	if permMode := os.Getenv("CLAUDE_CODE_PERMISSION_MODE"); permMode != "" {
		cfg.rawSettings["permission_mode"] = permMode
	}

	return cfg, nil
}

// GetString returns a string setting value with optional default
func (c *Config) GetString(key string, defaultVal string) string {
	if v, ok := c.rawSettings[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// GetInt returns an int setting value with optional default
func (c *Config) GetInt(key string, defaultVal int) int {
	if v, ok := c.rawSettings[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

// GetBool returns a bool setting value with optional default
func (c *Config) GetBool(key string, defaultVal bool) bool {
	if v, ok := c.rawSettings[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true" || val == "1" || val == "yes"
		case float64:
			return val != 0
		}
	}
	return defaultVal
}

// GetEnv returns an environment variable value from loaded env files
func (c *Config) GetEnv(key string) string {
	if v, ok := c.rawEnv[key]; ok {
		return v
	}
	return os.Getenv(key)
}

// Set sets a setting value
func (c *Config) Set(key string, value any) {
	c.rawSettings[key] = value
}

// All returns all raw settings
func (c *Config) All() map[string]any {
	return c.rawSettings
}

// Model returns the configured model
func (c *Config) Model() string {
	return c.GetString("model", "claude-sonnet-4-20250514")
}

// MaxTokens returns the configured max tokens
func (c *Config) MaxTokens() int {
	return c.GetInt("max_tokens", 8192)
}

// PermissionMode returns the configured permission mode
func (c *Config) PermissionMode() string {
	return c.GetString("permission_mode", "default")
}

// APIKey returns the API key from env
func (c *Config) APIKey() string {
	return c.GetEnv("ANTHROPIC_API_KEY")
}

// ShowWelcomeBanner returns whether to show the welcome banner
func (c *Config) ShowWelcomeBanner() bool {
	return c.GetBool("showWelcomeBanner", true)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return "."
}

func parseEnvFile(content string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = strings.Trim(value, "\"'")
			result[key] = value
		}
	}
	return result
}
