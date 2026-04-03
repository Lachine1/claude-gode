package cli

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Lachine1/claude-gode/internal/bootstrap"
	"github.com/Lachine1/claude-gode/internal/tui"
)

// Run is the main entry point for the CLI
func Run(args []string, version, buildTime string) error {
	// Handle --version
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("claude-gode %s", version)
			if buildTime != "" {
				fmt.Printf(" (built %s)", buildTime)
			}
			fmt.Println()
			return nil
		}
		if arg == "--help" || arg == "-h" {
			printHelp()
			return nil
		}
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(nil, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Bootstrap: load config, auth, settings
	state, err := bootstrap.Initialize()
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Launch the TUI
	return tui.Run(ctx, state, args)
}

func printHelp() {
	fmt.Println(`Claude Code - AI-powered coding assistant

Usage:
  claude-gode [prompt]
  claude-gode [flags]

Flags:
  -h, --help       Show this help
  -v, --version    Show version
  -p, --print      Non-interactive output
  -c, --continue   Continue the last session
  --resume <id>    Resume a specific session
  --model <model>  Override the model
  --debug          Enable debug mode
  --settings <path> Path to settings file

Permission Modes:
  --permission-mode default    Ask for permission before executing tools
  --permission-mode auto       Auto-approve safe tools
  --permission-mode yolo       Approve all tools without asking
  --permission-mode plan       Plan mode - no tool execution

Examples:
  claude-gode "Refactor the auth module"
  claude-gode --permission-mode yolo "Fix all lint errors"
  claude-gode -c "Continue from where we left off"`)
}
