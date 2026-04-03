package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
			return handleSettings(ctx, cfg, args)
		},
	}
}

func handleSettings(ctx *types.CommandContext, cfg *svcconfig.Config, args []string) error {
	settingsPath := filepath.Join(homeDir(), ".claude", "settings.json")

	if len(args) == 0 {
		return showSettings(ctx, cfg, settingsPath)
	}

	if len(args) == 1 {
		return getSetting(ctx, settingsPath, args[0])
	}

	value := args[1]
	for _, a := range args[2:] {
		value += " " + a
	}
	return setSetting(ctx, settingsPath, args[0], value)
}

func showSettings(ctx *types.CommandContext, cfg *svcconfig.Config, path string) error {
	settings := loadSettings(path)
	w := ctx.WriteOutput

	w("")
	w("  Settings")
	w("  ═══════════════════════════════════════")
	w("")
	w(fmt.Sprintf("  %-25s %s", "Model:", cfg.Model()))
	w(fmt.Sprintf("  %-25s %d", "Max Tokens:", cfg.MaxTokens()))
	w(fmt.Sprintf("  %-25s %s", "Permission Mode:", cfg.PermissionMode()))
	w("")

	if len(settings) > 0 {
		w("  Custom Settings:")
		for k, v := range settings {
			w(fmt.Sprintf("    %-23s %v", k, v))
		}
		w("")
	}

	w("  Use /settings <key> <value> to set a value.")
	w("  Use /settings <key> to view a specific value.")
	w("")
	return nil
}

func getSetting(ctx *types.CommandContext, path, key string) error {
	settings := loadSettings(path)
	w := ctx.WriteOutput

	val, ok := settings[key]
	if !ok {
		w("  Setting '" + key + "' not found.")
		return nil
	}

	w(fmt.Sprintf("  %s = %v", key, val))
	return nil
}

func setSetting(ctx *types.CommandContext, path, key, value string) error {
	w := ctx.WriteOutput
	settings := loadSettings(path)
	settings[key] = value

	if err := saveSettings(path, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	w("  Set " + key + " = " + value)
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
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return "."
}
