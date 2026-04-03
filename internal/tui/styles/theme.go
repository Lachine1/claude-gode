package styles

import (
	"charm.land/lipgloss/v2"
)

// Colors used throughout the TUI
const (
	ColorBackground       = "#1a1b26"
	ColorSurface          = "#1f2335"
	ColorBorder           = "#292e42"
	ColorText             = "#c0caf5"
	ColorTextMuted        = "#565f89"
	ColorPrimary          = "#7aa2f7"
	ColorSecondary        = "#bb9af7"
	ColorSuccess          = "#9ece6a"
	ColorWarning          = "#e0af68"
	ColorError            = "#f7768e"
	ColorInfo             = "#7dcfff"
	ColorUserBg           = "#1e2a3a"
	ColorAssistantBg      = "#1a1b26"
	ColorToolBg           = "#24283b"
	ColorThinkingBg       = "#1e2030"
	ColorPermissionBg     = "#2d1f3d"
	ColorPermissionBorder = "#bb9af7"
	ColorStatusBarBg      = "#16161e"
	ColorStatusBarFg      = "#565f89"
	ColorInputBg          = "#1f2335"
	ColorWelcomeBg        = "#1a1b26"
)

// Theme holds all lipgloss styles for the application
type Theme struct {
	Base               lipgloss.Style
	UserMessage        lipgloss.Style
	AssistantMessage   lipgloss.Style
	ToolCall           lipgloss.Style
	ToolCallRunning    lipgloss.Style
	ToolCallSuccess    lipgloss.Style
	ToolCallError      lipgloss.Style
	ToolResult         lipgloss.Style
	ToolResultError    lipgloss.Style
	ThinkingBlock      lipgloss.Style
	Error              lipgloss.Style
	Success            lipgloss.Style
	StatusBar          lipgloss.Style
	StatusBarModel     lipgloss.Style
	StatusBarTokens    lipgloss.Style
	StatusBarCost      lipgloss.Style
	StatusBarPerm      lipgloss.Style
	StatusBarGit       lipgloss.Style
	Input              lipgloss.Style
	InputPrompt        lipgloss.Style
	InputCursor        lipgloss.Style
	Welcome            lipgloss.Style
	WelcomeTitle       lipgloss.Style
	WelcomeTip         lipgloss.Style
	WelcomeKey         lipgloss.Style
	PermissionDialog   lipgloss.Style
	PermissionTitle    lipgloss.Style
	PermissionOption   lipgloss.Style
	PermissionSelected lipgloss.Style
	Border             lipgloss.Style
	Scrollable         lipgloss.Style
}

// DefaultTheme returns the default theme for the application
func DefaultTheme() Theme {
	t := Theme{}

	t.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorBackground))

	t.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorUserBg)).
		Padding(1, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary))

	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorAssistantBg)).
		Padding(1, 2).
		MarginBottom(1)

	t.ToolCall = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorToolBg)).
		Padding(0, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.ToolCallRunning = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorWarning))

	t.ToolCallSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))

	t.ToolCallError = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError))

	t.ToolResult = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorToolBg)).
		Padding(0, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.ToolResultError = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError)).
		Background(lipgloss.Color(ColorToolBg)).
		Padding(0, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorError))

	t.ThinkingBlock = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		Background(lipgloss.Color(ColorThinkingBg)).
		Padding(1, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorSecondary))

	t.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError)).
		Bold(true)

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)).
		Bold(true)

	t.StatusBar = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorStatusBarFg)).
		Padding(0, 1)

	t.StatusBarModel = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		Padding(0, 1)

	t.StatusBarTokens = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorInfo)).
		Padding(0, 1)

	t.StatusBarCost = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorWarning)).
		Padding(0, 1)

	t.StatusBarPerm = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorSuccess)).
		Padding(0, 1)

	t.StatusBarGit = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorStatusBarBg)).
		Foreground(lipgloss.Color(ColorSecondary)).
		Padding(0, 1)

	t.Input = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorInputBg)).
		Foreground(lipgloss.Color(ColorText)).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.InputPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)

	t.InputCursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)

	t.Welcome = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorWelcomeBg)).
		Padding(2, 4)

	t.WelcomeTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		PaddingBottom(1)

	t.WelcomeTip = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextMuted)).
		PaddingLeft(2)

	t.WelcomeKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSecondary)).
		Bold(true)

	t.PermissionDialog = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPermissionBg)).
		Padding(1, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPermissionBorder))

	t.PermissionTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPermissionBorder)).
		Bold(true).
		PaddingBottom(1)

	t.PermissionOption = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		PaddingLeft(2)

	t.PermissionSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		PaddingLeft(2)

	t.Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))

	t.Scrollable = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBackground))

	return t
}
