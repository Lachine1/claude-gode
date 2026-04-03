package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ServerConfig holds the merged configuration for MCP servers
type ServerConfig struct {
	Servers map[string]MCPConfig `json:"mcpServers"`
}

// LoadUserConfig loads MCP server configuration from ~/.claude/mcp.json
func LoadUserConfig() (*ServerConfig, error) {
	path := filepath.Join(homeDir(), ".claude", "mcp.json")
	return loadConfigFile(path)
}

// LoadProjectConfig loads MCP server configuration from .mcp.json in the given directory
func LoadProjectConfig(cwd string) (*ServerConfig, error) {
	path := filepath.Join(cwd, ".mcp.json")
	return loadConfigFile(path)
}

// LoadMergedConfig loads and merges MCP configurations with priority: project > user > default
func LoadMergedConfig(cwd string) (*ServerConfig, error) {
	merged := &ServerConfig{
		Servers: make(map[string]MCPConfig),
	}

	userConfig, err := LoadUserConfig()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load user config: %w", err)
	}
	if userConfig != nil {
		for name, cfg := range userConfig.Servers {
			merged.Servers[name] = cfg
		}
	}

	projectConfig, err := LoadProjectConfig(cwd)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}
	if projectConfig != nil {
		for name, cfg := range projectConfig.Servers {
			merged.Servers[name] = cfg
		}
	}

	return merged, nil
}

// ValidateServer checks if an MCP server configuration is valid
func ValidateServer(config MCPConfig) error {
	if config.Command == "" {
		return fmt.Errorf("command is required")
	}

	path, err := exec.LookPath(config.Command)
	if err != nil {
		return fmt.Errorf("command %q not found in PATH: %w", config.Command, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access command %q: %w", config.Command, err)
	}

	if info.IsDir() {
		return fmt.Errorf("command %q is a directory, not an executable", config.Command)
	}

	return nil
}

// ValidateConfig validates all servers in a configuration
func (c *ServerConfig) ValidateConfig() map[string]error {
	errors := make(map[string]error)
	for name, cfg := range c.Servers {
		if err := ValidateServer(cfg); err != nil {
			errors[name] = err
		}
	}
	return errors
}

func loadConfigFile(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]MCPConfig)
	}

	return &cfg, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
