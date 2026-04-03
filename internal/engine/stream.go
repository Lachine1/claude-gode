package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// streamEventType represents the type of an SSE event
type streamEventType string

const (
	eventMessageStart      streamEventType = "message_start"
	eventContentBlockStart streamEventType = "content_block_start"
	eventContentBlockDelta streamEventType = "content_block_delta"
	eventContentBlockStop  streamEventType = "content_block_stop"
	eventMessageDelta      streamEventType = "message_delta"
	eventMessageStop       streamEventType = "message_stop"
)

// sseParser parses Anthropic's SSE format
type sseParser struct {
	onToken    func(token string)
	onToolCall func(toolCall ToolCall)
	onThinking func(text string)
	onUsage    func(usage types.Usage)
	onDone     func(stopReason string)
}

// newSSEParser creates a new SSE parser
func newSSEParser(
	onToken func(token string),
	onToolCall func(toolCall ToolCall),
	onThinking func(text string),
) *sseParser {
	return &sseParser{
		onToken:    onToken,
		onToolCall: onToolCall,
		onThinking: onThinking,
	}
}

// parse reads from the response body and processes SSE events
func (p *sseParser) parse(body io.ReadCloser) error {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var currentEvent string
	var dataBuffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if currentEvent != "" && dataBuffer.Len() > 0 {
				if err := p.handleEvent(currentEvent, dataBuffer.Bytes()); err != nil {
					return err
				}
				currentEvent = ""
				dataBuffer.Reset()
			}
			continue
		}

		if len(line) > 6 && line[:6] == "event:" {
			currentEvent = line[6:]
			if len(currentEvent) > 0 && currentEvent[0] == ' ' {
				currentEvent = currentEvent[1:]
			}
			continue
		}

		if len(line) > 5 && line[:5] == "data:" {
			data := line[5:]
			if len(data) > 0 && data[0] == ' ' {
				data = data[1:]
			}
			dataBuffer.WriteString(data)
			continue
		}
	}

	if currentEvent != "" && dataBuffer.Len() > 0 {
		if err := p.handleEvent(currentEvent, dataBuffer.Bytes()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	return nil
}

// handleEvent processes a single SSE event
func (p *sseParser) handleEvent(eventType string, data []byte) error {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	switch base.Type {
	case "message_start":
		return p.handleMessageStart(data)
	case "content_block_start":
		return p.handleContentBlockStart(data)
	case "content_block_delta":
		return p.handleContentBlockDelta(data)
	case "content_block_stop":
		return p.handleContentBlockStop(data)
	case "message_delta":
		return p.handleMessageDelta(data)
	case "message_stop":
		return p.handleMessageStop()
	case "error":
		return p.handleAPIError(data)
	case "ping":
		return nil
	}

	return nil
}

func (p *sseParser) handleMessageStart(data []byte) error {
	var event struct {
		Message struct {
			ID    string `json:"id"`
			Usage struct {
				InputTokens int `json:"input_tokens"`
				CacheRead   int `json:"cache_read_input_tokens"`
				CacheWrite  int `json:"cache_creation_input_tokens"`
			} `json:"usage"`
		} `json:"message"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse message_start: %w", err)
	}

	usage := types.Usage{
		InputTokens: event.Message.Usage.InputTokens,
		CacheRead:   event.Message.Usage.CacheRead,
		CacheWrite:  event.Message.Usage.CacheWrite,
	}
	if p.onUsage != nil {
		p.onUsage(usage)
	}
	return nil
}

func (p *sseParser) handleContentBlockStart(data []byte) error {
	var event struct {
		Index        int `json:"index"`
		ContentBlock struct {
			Type     string                 `json:"type"`
			Text     string                 `json:"text,omitempty"`
			ID       string                 `json:"id,omitempty"`
			Name     string                 `json:"name,omitempty"`
			Input    map[string]interface{} `json:"input,omitempty"`
			Thinking string                 `json:"thinking,omitempty"`
		} `json:"content_block"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse content_block_start: %w", err)
	}

	switch event.ContentBlock.Type {
	case "tool_use":
		inputJSON, err := json.Marshal(event.ContentBlock.Input)
		if err != nil {
			return fmt.Errorf("failed to marshal tool input: %w", err)
		}
		toolCall := ToolCall{
			ToolUseID: event.ContentBlock.ID,
			Name:      event.ContentBlock.Name,
			Input:     inputJSON,
		}
		if p.onToolCall != nil {
			p.onToolCall(toolCall)
		}
	case "thinking":
		if event.ContentBlock.Thinking != "" && p.onThinking != nil {
			p.onThinking(event.ContentBlock.Thinking)
		}
	}

	return nil
}

func (p *sseParser) handleContentBlockDelta(data []byte) error {
	var event struct {
		Index int `json:"index"`
		Delta struct {
			Type     string `json:"type"`
			Text     string `json:"text,omitempty"`
			Thinking string `json:"thinking,omitempty"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse content_block_delta: %w", err)
	}

	switch event.Delta.Type {
	case "text_delta":
		if event.Delta.Text != "" && p.onToken != nil {
			p.onToken(event.Delta.Text)
		}
	case "thinking_delta":
		if event.Delta.Thinking != "" && p.onThinking != nil {
			p.onThinking(event.Delta.Thinking)
		}
	}

	return nil
}

func (p *sseParser) handleContentBlockStop(data []byte) error {
	return nil
}

func (p *sseParser) handleMessageDelta(data []byte) error {
	var event struct {
		Delta struct {
			StopReason   string `json:"stop_reason"`
			StopSequence string `json:"stop_sequence"`
		} `json:"delta"`
		Usage struct {
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse message_delta: %w", err)
	}

	if event.Usage.OutputTokens > 0 && p.onUsage != nil {
		p.onUsage(types.Usage{OutputTokens: event.Usage.OutputTokens})
	}

	if event.Delta.StopReason != "" && p.onDone != nil {
		p.onDone(event.Delta.StopReason)
	}

	return nil
}

func (p *sseParser) handleMessageStop() error {
	return nil
}

func (p *sseParser) handleAPIError(data []byte) error {
	var apiErr struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &apiErr); err != nil {
		return fmt.Errorf("API error: %s", string(data))
	}
	return fmt.Errorf("API error (%s): %s", apiErr.Type, apiErr.Message)
}

// doAPIRequest makes the HTTP request to the Anthropic API and returns the response body
func doAPIRequest(
	cfg *types.APIConfig,
	messages []types.Message,
	systemPrompt string,
	tools []types.Tool,
) (io.ReadCloser, error) {
	body, err := buildRequestBody(cfg, messages, systemPrompt, tools)
	if err != nil {
		return nil, err
	}

	var lastErr error
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com/v1"
		}

		req, err := http.NewRequest("POST", baseURL+"/messages", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", cfg.APIKey)
		req.Header.Set("anthropic-version", "2024-02-15")
		req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

		client := &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		}
		if cfg.Timeout == 0 {
			client.Timeout = 5 * time.Minute
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
		}

		return resp.Body, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
	}
	return nil, fmt.Errorf("all %d retries exhausted", maxRetries)
}

// buildRequestBody constructs the JSON body for the API request
func buildRequestBody(
	cfg *types.APIConfig,
	messages []types.Message,
	systemPrompt string,
	tools []types.Tool,
) ([]byte, error) {
	type toolDef struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		InputSchema map[string]interface{} `json:"input_schema"`
	}

	var toolDefs []toolDef
	for _, t := range tools {
		toolDefs = append(toolDefs, toolDef{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.JSONSchema(),
		})
	}

	type systemBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8192
	}

	req := struct {
		Model     string          `json:"model"`
		MaxTokens int             `json:"max_tokens"`
		Messages  []types.Message `json:"messages"`
		System    []systemBlock   `json:"system"`
		Tools     []toolDef       `json:"tools,omitempty"`
		Stream    bool            `json:"stream"`
	}{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  messages,
		System:    []systemBlock{{Type: "text", Text: systemPrompt}},
		Tools:     toolDefs,
		Stream:    true,
	}

	return json.Marshal(req)
}
