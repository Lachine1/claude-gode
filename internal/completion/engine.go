package completion

import (
	"strings"
)

// CompletionEngine holds command metadata and usage stats
type CompletionEngine struct {
	commands   []CommandInfo
	recentCmds map[string]int
}

// NewEngine creates a new completion engine
func NewEngine(commands []CommandInfo) *CompletionEngine {
	return &CompletionEngine{
		commands:   commands,
		recentCmds: make(map[string]int),
	}
}

// GetSuggestions returns suggestions based on current input context
func (e *CompletionEngine) GetSuggestions(input string, cwd string) []SuggestionItem {
	if input == "" {
		return nil
	}

	// Command suggestions: input starts with "/"
	if strings.HasPrefix(input, "/") {
		query := strings.TrimPrefix(input, "/")
		if spaceIdx := strings.Index(query, " "); spaceIdx != -1 {
			query = query[:spaceIdx]
		}
		return e.SuggestCommands(query)
	}

	// File/path suggestions: look for @tokens
	if idx := strings.LastIndex(input, "@"); idx != -1 {
		if idx > 0 && input[idx-1] != ' ' && input[idx-1] != '\t' {
			return nil
		}
		token := input[idx:]
		if len(token) <= 1 {
			return SuggestPaths("", cwd)
		}
		return SuggestPaths(token, cwd)
	}

	return nil
}

// GetGhostText returns inline ghost text for the current input
func (e *CompletionEngine) GetGhostText(input string, cwd string) *GhostText {
	if input == "" {
		return nil
	}

	// Shell history ghost text for bash mode
	if strings.HasPrefix(input, "!") {
		return GetShellHistoryCompletion(input)
	}

	// Ghost text for commands
	if strings.HasPrefix(input, "/") {
		query := strings.TrimPrefix(input, "/")
		if spaceIdx := strings.Index(query, " "); spaceIdx == -1 {
			suggestions := e.SuggestCommands(query)
			if len(suggestions) > 0 {
				suffix := suggestions[0].DisplayText[len(input):]
				if suffix != "" {
					return &GhostText{
						Text:           suffix,
						FullCommand:    suggestions[0].DisplayText,
						InsertPosition: len(input),
					}
				}
			}
		}
	}

	return nil
}

// AcceptSuggestion applies a suggestion to the input
func (e *CompletionEngine) AcceptSuggestion(input string, suggestion SuggestionItem) string {
	if strings.HasPrefix(input, "/") {
		query := strings.TrimPrefix(input, "/")
		if spaceIdx := strings.Index(query, " "); spaceIdx != -1 {
			return "/" + suggestion.ID + query[spaceIdx:]
		}
		return "/" + suggestion.ID
	}

	if idx := strings.LastIndex(input, "@"); idx != -1 {
		token := extractAtToken(input, idx)
		if token != "" {
			prefix := input[:idx]
			suffix := input[idx+len(token):]
			return prefix + suggestion.ID + suffix
		}
	}

	return input
}

func extractAtToken(input string, atIdx int) string {
	if atIdx > 0 {
		prev := input[atIdx-1]
		if prev != ' ' && prev != '\t' {
			return ""
		}
	}

	rest := input[atIdx:]
	if len(rest) <= 1 {
		return "@"
	}

	end := len(rest)
	for i := 1; i < len(rest); i++ {
		if rest[i] == ' ' || rest[i] == '\t' {
			end = i
			break
		}
	}
	return rest[:end]
}
