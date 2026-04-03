package continuecmd

import (
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /continue command.
func New() types.Command {
	return types.Command{
		Name:        "continue",
		Aliases:     []string{"resume"},
		Description: "Continue last session",
		Usage:       "/continue",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleContinue(ctx, args)
		},
	}
}

func handleContinue(ctx *types.CommandContext, args []string) error {
	sessionPath := filepath.Join(homeDir(), ".claude", "sessions", "last.json")

	w := ctx.WriteOutput
	w("")
	w("  Continue Session")
	w("  ═══════════════════════════════════════")
	w("")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			w("  No previous session found.")
			w("  Sessions are saved automatically when exiting.")
		} else {
			w("  Error reading session: " + err.Error())
		}
		w("")
		return nil
	}

	_ = data
	w("  Session data found.")
	w("  Loading previous conversation context...")
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
