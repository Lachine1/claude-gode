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
		return listAgents()
	default:
		return fmt.Errorf("unknown agents action: %s (use list or status)", action)
	}
}

func listAgents() error {
	fmt.Println()
	fmt.Println("  Agents")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Println("  Main Agent (active)")
	fmt.Println("    Status:    running")
	fmt.Println("    Model:     claude-sonnet-4-20250514")
	fmt.Println("    Context:   primary conversation")
	fmt.Println()
	fmt.Println("  Sub-agents can be spawned for parallel tasks.")
	fmt.Println("  Use the Task tool to delegate work to sub-agents.")
	fmt.Println()
	return nil
}
