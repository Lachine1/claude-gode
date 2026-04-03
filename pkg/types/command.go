package types

// Command represents a slash command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Handler     CommandHandler
}

// CommandHandler is the function signature for command execution
type CommandHandler func(ctx *CommandContext, args []string) error

// CommandContext provides context for command execution
type CommandContext struct {
	Cwd         string
	GetMessages func() []Message
	SetMessages func([]Message)
	WriteOutput func(string)
	Debug       bool
}

// CommandError represents an error from command execution
type CommandError struct {
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}
