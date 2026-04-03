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

	fmt.Println()
	fmt.Println("  Clear Conversation")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Cleared %d messages from history.\n", len(messages))
	fmt.Println()

	ctx.SetMessages([]types.Message{})

	return nil
}
