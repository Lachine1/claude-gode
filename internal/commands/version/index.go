package version

import (
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
	w := ctx.WriteOutput
	w("")
	w("  Claude Code (Go)")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Version:   " + types.Version)
	if types.BuildTime != "" {
		w("  Built:     " + types.BuildTime)
	}
	w("  Go:        " + runtime.Version())
	w("  OS/Arch:   " + runtime.GOOS + "/" + runtime.GOARCH)
	w("")
	return nil
}
