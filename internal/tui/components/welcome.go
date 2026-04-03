package components

import (
	"fmt"

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

	logo := `
  ╔═╗╔═╗╔╦╗╔═╗╦ ╦╔═╗╦═╗
  ╠╣ ║ ║ ║║║╣ ║║║║ ║╠╦╝
  ╚  ╚═╝═╩╝╚═╝╚╩╝╚═╝╚╚
`

	title := w.Theme.WelcomeTitle.Render("Claude Code — Go Edition")
	tips := []string{
		"Type your message and press Enter to start",
		"Use /help to see all available commands",
		"Use /models to switch AI models",
		"Use /settings to configure preferences",
		"Press Ctrl+C or q to quit",
	}

	var tipLines string
	for i, tip := range tips {
		num := w.Theme.WelcomeKey.Render(fmt.Sprintf("%d.", i+1))
		tipLines += fmt.Sprintf("  %s %s\n", num, w.Theme.WelcomeTip.Render(tip))
	}

	footer := w.Theme.WelcomeKey.Render("Press any key to dismiss")

	content := logo + "\n" + title + "\n\n" +
		w.Theme.WelcomeTitle.Render("Quick Tips:") + "\n\n" + tipLines + "\n" + footer

	return w.Theme.Welcome.Render(content)
}
