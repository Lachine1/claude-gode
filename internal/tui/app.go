package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Lachine1/claude-gode/internal/bootstrap"
	"github.com/Lachine1/claude-gode/internal/commands"
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

type appState int

const (
	stateWelcome appState = iota
	stateChat
	statePermissionDialog
)

type appModel struct {
	state          appState
	theme          styles.Theme
	bootstrap      *bootstrap.State
	welcome        *components.WelcomeScreen
	messageList    *components.MessageList
	input          *components.Input
	statusBar      *components.StatusBar
	permission     *components.PermissionDialog
	messages       []types.Message
	usage          types.Usage
	cost           float64
	permissionMode string
	width          int
	height         int
}

func stylesDefaultTheme() styles.Theme {
	return styles.DefaultTheme()
}

func newAppModel(state *bootstrap.State, args []string, theme styles.Theme) appModel {
	welcome := components.NewWelcomeScreen(theme)
	input := components.NewInput(theme)
	statusBar := components.NewStatusBar(theme)

	permMode := "default"
	if state.Config.PermissionMode != "" {
		permMode = state.Config.PermissionMode
	}

	statusBar.Update(state.Config.Model, 0, 0, 0, 0, 0.0, permMode, "")

	return appModel{
		state:          stateWelcome,
		theme:          theme,
		bootstrap:      state,
		welcome:        welcome,
		messageList:    &components.MessageList{Theme: theme, Height: 20},
		input:          input,
		statusBar:      statusBar,
		messages:       make([]types.Message, 0),
		permissionMode: permMode,
		width:          80,
		height:         24,
	}
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 4 {
			m.messageList.Height = m.height - 4
		}
		return m, nil

	case tea.KeyMsg:
		if m.state == stateWelcome {
			m.welcome.Dismissed = true
			m.state = stateChat
			return m, nil
		}

		if m.state == statePermissionDialog && m.permission != nil {
			if m.permission.HandleKey(msg.String()) {
				m.state = stateChat
				m.permission = nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
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
		case "esc":
			if m.state == statePermissionDialog {
				m.permission.HandleKey("esc")
				m.state = stateChat
				m.permission = nil
			}
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

	ctx := &types.CommandContext{
		Cwd:         m.bootstrap.Cwd,
		GetMessages: func() []types.Message { return m.messages },
		SetMessages: func(msgs []types.Message) { m.messages = msgs },
		Debug:       false,
	}

	if err := cmd.Handler(ctx, args); err != nil {
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "error",
			Content: fmt.Sprintf("Error: %v", err),
			Theme:   m.theme,
		})
		m.messageList.ScrollToBottom()
	}

	return m, nil
}

func (m appModel) handleUserInput(input string) (tea.Model, tea.Cmd) {
	m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
		Type:    "user",
		Content: input,
		Theme:   m.theme,
	})

	m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
		Type:    "assistant",
		Content: "Processing...",
		Theme:   m.theme,
	})
	m.messageList.ScrollToBottom()

	return m, tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
		return nil
	})
}

func (m appModel) View() tea.View {
	if m.state == stateWelcome {
		return tea.NewView(m.welcome.Render(m.width))
	}

	var content strings.Builder

	msgHeight := m.height - 4
	if msgHeight > 0 {
		m.messageList.Height = msgHeight
		content.WriteString(m.messageList.Render(m.width))
		content.WriteString("\n")
	}

	if m.state == statePermissionDialog && m.permission != nil {
		content.WriteString("\n")
		content.WriteString(m.permission.Render(m.width))
		content.WriteString("\n")
	}

	content.WriteString(m.input.Render(m.width))
	content.WriteString("\n")
	content.WriteString(m.statusBar.Render(m.width))

	return tea.NewView(content.String())
}
