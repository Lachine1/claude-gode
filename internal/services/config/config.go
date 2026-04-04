package config

import (
	"os"
	"strconv"
)

// Config holds application configuration (wraps Settings for compatibility)
type Config struct {
	Settings *Settings
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Settings: DefaultSettings(),
	}
}

// Load loads configuration from files and environment
func Load(cwd string) (*Config, error) {
	settings, err := LoadSettings(cwd)
	if err != nil {
		return nil, err
	}
	return &Config{Settings: settings}, nil
}

// GetString returns a string setting value with optional default
func (c *Config) GetString(key string, defaultVal string) string {
	if v, ok := c.Settings.Raw[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// GetInt returns an int setting value with optional default
func (c *Config) GetInt(key string, defaultVal int) int {
	if v, ok := c.Settings.Raw[key]; ok {
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
	if v, ok := c.Settings.Raw[key]; ok {
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

// GetEnv returns an environment variable value
func (c *Config) GetEnv(key string) string {
	if c.Settings.Env != nil {
		if v, ok := c.Settings.Env[key]; ok {
			return v
		}
	}
	return os.Getenv(key)
}

// Set sets a setting value
func (c *Config) Set(key string, value any) {
	c.Settings.Set(key, value)
}

// All returns all raw settings
func (c *Config) All() map[string]any {
	return c.Settings.Raw
}

// Model returns the configured model
func (c *Config) Model() string {
	return c.Settings.Model
}

// MaxTokens returns the configured max tokens
func (c *Config) MaxTokens() int {
	return c.Settings.MaxTokens()
}

// PermissionMode returns the configured permission mode
func (c *Config) PermissionMode() string {
	return c.Settings.PermissionMode
}

// APIKey returns the API key from env
func (c *Config) APIKey() string {
	return c.Settings.APIKey()
}

// ShowWelcomeBanner returns whether to show the welcome banner
func (c *Config) ShowWelcomeBanner() bool {
	return c.Settings.ShowWelcomeBanner
}
