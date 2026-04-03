package glob

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	matches, err := globWalk(searchPath, params.Pattern)
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

// globWalk walks the directory tree and matches files against the pattern.
// It supports ** patterns which match zero or more directories.
func globWalk(root string, pattern string) ([]string, error) {
	pattern = filepath.ToSlash(pattern)
	segments := strings.Split(pattern, "/")

	var matches []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		if rel == "." {
			return nil
		}

		relSlash := filepath.ToSlash(rel)
		relSegments := strings.Split(relSlash, "/")

		if matchSegments(segments, relSegments) {
			matches = append(matches, path)
		}

		return nil
	})

	return matches, err
}

// matchSegments checks if the pattern segments match the path segments.
// It supports ** which matches zero or more path segments.
func matchSegments(patternSegs, pathSegs []string) bool {
	return matchRecursive(patternSegs, 0, pathSegs, 0)
}

// matchRecursive recursively matches pattern segments against path segments.
func matchRecursive(patternSegs []string, pi int, pathSegs []string, si int) bool {
	for pi < len(patternSegs) {
		pat := patternSegs[pi]

		if pat == "**" {
			pi++
			if pi == len(patternSegs) {
				return true
			}
			for i := si; i <= len(pathSegs); i++ {
				if matchRecursive(patternSegs, pi, pathSegs, i) {
					return true
				}
			}
			return false
		}

		if si >= len(pathSegs) {
			return false
		}

		matched, err := filepath.Match(pat, pathSegs[si])
		if err != nil || !matched {
			return false
		}

		pi++
		si++
	}

	return si == len(pathSegs)
}
