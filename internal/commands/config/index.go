package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Command returns the /config command
func Command() types.Command {
	return types.Command{
		Name:        "config",
		Aliases:     []string{},
		Description: "Show/edit config",
		Usage:       "/config [key] [value]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleConfig(ctx, args)
		},
	}
}

func handleConfig(ctx *types.CommandContext, args []string) error {
	configPath := filepath.Join(homeDir(), ".claude", "settings.json")

	if len(args) == 0 {
		return showConfig(ctx, configPath)
	}

	if len(args) == 1 {
		return getConfigValue(configPath, args[0])
	}

	if len(args) >= 2 {
		return setConfigValue(configPath, args[0], strings.Join(args[1:], " "))
	}

	return nil
}

func showConfig(ctx *types.CommandContext, path string) error {
	cfg := loadConfig(path)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  Configuration\n")
	sb.WriteString("  ═══════════════════════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("  CWD:              %s\n", ctx.Cwd))

	if cfg.APIKey != "" {
		masked := cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
		sb.WriteString(fmt.Sprintf("  API Key:          %s\n", masked))
	} else {
		sb.WriteString("  API Key:          (not set)\n")
	}

	sb.WriteString(fmt.Sprintf("  Model:            %s\n", cfg.Model))
	sb.WriteString(fmt.Sprintf("  Max Tokens:       %d\n", cfg.MaxTokens))
	sb.WriteString(fmt.Sprintf("  Permission Mode:  %s\n", cfg.PermissionMode))

	if len(cfg.Settings) > 0 {
		sb.WriteString("\n  Custom Settings:\n")
		for k, v := range cfg.Settings {
			sb.WriteString(fmt.Sprintf("    %-20s %v\n", k, v))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("  Config file: ~/.claude/settings.json\n")
	sb.WriteString("\n")

	fmt.Println(sb.String())
	return nil
}

func getConfigValue(path, key string) error {
	cfg := loadConfig(path)

	switch key {
	case "model":
		fmt.Printf("  model = %s\n", cfg.Model)
	case "max_tokens":
		fmt.Printf("  max_tokens = %d\n", cfg.MaxTokens)
	case "permission_mode":
		fmt.Printf("  permission_mode = %s\n", cfg.PermissionMode)
	case "api_key":
		if cfg.APIKey != "" {
			masked := cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
			fmt.Printf("  api_key = %s\n", masked)
		} else {
			fmt.Println("  api_key = (not set)")
		}
	default:
		if val, ok := cfg.Settings[key]; ok {
			fmt.Printf("  %s = %v\n", key, val)
		} else {
			fmt.Printf("  Unknown config key: %s\n", key)
		}
	}

	return nil
}

func setConfigValue(path, key, value string) error {
	cfg := loadConfig(path)

	switch key {
	case "model":
		cfg.Model = value
	case "max_tokens":
		fmt.Sscanf(value, "%d", &cfg.MaxTokens)
	case "permission_mode":
		cfg.PermissionMode = value
	default:
		if cfg.Settings == nil {
			cfg.Settings = make(map[string]any)
		}
		cfg.Settings[key] = value
	}

	if err := saveConfig(path, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("  Set %s = %s\n", key, value)
	return nil
}

type configData struct {
	APIKey         string         `json:"api_key,omitempty"`
	Model          string         `json:"model,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	PermissionMode string         `json:"permission_mode,omitempty"`
	Settings       map[string]any `json:"settings,omitempty"`
}

func loadConfig(path string) configData {
	cfg := configData{}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(path string, cfg configData) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
