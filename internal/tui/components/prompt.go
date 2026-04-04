package components

import (
	"strings"

	"github.com/Lachine1/claude-gode/internal/completion"
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

	// Completion
	Suggestions        []completion.SuggestionItem
	SelectedSuggestion int
	GhostText          *completion.GhostText
	ShowSuggestions    bool
}

func NewPromptInput(theme styles.Theme) *PromptInput {
	return &PromptInput{
		Theme:   theme,
		HistIdx: -1,
		Mode:    "prompt",
	}
}

func (p *PromptInput) Insert(r rune) {
	buf := []rune(p.Buffer)
	if p.Cursor < 0 || p.Cursor > len(buf) {
		p.Cursor = len(buf)
	}
	buf = append(buf[:p.Cursor], append([]rune{r}, buf[p.Cursor:]...)...)
	p.Buffer = string(buf)
	p.Cursor++
}

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

func (p *PromptInput) UpdateSuggestions(items []completion.SuggestionItem) {
	p.Suggestions = items
	p.SelectedSuggestion = 0
	p.ShowSuggestions = len(items) > 0
}

func (p *PromptInput) AcceptSuggestion() {
	if !p.ShowSuggestions || len(p.Suggestions) == 0 {
		return
	}
	s := p.Suggestions[p.SelectedSuggestion]
	if strings.HasPrefix(p.Buffer, "/") {
		query := strings.TrimPrefix(p.Buffer, "/")
		if spaceIdx := strings.Index(query, " "); spaceIdx != -1 {
			p.Buffer = "/" + s.ID + query[spaceIdx:]
		} else {
			p.Buffer = "/" + s.ID
		}
	} else if idx := strings.LastIndex(p.Buffer, "@"); idx != -1 {
		p.Buffer = p.Buffer[:idx] + s.ID
	} else {
		p.Buffer = s.DisplayText
	}
	p.Cursor = len([]rune(p.Buffer))
	p.DismissSuggestions()
}

func (p *PromptInput) AcceptGhostText() {
	if p.GhostText == nil {
		return
	}
	p.Buffer = p.GhostText.FullCommand
	p.Cursor = len([]rune(p.Buffer))
	p.GhostText = nil
}

func (p *PromptInput) NextSuggestion() {
	if !p.ShowSuggestions || len(p.Suggestions) == 0 {
		return
	}
	p.SelectedSuggestion = (p.SelectedSuggestion + 1) % len(p.Suggestions)
}

func (p *PromptInput) PrevSuggestion() {
	if !p.ShowSuggestions || len(p.Suggestions) == 0 {
		return
	}
	p.SelectedSuggestion = (p.SelectedSuggestion - 1 + len(p.Suggestions)) % len(p.Suggestions)
}

func (p *PromptInput) DismissSuggestions() {
	p.Suggestions = nil
	p.SelectedSuggestion = 0
	p.ShowSuggestions = false
}

func (p *PromptInput) RenderSuggestions(width int) string {
	if !p.ShowSuggestions || len(p.Suggestions) == 0 || width <= 0 {
		return ""
	}

	var lines []string
	maxLines := 3
	if len(p.Suggestions) < maxLines {
		maxLines = len(p.Suggestions)
	}

	for i := 0; i < maxLines; i++ {
		s := p.Suggestions[i]
		var prefix string
		if i == p.SelectedSuggestion {
			prefix = "❯ "
		} else if i == p.SelectedSuggestion-1 {
			prefix = "↑ "
		} else if i == p.SelectedSuggestion+1 {
			prefix = "↓ "
		} else {
			prefix = "  "
		}

		display := s.DisplayText
		if s.Description != "" {
			display += " - " + s.Description
		}
		tag := "[" + s.Tag + "]"

		line := prefix + display
		if len(line) > width-len(tag)-2 {
			line = line[:width-len(tag)-2]
		}
		padding := width - len(line) - len(tag)
		if padding > 0 {
			line += strings.Repeat(" ", padding)
		}
		line += tag

		if i == p.SelectedSuggestion {
			lines = append(lines, p.Theme.Suggestion.Render(line))
		} else {
			lines = append(lines, p.Theme.Subtle.Render(line))
		}
	}

	return strings.Join(lines, "\n")
}

func (p *PromptInput) Render(width int) string {
	if width <= 0 {
		return ""
	}

	var prefix string
	switch p.Mode {
	case "bash":
		prefix = "! "
	case "memory":
		prefix = "# "
	default:
		prefix = "❯ "
	}

	buf := []rune(p.Buffer)
	cursorPos := p.Cursor
	if cursorPos > len(buf) {
		cursorPos = len(buf)
	}

	var inputLine string
	left := string(buf[:cursorPos])
	right := string(buf[cursorPos:])

	if p.Focused {
		if p.GhostText != nil && p.GhostText.InsertPosition == len(buf) {
			inputLine = prefix + left + "█" + p.Theme.Subtle.Render(p.GhostText.Text)
		} else {
			inputLine = prefix + left + "█" + right
		}
	} else {
		inputLine = prefix + p.Buffer
	}

	if len(inputLine) > width-2 {
		inputLine = inputLine[:width-2]
	}

	border := p.Theme.PromptBorder.Width(width).Render(" ")
	footer := p.Theme.PromptFooter.Render(p.Model + " | " + p.PermMode + " | Tab to complete · Ctrl+O for help")

	result := inputLine + "\n" + border + "\n" + footer

	if p.ShowSuggestions && len(p.Suggestions) > 0 {
		suggestions := p.RenderSuggestions(width)
		if suggestions != "" {
			result = suggestions + "\n" + result
		}
	}

	return result
}
