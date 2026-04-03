package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	svcconfig "github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

var availableModels = []string{
	"claude-sonnet-4-20250514",
	"claude-sonnet-4-5-20250929",
	"claude-opus-4-20250514",
	"claude-opus-4-1-20250805",
	"claude-haiku-4-20250514",
}

var modelDescriptions = map[string]string{
	"claude-sonnet-4-20250514":   "Balanced performance and speed (default)",
	"claude-sonnet-4-5-20250929": "Latest Sonnet - improved reasoning",
	"claude-opus-4-20250514":     "Most capable model for complex tasks",
	"claude-opus-4-1-20250805":   "Latest Opus - enhanced capabilities",
	"claude-haiku-4-20250514":    "Fastest model for quick tasks",
}

// New creates the /models command.
func New(cfg *svcconfig.Config) types.Command {
	return types.Command{
		Name:        "models",
		Aliases:     []string{"model"},
		Description: "List and switch models",
		Usage:       "/models [model-name]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleModels(cfg, args)
		},
	}
}

func handleModels(cfg *svcconfig.Config, args []string) error {
	if len(args) > 0 {
		return switchModel(cfg, args[0])
	}
	return listModels(cfg)
}

func listModels(cfg *svcconfig.Config) error {
	currentModel := cfg.Model

	fmt.Println()
	fmt.Println("  Available Models")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	for _, model := range availableModels {
		marker := " "
		if model == currentModel {
			marker = "→"
		}
		desc := modelDescriptions[model]
		fmt.Printf("  %s %-35s %s\n", marker, model, desc)
	}

	fmt.Println()
	fmt.Println("  Use /models <model-name> to switch.")
	fmt.Println()
	return nil
}

func switchModel(cfg *svcconfig.Config, modelName string) error {
	found := false
	for _, m := range availableModels {
		if m == modelName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unknown model: %s\nAvailable models: %s", modelName, strings.Join(availableModels, ", "))
	}

	settingsPath := filepath.Join(homeDir(), ".claude", "settings.json")
	settings := loadSettings(settingsPath)
	settings["model"] = modelName

	if err := saveSettings(settingsPath, settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Printf("  Switched to model: %s\n", modelName)
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
