package components

import (
	"strings"

	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type Input struct {
	Buffer  string
	Cursor  int
	History []string
	HistIdx int
	Theme   styles.Theme
	Width   int
	Focused bool
}

func NewInput(theme styles.Theme) *Input {
	return &Input{
		Theme:   theme,
		HistIdx: -1,
	}
}

func (i *Input) Insert(r rune) {
	if i.Cursor < 0 {
		i.Cursor = 0
	}
	buf := []rune(i.Buffer)
	if i.Cursor > len(buf) {
		i.Cursor = len(buf)
	}
	buf = append(buf[:i.Cursor], append([]rune{r}, buf[i.Cursor:]...)...)
	i.Buffer = string(buf)
	i.Cursor++
}

func (i *Input) Backspace() {
	if i.Cursor <= 0 {
		return
	}
	buf := []rune(i.Buffer)
	if i.Cursor > len(buf) {
		i.Cursor = len(buf)
	}
	buf = append(buf[:i.Cursor-1], buf[i.Cursor:]...)
	i.Buffer = string(buf)
	i.Cursor--
}

func (i *Input) Delete() {
	buf := []rune(i.Buffer)
	if i.Cursor >= len(buf) {
		return
	}
	buf = append(buf[:i.Cursor], buf[i.Cursor+1:]...)
	i.Buffer = string(buf)
}

func (i *Input) MoveLeft() {
	if i.Cursor > 0 {
		i.Cursor--
	}
}

func (i *Input) MoveRight() {
	buf := []rune(i.Buffer)
	if i.Cursor < len(buf) {
		i.Cursor++
	}
}

func (i *Input) MoveHome() {
	i.Cursor = 0
}

func (i *Input) MoveEnd() {
	i.Cursor = len([]rune(i.Buffer))
}

func (i *Input) Submit() string {
	text := strings.TrimSpace(i.Buffer)
	if text == "" {
		return ""
	}
	i.History = append(i.History, text)
	i.HistIdx = -1
	i.Buffer = ""
	i.Cursor = 0
	return text
}

func (i *Input) HistoryUp() {
	if len(i.History) == 0 {
		return
	}
	if i.HistIdx == -1 {
		i.HistIdx = len(i.History) - 1
	} else if i.HistIdx > 0 {
		i.HistIdx--
	} else {
		return
	}
	i.Buffer = i.History[i.HistIdx]
	i.Cursor = len([]rune(i.Buffer))
}

func (i *Input) HistoryDown() {
	if i.HistIdx == -1 {
		return
	}
	if i.HistIdx < len(i.History)-1 {
		i.HistIdx++
		i.Buffer = i.History[i.HistIdx]
		i.Cursor = len([]rune(i.Buffer))
	} else {
		i.HistIdx = -1
		i.Buffer = ""
		i.Cursor = 0
	}
}

func (i *Input) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}

	prompt := i.Theme.InputPrompt.Render("❯ ")

	buf := []rune(i.Buffer)
	cursorPos := i.Cursor
	if cursorPos > len(buf) {
		cursorPos = len(buf)
	}

	var display string
	if i.Focused {
		left := string(buf[:cursorPos])
		right := string(buf[cursorPos:])
		cursor := i.Theme.InputCursor.Render("█")
		display = prompt + left + cursor + right
	} else {
		display = prompt + i.Buffer
	}

	return i.Theme.Input.Width(w).Render(display)
}
