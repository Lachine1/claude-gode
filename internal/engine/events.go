package engine

import (
	"encoding/json"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Event is the base interface for all engine events
type Event interface {
	isEvent()
}

// TextEvent is emitted when text tokens are streamed
type TextEvent struct {
	Token string
}

func (TextEvent) isEvent() {}

// ToolCallEvent is emitted when the LLM requests a tool call
type ToolCallEvent struct {
	ToolUseID string
	Name      string
	Input     json.RawMessage
}

func (ToolCallEvent) isEvent() {}

// ToolResultEvent is emitted after a tool has been executed
type ToolResultEvent struct {
	ToolUseID string
	Result    string
	IsError   bool
}

func (ToolResultEvent) isEvent() {}

// ThinkingEvent is emitted when the model produces thinking content
type ThinkingEvent struct {
	Text string
}

func (ThinkingEvent) isEvent() {}

// UsageEvent is emitted when usage information is received
type UsageEvent struct {
	Usage types.Usage
}

func (UsageEvent) isEvent() {}

// DoneEvent is emitted when the response is complete
type DoneEvent struct {
	StopReason string
}

func (DoneEvent) isEvent() {}

// ErrorEvent is emitted when an error occurs
type ErrorEvent struct {
	Err error
}

func (ErrorEvent) isEvent() {}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ToolUseID string
	Name      string
	Input     json.RawMessage
}

// QueryResponse represents the complete response from a Query call
type QueryResponse struct {
	ID         string
	Content    []types.ContentBlock
	StopReason string
	Usage      types.Usage
}
