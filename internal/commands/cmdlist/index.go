package cmdlist

import (
	"fmt"
	"strings"

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
			return handleCommands(ctx, cmds, args)
		},
	}
}

func handleCommands(ctx *types.CommandContext, cmds []types.Command, args []string) error {
	w := ctx.WriteOutput
	w("")
	w("  Available Commands")
	w("  ═══════════════════════════════════════")
	w("")

	for _, cmd := range cmds {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = " (" + strings.Join(cmd.Aliases, ", ") + ")"
		}
		w(fmt.Sprintf("  /%-20s %s", cmd.Name+aliases, cmd.Description))
	}

	w("")
	w(fmt.Sprintf("  Total: %d commands", len(cmds)))
	w("")
	return nil
}
