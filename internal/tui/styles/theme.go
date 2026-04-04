package styles

import (
	"charm.land/lipgloss/v2"
)

// Claude Code exact color palette (dark theme)
// Source: references/claude-code/src/utils/theme.ts
const (
	ColorClaude          = "#D77757" // rgb(215,119,87)
	ColorClaudeShimmer   = "#EB9F7F" // rgb(235,159,127)
	ColorPermission      = "#B1B9F9" // rgb(177,185,249)
	ColorPermissionShim  = "#CFD7FF" // rgb(207,215,255)
	ColorSuccess         = "#4EBA65" // rgb(78,186,101)
	ColorError           = "#FF6B80" // rgb(255,107,128)
	ColorWarning         = "#FFC107" // rgb(255,193,7)
	ColorText            = "#FFFFFF" // rgb(255,255,255)
	ColorSubtle          = "#505050" // rgb(80,80,80)
	ColorInactive        = "#999999" // rgb(153,153,153)
	ColorSuggestion      = "#B1B9F9" // rgb(177,185,249)
	ColorUserBg          = "#373737" // rgb(55,55,55)
	ColorUserBgHover     = "#464646" // rgb(70,70,70)
	ColorBashBorder      = "#FD5DB1" // rgb(253,93,177)
	ColorPromptBorder    = "#888888" // rgb(136,136,136)
	ColorBg              = "#0D0D0D" // rgb(13,13,13)
	ColorMsgActionsBg    = "#2C323E" // rgb(44,50,62)
	ColorDiffAdded       = "#4EBA65"
	ColorDiffRemoved     = "#FF6B80"
	ColorDiffAddedDim    = "#2A6B3A"
	ColorDiffRemovedDim  = "#6B2A3A"
	ColorThinkingShimmer = "#B1B9F9" // rgb(177,185,249)
)

type Theme struct {
	Base             lipgloss.Style
	UserMessage      lipgloss.Style
	AssistantMessage lipgloss.Style
	ToolCall         lipgloss.Style
	ToolCallRunning  lipgloss.Style
	ToolCallSuccess  lipgloss.Style
	ToolCallError    lipgloss.Style
	ToolResult       lipgloss.Style
	ThinkingBlock    lipgloss.Style
	Error            lipgloss.Style
	Success          lipgloss.Style
	Subtle           lipgloss.Style
	Inactive         lipgloss.Style
	Suggestion       lipgloss.Style
	CommandOutput    lipgloss.Style
	PromptBorder     lipgloss.Style
	PromptPrefix     lipgloss.Style
	PromptFooter     lipgloss.Style
	SpinnerGlyph     lipgloss.Style
	SpinnerText      lipgloss.Style
	SpinnerStatus    lipgloss.Style
	SpinnerTimer     lipgloss.Style
	PermissionBorder lipgloss.Style
	PermissionTitle  lipgloss.Style
	PermissionSub    lipgloss.Style
	PermOptionFocus  lipgloss.Style
	PermOptionBlur   lipgloss.Style
	PermOptionIdx    lipgloss.Style
	PermOptionCheck  lipgloss.Style
	PermCancel       lipgloss.Style
	Scrollable       lipgloss.Style
	NewMsgPill       lipgloss.Style
	StickyHeader     lipgloss.Style
	ModalTopBorder   lipgloss.Style
}

func DefaultTheme() Theme {
	t := Theme{}

	t.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorBg))

	// User messages: gray background, no border, full width
	// "You: <message>" text, compact
	t.UserMessage = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorUserBg)).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1)

	// Assistant messages: no background, plain white text
	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		MarginTop(1).
		MarginBottom(1)

	// Tool calls: left border only (│), blue
	t.ToolCall = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderLeft(true).
		BorderRight(false).
		BorderBottom(false).
		BorderForeground(lipgloss.Color(ColorPermission)).
		PaddingLeft(1).
		MarginTop(1).
		MarginBottom(1)

	t.ToolCallRunning = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning))
	t.ToolCallSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
	t.ToolCallError = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError))

	// Tool results: dimmed, collapsed preview
	t.ToolResult = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInactive)).
		PaddingLeft(2).
		MarginTop(1).
		MarginBottom(1)

	// Thinking blocks: subtle, dimmed, italic
	t.ThinkingBlock = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSubtle)).
		Italic(true).
		PaddingLeft(2).
		MarginTop(1).
		MarginBottom(1)

	t.Error = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Bold(true)
	t.Success = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Bold(true)
	t.Subtle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))
	t.Inactive = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInactive))
	t.Suggestion = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuggestion))

	// Command output: muted text
	t.CommandOutput = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInactive)).
		MarginTop(1).
		MarginBottom(1)

	// Prompt input: bottom + right border only, round style
	t.PromptBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(true).
		BorderBottom(true).
		BorderForeground(lipgloss.Color(ColorPromptBorder))

	t.PromptPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText))
	t.PromptFooter = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))

	// Spinner
	t.SpinnerGlyph = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorClaude))
	t.SpinnerText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorClaude))
	t.SpinnerStatus = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))
	t.SpinnerTimer = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))

	// Permission dialog: TOP border only, round
	t.PermissionBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(true).
		BorderLeft(false).
		BorderRight(false).
		BorderBottom(false).
		BorderForeground(lipgloss.Color(ColorPermission))

	t.PermissionTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPermission)).
		Bold(true)

	t.PermissionSub = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))

	// Permission options
	t.PermOptionFocus = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuggestion))
	t.PermOptionBlur = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText))
	t.PermOptionIdx = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))
	t.PermOptionCheck = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
	t.PermCancel = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtle))

	t.Scrollable = lipgloss.NewStyle().Background(lipgloss.Color(ColorBg))

	// New messages pill
	t.NewMsgPill = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorUserBg)).
		Foreground(lipgloss.Color(ColorSubtle)).
		Padding(0, 1)

	// Sticky header (when scrolled up)
	t.StickyHeader = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorUserBg)).
		Foreground(lipgloss.Color(ColorSubtle))

	// Modal top border
	t.ModalTopBorder = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPermission))

	return t
}
