package glob

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// GlobInput represents the input parameters for the Glob tool
type GlobInput struct {
	Path    string `json:"path"`
	Pattern string `json:"pattern"`
}

// GlobOutput represents the output from the Glob tool
type GlobOutput struct {
	Matches []string `json:"matches"`
	Count   int      `json:"count"`
}

// GlobTool implements the types.Tool interface for finding files by glob pattern
type GlobTool struct{}

// New creates a new GlobTool
func New() *GlobTool {
	return &GlobTool{}
}

// Name returns the tool name
func (t *GlobTool) Name() string {
	return "glob"
}

// Description returns the tool description
func (t *GlobTool) Description() string {
	return "Find files matching a glob pattern. Supports standard glob patterns like *, **, ?, [abc]."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *GlobTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The glob pattern to match files against",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory to search in (defaults to current working directory)",
			},
		},
		"required": []string{"pattern"},
	}
}

// Execute finds files matching the glob pattern
func (t *GlobTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params GlobInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid input: %v", err),
		}, nil
	}

	if params.Pattern == "" {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: "Pattern cannot be empty",
		}, nil
	}

	searchPath := ctx.Cwd
	if params.Path != "" {
		if !filepath.IsAbs(params.Path) {
			searchPath = filepath.Join(ctx.Cwd, params.Path)
		} else {
			searchPath = params.Path
		}
	}

	searchPath = filepath.Clean(searchPath)

	info, err := os.Stat(searchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("Path not found: %s", params.Path),
			}, nil
		}
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Cannot access path: %v", err),
		}, nil
	}

	if !info.IsDir() {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Path is not a directory: %s", params.Path),
		}, nil
	}

	globPattern := filepath.Join(searchPath, params.Pattern)

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid glob pattern: %v", err),
		}, nil
	}

	relPaths := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, err := filepath.Rel(ctx.Cwd, match)
		if err != nil {
			rel = match
		}
		relPaths = append(relPaths, rel)
	}

	output := GlobOutput{
		Matches: relPaths,
		Count:   len(relPaths),
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}
