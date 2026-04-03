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
			return handleConfig(cfg, args)
		},
	}
}

func handleConfig(cfg *svcconfig.Config, args []string) error {
	fmt.Println()
	fmt.Println("  Current Configuration")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  %-25s %s\n", "Model:", cfg.Model)
	fmt.Printf("  %-25s %d\n", "Max Tokens:", cfg.MaxTokens)
	fmt.Printf("  %-25s %s\n", "Permission Mode:", cfg.PermissionMode)

	if cfg.APIKey != "" {
		masked := cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
		fmt.Printf("  %-25s %s\n", "API Key:", masked)
	} else {
		fmt.Printf("  %-25s %s\n", "API Key:", "(not set)")
	}

	if len(cfg.Settings) > 0 {
		fmt.Println()
		fmt.Println("  Settings:")
		for k, v := range cfg.Settings {
			fmt.Printf("    %-23s %v\n", k, v)
		}
	}

	fmt.Println()

	configPath := filepath.Join(homeDir(), ".claude", "settings.json")
	fmt.Printf("  Config file: %s\n", configPath)
	fmt.Println()
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
