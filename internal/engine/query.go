package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Query sends a message to the Anthropic API and streams the response.
// It handles retries with exponential backoff for rate limits.
func Query(
	ctx context.Context,
	cfg *types.APIConfig,
	messages []types.Message,
	systemPrompt string,
	tools []types.Tool,
	onToken func(token string),
	onToolCall func(toolCall ToolCall),
	onThinking func(text string),
) (*QueryResponse, error) {
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := doQueryOnce(ctx, cfg, messages, systemPrompt, tools, onToken, onToolCall, onThinking)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if isRetryableError(err) {
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("query failed after %d retries: %w", maxRetries, lastErr)
}

func doQueryOnce(
	ctx context.Context,
	cfg *types.APIConfig,
	messages []types.Message,
	systemPrompt string,
	tools []types.Tool,
	onToken func(token string),
	onToolCall func(toolCall ToolCall),
	onThinking func(text string),
) (*QueryResponse, error) {
	body, err := doAPIRequest(cfg, messages, systemPrompt, tools)
	if err != nil {
		return nil, err
	}

	resp := &QueryResponse{}

	parser := newSSEParser(onToken, onToolCall, onThinking)
	parser.onUsage = func(usage types.Usage) {
		resp.Usage.InputTokens += usage.InputTokens
		resp.Usage.OutputTokens += usage.OutputTokens
		resp.Usage.CacheRead += usage.CacheRead
		resp.Usage.CacheWrite += usage.CacheWrite
	}
	parser.onDone = func(stopReason string) {
		resp.StopReason = stopReason
	}

	if err := parser.parse(body); err != nil {
		return resp, err
	}

	return resp, nil
}

func isRetryableError(err error) bool {
	msg := err.Error()
	retryable := []string{
		"429",
		"rate_limit",
		"500",
		"502",
		"503",
		"529",
		"overloaded",
		"retries exhausted",
	}
	for _, sub := range retryable {
		if strings.Contains(msg, sub) {
			return true
		}
	}
	return false
}
