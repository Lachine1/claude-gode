package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type UserMessage struct {
	Content string
	Theme   styles.Theme
}

func (m UserMessage) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}
	return m.Theme.UserMessage.Width(w).Render(m.Content)
}

type AssistantMessage struct {
	Content string
	Theme   styles.Theme
}

func (m AssistantMessage) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}
	return m.Theme.AssistantMessage.Width(w).Render(m.Content)
}

type ToolCallDisplay struct {
	ToolName  string
	Input     string
	ToolUseID string
	Status    string
	Theme     styles.Theme
	Spinner   string
}

func (t ToolCallDisplay) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}

	var statusStyle lipgloss.Style
	var statusIcon string

	switch t.Status {
	case "running":
		statusStyle = t.Theme.ToolCallRunning
		statusIcon = "⟳ " + t.Spinner
	case "success":
		statusStyle = t.Theme.ToolCallSuccess
		statusIcon = "✓"
	case "error":
		statusStyle = t.Theme.ToolCallError
		statusIcon = "✗"
	default:
		statusStyle = t.Theme.ToolCallRunning
		statusIcon = "◌"
	}

	toolName := lipgloss.NewStyle().Bold(true).Render(t.ToolName)
	statusStr := statusStyle.Render(statusIcon)
	inputSummary := t.Input
	if len(inputSummary) > 80 {
		inputSummary = inputSummary[:77] + "..."
	}

	header := fmt.Sprintf("%s  %s", statusStr, toolName)
	body := ""
	if inputSummary != "" {
		body = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorTextMuted)).
			Render(inputSummary)
	}

	content := header
	if body != "" {
		content += "\n" + body
	}

	return t.Theme.ToolCall.Width(w).Render(content)
}

type ToolResultDisplay struct {
	ToolName  string
	ToolUseID string
	Output    string
	IsError   bool
	Collapsed bool
	Theme     styles.Theme
}

func (t ToolResultDisplay) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}

	statusIcon := "✓"
	var baseStyle lipgloss.Style

	if t.IsError {
		statusIcon = "✗"
		baseStyle = t.Theme.ToolResultError
	} else {
		baseStyle = t.Theme.ToolResult
	}

	header := fmt.Sprintf("%s  %s result", statusIcon, t.ToolName)

	if t.Collapsed {
		preview := t.Output
		if len(preview) > 120 {
			preview = preview[:117] + "..."
		}
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorTextMuted)).
			Italic(true)
		content := header + "\n" + previewStyle.Render(preview)
		return baseStyle.Width(w).Render(content)
	}

	outputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted)).
		MarginTop(1)

	content := header + "\n" + outputStyle.Width(w-4).Render(t.Output)
	return baseStyle.Width(w).Render(content)
}

type ThinkingBlock struct {
	Content   string
	Collapsed bool
	Theme     styles.Theme
}

func (t ThinkingBlock) Render(width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorSecondary)).
		Bold(true).
		Render("Thinking...")

	if t.Collapsed {
		preview := t.Content
		if len(preview) > 80 {
			preview = preview[:77] + "..."
		}
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorTextMuted)).
			Italic(true)
		content := header + "  " + previewStyle.Render(preview)
		return t.Theme.ThinkingBlock.Width(w).Render(content)
	}

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted)).
		MarginTop(1)

	content := header + "\n" + contentStyle.Width(w-4).Render(t.Content)
	return t.Theme.ThinkingBlock.Width(w).Render(content)
}

type DisplayMessage struct {
	Type      string
	Content   string
	ToolName  string
	ToolID    string
	Status    string
	Output    string
	IsError   bool
	Collapsed bool
	Spinner   string
	Theme     styles.Theme
}

type MessageList struct {
	Messages []DisplayMessage
	Scroll   int
	Theme    styles.Theme
	Height   int
}

func (ml MessageList) Render(width int) string {
	if ml.Height <= 0 {
		ml.Height = 20
	}

	var lines []string
	for _, msg := range ml.Messages {
		switch msg.Type {
		case "user":
			lines = append(lines, UserMessage{Content: msg.Content, Theme: msg.Theme}.Render(width))
		case "assistant":
			lines = append(lines, AssistantMessage{Content: msg.Content, Theme: msg.Theme}.Render(width))
		case "tool_call":
			lines = append(lines, ToolCallDisplay{
				ToolName:  msg.ToolName,
				Input:     msg.Content,
				ToolUseID: msg.ToolID,
				Status:    msg.Status,
				Theme:     msg.Theme,
				Spinner:   msg.Spinner,
			}.Render(width))
		case "tool_result":
			lines = append(lines, ToolResultDisplay{
				ToolName:  msg.ToolName,
				ToolUseID: msg.ToolID,
				Output:    msg.Output,
				IsError:   msg.IsError,
				Collapsed: msg.Collapsed,
				Theme:     msg.Theme,
			}.Render(width))
		case "thinking":
			lines = append(lines, ThinkingBlock{
				Content:   msg.Content,
				Collapsed: msg.Collapsed,
				Theme:     msg.Theme,
			}.Render(width))
		case "error":
			lines = append(lines, msg.Theme.Error.Render(fmt.Sprintf("Error: %s", msg.Content)))
		}
	}

	allContent := strings.Join(lines, "\n")
	totalLines := strings.Count(allContent, "\n") + 1

	if totalLines <= ml.Height {
		return ml.Theme.Scrollable.Width(width).Height(ml.Height).Render(allContent)
	}

	if ml.Scroll < 0 {
		ml.Scroll = 0
	}
	if ml.Scroll > totalLines-ml.Height {
		ml.Scroll = totalLines - ml.Height
	}

	allLines := strings.Split(allContent, "\n")
	start := ml.Scroll
	end := start + ml.Height
	if end > len(allLines) {
		end = len(allLines)
	}

	visible := strings.Join(allLines[start:end], "\n")
	return ml.Theme.Scrollable.Width(width).Height(ml.Height).Render(visible)
}

func (ml *MessageList) ScrollUp() {
	if ml.Scroll > 0 {
		ml.Scroll--
	}
}

func (ml *MessageList) ScrollDown(maxLines int) {
	totalLines := maxLines
	if ml.Scroll < totalLines-ml.Height {
		ml.Scroll++
	}
}

func (ml *MessageList) ScrollToBottom() {
	ml.Scroll = 999999
}

func (ml *MessageList) PageUp() {
	ml.Scroll -= ml.Height
	if ml.Scroll < 0 {
		ml.Scroll = 0
	}
}

func (ml *MessageList) PageDown(maxLines int) {
	ml.Scroll += ml.Height
	if ml.Scroll > maxLines-ml.Height {
		ml.Scroll = maxLines - ml.Height
	}
	if ml.Scroll < 0 {
		ml.Scroll = 0
	}
}
