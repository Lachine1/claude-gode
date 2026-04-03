package configcmd

import (
	"fmt"
	"os"
	"path/filepath"

	svcconfig "github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /config command.
func New(cfg *svcconfig.Config) types.Command {
	return types.Command{
		Name:        "config",
		Aliases:     []string{"cfg"},
		Description: "Show current config",
		Usage:       "/config",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleConfig(ctx, cfg, args)
		},
	}
}

func handleConfig(ctx *types.CommandContext, cfg *svcconfig.Config, args []string) error {
	w := ctx.WriteOutput
	w("")
	w("  Current Configuration")
	w("  ═══════════════════════════════════════")
	w("")
	w(fmt.Sprintf("  %-25s %s", "Model:", cfg.Model()))
	w(fmt.Sprintf("  %-25s %d", "Max Tokens:", cfg.MaxTokens()))
	w(fmt.Sprintf("  %-25s %s", "Permission Mode:", cfg.PermissionMode()))

	apiKey := cfg.APIKey()
	if apiKey != "" {
		var masked string
		if len(apiKey) <= 8 {
			masked = "****"
		} else {
			masked = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		}
		w(fmt.Sprintf("  %-25s %s", "API Key:", masked))
	} else {
		w(fmt.Sprintf("  %-25s %s", "API Key:", "(not set)"))
	}

	settings := cfg.All()
	if len(settings) > 0 {
		w("")
		w("  Settings:")
		for k, v := range settings {
			w(fmt.Sprintf("    %-23s %v", k, v))
		}
	}

	w("")

	configPath := filepath.Join(homeDir(), ".claude", "settings.json")
	w(fmt.Sprintf("  Config file: %s", configPath))
	w("")
	return nil
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
