package plan

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /plan command.
func New() types.Command {
	return types.Command{
		Name:        "plan",
		Aliases:     []string{},
		Description: "Enter plan mode",
		Usage:       "/plan",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handlePlan(ctx, args)
		},
	}
}

func handlePlan(ctx *types.CommandContext, args []string) error {
	fmt.Println()
	fmt.Println("  Plan Mode")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Println("  Switched to plan mode.")
	fmt.Println()
	fmt.Println("  In plan mode, the assistant will:")
	fmt.Println("  • Analyze the task and propose a plan")
	fmt.Println("  • Not execute any tools or make changes")
	fmt.Println("  • Wait for your approval before proceeding")
	fmt.Println()
	fmt.Println("  Use /plan to exit plan mode and return to normal.")
	fmt.Println()
	return nil
}
