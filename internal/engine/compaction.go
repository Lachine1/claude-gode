package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// compactIfNeeded checks if compaction is needed and performs the appropriate tier.
func (e *QueryEngine) compactIfNeeded(ctx context.Context) error {
	if !e.shouldCompact() {
		return nil
	}

	tokens := e.estimateTokenCount(e.messages)

	if tokens < microCompactThreshold {
		return nil
	}

	if tokens < memoryCompactThreshold {
		if e.config.Debug {
			fmt.Printf("[engine] microcompact: estimated %d tokens\n", tokens)
		}
		e.microcompact()
		return nil
	}

	if tokens < fullCompactThreshold {
		if e.config.Debug {
			fmt.Printf("[engine] session memory compact: estimated %d tokens\n", tokens)
		}
		return e.sessionMemoryCompact()
	}

	if e.config.Debug {
		fmt.Printf("[engine] full compact: estimated %d tokens\n", tokens)
	}
	return e.fullCompact(ctx)
}

// microcompact removes messages from the local cache without invalidating the API context.
// This is the lightest form of compaction - it trims tool result content that is no longer needed.
func (e *QueryEngine) microcompact() {
	if len(e.messages) <= 2 {
		return
	}

	kept := 1
	for i := len(e.messages) - 1; i >= 0; i-- {
		msg := e.messages[i]
		if msg.Role == types.RoleUser {
			kept++
		}
		if kept > 10 {
			break
		}
	}

	if len(e.messages) > 20 {
		trimmed := e.messages[len(e.messages)-20:]

		firstUserIdx := -1
		for i, msg := range trimmed {
			if msg.Role == types.RoleUser {
				firstUserIdx = i
				break
			}
		}

		if firstUserIdx > 0 {
			trimmed = trimmed[firstUserIdx:]
		}

		e.messages = trimmed
	}
}

// sessionMemoryCompact creates a summary of the session to reduce context.
// It replaces older messages with a condensed summary while preserving the most recent exchanges.
func (e *QueryEngine) sessionMemoryCompact() error {
	if len(e.messages) <= 6 {
		return nil
	}

	keepCount := 6
	if keepCount > len(e.messages) {
		keepCount = len(e.messages)
	}

	oldMessages := e.messages[:len(e.messages)-keepCount]

	summary := e.buildSessionSummary(oldMessages)

	summaryMsg := types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{
				Type: types.ContentTypeText,
				Text: fmt.Sprintf("[Session summary of previous conversation]\n%s", summary),
			},
		},
	}

	e.messages = append([]types.Message{summaryMsg}, e.messages[len(e.messages)-keepCount:]...)

	return nil
}

// fullCompact uses a sub-query to summarize the conversation into a structured format.
// This is the most aggressive compaction tier.
func (e *QueryEngine) fullCompact(ctx context.Context) error {
	if len(e.messages) <= 4 {
		return nil
	}

	keepCount := 4
	oldMessages := e.messages[:len(e.messages)-keepCount]

	conversationText := e.formatConversationForSummary(oldMessages)

	summaryPrompt := fmt.Sprintf(
		`Summarize the following conversation. Focus on:
1. What the user was trying to accomplish
2. What tools were called and their key results
3. Any important findings or decisions made
4. Current state of the work

Keep the summary concise but complete. The summary will be used as context for continuing the conversation.

Conversation:
%s`,
		conversationText,
	)

	summaryMsg := types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{Type: types.ContentTypeText, Text: summaryPrompt},
		},
	}

	apiCfg := &types.APIConfig{
		APIKey:     e.config.APIKey,
		Model:      e.config.Model,
		MaxTokens:  4096,
		MaxRetries: 2,
		BaseURL:    e.config.BaseURL,
	}
	if apiCfg.Model == "" {
		apiCfg.Model = "claude-sonnet-4-20250514"
	}

	var summaryBuilder strings.Builder
	_, err := Query(
		ctx,
		apiCfg,
		[]types.Message{summaryMsg},
		"You are a helpful assistant that summarizes conversations concisely.",
		nil,
		func(token string) {
			summaryBuilder.WriteString(token)
		},
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("compact query failed: %w", err)
	}

	compactMsg := types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{
				Type: types.ContentTypeText,
				Text: fmt.Sprintf("[Compact conversation summary]\n%s", summaryBuilder.String()),
			},
		},
	}

	e.messages = append([]types.Message{compactMsg}, e.messages[len(e.messages)-keepCount:]...)

	return nil
}

// buildSessionSummary creates a brief summary of old messages.
func (e *QueryEngine) buildSessionSummary(messages []types.Message) string {
	var sb strings.Builder

	toolCallCount := 0
	userMsgCount := 0
	assistantMsgCount := 0

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			userMsgCount++
		case types.RoleAssistant:
			assistantMsgCount++
			for _, block := range msg.Content {
				if block.Type == types.ContentTypeToolUse {
					toolCallCount++
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("Previous conversation had %d user messages, %d assistant messages, and %d tool calls.",
		userMsgCount, assistantMsgCount, toolCallCount))

	if len(messages) > 0 {
		firstMsg := messages[0]
		if len(firstMsg.Content) > 0 && firstMsg.Content[0].Type == types.ContentTypeText {
			preview := firstMsg.Content[0].Text
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("\nFirst user message: %s", preview))
		}
	}

	return sb.String()
}

// formatConversationForSummary formats messages into a readable format for summarization.
func (e *QueryEngine) formatConversationForSummary(messages []types.Message) string {
	var sb strings.Builder

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			sb.WriteString("User: ")
			for _, block := range msg.Content {
				if block.Type == types.ContentTypeText {
					sb.WriteString(block.Text)
				} else if block.Type == types.ContentTypeToolResult && block.ToolResult != nil {
					sb.WriteString("[Tool result]\n")
				}
			}
			sb.WriteString("\n\n")
		case types.RoleAssistant:
			sb.WriteString("Assistant: ")
			for _, block := range msg.Content {
				if block.Type == types.ContentTypeText {
					sb.WriteString(block.Text)
				} else if block.Type == types.ContentTypeToolUse && block.ToolUse != nil {
					sb.WriteString(fmt.Sprintf("[Called tool: %s]", block.ToolUse.Name))
				}
			}
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// estimateTokenCount provides a rough estimate of token count.
// Uses ~4 characters per token as a rough heuristic.
func (e *QueryEngine) estimateTokenCount(messages []types.Message) int {
	totalChars := 0
	for _, msg := range messages {
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentTypeText:
				totalChars += len(block.Text)
			case types.ContentTypeToolUse:
				totalChars += len(block.ToolUse.Name) * 2
			case types.ContentTypeToolResult:
				for _, inner := range block.ToolResult.Content {
					if inner.Type == types.ContentTypeText {
						totalChars += len(inner.Text)
					}
				}
			}
		}
	}

	return totalChars / 4
}

// shouldCompact checks if the context is approaching token limits.
func (e *QueryEngine) shouldCompact() bool {
	tokens := e.estimateTokenCount(e.messages)
	return tokens > microCompactThreshold
}

// Compact forces context compaction regardless of thresholds.
func (e *QueryEngine) Compact(ctx context.Context) error {
	if len(e.messages) <= 4 {
		return nil
	}
	return e.fullCompact(ctx)
}
