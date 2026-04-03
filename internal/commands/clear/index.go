package clear

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /clear command.
func New() types.Command {
	return types.Command{
		Name:        "clear",
		Aliases:     []string{"reset"},
		Description: "Clear conversation history",
		Usage:       "/clear",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleClear(ctx, args)
		},
	}
}

func handleClear(ctx *types.CommandContext, args []string) error {
	messages := ctx.GetMessages()

	w := ctx.WriteOutput
	w("")
	w("  Clear Conversation")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Cleared " + fmt.Sprint(len(messages)) + " messages from history.")
	w("")

	ctx.SetMessages([]types.Message{})

	return nil
}
