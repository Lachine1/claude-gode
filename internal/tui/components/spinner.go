package components

import (
	"fmt"
	"time"

	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

// Spinner frames (braille, 120ms interval)
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner shows loading/processing status
// Format: ⠋ Working... (↓ 1,234 tokens · 0:05)
type Spinner struct {
	Theme        styles.Theme
	Mode         string // "requesting", "tool_use", "responding", "thinking"
	Frame        int
	StartTime    time.Time
	TokenCount   int
	Elapsed      time.Duration
	Suffix       string
	Thinking     bool
	ThinkingDone bool
	ThinkingDur  time.Duration
	Stalled      bool
	OverrideMsg  string
}

func NewSpinner(theme styles.Theme) *Spinner {
	return &Spinner{
		Theme:     theme,
		Mode:      "responding",
		StartTime: time.Now(),
	}
}

func (s *Spinner) Tick() {
	s.Frame = (s.Frame + 1) % len(SpinnerFrames)
	s.Elapsed = time.Since(s.StartTime)

	// Stalled detection: 3s with no progress
	if s.Elapsed > 3*time.Second && s.TokenCount == 0 {
		s.Stalled = true
	}
}

func (s *Spinner) Render(width int) string {
	if width <= 0 {
		return ""
	}

	glyph := SpinnerFrames[s.Frame]

	if s.Stalled {
		glyph = s.Theme.ToolCallError.Render(glyph)
	} else {
		glyph = s.Theme.SpinnerGlyph.Render(glyph)
	}

	var msg string
	if s.OverrideMsg != "" {
		msg = s.OverrideMsg
	} else {
		switch s.Mode {
		case "requesting":
			msg = "Requesting..."
		case "tool_use":
			msg = "Using tools..."
		case "thinking":
			msg = "Thinking..."
		default:
			msg = "Working..."
		}
	}

	if s.Stalled {
		msg = s.Theme.ToolCallError.Render(msg)
	} else {
		msg = s.Theme.SpinnerText.Render(msg)
	}

	var parts []string
	if s.Suffix != "" {
		parts = append(parts, s.Theme.SpinnerStatus.Render(s.Suffix))
	}
	if s.Elapsed > 0 {
		parts = append(parts, s.Theme.SpinnerTimer.Render(formatDuration(s.Elapsed)))
	}
	if s.TokenCount > 0 {
		tokenStr := s.Theme.SpinnerStatus.Render(fmt.Sprintf("↓ %s tokens", formatNumber(s.TokenCount)))
		parts = append(parts, tokenStr)
	}
	if s.Thinking && !s.ThinkingDone {
		parts = append(parts, s.Theme.SpinnerStatus.Render("thinking"))
	} else if s.ThinkingDone {
		parts = append(parts, s.Theme.SpinnerTimer.Render(fmt.Sprintf("thought for %s", formatDuration(s.ThinkingDur))))
	}

	var statusStr string
	if len(parts) > 0 {
		statusStr = " (" + joinParts(parts) + ")"
	}

	return glyph + " " + msg + statusStr
}

func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	mins := total / 60
	secs := total % 60
	if mins > 0 {
		return fmt.Sprintf("%d:%02d", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " · "
		}
		result += p
	}
	return result
}
