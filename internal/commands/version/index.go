package version

import (
	"fmt"
	"runtime"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /version command.
func New() types.Command {
	return types.Command{
		Name:        "version",
		Aliases:     []string{"ver", "-v", "--version"},
		Description: "Show version",
		Usage:       "/version",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleVersion(ctx, args)
		},
	}
}

func handleVersion(ctx *types.CommandContext, args []string) error {
	fmt.Println()
	fmt.Println("  Claude Code (Go)")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Version:   %s\n", types.Version)
	if types.BuildTime != "" {
		fmt.Printf("  Built:     %s\n", types.BuildTime)
	}
	fmt.Printf("  Go:        %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	return nil
}
