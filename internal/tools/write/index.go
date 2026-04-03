package write

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// WriteInput represents the input parameters for the Write tool
type WriteInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteOutput represents the output from the Write tool
type WriteOutput struct {
	Path         string `json:"path"`
	BytesWritten int    `json:"bytes_written"`
}

// WriteTool implements the types.Tool interface for writing file contents
type WriteTool struct{}

// New creates a new WriteTool
func New() *WriteTool {
	return &WriteTool{}
}

// Name returns the tool name
func (t *WriteTool) Name() string {
	return "write"
}

// Description returns the tool description
func (t *WriteTool) Description() string {
	return "Write content to a file. Creates parent directories if they don't exist. Overwrites existing files."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *WriteTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

// Execute writes content to the file
func (t *WriteTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params WriteInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid input: %v", err),
		}, nil
	}

	if params.Path == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "Path cannot be empty",
		}, nil
	}

	targetPath := params.Path
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(ctx.Cwd, params.Path)
	}

	targetPath = filepath.Clean(targetPath)

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Failed to create directory: %v", err),
		}, nil
	}

	content := []byte(params.Content)
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Failed to write file: %v", err),
		}, nil
	}

	output := WriteOutput{
		Path:         params.Path,
		BytesWritten: len(content),
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}
