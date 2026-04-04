package components

import (
	"strings"

	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

// PromptInput is the bottom input area with ❯ prefix and footer
type PromptInput struct {
	Buffer    string
	Cursor    int
	History   []string
	HistIdx   int
	Theme     styles.Theme
	Width     int
	Focused   bool
	Mode      string // "prompt", "bash", "memory"
	Model     string
	PermMode  string
	IsLoading bool
}

func NewPromptInput(theme styles.Theme) *PromptInput {
	return &PromptInput{
		Theme:   theme,
		HistIdx: -1,
		Mode:    "prompt",
	}
}

// Insert adds a character at the cursor position
func (p *PromptInput) Insert(r rune) {
	buf := []rune(p.Buffer)
	if p.Cursor < 0 || p.Cursor > len(buf) {
		p.Cursor = len(buf)
	}
	buf = append(buf[:p.Cursor], append([]rune{r}, buf[p.Cursor:]...)...)
	p.Buffer = string(buf)
	p.Cursor++
}

// Backspace deletes character before cursor
func (p *PromptInput) Backspace() {
	if p.Cursor <= 0 {
		return
	}
	buf := []rune(p.Buffer)
	if p.Cursor > len(buf) {
		p.Cursor = len(buf)
	}
	buf = append(buf[:p.Cursor-1], buf[p.Cursor:]...)
	p.Buffer = string(buf)
	p.Cursor--
}

// Delete deletes character at cursor
func (p *PromptInput) Delete() {
	buf := []rune(p.Buffer)
	if p.Cursor >= len(buf) {
		return
	}
	buf = append(buf[:p.Cursor], buf[p.Cursor+1:]...)
	p.Buffer = string(buf)
}

func (p *PromptInput) MoveLeft() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

func (p *PromptInput) MoveRight() {
	buf := []rune(p.Buffer)
	if p.Cursor < len(buf) {
		p.Cursor++
	}
}

func (p *PromptInput) MoveHome() {
	p.Cursor = 0
}

func (p *PromptInput) MoveEnd() {
	p.Cursor = len([]rune(p.Buffer))
}

// Submit returns the input text and clears the buffer
func (p *PromptInput) Submit() string {
	text := strings.TrimSpace(p.Buffer)
	if text == "" {
		return ""
	}
	p.History = append(p.History, text)
	p.HistIdx = -1
	p.Buffer = ""
	p.Cursor = 0
	return text
}

func (p *PromptInput) HistoryUp() {
	if len(p.History) == 0 {
		return
	}
	if p.HistIdx == -1 {
		p.HistIdx = len(p.History) - 1
	} else if p.HistIdx > 0 {
		p.HistIdx--
	} else {
		return
	}
	p.Buffer = p.History[p.HistIdx]
	p.Cursor = len([]rune(p.Buffer))
}

func (p *PromptInput) HistoryDown() {
	if p.HistIdx == -1 {
		return
	}
	if p.HistIdx < len(p.History)-1 {
		p.HistIdx++
		p.Buffer = p.History[p.HistIdx]
		p.Cursor = len([]rune(p.Buffer))
	} else {
		p.HistIdx = -1
		p.Buffer = ""
		p.Cursor = 0
	}
}

// Render renders the full prompt input area with footer
// Format:
// ❯ <input text with cursor>
// ╰────────────────────────────────────── (bottom border)
// <model> | <perm_mode> | Ctrl+O for help
func (p *PromptInput) Render(width int) string {
	if width <= 0 {
		return ""
	}

	// Mode indicator prefix: ❯ for prompt, ! for bash, # for memory
	var prefix string
	switch p.Mode {
	case "bash":
		prefix = "! "
	case "memory":
		prefix = "# "
	default:
		prefix = "❯ "
	}

	// Build input line with cursor
	buf := []rune(p.Buffer)
	cursorPos := p.Cursor
	if cursorPos > len(buf) {
		cursorPos = len(buf)
	}

	var inputLine string
	left := string(buf[:cursorPos])
	right := string(buf[cursorPos:])

	if p.Focused {
		// Cursor shown as inverted space or block
		inputLine = prefix + left + "█" + right
	} else {
		inputLine = prefix + p.Buffer
	}

	// Truncate to width
	if len(inputLine) > width-2 {
		inputLine = inputLine[:width-2]
	}

	// Bottom border: ╰─────────────────╯
	border := p.Theme.PromptBorder.Width(width).Render(" ")

	// Footer: model | perm_mode | Ctrl+O for help
	footer := p.Theme.PromptFooter.Render(p.Model + " | " + p.PermMode + " | Ctrl+O for help")

	return inputLine + "\n" + border + "\n" + footer
}
