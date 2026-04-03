package types

// Message represents a message in the conversation
type Message struct {
	ID        string         `json:"id"`
	Role      MessageRole    `json:"role"`
	Content   []ContentBlock `json:"content"`
	Timestamp int64          `json:"timestamp,omitempty"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

// ContentBlock represents a block of content in a message
type ContentBlock struct {
	Type       ContentType        `json:"type"`
	Text       string             `json:"text,omitempty"`
	ToolUse    *ToolUseContent    `json:"tool_use,omitempty"`
	ToolResult *ToolResultContent `json:"tool_result,omitempty"`
	Thinking   *ThinkingContent   `json:"thinking,omitempty"`
	Source     *SourceContent     `json:"source,omitempty"`
}

// ContentType represents the type of content block
type ContentType string

const (
	ContentTypeText       ContentType = "text"
	ContentTypeToolUse    ContentType = "tool_use"
	ContentTypeToolResult ContentType = "tool_result"
	ContentTypeThinking   ContentType = "thinking"
	ContentTypeImage      ContentType = "image"
)

// ToolUseContent represents a tool use content block
type ToolUseContent struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResultContent represents a tool result content block
type ToolResultContent struct {
	ToolUseID string         `json:"tool_use_id"`
	Content   []ContentBlock `json:"content"`
	IsError   bool           `json:"is_error,omitempty"`
}

// ThinkingContent represents a thinking/reasoning block
type ThinkingContent struct {
	Text      string `json:"text"`
	Signature string `json:"signature,omitempty"`
}

// SourceContent represents an image or file source
type SourceContent struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}
