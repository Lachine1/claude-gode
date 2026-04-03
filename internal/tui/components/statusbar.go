package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type StatusBar struct {
	Model      string
	InputTok   int
	OutputTok  int
	CacheRead  int
	CacheWrite int
	Cost       float64
	PermMode   string
	GitBranch  string
	Theme      styles.Theme
	Width      int
}

func NewStatusBar(theme styles.Theme) *StatusBar {
	return &StatusBar{
		Theme:    theme,
		Model:    "claude-sonnet-4-20250514",
		PermMode: "default",
	}
}

func (s *StatusBar) Update(model string, inputTok, outputTok, cacheRead, cacheWrite int, cost float64, permMode, gitBranch string) {
	if model != "" {
		s.Model = model
	}
	s.InputTok = inputTok
	s.OutputTok = outputTok
	s.CacheRead = cacheRead
	s.CacheWrite = cacheWrite
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

	modelStr := s.Theme.StatusBarModel.Render(fmt.Sprintf(" %s ", s.Model))
	tokStr := s.Theme.StatusBarTokens.Render(fmt.Sprintf(" tokens: %d in / %d out ", s.InputTok, s.OutputTok))
	costStr := s.Theme.StatusBarCost.Render(fmt.Sprintf(" cost: $%.4f ", s.Cost))

	var permStyle lipgloss.Style
	switch s.PermMode {
	case "accept":
		permStyle = s.Theme.StatusBarPerm
	case "smart":
		permStyle = s.Theme.StatusBarCost
	case "edit":
		permStyle = s.Theme.StatusBarModel
	default:
		permStyle = s.Theme.StatusBarPerm
	}
	permStr := permStyle.Render(fmt.Sprintf(" perm: %s ", s.PermMode))

	gitStr := ""
	if s.GitBranch != "" {
		gitStr = s.Theme.StatusBarGit.Render(fmt.Sprintf(" git: %s ", s.GitBranch))
	}

	parts := []string{modelStr, tokStr, costStr, permStr}
	if gitStr != "" {
		parts = append(parts, gitStr)
	}

	bar := ""
	for _, p := range parts {
		bar += p
	}

	bar = lipgloss.NewStyle().
		MaxWidth(width).
		Render(bar)

	return s.Theme.StatusBar.Width(width).Render(bar)
}
