package types

import "encoding/json"

// ToolResult is the result of executing a tool
type ToolResult[T any] struct {
	Data         T         `json:"data"`
	NewMessages  []Message `json:"new_messages,omitempty"`
	IsError      bool      `json:"is_error,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// ToolCallProgress is a callback for reporting tool progress
type ToolCallProgress func(progress ToolProgress)

// ToolProgress represents progress of a tool execution
type ToolProgress struct {
	ToolUseID string          `json:"tool_use_id"`
	Data      json.RawMessage `json:"data"`
}

// Tool represents a tool that can be called by the LLM
type Tool interface {
	// Name returns the tool name as exposed to the LLM
	Name() string

	// Description returns the tool description for the system prompt
	Description() string

	// JSONSchema returns the JSON schema for the tool's input parameters
	JSONSchema() map[string]interface{}

	// Execute runs the tool with the given input
	Execute(ctx *ToolContext, input json.RawMessage, progress ToolCallProgress) (*ToolResult[json.RawMessage], error)
}

// ToolContext provides context for tool execution
type ToolContext struct {
	Cwd           string
	AbortSignal   <-chan struct{}
	GetAppState   func() map[string]interface{}
	SetAppState   func(map[string]interface{})
	Messages      []Message
	Debug         bool
	Verbose       bool
	MainLoopModel string
}
