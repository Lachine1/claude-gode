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

	fmt.Println()
	fmt.Println("  Session Usage")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Messages:      %d\n", messages)
	fmt.Printf("  Input tokens:  %d\n", usage.InputTokens)
	fmt.Printf("  Output tokens: %d\n", usage.OutputTokens)
	fmt.Printf("  Cache read:    %d\n", usage.CacheRead)
	fmt.Printf("  Cache write:   %d\n", usage.CacheWrite)

	totalTokens := usage.InputTokens + usage.OutputTokens
	fmt.Printf("  Total tokens:  %d\n", totalTokens)
	fmt.Println()
	fmt.Printf("  Estimated cost: $%.4f\n", totalCost)
	fmt.Println()
	return nil
}
