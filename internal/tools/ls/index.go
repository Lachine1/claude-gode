package ls

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

const maxEntries = 100

// LSInput represents the input parameters for the LS tool
type LSInput struct {
	Path string `json:"path,omitempty"`
}

// LSFile represents a file or directory entry
type LSFile struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
}

// LSOutput represents the output from the LS tool
type LSOutput struct {
	Path    string   `json:"path"`
	Entries []LSFile `json:"entries"`
	Count   int      `json:"count"`
}

// LSTool implements the types.Tool interface for listing directory contents
type LSTool struct{}

// New creates a new LSTool
func New() *LSTool {
	return &LSTool{}
}

// Name returns the tool name
func (t *LSTool) Name() string {
	return "ls"
}

// Description returns the tool description
func (t *LSTool) Description() string {
	return "List the contents of a directory. Shows files and directories with type indicators. Limited to 100 entries."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *LSTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory path to list (defaults to current working directory)",
			},
		},
	}
}

// Execute lists the directory contents
func (t *LSTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params LSInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid input: %v", err),
		}, nil
	}

	listPath := ctx.Cwd
	if params.Path != "" {
		if !filepath.IsAbs(params.Path) {
			listPath = filepath.Join(ctx.Cwd, params.Path)
		} else {
			listPath = params.Path
		}
	}

	listPath = filepath.Clean(listPath)

	info, err := os.Stat(listPath)
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

	entries, err := os.ReadDir(listPath)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Failed to read directory: %v", err),
		}, nil
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir() && !entries[j].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	limit := len(entries)
	if limit > maxEntries {
		limit = maxEntries
	}

	resultEntries := make([]LSFile, 0, limit)
	for i := 0; i < limit; i++ {
		entry := entries[i]
		info, err := entry.Info()
		if err != nil {
			continue
		}

		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		} else if entry.Type()&os.ModeSymlink != 0 {
			entryType = "symlink"
		} else if info.Mode()&os.ModeSocket != 0 {
			entryType = "socket"
		}

		resultEntries = append(resultEntries, LSFile{
			Name: entry.Name(),
			Type: entryType,
			Size: info.Size(),
		})
	}

	output := LSOutput{
		Path:    params.Path,
		Entries: resultEntries,
		Count:   len(resultEntries),
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}

// FormatEntry formats a single entry for display
func FormatEntry(entry LSFile) string {
	typeIndicator := " "
	switch entry.Type {
	case "directory":
		typeIndicator = "d"
	case "symlink":
		typeIndicator = "l"
	case "socket":
		typeIndicator = "s"
	}
	return fmt.Sprintf("[%s] %s", typeIndicator, entry.Name)
}

// FormatEntries formats all entries for display
func FormatEntries(entries []LSFile) string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, FormatEntry(e))
	}
	return strings.Join(lines, "\n")
}
