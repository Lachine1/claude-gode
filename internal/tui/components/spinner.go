package components

import (
	"fmt"

	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type Spinner struct {
	Frame   int
	Count   int
	Theme   styles.Theme
	Running bool
}

var BrailleSpinner = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

func NewSpinner(theme styles.Theme) *Spinner {
	return &Spinner{
		Theme: theme,
	}
}

func (s *Spinner) Tick() {
	s.Frame = (s.Frame + 1) % len(BrailleSpinner)
}

func (s *Spinner) Render() string {
	if !s.Running {
		return ""
	}

	frame := BrailleSpinner[s.Frame]
	if s.Count > 0 {
		return fmt.Sprintf("%s %d tokens", frame, s.Count)
	}
	return frame
}

func (s *Spinner) SetCount(count int) {
	s.Count = count
}

func (s *Spinner) Start() {
	s.Running = true
	s.Frame = 0
}

func (s *Spinner) Stop() {
	s.Running = false
	s.Count = 0
}
