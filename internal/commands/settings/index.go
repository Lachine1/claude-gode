package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	svcconfig "github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /settings command.
func New(cfg *svcconfig.Config) types.Command {
	return types.Command{
		Name:        "settings",
		Aliases:     []string{"set"},
		Description: "View/edit settings",
		Usage:       "/settings [key] [value]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleSettings(cfg, args)
		},
	}
}

func handleSettings(cfg *svcconfig.Config, args []string) error {
	settingsPath := filepath.Join(homeDir(), ".claude", "settings.json")

	if len(args) == 0 {
		return showSettings(cfg, settingsPath)
	}

	if len(args) == 1 {
		return getSetting(settingsPath, args[0])
	}

	return setSetting(settingsPath, args[0], strings.Join(args[1:], " "))
}

func showSettings(cfg *svcconfig.Config, path string) error {
	settings := loadSettings(path)

	fmt.Println()
	fmt.Println("  Settings")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  %-25s %s\n", "Model:", cfg.Model)
	fmt.Printf("  %-25s %d\n", "Max Tokens:", cfg.MaxTokens)
	fmt.Printf("  %-25s %s\n", "Permission Mode:", cfg.PermissionMode)
	fmt.Println()

	if len(settings) > 0 {
		fmt.Println("  Custom Settings:")
		for k, v := range settings {
			fmt.Printf("    %-23s %v\n", k, v)
		}
		fmt.Println()
	}

	fmt.Println("  Use /settings <key> <value> to set a value.")
	fmt.Println("  Use /settings <key> to view a specific value.")
	fmt.Println()
	return nil
}

func getSetting(path, key string) error {
	settings := loadSettings(path)

	val, ok := settings[key]
	if !ok {
		fmt.Printf("  Setting '%s' not found.\n", key)
		return nil
	}

	fmt.Printf("  %s = %v\n", key, val)
	return nil
}

func setSetting(path, key, value string) error {
	settings := loadSettings(path)
	settings[key] = value

	if err := saveSettings(path, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("  Set %s = %s\n", key, value)
	return nil
}

func loadSettings(path string) map[string]interface{} {
	settings := make(map[string]interface{})

	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	_ = json.Unmarshal(data, &settings)
	return settings
}

func saveSettings(path string, settings map[string]interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
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
