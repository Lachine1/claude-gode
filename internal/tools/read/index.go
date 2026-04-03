package read

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

// ReadInput represents the input parameters for the Read tool
type ReadInput struct {
	Path   string `json:"path"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}

// ReadOutput represents the output from the Read tool
type ReadOutput struct {
	Content   string `json:"content"`
	Path      string `json:"path"`
	Lines     int    `json:"lines"`
	Truncated bool   `json:"truncated,omitempty"`
}

// ReadTool implements the types.Tool interface for reading file contents
type ReadTool struct{}

// New creates a new ReadTool
func New() *ReadTool {
	return &ReadTool{}
}

// Name returns the tool name
func (t *ReadTool) Name() string {
	return "read"
}

// Description returns the tool description
func (t *ReadTool) Description() string {
	return "Read the contents of a file. Returns content with line numbers. Supports reading specific line ranges with offset and limit."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *ReadTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "The line number to start reading from (1-indexed)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "The maximum number of lines to read",
			},
		},
		"required": []string{"path"},
	}
}

// isBinary checks if the file appears to be binary
func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// Execute reads the file content
func (t *ReadTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params ReadInput
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
		targetPath = filepath.Join(ctx.Cwd, targetPath)
	}

	targetPath = filepath.Clean(targetPath)

	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("File not found: %s", params.Path),
			}, nil
		}
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Cannot access file: %v", err),
		}, nil
	}

	if info.IsDir() {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Path is a directory: %s", params.Path),
		}, nil
	}

	if info.Size() > maxFileSize {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("File too large (%d bytes, max %d bytes)", info.Size(), maxFileSize),
		}, nil
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	if isBinary(data) {
		output := ReadOutput{
			Content: fmt.Sprintf("[Binary file, %d bytes]", len(data)),
			Path:    params.Path,
			Lines:   0,
		}
		resultJSON, _ := json.Marshal(output)
		return &types.ToolResult[json.RawMessage]{
			Data: resultJSON,
		}, nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	startLine := 0
	if params.Offset != nil && *params.Offset > 0 {
		startLine = *params.Offset - 1
		if startLine >= len(lines) {
			startLine = len(lines) - 1
		}
	}

	endLine := len(lines)
	if params.Limit != nil && *params.Limit > 0 {
		end := startLine + *params.Limit
		if end < endLine {
			endLine = end
		}
	}

	var sb strings.Builder
	truncated := false
	for i := startLine; i < endLine; i++ {
		sb.WriteString(fmt.Sprintf("%d: %s\n", i+1, lines[i]))
	}

	if endLine < len(lines) {
		truncated = true
		sb.WriteString(fmt.Sprintf("... (%d more lines)\n", len(lines)-endLine))
	}

	output := ReadOutput{
		Content:   sb.String(),
		Path:      params.Path,
		Lines:     endLine - startLine,
		Truncated: truncated,
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}

// ReadFile is a helper that reads a file and returns content with line numbers
func ReadFile(path string, offset, limit int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}

	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file too large (%d bytes, max %d bytes)", info.Size(), maxFileSize)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(file)
	lineNum := 0
	count := 0

	for scanner.Scan() {
		lineNum++
		if offset > 0 && lineNum < offset {
			continue
		}
		if limit > 0 && count >= limit {
			break
		}
		sb.WriteString(fmt.Sprintf("%d: %s\n", lineNum, scanner.Text()))
		count++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return sb.String(), nil
}
