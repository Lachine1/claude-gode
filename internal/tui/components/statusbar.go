package components

import (
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

// StatusBar renders at the very bottom
// Format: <model> | <perm_mode> | <git_branch>
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

	modelStr := s.Theme.PromptFooter.Render(s.Model)
	permStr := s.Theme.PromptFooter.Render(s.PermMode)

	var parts []string
	parts = append(parts, modelStr)
	parts = append(parts, permStr)
	if s.GitBranch != "" {
		parts = append(parts, s.Theme.PromptFooter.Render(s.GitBranch))
	}

	bar := ""
	for i, p := range parts {
		if i > 0 {
			bar += " | "
		}
		bar += p
	}

	return s.Theme.PromptFooter.Width(width).Render(bar)
}
