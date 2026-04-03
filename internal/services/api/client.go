package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

const (
	defaultBaseURL  = "https://api.anthropic.com/v1"
	defaultTimeout  = 5 * time.Minute
	defaultMaxRetry = 3
	apiVersion      = "2024-02-15"
)

// ClientOption configures the API client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithTimeout sets the request timeout
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithDebug enables request/response logging
func WithDebug() ClientOption {
	return func(c *Client) {
		c.debug = true
	}
}

// Client is the Anthropic Messages API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	maxRetries int
	timeout    time.Duration
	debug      bool
}

// NewClient creates a new API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		maxRetries: defaultMaxRetry,
		timeout:    defaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// MessageRequest represents a request to the Messages API
type MessageRequest struct {
	Model         string                 `json:"model"`
	MaxTokens     int                    `json:"max_tokens"`
	Messages      []types.Message        `json:"messages"`
	System        []SystemBlock          `json:"system,omitempty"`
	Tools         []ToolDef              `json:"tools,omitempty"`
	ToolChoice    map[string]interface{} `json:"tool_choice,omitempty"`
	Temperature   *float64               `json:"temperature,omitempty"`
	TopP          *float64               `json:"top_p,omitempty"`
	TopK          *int                   `json:"top_k,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Stream        bool                   `json:"stream"`
	Metadata      *RequestMetadata       `json:"metadata,omitempty"`
}

// SystemBlock represents a system prompt block
type SystemBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolDef represents a tool definition for the API
type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// RequestMetadata contains optional request metadata
type RequestMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// MessageResponse represents a response from the Messages API
type MessageResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []types.ContentBlock `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason,omitempty"`
	StopSequence string               `json:"stop_sequence,omitempty"`
	Usage        types.Usage          `json:"usage"`
}

// APIError represents an error response from the API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%s): %s", e.Type, e.Message)
}

// isRetryable returns true if the error is retryable
func (e *APIError) isRetryable() bool {
	return e.Code == http.StatusTooManyRequests ||
		e.Code == http.StatusServiceUnavailable ||
		e.Code == http.StatusGatewayTimeout ||
		e.Code == http.StatusInternalServerError
}

// CreateMessage sends a non-streaming message request
func (c *Client) CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	req.Stream = false
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp MessageResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

// CreateMessageStream sends a streaming message request
func (c *Client) CreateMessageStream(
	ctx context.Context,
	req *MessageRequest,
	onToken func(string),
	onToolCall func(ToolCall),
) (*MessageResponse, error) {
	req.Stream = true

	body, err := c.doStreamRequest(ctx, req, onToken, onToolCall)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	resp, err := c.parseStream(body, onToken, onToolCall)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountTokens estimates token count for messages
func (c *Client) CountTokens(ctx context.Context, messages []types.Message) (int, error) {
	total := 0
	for _, msg := range messages {
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentTypeText:
				total += estimateTokens(block.Text)
			case types.ContentTypeToolUse:
				if block.ToolUse != nil {
					total += estimateTokens(block.ToolUse.Name)
					data, _ := json.Marshal(block.ToolUse.Input)
					total += estimateTokens(string(data))
				}
			case types.ContentTypeToolResult:
				if block.ToolResult != nil {
					total += estimateTokens(block.ToolResult.ToolUseID)
					for _, inner := range block.ToolResult.Content {
						if inner.Type == types.ContentTypeText {
							total += estimateTokens(inner.Text)
						}
					}
				}
			}
		}
		total += 4
	}
	return total, nil
}

// doRequest makes an HTTP request with retries
func (c *Client) doRequest(ctx context.Context, req *MessageRequest) (io.ReadCloser, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if c.debug {
				log.Printf("[api] retry %d/%d after %v", attempt, c.maxRetries, backoff)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.setHeaders(httpReq)

		if c.debug {
			log.Printf("[api] POST %s/messages (attempt %d)", c.baseURL, attempt+1)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return resp.Body, nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		apiErr := &APIError{Code: resp.StatusCode}
		if err := json.Unmarshal(respBody, apiErr); err != nil {
			apiErr.Message = string(respBody)
			apiErr.Type = "unknown"
		}

		if !apiErr.isRetryable() {
			return nil, apiErr
		}

		lastErr = apiErr
	}

	return nil, fmt.Errorf("all %d retries exhausted: %w", c.maxRetries, lastErr)
}

// doStreamRequest makes a streaming HTTP request
func (c *Client) doStreamRequest(ctx context.Context, req *MessageRequest, onToken func(string), onToolCall func(ToolCall)) (io.ReadCloser, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		apiErr := &APIError{Code: resp.StatusCode}
		if err := json.Unmarshal(respBody, apiErr); err != nil {
			apiErr.Message = string(respBody)
			apiErr.Type = "unknown"
		}
		return nil, apiErr
	}

	return resp.Body, nil
}

// setHeaders sets the required HTTP headers
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")
}

// ToolCall represents a tool call from the API
type ToolCall struct {
	ToolUseID string
	Name      string
	Input     json.RawMessage
}

// parseStream reads the SSE stream and returns the final response
func (c *Client) parseStream(body io.ReadCloser, onToken func(string), onToolCall func(ToolCall)) (*MessageResponse, error) {
	defer body.Close()

	parser := newSSEParser(onToken, onToolCall)
	resp, err := parser.parse(body)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// estimateTokens provides a rough token count from text
func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text) / 4
}
