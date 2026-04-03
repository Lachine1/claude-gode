package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// sseParser parses Anthropic's SSE format
type sseParser struct {
	onToken    func(string)
	onToolCall func(ToolCall)
	messageID  string
	model      string
	stopReason string
	stopSeq    string
	usage      types.Usage
	content    []types.ContentBlock
}

func newSSEParser(onToken func(string), onToolCall func(ToolCall)) *sseParser {
	return &sseParser{
		onToken:    onToken,
		onToolCall: onToolCall,
		content:    make([]types.ContentBlock, 0),
	}
}

func (p *sseParser) parse(body io.ReadCloser) (*MessageResponse, error) {
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
					return nil, err
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
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("stream read error: %w", err)
	}

	return &MessageResponse{
		ID:           p.messageID,
		Type:         "message",
		Role:         "assistant",
		Content:      p.content,
		Model:        p.model,
		StopReason:   p.stopReason,
		StopSequence: p.stopSeq,
		Usage:        p.usage,
	}, nil
}

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
		return nil
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
			Model string `json:"model"`
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

	p.messageID = event.Message.ID
	p.model = event.Message.Model
	p.usage.InputTokens = event.Message.Usage.InputTokens
	p.usage.CacheRead = event.Message.Usage.CacheRead
	p.usage.CacheWrite = event.Message.Usage.CacheWrite
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
		p.content = append(p.content, types.ContentBlock{
			Type: types.ContentTypeToolUse,
			ToolUse: &types.ToolUseContent{
				ID:    event.ContentBlock.ID,
				Name:  event.ContentBlock.Name,
				Input: event.ContentBlock.Input,
			},
		})
	case "thinking":
		p.content = append(p.content, types.ContentBlock{
			Type:     types.ContentTypeThinking,
			Thinking: &types.ThinkingContent{Text: event.ContentBlock.Thinking},
		})
	case "text":
		p.content = append(p.content, types.ContentBlock{
			Type: types.ContentTypeText,
			Text: event.ContentBlock.Text,
		})
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
		if len(p.content) > 0 {
			last := &p.content[len(p.content)-1]
			if last.Type == types.ContentTypeText {
				last.Text += event.Delta.Text
			}
		}
	case "thinking_delta":
		if event.Delta.Thinking != "" && len(p.content) > 0 {
			last := &p.content[len(p.content)-1]
			if last.Type == types.ContentTypeThinking && last.Thinking != nil {
				last.Thinking.Text += event.Delta.Thinking
			}
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

	p.usage.OutputTokens = event.Usage.OutputTokens
	p.stopReason = event.Delta.StopReason
	p.stopSeq = event.Delta.StopSequence
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
	return &APIError{
		Type:    apiErr.Type,
		Message: apiErr.Message,
	}
}
