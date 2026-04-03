package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

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

func (ml *MessageList) Render(width int) string {
	if ml.Height <= 0 {
		ml.Height = 20
	}

	var lines []string

	for _, msg := range ml.Messages {
		rendered := renderMessage(msg, width)
		lines = append(lines, rendered)
	}

	allContent := strings.Join(lines, "\n")
	allLines := strings.Split(allContent, "\n")

	totalLines := len(allLines)
	if totalLines <= ml.Height {
		return lipgloss.NewStyle().
			Width(width).
			Height(ml.Height).
			Background(lipgloss.Color(styles.ColorBackground)).
			Render(allContent)
	}

	if ml.Scroll < 0 {
		ml.Scroll = 0
	}
	maxScroll := totalLines - ml.Height
	if ml.Scroll > maxScroll {
		ml.Scroll = maxScroll
	}

	start := ml.Scroll
	end := start + ml.Height
	if end > totalLines {
		end = totalLines
	}

	visible := strings.Join(allLines[start:end], "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(ml.Height).
		Background(lipgloss.Color(styles.ColorBackground)).
		Render(visible)
}

func renderMessage(msg DisplayMessage, width int) string {
	w := width - 4
	if w < 20 {
		w = 20
	}

	switch msg.Type {
	case "user":
		return renderUserMessage(msg, w)
	case "assistant":
		return renderAssistantMessage(msg, w)
	case "command_output":
		return renderCommandOutput(msg, w)
	case "tool_call":
		return renderToolCall(msg, w)
	case "tool_result":
		return renderToolResult(msg, w)
	case "thinking":
		return renderThinking(msg, w)
	case "error":
		return msg.Theme.Error.Render("Error: " + msg.Content)
	default:
		return msg.Content
	}
}

func renderUserMessage(msg DisplayMessage, width int) string {
	prefix := msg.Theme.UserPrefix.Render("You:")
	content := wrapText(msg.Content, width-8)
	lines := strings.Split(content, "\n")
	var builder strings.Builder
	builder.WriteString(prefix + " ")
	for i, line := range lines {
		if i == 0 {
			builder.WriteString(line + "\n")
		} else {
			builder.WriteString("      " + line + "\n")
		}
	}
	return msg.Theme.UserMessage.Width(width).Render(strings.TrimSuffix(builder.String(), "\n"))
}

func renderAssistantMessage(msg DisplayMessage, width int) string {
	prefix := msg.Theme.AssistantPrefix.Render("Claude:")
	content := wrapText(msg.Content, width-8)
	lines := strings.Split(content, "\n")
	var builder strings.Builder
	builder.WriteString(prefix + " ")
	for i, line := range lines {
		if i == 0 {
			builder.WriteString(line + "\n")
		} else {
			builder.WriteString("       " + line + "\n")
		}
	}
	return msg.Theme.AssistantMessage.Width(width).Render(strings.TrimSuffix(builder.String(), "\n"))
}

func renderCommandOutput(msg DisplayMessage, width int) string {
	lines := strings.Split(msg.Content, "\n")
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString("  " + line + "\n")
	}
	return msg.Theme.CommandOutput.Width(width).Render(strings.TrimSuffix(builder.String(), "\n"))
}

func renderToolCall(msg DisplayMessage, width int) string {
	var statusIcon string
	var statusStyle lipgloss.Style

	switch msg.Status {
	case "running":
		statusIcon = msg.Spinner
		statusStyle = msg.Theme.ToolCallRunning
	case "success":
		statusIcon = "✓"
		statusStyle = msg.Theme.ToolCallSuccess
	case "error":
		statusIcon = "✗"
		statusStyle = msg.Theme.ToolCallError
	default:
		statusIcon = "◌"
		statusStyle = msg.Theme.ToolCallRunning
	}

	toolName := msg.ToolName
	if toolName == "" {
		toolName = msg.Content
	}

	input := msg.Output
	if input == "" {
		input = msg.Content
	}

	header := fmt.Sprintf("%s %s", statusStyle.Render(statusIcon), lipgloss.NewStyle().Bold(true).Render(toolName))

	var body string
	if input != "" && input != toolName {
		preview := input
		if len(preview) > 100 {
			preview = preview[:97] + "..."
		}
		body = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorTextMuted)).
			Render(preview)
	}

	content := header
	if body != "" {
		content = header + "\n" + body
	}

	return msg.Theme.ToolCall.Width(width).Render(content)
}

func renderToolResult(msg DisplayMessage, width int) string {
	icon := "✓"
	if msg.IsError {
		icon = "✗"
	}

	header := fmt.Sprintf("%s %s", icon, msg.ToolName)

	output := msg.Output
	if len(output) > 200 {
		output = output[:197] + "..."
	}

	content := header + "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted)).
		Render(output)

	return msg.Theme.ToolResult.Width(width).Render(content)
}

func renderThinking(msg DisplayMessage, width int) string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorSecondary)).
		Bold(true).
		Render("Thinking...")

	preview := msg.Content
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}

	content := header + " " + lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ColorTextMuted)).
		Italic(true).
		Render(preview)

	return msg.Theme.ThinkingBlock.Width(width).Render(content)
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line + "\n")
			continue
		}

		words := strings.Fields(line)
		currentLen := 0

		for _, word := range words {
			wordLen := len(word) + 1
			if currentLen+wordLen > width && currentLen > 0 {
				result.WriteString("\n")
				currentLen = 0
			}
			if currentLen > 0 {
				result.WriteString(" ")
			}
			result.WriteString(word)
			currentLen += wordLen
		}
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func (ml *MessageList) ScrollUp() {
	if ml.Scroll > 0 {
		ml.Scroll--
	}
}

func (ml *MessageList) ScrollDown() {
	ml.Scroll++
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

func (ml *MessageList) PageDown() {
	ml.Scroll += ml.Height
}
