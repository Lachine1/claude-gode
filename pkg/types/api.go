package types

// APIConfig holds configuration for the Anthropic API
type APIConfig struct {
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	Timeout     int
	MaxRetries  int
}

// Usage tracks token usage for a session
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheRead    int `json:"cache_read_tokens,omitempty"`
	CacheWrite   int `json:"cache_write_tokens,omitempty"`
}

// APIResponse represents a streaming response from the API
type APIResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// APIError represents an error from the API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
