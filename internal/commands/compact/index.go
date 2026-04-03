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

	w := ctx.WriteOutput
	messages := eng.GetMessages()
	usage := eng.GetUsage()
	cost := eng.GetTotalCost()
	tokens := estimateTokens(messages)

	w("")
	w("  Context Compaction")
	w("  ═══════════════════════════════════════")
	w("")
	w(fmt.Sprintf("  Messages:    %d", len(messages)))
	w(fmt.Sprintf("  Est. tokens: ~%d", tokens))
	w(fmt.Sprintf("  Input:       %d tokens", usage.InputTokens))
	w(fmt.Sprintf("  Output:      %d tokens", usage.OutputTokens))
	w(fmt.Sprintf("  Cache read:  %d tokens", usage.CacheRead))
	w(fmt.Sprintf("  Cache write: %d tokens", usage.CacheWrite))
	w(fmt.Sprintf("  Est. cost:   $%.4f", cost))
	w("")

	if tokens < 100000 {
		w("  Context is within limits. No compaction needed.")
		w("")
		return nil
	}

	w("  Running compaction...")
	if err := eng.Compact(context.Background()); err != nil {
		w("  Compaction error: " + err.Error())
		w("")
		return nil
	}

	newMessages := eng.GetMessages()
	w(fmt.Sprintf("  Compaction complete. Messages: %d -> %d", len(messages), len(newMessages)))
	w("")
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
