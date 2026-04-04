package completion

// SuggestionType categorizes the kind of suggestion
type SuggestionType string

const (
	SuggestionCommand   SuggestionType = "command"
	SuggestionFile      SuggestionType = "file"
	SuggestionDirectory SuggestionType = "directory"
	SuggestionAgent     SuggestionType = "agent"
	SuggestionShell     SuggestionType = "shell"
	SuggestionNone      SuggestionType = "none"
)

// SuggestionItem represents a single completion suggestion
type SuggestionItem struct {
	ID          string
	DisplayText string
	Tag         string
	Description string
	Type        SuggestionType
	Color       string
}

// GhostText represents inline ghost text (gray suffix)
type GhostText struct {
	Text           string
	FullCommand    string
	InsertPosition int
}

// CommandInfo is a lightweight command descriptor for completion
type CommandInfo struct {
	Name        string
	Aliases     []string
	Description string
	Category    string
	HasArgs     bool
	ArgNames    []string
}
