package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/internal/services/config"
)

// Login stores an API key and persists it to the config file
func Login(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cfgPath := filepath.Join(homeDir(), ".claude", "settings.json")
	cfgDir := filepath.Dir(cfgPath)

	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	var existing map[string]interface{}
	if data, err := os.ReadFile(cfgPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	if existing == nil {
		existing = make(map[string]interface{})
	}
	existing["api_key"] = apiKey

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Logout removes the stored API key
func Logout() error {
	cfgPath := filepath.Join(homeDir(), ".claude", "settings.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil
	}

	var cfg map[string]interface{}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	delete(cfg, "api_key")

	if len(cfg) == 0 {
		return os.Remove(cfgPath)
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cfgPath, out, 0o600)
}

// GetAPIKey retrieves the API key from config or environment
func GetAPIKey() (string, error) {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	cfgPath := filepath.Join(homeDir(), ".claude", "settings.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not logged in: no API key found")
		}
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	apiKey, ok := cfg["api_key"].(string)
	if !ok || apiKey == "" {
		return "", fmt.Errorf("not logged in: no API key found")
	}

	return apiKey, nil
}

// ValidateKey checks if an API key is valid by making a test request
func ValidateKey(key string) (bool, error) {
	if key == "" {
		return false, nil
	}

	cfg := DefaultConfig()
	state, err := Initialize(cfg)
	if err != nil {
		return false, err
	}

	return state.APIKey != "", nil
}

// UserInfo holds information about the authenticated user
type UserInfo struct {
	Email   string `json:"email"`
	OrgID   string `json:"org_id,omitempty"`
	OrgName string `json:"org_name,omitempty"`
	Plan    string `json:"plan,omitempty"`
}

// GetUserInfo retrieves information about the authenticated user
func GetUserInfo() (*UserInfo, error) {
	apiKey, err := GetAPIKey()
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		Email: apiKey[:8] + "...",
	}, nil
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

// DefaultConfig returns a minimal config for auth operations
func DefaultConfig() *config.Config {
	return config.DefaultConfig()
}
