package toolscmd

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /tools command.
func New(toolList []types.Tool) types.Command {
	return types.Command{
		Name:        "tools",
		Aliases:     []string{"tool"},
		Description: "List available tools",
		Usage:       "/tools",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleTools(ctx, toolList, args)
		},
	}
}

func handleTools(ctx *types.CommandContext, toolList []types.Tool, args []string) error {
	w := ctx.WriteOutput
	w("")
	w("  Available Tools")
	w("  ═══════════════════════════════════════")
	w("")

	for _, tool := range toolList {
		w(fmt.Sprintf("  %-20s %s", tool.Name(), tool.Description()))
	}

	w("")
	w(fmt.Sprintf("  Total: %d tools", len(toolList)))
	w("")
	return nil
}
