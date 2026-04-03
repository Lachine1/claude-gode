package styles

import (
	"charm.land/lipgloss/v2"
)

// Claude Code exact colors (converted from RGB)
const (
	// Background colors
	ColorBackground = "#0d0d0d"
	ColorSurface    = "#1a1a1a"
	ColorBorder     = "#333333"

	// Text colors
	ColorText      = "#e5e5e5"
	ColorTextMuted = "#737373"

	// Accent colors
	ColorPrimary   = "#d4a373" // Orange/amber for assistant
	ColorSecondary = "#4a9eff" // Blue for tool calls
	ColorSuccess   = "#98c379"
	ColorWarning   = "#d4a373"
	ColorError     = "#e06c75"
	ColorInfo      = "#56b6c2"

	// Message colors
	ColorUserBg      = "#262626" // Gray bg for user
	ColorAssistantFg = "#d4a373" // Orange text for assistant
	ColorToolBorder  = "#4a9eff" // Blue border for tool calls
	ColorToolBg      = "#1a1a1a"

	// Permission dialog
	ColorPermBg     = "#1a1a1a"
	ColorPermTitle  = "#d4a373"
	ColorPermBorder = "#4a9eff"

	// Status bar
	ColorStatusBg = "#0d0d0d"
	ColorStatusFg = "#737373"

	// Input
	ColorInputBorder = "#333333"
	ColorInputBg     = "#0d0d0d"
)

type Theme struct {
	Base               lipgloss.Style
	UserMessage        lipgloss.Style
	UserPrefix         lipgloss.Style
	AssistantMessage   lipgloss.Style
	AssistantPrefix    lipgloss.Style
	CommandOutput      lipgloss.Style
	ToolCall           lipgloss.Style
	ToolCallRunning    lipgloss.Style
	ToolCallSuccess    lipgloss.Style
	ToolCallError      lipgloss.Style
	ToolResult         lipgloss.Style
	ThinkingBlock      lipgloss.Style
	Error              lipgloss.Style
	Success            lipgloss.Style
	StatusBar          lipgloss.Style
	StatusModel        lipgloss.Style
	StatusTokens       lipgloss.Style
	StatusMode         lipgloss.Style
	InputBorder        lipgloss.Style
	InputPrompt        lipgloss.Style
	InputText          lipgloss.Style
	PermissionDialog   lipgloss.Style
	PermissionTitle    lipgloss.Style
	PermissionOption   lipgloss.Style
	PermissionSelected lipgloss.Style
	Spinner            lipgloss.Style
}

func DefaultTheme() Theme {
	t := Theme{}

	t.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorBackground))

	// User message: gray background
	t.UserMessage = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorUserBg)).
		Foreground(lipgloss.Color(ColorText)).
		Padding(0, 1)

	t.UserPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Bold(true)

	// Assistant message: orange text
	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAssistantFg)).
		Padding(0, 1)

	t.AssistantPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAssistantFg)).
		Bold(true)

	// Command output: muted
	t.CommandOutput = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Padding(0, 1)

	// Tool call: blue border
	t.ToolCall = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorToolBg)).
		Foreground(lipgloss.Color(ColorText)).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorToolBorder))

	t.ToolCallRunning = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSecondary))

	t.ToolCallSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))

	t.ToolCallError = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError))

	t.ToolResult = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorToolBg)).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.ThinkingBlock = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorSurface)).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError))

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))

	// Status bar: dark with muted text
	t.StatusBar = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBg)).
		Foreground(lipgloss.Color(ColorStatusFg)).
		Padding(0, 1)

	t.StatusModel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)

	t.StatusTokens = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInfo))

	t.StatusMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))

	// Input: simple with gray border
	t.InputBorder = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorInputBg)).
		Foreground(lipgloss.Color(ColorText)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorInputBorder)).
		Padding(0, 1)

	t.InputPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted))

	t.InputText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))

	// Permission dialog: blue border, orange title
	t.PermissionDialog = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPermBg)).
		Foreground(lipgloss.Color(ColorText)).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorPermBorder))

	t.PermissionTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPermTitle)).
		Bold(true)

	t.PermissionOption = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))

	t.PermissionSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)

	t.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSecondary))

	return t
}
