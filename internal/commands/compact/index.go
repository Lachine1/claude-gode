package compact

import (
	"context"
	"fmt"

	"github.com/Lachine1/claude-gode/internal/engine"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /compact command.
func New(eng *engine.QueryEngine) types.Command {
	return types.Command{
		Name:        "compact",
		Aliases:     []string{},
		Description: "Trigger context compaction",
		Usage:       "/compact",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleCompact(ctx, eng)
		},
	}
}

func handleCompact(ctx *types.CommandContext, eng *engine.QueryEngine) error {
	if eng == nil {
		return fmt.Errorf("query engine not initialized")
	}

	messages := eng.GetMessages()
	usage := eng.GetUsage()
	cost := eng.GetTotalCost()
	tokens := estimateTokens(messages)

	fmt.Println()
	fmt.Println("  Context Compaction")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Messages:    %d\n", len(messages))
	fmt.Printf("  Est. tokens: ~%d\n", tokens)
	fmt.Printf("  Input:       %d tokens\n", usage.InputTokens)
	fmt.Printf("  Output:      %d tokens\n", usage.OutputTokens)
	fmt.Printf("  Cache read:  %d tokens\n", usage.CacheRead)
	fmt.Printf("  Cache write: %d tokens\n", usage.CacheWrite)
	fmt.Printf("  Est. cost:   $%.4f\n", cost)
	fmt.Println()

	if tokens < 100000 {
		fmt.Println("  Context is within limits. No compaction needed.")
		fmt.Println()
		return nil
	}

	fmt.Println("  Running compaction...")
	if err := eng.Compact(context.Background()); err != nil {
		fmt.Printf("  Compaction error: %v\n", err)
		fmt.Println()
		return nil
	}

	newMessages := eng.GetMessages()
	fmt.Printf("  Compaction complete. Messages: %d -> %d\n", len(messages), len(newMessages))
	fmt.Println()
	return nil
}

func estimateTokens(messages []types.Message) int {
	totalChars := 0
	for _, msg := range messages {
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentTypeText:
				totalChars += len(block.Text)
			case types.ContentTypeToolUse:
				if block.ToolUse != nil {
					totalChars += len(block.ToolUse.Name) * 2
				}
			case types.ContentTypeToolResult:
				if block.ToolResult != nil {
					for _, inner := range block.ToolResult.Content {
						if inner.Type == types.ContentTypeText {
							totalChars += len(inner.Text)
						}
					}
				}
			}
		}
	}
	return totalChars / 4
}
