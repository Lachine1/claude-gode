package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

// DisplayMessage represents a rendered message in the UI
type DisplayMessage struct {
	Type      string // "user", "assistant", "tool_call", "tool_result", "thinking", "error", "command_output"
	Content   string
	ToolName  string
	ToolID    string
	Status    string // "running", "success", "error"
	Output    string
	IsError   bool
	Collapsed bool
	Spinner   string
	Theme     styles.Theme
}

// MessageList holds all messages and renders them scrollable
type MessageList struct {
	Messages []DisplayMessage
	Scroll   int
	Theme    styles.Theme
	Height   int
}

func NewMessageList(theme styles.Theme) *MessageList {
	return &MessageList{Theme: theme, Height: 20}
}

func (ml *MessageList) AddUserMessage(text string) {
	ml.Messages = append(ml.Messages, DisplayMessage{
		Type:    "user",
		Content: text,
		Theme:   ml.Theme,
	})
	ml.ScrollToBottom()
}

func (ml *MessageList) AddAssistantMessage(text string) {
	ml.Messages = append(ml.Messages, DisplayMessage{
		Type:    "assistant",
		Content: text,
		Theme:   ml.Theme,
	})
	ml.ScrollToBottom()
}

func (ml *MessageList) AppendToAssistant(text string) {
	for i := len(ml.Messages) - 1; i >= 0; i-- {
		if ml.Messages[i].Type == "assistant" {
			ml.Messages[i].Content += text
			break
		}
	}
	ml.ScrollToBottom()
}

func (ml *MessageList) AddToolCall(toolName, input, toolID string) {
	ml.Messages = append(ml.Messages, DisplayMessage{
		Type:     "tool_call",
		ToolName: toolName,
		Content:  input,
		ToolID:   toolID,
		Status:   "running",
		Spinner:  "⠋",
		Theme:    ml.Theme,
	})
	ml.ScrollToBottom()
}

func (ml *MessageList) CompleteToolCall(toolID, output string, isError bool) {
	for i := len(ml.Messages) - 1; i >= 0; i-- {
		if ml.Messages[i].Type == "tool_call" && ml.Messages[i].ToolID == toolID {
			ml.Messages[i].Status = "success"
			if isError {
				ml.Messages[i].Status = "error"
			}
			ml.Messages[i].Output = output
			break
		}
	}
	ml.ScrollToBottom()
}

func (ml *MessageList) AddSystemMessage(text string) {
	ml.Messages = append(ml.Messages, DisplayMessage{
		Type:    "command_output",
		Content: text,
		Theme:   ml.Theme,
	})
	ml.ScrollToBottom()
}

func (ml *MessageList) StartAssistant() {
	ml.Messages = append(ml.Messages, DisplayMessage{
		Type:    "assistant",
		Content: "",
		Theme:   ml.Theme,
	})
	ml.ScrollToBottom()
}

func (ml *MessageList) ScrollUp() {
	if ml.Scroll > 0 {
		ml.Scroll--
	}
}

func (ml *MessageList) ScrollDown(maxLines int) {
	if ml.Scroll < maxLines-ml.Height {
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

func (ml *MessageList) Render(width int) string {
	if ml.Height <= 0 {
		ml.Height = 20
	}

	var lines []string
	for _, msg := range ml.Messages {
		rendered := renderMessage(msg, width)
		if rendered != "" {
			lines = append(lines, rendered)
		}
	}

	allContent := strings.Join(lines, "\n")
	if allContent == "" {
		return ""
	}

	totalLines := strings.Count(allContent, "\n") + 1

	if totalLines <= ml.Height {
		return ml.Theme.Scrollable.Width(width).Render(allContent)
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
	return ml.Theme.Scrollable.Width(width).Render(visible)
}

func renderMessage(msg DisplayMessage, width int) string {
	switch msg.Type {
	case "user":
		return renderUserMessage(msg, width)
	case "assistant":
		return renderAssistantMessage(msg, width)
	case "tool_call":
		return renderToolCall(msg, width)
	case "tool_result":
		return renderToolResult(msg, width)
	case "thinking":
		return renderThinking(msg, width)
	case "error":
		return msg.Theme.Error.Render(fmt.Sprintf("Error: %s", msg.Content))
	case "command_output":
		return msg.Theme.CommandOutput.Render(msg.Content)
	default:
		return msg.Content
	}
}

func ensureTheme(t styles.Theme) styles.Theme {
	if t.UserMessage.GetBackground() == nil {
		t.UserMessage = t.UserMessage.
			Background(lipgloss.Color(styles.ColorUserBg)).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)
	}
	if t.AssistantMessage.GetForeground() == nil {
		t.AssistantMessage = t.AssistantMessage.
			Foreground(lipgloss.Color(styles.ColorText)).
			MarginTop(1).
			MarginBottom(1)
	}
	if !t.ToolCall.GetBorderLeft() {
		t.ToolCall = t.ToolCall.
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(false).
			BorderLeft(true).
			BorderRight(false).
			BorderBottom(false).
			BorderForeground(lipgloss.Color(styles.ColorPermission)).
			PaddingLeft(1).
			MarginTop(1).
			MarginBottom(1)
	}
	return t
}

// renderUserMessage: gray background, NO prefix, just the content
func renderUserMessage(msg DisplayMessage, width int) string {
	t := ensureTheme(msg.Theme)
	return t.UserMessage.Width(width).Render(msg.Content)
}

// renderAssistantMessage: plain white text, NO prefix
func renderAssistantMessage(msg DisplayMessage, width int) string {
	t := ensureTheme(msg.Theme)
	if msg.Content == "" {
		return t.AssistantMessage.Width(width).Render("●")
	}
	return t.AssistantMessage.Width(width).Render(msg.Content)
}

// renderToolCall: left border (│), format: "⟳ tool_name — input_summary"
func renderToolCall(msg DisplayMessage, width int) string {
	t := ensureTheme(msg.Theme)

	var statusIcon string
	var statusText string

	switch msg.Status {
	case "running":
		statusIcon = msg.Spinner
		statusText = t.ToolCallRunning.Render(msg.ToolName)
	case "success":
		statusIcon = "✓"
		statusText = t.ToolCallSuccess.Render(msg.ToolName)
	case "error":
		statusIcon = "✗"
		statusText = t.ToolCallError.Render(msg.ToolName)
	default:
		statusIcon = "◌"
		statusText = t.ToolCallRunning.Render(msg.ToolName)
	}

	inputSummary := msg.Content
	if len(inputSummary) > 80 {
		inputSummary = inputSummary[:77] + "..."
	}

	var content string
	if inputSummary != "" && inputSummary != msg.ToolName {
		content = statusIcon + " " + statusText + " — " + inputSummary
	} else {
		content = statusIcon + " " + statusText
	}

	return t.ToolCall.Width(width).Render(content)
}

// renderToolResult: dimmed, collapsed preview
func renderToolResult(msg DisplayMessage, width int) string {
	if msg.Collapsed {
		preview := msg.Output
		if len(preview) > 120 {
			preview = preview[:117] + "..."
		}
		return msg.Theme.ToolResult.Width(width).Render(preview)
	}
	return msg.Theme.ToolResult.Width(width).Render(msg.Output)
}

// renderThinking: dimmed, italic, subtle
func renderThinking(msg DisplayMessage, width int) string {
	return msg.Theme.ThinkingBlock.Width(width).Render(msg.Content)
}
