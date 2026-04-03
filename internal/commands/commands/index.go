package commands

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /commands command.
func New(cmds []types.Command) types.Command {
	return types.Command{
		Name:        "commands",
		Aliases:     []string{"cmds"},
		Description: "List available commands",
		Usage:       "/commands",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleCommands(ctx, cmds)
		},
	}
}

func handleCommands(ctx *types.CommandContext, cmds []types.Command) error {
	w := ctx.WriteOutput
	w("")
	w("  Available Commands")
	w("  ═══════════════════════════════════════")
	w("")

	for _, cmd := range cmds {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = " (" + joinAliases(cmd.Aliases) + ")"
		}
		w(fmt.Sprintf("  /%-20s %s", cmd.Name+aliases, cmd.Description))
	}

	w("")
	w(fmt.Sprintf("  Total: %d commands", len(cmds)))
	w("")
	return nil
}

func joinAliases(aliases []string) string {
	result := ""
	for i, a := range aliases {
		if i > 0 {
			result += ", "
		}
		result += a
	}
	return result
}
