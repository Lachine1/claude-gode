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
			return handleTools(toolList, args)
		},
	}
}

func handleTools(toolList []types.Tool, args []string) error {
	fmt.Println()
	fmt.Println("  Available Tools")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	for _, tool := range toolList {
		fmt.Printf("  %-20s %s\n", tool.Name(), tool.Description())
	}

	fmt.Println()
	fmt.Printf("  Total: %d tools\n", len(toolList))
	fmt.Println()
	return nil
}
