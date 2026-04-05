package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

const (
	defaultMaxTokens       = 8192
	contextWindowTokens    = 200000
	microCompactThreshold  = 100000
	memoryCompactThreshold = 140000
	fullCompactThreshold   = 170000
)

// QueryEngine manages the conversation loop with LLM API calls, tool execution, and compaction.
type QueryEngine struct {
	config    EngineConfig
	messages  []types.Message
	usage     types.Usage
	totalCost float64
	abortCtrl *context.CancelFunc
	fileCache *FileStateCache
	toolMap   map[string]types.Tool
}

// EngineConfig holds configuration for the QueryEngine.
type EngineConfig struct {
	Cwd          string
	Tools        []types.Tool
	Model        string
	MaxTokens    int
	MaxBudgetUSD float64
	CustomPrompt string
	AppendPrompt string
	Debug        bool
	Verbose      bool
	APIKey       string
	BaseURL      string
}

// NewQueryEngine creates a new QueryEngine with the given configuration.
func NewQueryEngine(config EngineConfig) *QueryEngine {
	toolMap := make(map[string]types.Tool)
	for _, t := range config.Tools {
		toolMap[t.Name()] = t
	}

	return &QueryEngine{
		config:    config,
		messages:  make([]types.Message, 0),
		toolMap:   toolMap,
		fileCache: NewFileStateCache(),
	}
}

// SetModel updates the model used by the QueryEngine at runtime.
func (e *QueryEngine) SetModel(model string) {
	e.config.Model = model
}

// SubmitMessage is the main entry point. It adds a user message, calls the API,
// executes any tool calls, and handles compaction.
func (e *QueryEngine) SubmitMessage(ctx context.Context, userMessage string, onEvent func(Event)) error {
	if e.abortCtrl != nil {
		(*e.abortCtrl)()
	}

	ctx, cancel := context.WithCancel(ctx)
	e.abortCtrl = &cancel

	userMsg := types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{Type: types.ContentTypeText, Text: userMessage},
		},
		Timestamp: time.Now().Unix(),
	}
	e.messages = append(e.messages, userMsg)

	if err := e.runLoop(ctx, onEvent); err != nil {
		if onEvent != nil {
			onEvent(ErrorEvent{Err: err})
		}
		return err
	}

	if err := e.compactIfNeeded(ctx); err != nil {
		if e.config.Debug {
			fmt.Printf("[engine] compaction error: %v\n", err)
		}
	}

	return nil
}

// Abort cancels the current operation.
func (e *QueryEngine) Abort() {
	if e.abortCtrl != nil {
		(*e.abortCtrl)()
	}
}

// GetMessages returns the current message history.
func (e *QueryEngine) GetMessages() []types.Message {
	return e.messages
}

// GetUsage returns the accumulated token usage.
func (e *QueryEngine) GetUsage() types.Usage {
	return e.usage
}

// GetTotalCost returns the accumulated cost in USD.
func (e *QueryEngine) GetTotalCost() float64 {
	return e.totalCost
}

func (e *QueryEngine) runLoop(ctx context.Context, onEvent func(Event)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if e.config.MaxBudgetUSD > 0 && e.totalCost >= e.config.MaxBudgetUSD {
			return fmt.Errorf("budget exceeded: $%.2f / $%.2f", e.totalCost, e.config.MaxBudgetUSD)
		}

		apiCfg := &types.APIConfig{
			Model:      e.config.Model,
			MaxTokens:  e.config.MaxTokens,
			MaxRetries: 3,
			APIKey:     e.config.APIKey,
			BaseURL:    e.config.BaseURL,
		}
		if apiCfg.Model == "" {
			apiCfg.Model = "claude-sonnet-4-20250514"
		}
		if apiCfg.MaxTokens == 0 {
			apiCfg.MaxTokens = defaultMaxTokens
		}

		systemPrompt := e.buildSystemPrompt()
		apiMessages := e.buildAPIMessages()

		var textBuffer strings.Builder
		var toolCalls []ToolCall
		var toolCallResults []types.ContentBlock

		resp, err := Query(
			ctx,
			apiCfg,
			apiMessages,
			systemPrompt,
			e.config.Tools,
			func(token string) {
				textBuffer.WriteString(token)
				if onEvent != nil {
					onEvent(TextEvent{Token: token})
				}
			},
			func(tc ToolCall) {
				toolCalls = append(toolCalls, tc)
				if onEvent != nil {
					onEvent(ToolCallEvent{
						ToolUseID: tc.ToolUseID,
						Name:      tc.Name,
						Input:     tc.Input,
					})
				}
			},
			func(text string) {
				if onEvent != nil {
					onEvent(ThinkingEvent{Text: text})
				}
			},
		)
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		e.usage.InputTokens += resp.Usage.InputTokens
		e.usage.OutputTokens += resp.Usage.OutputTokens
		e.usage.CacheRead += resp.Usage.CacheRead
		e.usage.CacheWrite += resp.Usage.CacheWrite
		e.totalCost = estimateCost(e.usage, e.config.Model)

		if onEvent != nil {
			onEvent(UsageEvent{Usage: e.usage})
		}

		assistantContent := make([]types.ContentBlock, 0)

		if textBuffer.Len() > 0 {
			assistantContent = append(assistantContent, types.ContentBlock{
				Type: types.ContentTypeText,
				Text: textBuffer.String(),
			})
		}

		for _, tc := range toolCalls {
			assistantContent = append(assistantContent, types.ContentBlock{
				Type: types.ContentTypeToolUse,
				ToolUse: &types.ToolUseContent{
					ID:    tc.ToolUseID,
					Name:  tc.Name,
					Input: e.parseToolInput(tc.Input),
				},
			})
		}

		assistantMsg := types.Message{
			Role:      types.RoleAssistant,
			Content:   assistantContent,
			Timestamp: time.Now().Unix(),
		}
		e.messages = append(e.messages, assistantMsg)

		if len(toolCalls) == 0 {
			if onEvent != nil {
				onEvent(DoneEvent{StopReason: resp.StopReason})
			}
			return nil
		}

		toolResultBlocks := make([]types.ContentBlock, 0)

		for _, tc := range toolCalls {
			result, err := e.executeToolCall(ctx, tc)
			if err != nil {
				result = &types.ToolResult[json.RawMessage]{
					IsError:      true,
					ErrorMessage: err.Error(),
				}
			}

			contentText := string(result.Data)
			if result.IsError && result.ErrorMessage != "" {
				contentText = result.ErrorMessage
			}

			toolResultBlocks = append(toolResultBlocks, types.ContentBlock{
				Type: types.ContentTypeToolResult,
				ToolResult: &types.ToolResultContent{
					ToolUseID: tc.ToolUseID,
					Content: []types.ContentBlock{
						{Type: types.ContentTypeText, Text: contentText},
					},
					IsError: result.IsError,
				},
			})

			if onEvent != nil {
				onEvent(ToolResultEvent{
					ToolUseID: tc.ToolUseID,
					Result:    contentText,
					IsError:   result.IsError,
				})
			}
		}

		toolMsg := types.Message{
			Role:      types.RoleUser,
			Content:   toolResultBlocks,
			Timestamp: time.Now().Unix(),
		}
		e.messages = append(e.messages, toolMsg)
		_ = toolCallResults
	}
}

func (e *QueryEngine) executeToolCall(ctx context.Context, tc ToolCall) (*types.ToolResult[json.RawMessage], error) {
	tool, ok := e.toolMap[tc.Name]
	if !ok {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Unknown tool: %s", tc.Name),
		}, nil
	}

	abortCh := ctx.Done()
	abortSignal := make(chan struct{})
	go func() {
		<-abortCh
		close(abortSignal)
	}()

	toolCtx := &types.ToolContext{
		Cwd:           e.config.Cwd,
		AbortSignal:   abortSignal,
		GetAppState:   func() map[string]interface{} { return nil },
		SetAppState:   func(map[string]interface{}) {},
		Messages:      e.messages,
		Debug:         e.config.Debug,
		Verbose:       e.config.Verbose,
		MainLoopModel: e.config.Model,
	}

	result, err := tool.Execute(toolCtx, tc.Input, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e *QueryEngine) parseToolInput(raw json.RawMessage) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return map[string]interface{}{"raw": string(raw)}
	}
	return m
}

func estimateCost(usage types.Usage, model string) float64 {
	var inputCost, outputCost, cacheReadCost, cacheWriteCost float64

	switch {
	case strings.Contains(model, "opus"):
		inputCost = 15.0
		outputCost = 75.0
		cacheReadCost = 1.875
		cacheWriteCost = 18.75
	case strings.Contains(model, "sonnet"):
		inputCost = 3.0
		outputCost = 15.0
		cacheReadCost = 0.30
		cacheWriteCost = 3.75
	case strings.Contains(model, "haiku"):
		inputCost = 0.80
		outputCost = 4.0
		cacheReadCost = 0.08
		cacheWriteCost = 1.0
	default:
		inputCost = 3.0
		outputCost = 15.0
		cacheReadCost = 0.30
		cacheWriteCost = 3.75
	}

	total := 0.0
	total += float64(usage.InputTokens) / 1_000_000 * inputCost
	total += float64(usage.OutputTokens) / 1_000_000 * outputCost
	total += float64(usage.CacheRead) / 1_000_000 * cacheReadCost
	total += float64(usage.CacheWrite) / 1_000_000 * cacheWriteCost
	return total
}
