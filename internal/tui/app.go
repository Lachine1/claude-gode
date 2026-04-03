package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Lachine1/claude-gode/internal/bootstrap"
	"github.com/Lachine1/claude-gode/internal/commands"
	"github.com/Lachine1/claude-gode/internal/engine"
	"github.com/Lachine1/claude-gode/internal/tui/components"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// Run starts the terminal user interface
func Run(ctx context.Context, state *bootstrap.State, args []string) error {
	theme := styles.DefaultTheme()
	model := newAppModel(state, args, theme)
	p := tea.NewProgram(model, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

type appModel struct {
	theme          styles.Theme
	bootstrap      *bootstrap.State
	messageList    *components.MessageList
	input          *components.Input
	statusBar      *components.StatusBar
	messages       []types.Message
	usage          types.Usage
	cost           float64
	permissionMode string
	gitBranch      string
	width          int
	height         int
	spinnerFrame   int
	spinnerTicker  *time.Ticker
	streamingText  strings.Builder
	streamingTool  strings.Builder
}

func newAppModel(state *bootstrap.State, args []string, theme styles.Theme) appModel {
	input := components.NewInput(theme)
	input.Focused = true
	statusBar := components.NewStatusBar(theme)

	permMode := "default"
	if state.Config.PermissionMode() != "" {
		permMode = state.Config.PermissionMode()
	}

	statusBar.Update(state.Config.Model(), 0, 0, 0, 0, 0.0, permMode, "")

	m := appModel{
		theme:          theme,
		bootstrap:      state,
		messageList:    &components.MessageList{Theme: theme, Height: 20},
		input:          input,
		statusBar:      statusBar,
		messages:       make([]types.Message, 0),
		permissionMode: permMode,
		width:          80,
		height:         24,
		spinnerFrame:   0,
	}

	if len(args) > 0 {
		m.messages = append(m.messages, types.Message{
			Role: types.RoleUser,
			Content: []types.ContentBlock{
				{Type: types.ContentTypeText, Text: strings.Join(args, " ")},
			},
		})
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "user",
			Content: strings.Join(args, " "),
			Theme:   theme,
		})
	}

	return m
}

func (m appModel) Init() tea.Cmd {
	return tea.Tick(time.Second/10, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

type spinnerTickMsg struct{}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 3 {
			m.messageList.Height = m.height - 3
		}
		return m, nil

	case spinnerTickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % 4
		for i := range m.messageList.Messages {
			if m.messageList.Messages[i].Type == "tool_call" && m.messageList.Messages[i].Status == "running" {
				m.messageList.Messages[i].Spinner = spinnerFrames[m.spinnerFrame]
			}
		}
		return m, tea.Tick(time.Second/10, func(time.Time) tea.Msg {
			return spinnerTickMsg{}
		})

	case engineResultMsg:
		return m.handleEngineResult(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.input.Buffer == "" {
				return m, tea.Quit
			}
		case "enter":
			return m.processInput()
		case "up":
			m.input.HistoryUp()
			return m, nil
		case "down":
			m.input.HistoryDown()
			return m, nil
		case "pgup":
			m.messageList.PageUp()
			return m, nil
		case "pgdown":
			m.messageList.PageDown(9999)
			return m, nil
		case "left":
			m.input.MoveLeft()
			return m, nil
		case "right":
			m.input.MoveRight()
			return m, nil
		case "home":
			m.input.MoveHome()
			return m, nil
		case "end":
			m.input.MoveEnd()
			return m, nil
		case "backspace":
			m.input.Backspace()
			return m, nil
		case "delete":
			m.input.Delete()
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.input.Insert(rune(msg.String()[0]))
			}
			return m, nil
		}
	}

	return m, nil
}

func (m appModel) processInput() (tea.Model, tea.Cmd) {
	input := m.input.Submit()
	if input == "" {
		return m, nil
	}

	if strings.HasPrefix(input, "/") {
		return m.handleCommand(input)
	}

	return m.handleUserInput(input)
}

func (m appModel) handleCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	cmd := commands.FindCommand(cmdName, m.bootstrap.Commands)
	if cmd == nil {
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "error",
			Content: fmt.Sprintf("Unknown command: /%s. Type /help for available commands.", cmdName),
			Theme:   m.theme,
		})
		m.messageList.ScrollToBottom()
		return m, nil
	}

	var output strings.Builder
	ctx := &types.CommandContext{
		Cwd:         m.bootstrap.Cwd,
		GetMessages: func() []types.Message { return m.messages },
		SetMessages: func(msgs []types.Message) { m.messages = msgs },
		Debug:       false,
		WriteOutput: func(s string) {
			output.WriteString(s)
			output.WriteString("\n")
		},
	}

	if err := cmd.Handler(ctx, args); err != nil {
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "error",
			Content: fmt.Sprintf("Error: %v", err),
			Theme:   m.theme,
		})
		m.messageList.ScrollToBottom()
		return m, nil
	}

	outStr := output.String()
	if outStr != "" {
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "command_output",
			Content: strings.TrimRight(outStr, "\n"),
			Theme:   m.theme,
		})
		m.messageList.ScrollToBottom()
	}

	return m, nil
}

func (m appModel) handleUserInput(input string) (tea.Model, tea.Cmd) {
	m.messages = append(m.messages, types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{Type: types.ContentTypeText, Text: input},
		},
	})

	m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
		Type:    "user",
		Content: input,
		Theme:   m.theme,
	})

	m.messageList.ScrollToBottom()

	m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
		Type:    "assistant",
		Content: "",
		Status:  "running",
		Theme:   m.theme,
	})
	m.messageList.ScrollToBottom()

	return m, m.submitQuery(input)
}

type engineResultMsg struct {
	text      string
	toolCalls []string
	err       error
	usage     types.Usage
}

func (m appModel) submitQuery(input string) tea.Cmd {
	return func() tea.Msg {
		var textBuffer strings.Builder
		var toolCalls []string
		var lastUsage types.Usage

		ctx := context.Background()
		err := m.bootstrap.QueryEngine.SubmitMessage(ctx, input, func(ev engine.Event) {
			switch e := ev.(type) {
			case engine.TextEvent:
				textBuffer.WriteString(e.Token)
			case engine.ToolCallEvent:
				toolCalls = append(toolCalls, e.Name)
			case engine.UsageEvent:
				lastUsage = e.Usage
			}
		})

		return engineResultMsg{
			text:      textBuffer.String(),
			toolCalls: toolCalls,
			err:       err,
			usage:     lastUsage,
		}
	}
}

func (m appModel) handleEngineResult(msg engineResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "error",
			Content: fmt.Sprintf("Error: %v", msg.err),
			Theme:   m.theme,
		})
		m.messageList.ScrollToBottom()
		return m, nil
	}

	if msg.text != "" {
		m.messages = append(m.messages, types.Message{
			Role: types.RoleAssistant,
			Content: []types.ContentBlock{
				{Type: types.ContentTypeText, Text: msg.text},
			},
		})

		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "assistant",
			Content: msg.text,
			Theme:   m.theme,
		})
	}

	if len(msg.toolCalls) > 0 {
		toolMsg := "Tools called: " + strings.Join(msg.toolCalls, ", ")
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "tool_call",
			Content: toolMsg,
			Theme:   m.theme,
		})
	}

	m.usage = msg.usage
	m.messageList.ScrollToBottom()
	return m, nil
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸"}

func (m appModel) View() tea.View {
	var content strings.Builder

	msgHeight := m.height - 3
	if msgHeight > 0 {
		m.messageList.Height = msgHeight
		content.WriteString(m.messageList.Render(m.width))
		content.WriteString("\n")
	}

	inputLine := m.theme.InputPrompt.Render("> ") + m.input.RenderInline()
	content.WriteString(inputLine)
	content.WriteString("\n")
	content.WriteString(m.statusBar.Render(m.width))

	return tea.NewView(content.String())
}
