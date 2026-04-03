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
			return handleCommands(cmds, args)
		},
	}
}

func handleCommands(cmds []types.Command, args []string) error {
	fmt.Println()
	fmt.Println("  Available Commands")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	for _, cmd := range cmds {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = " (" + strings.Join(cmd.Aliases, ", ") + ")"
		}
		fmt.Printf("  /%-20s %s\n", cmd.Name+aliases, cmd.Description)
	}

	fmt.Println()
	fmt.Printf("  Total: %d commands\n", len(cmds))
	fmt.Println()
	return nil
}
