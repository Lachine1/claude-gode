package plan

import (
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
	w := ctx.WriteOutput
	w("")
	w("  Plan Mode")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Switched to plan mode.")
	w("")
	w("  In plan mode, the assistant will:")
	w("  • Analyze the task and propose a plan")
	w("  • Not execute any tools or make changes")
	w("  • Wait for your approval before proceeding")
	w("")
	w("  Use /plan to exit plan mode and return to normal.")
	w("")
	return nil
}
