package edit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// EditTool performs search and replace in files
type EditTool struct{}

// New creates a new EditTool
func New() *EditTool {
	return &EditTool{}
}

// Name returns the tool name
func (t *EditTool) Name() string {
	return "edit"
}

// Description returns the tool description
func (t *EditTool) Description() string {
	return "Edit a file by replacing occurrences of text. Use old_string to find text and new_string to replace it."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *EditTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to edit",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to find and replace",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace with",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

type editInput struct {
	Path      string `json:"path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// Execute performs the edit operation
func (t *EditTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params editInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("invalid input: %v", err),
		}, nil
	}

	if params.Path == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "path cannot be empty",
		}, nil
	}

	if params.OldString == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "old_string cannot be empty",
		}, nil
	}

	filePath := params.Path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(ctx.Cwd, filePath)
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("file not found: %s", params.Path),
			}, nil
		}
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	originalContent := string(content)

	// Check if old_string exists
	if !strings.Contains(originalContent, params.OldString) {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("old_string not found in file: %s", params.Path),
		}, nil
	}

	// Count occurrences
	count := strings.Count(originalContent, params.OldString)

	// Perform replacement
	newContent := strings.ReplaceAll(originalContent, params.OldString, params.NewString)

	// Write the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	// Generate diff
	diff := generateDiff(originalContent, newContent, params.Path)

	result := map[string]interface{}{
		"path":         params.Path,
		"replacements": count,
		"diff":         diff,
		"message":      fmt.Sprintf("Replaced %d occurrence(s) in %s", count, params.Path),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}

func generateDiff(old, new, path string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	oldIdx := 0
	newIdx := 0

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if oldIdx < len(oldLines) && newIdx < len(newLines) {
			if oldLines[oldIdx] == newLines[newIdx] {
				oldIdx++
				newIdx++
				continue
			}
		}

		if oldIdx < len(oldLines) {
			diff.WriteString(fmt.Sprintf("-%s\n", oldLines[oldIdx]))
			oldIdx++
		}
		if newIdx < len(newLines) {
			diff.WriteString(fmt.Sprintf("+%s\n", newLines[newIdx]))
			newIdx++
		}
	}

	return diff.String()
}
