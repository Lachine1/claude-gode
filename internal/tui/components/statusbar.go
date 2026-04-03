package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type StatusBar struct {
	Model     string
	InputTok  int
	OutputTok int
	Cost      float64
	PermMode  string
	GitBranch string
	Theme     styles.Theme
	Width     int
}

func NewStatusBar(theme styles.Theme) *StatusBar {
	return &StatusBar{
		Theme:    theme,
		Model:    "claude-sonnet-4-20250514",
		PermMode: "default",
	}
}

func (s *StatusBar) Update(model string, inputTok, outputTok int, cost float64, permMode, gitBranch string) {
	if model != "" {
		s.Model = model
	}
	s.InputTok = inputTok
	s.OutputTok = outputTok
	s.Cost = cost
	if permMode != "" {
		s.PermMode = permMode
	}
	s.GitBranch = gitBranch
}

func (s *StatusBar) Render(width int) string {
	if width <= 0 {
		return ""
	}

	// Model name (orange/bold)
	modelStr := s.Theme.StatusModel.Render(fmt.Sprintf(" %s ", s.Model))

	// Token count (cyan)
	var tokStr string
	if s.InputTok > 0 || s.OutputTok > 0 {
		tokStr = s.Theme.StatusTokens.Render(fmt.Sprintf(" %d in / %d out ", s.InputTok, s.OutputTok))
	} else {
		tokStr = ""
	}

	// Cost (warning color)
	var costStr string
	if s.Cost > 0 {
		costStr = fmt.Sprintf(" $%.2f ", s.Cost)
	}

	// Permission mode (green)
	permStr := s.Theme.StatusMode.Render(fmt.Sprintf(" %s ", s.PermMode))

	// Git branch (muted)
	var gitStr string
	if s.GitBranch != "" {
		gitStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorTextMuted)).
			Render(fmt.Sprintf(" %s ", s.GitBranch))
	}

	// Help hint
	helpStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted)).
		Render(" ? for help ")

	parts := []string{modelStr}
	if tokStr != "" {
		parts = append(parts, tokStr)
	}
	if costStr != "" {
		parts = append(parts, s.Theme.StatusTokens.Render(costStr))
	}
	parts = append(parts, permStr)
	if gitStr != "" {
		parts = append(parts, gitStr)
	}
	parts = append(parts, helpStr)

	bar := strings.Join(parts, "│")

	return s.Theme.StatusBar.Width(width).Render(bar)
}
