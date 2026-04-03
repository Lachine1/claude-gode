package permissionmode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	svcconfig "github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

var validModes = []string{
	"default",
	"acceptEdits",
	"bypassPermissions",
	"autoEdit",
	"plan",
}

var modeDescriptions = map[string]string{
	"default":           "Ask for permission before executing tools",
	"acceptEdits":       "Allow edits without asking, but confirm destructive actions",
	"bypassPermissions": "Never ask for permission (use with caution)",
	"autoEdit":          "Automatically accept edits and file changes",
	"plan":              "Only plan, never execute tools",
}

// New creates the /permission-mode command.
func New(cfg *svcconfig.Config) types.Command {
	return types.Command{
		Name:        "permission-mode",
		Aliases:     []string{"perm"},
		Description: "Change permission mode",
		Usage:       "/permission-mode [mode]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handlePermissionMode(ctx, args, cfg)
		},
	}
}

func handlePermissionMode(ctx *types.CommandContext, args []string, cfg *svcconfig.Config) error {
	if len(args) == 0 {
		return showPermissionModes(cfg)
	}
	return setPermissionMode(cfg, args[0])
}

func showPermissionModes(cfg *svcconfig.Config) error {
	currentMode := cfg.PermissionMode
	if currentMode == "" {
		currentMode = "default"
	}

	fmt.Println()
	fmt.Println("  Permission Modes")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	for _, mode := range validModes {
		marker := " "
		if mode == currentMode {
			marker = "→"
		}
		desc := modeDescriptions[mode]
		fmt.Printf("  %s %-25s %s\n", marker, mode, desc)
	}

	fmt.Println()
	fmt.Println("  Use /permission-mode <mode> to change.")
	fmt.Println()
	return nil
}

func setPermissionMode(cfg *svcconfig.Config, mode string) error {
	valid := false
	for _, m := range validModes {
		if m == mode {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid permission mode: %s\nValid modes: %s", mode, validModes)
	}

	cfg.PermissionMode = mode

	settingsPath := filepath.Join(homeDir(), ".claude", "settings.json")
	settings := loadSettings(settingsPath)
	settings["permission_mode"] = mode

	if err := saveSettings(settingsPath, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("  Permission mode set to: %s\n", mode)
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
