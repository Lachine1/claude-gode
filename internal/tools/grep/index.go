package grep

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// GrepInput represents the input parameters for the Grep tool
type GrepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Include string `json:"include,omitempty"`
}

// GrepMatch represents a single match result
type GrepMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// GrepOutput represents the output from the Grep tool
type GrepOutput struct {
	Matches []GrepMatch `json:"matches"`
	Count   int         `json:"count"`
}

// GrepTool implements the types.Tool interface for searching file contents
type GrepTool struct{}

// New creates a new GrepTool
func New() *GrepTool {
	return &GrepTool{}
}

// Name returns the tool name
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns the tool description
func (t *GrepTool) Description() string {
	return "Search file contents using regular expressions. Returns matching lines with file:line:content format."
}

// JSONSchema returns the JSON schema for the tool's input
func (t *GrepTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file or directory to search in (defaults to current working directory)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to filter files (e.g., '*.go', '*.py')",
			},
		},
		"required": []string{"pattern"},
	}
}

// searchFile searches a single file for the pattern
func searchFile(filePath string, re *regexp.Regexp) ([]GrepMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, GrepMatch{
				File:    filePath,
				Line:    lineNum,
				Content: line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return matches, nil
}

// searchDir recursively searches a directory for the pattern
func searchDir(dirPath string, re *regexp.Regexp, includePattern string) ([]GrepMatch, error) {
	var allMatches []GrepMatch

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".venv" {
				return filepath.SkipDir
			}
			return nil
		}

		if includePattern != "" {
			matched, err := filepath.Match(includePattern, info.Name())
			if err != nil || !matched {
				return nil
			}
		}

		matches, err := searchFile(path, re)
		if err != nil {
			return nil
		}

		allMatches = append(allMatches, matches...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allMatches, nil
}

// Execute searches for the pattern in files
func (t *GrepTool) Execute(ctx *types.ToolContext, input json.RawMessage, progress types.ToolCallProgress) (*types.ToolResult[json.RawMessage], error) {
	var params GrepInput
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

	re, err := regexp.Compile(params.Pattern)
	if err != nil {
		return &types.ToolResult[json.RawMessage]{
			IsError:      true,
			ErrorMessage: fmt.Sprintf("Invalid regex pattern: %v", err),
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

	var matches []GrepMatch

	if info.IsDir() {
		matches, err = searchDir(searchPath, re, params.Include)
		if err != nil {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("Error searching directory: %v", err),
			}, nil
		}
	} else {
		matches, err = searchFile(searchPath, re)
		if err != nil {
			return &types.ToolResult[json.RawMessage]{
				IsError:      true,
				ErrorMessage: fmt.Sprintf("Error searching file: %v", err),
			}, nil
		}
	}

	relMatches := make([]GrepMatch, 0, len(matches))
	for _, m := range matches {
		relFile := m.File
		if rel, err := filepath.Rel(ctx.Cwd, m.File); err == nil {
			relFile = rel
		}
		relMatches = append(relMatches, GrepMatch{
			File:    relFile,
			Line:    m.Line,
			Content: m.Content,
		})
	}

	output := GrepOutput{
		Matches: relMatches,
		Count:   len(relMatches),
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return &types.ToolResult[json.RawMessage]{
		Data: resultJSON,
	}, nil
}

// FormatMatch formats a match in file:line:content format
func FormatMatch(m GrepMatch) string {
	return fmt.Sprintf("%s:%d:%s", m.File, m.Line, m.Content)
}

// FormatMatches formats all matches in file:line:content format
func FormatMatches(matches []GrepMatch) string {
	lines := make([]string, 0, len(matches))
	for _, m := range matches {
		lines = append(lines, FormatMatch(m))
	}
	return strings.Join(lines, "\n")
}
