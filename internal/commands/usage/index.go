package usage

import (
	"fmt"

	"github.com/Lachine1/claude-gode/internal/engine"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /usage command.
func New(eng *engine.QueryEngine) types.Command {
	return types.Command{
		Name:        "usage",
		Aliases:     []string{"cost", "tokens"},
		Description: "Show token usage and cost",
		Usage:       "/usage",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleUsage(ctx, eng)
		},
	}
}

func handleUsage(ctx *types.CommandContext, eng *engine.QueryEngine) error {
	var usage types.Usage
	var totalCost float64
	var messages int

	if eng != nil {
		usage = eng.GetUsage()
		totalCost = eng.GetTotalCost()
		messages = len(eng.GetMessages())
	}

	w := ctx.WriteOutput
	w("")
	w("  Session Usage")
	w("  ═══════════════════════════════════════")
	w("")
	w(fmt.Sprintf("  Messages:      %d", messages))
	w(fmt.Sprintf("  Input tokens:  %d", usage.InputTokens))
	w(fmt.Sprintf("  Output tokens: %d", usage.OutputTokens))
	w(fmt.Sprintf("  Cache read:    %d", usage.CacheRead))
	w(fmt.Sprintf("  Cache write:   %d", usage.CacheWrite))

	totalTokens := usage.InputTokens + usage.OutputTokens
	w(fmt.Sprintf("  Total tokens:  %d", totalTokens))
	w("")
	w(fmt.Sprintf("  Estimated cost: $%.4f", totalCost))
	w("")
	return nil
}
