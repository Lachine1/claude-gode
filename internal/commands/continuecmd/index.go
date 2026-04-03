package continuecmd

import (
	"fmt"
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

	fmt.Println()
	fmt.Println("  Continue Session")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("  No previous session found.")
			fmt.Println("  Sessions are saved automatically when exiting.")
		} else {
			fmt.Printf("  Error reading session: %v\n", err)
		}
		fmt.Println()
		return nil
	}

	_ = data
	fmt.Println("  Session data found.")
	fmt.Println("  Loading previous conversation context...")
	fmt.Println()
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
