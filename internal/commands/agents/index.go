package agents

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /agents command.
func New() types.Command {
	return types.Command{
		Name:        "agents",
		Aliases:     []string{"agent"},
		Description: "List/manage agents",
		Usage:       "/agents [list|status]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleAgents(ctx, args)
		},
	}
}

func handleAgents(ctx *types.CommandContext, args []string) error {
	action := "list"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "list", "status":
		return listAgents(ctx)
	default:
		return fmt.Errorf("unknown agents action: %s (use list or status)", action)
	}
}

func listAgents(ctx *types.CommandContext) error {
	w := ctx.WriteOutput
	w("")
	w("  Agents")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Main Agent (active)")
	w("    Status:    running")
	w("    Model:     claude-sonnet-4-20250514")
	w("    Context:   primary conversation")
	w("")
	w("  Sub-agents can be spawned for parallel tasks.")
	w("  Use the Task tool to delegate work to sub-agents.")
	w("")
	return nil
}
