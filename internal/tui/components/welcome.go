package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type WelcomeScreen struct {
	Theme     styles.Theme
	Dismissed bool
}

func NewWelcomeScreen(theme styles.Theme) *WelcomeScreen {
	return &WelcomeScreen{Theme: theme}
}

func (w *WelcomeScreen) Render(width int) string {
	if w.Dismissed {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorPrimary)).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorSecondary)).
		Bold(true)

	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted))

	var b strings.Builder

	b.WriteString(titleStyle.Render("Claude Code — Go Edition") + "\n\n")
	b.WriteString(tipStyle.Render("Quick Tips:") + "\n")
	b.WriteString("  " + keyStyle.Render("/help") + " " + tipStyle.Render("- Show all commands") + "\n")
	b.WriteString("  " + keyStyle.Render("/clear") + " " + tipStyle.Render("- Clear conversation") + "\n")
	b.WriteString("  " + keyStyle.Render("/models") + " " + tipStyle.Render("- Switch models") + "\n")
	b.WriteString("\n")
	b.WriteString(tipStyle.Render("Press any key to dismiss"))

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Render(b.String())
}
