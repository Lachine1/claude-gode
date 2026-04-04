package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Lachine1/claude-gode/internal/bootstrap"
	"github.com/Lachine1/claude-gode/internal/commands"
	"github.com/Lachine1/claude-gode/internal/completion"
	"github.com/Lachine1/claude-gode/internal/engine"
	"github.com/Lachine1/claude-gode/internal/tui/components"
	"github.com/Lachine1/claude-gode/internal/tui/styles"
	"github.com/Lachine1/claude-gode/pkg/types"
)

func Run(ctx context.Context, state *bootstrap.State, args []string) error {
	theme := styles.DefaultTheme()
	model := newAppModel(state, args, theme)
	p := tea.NewProgram(model, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

type appModel struct {
	theme            styles.Theme
	bootstrap        *bootstrap.State
	messageList      *components.MessageList
	prompt           *components.PromptInput
	spinner          *components.Spinner
	messages         []types.Message
	usage            types.Usage
	cost             float64
	permissionMode   string
	gitBranch        string
	width            int
	height           int
	streaming        bool
	streamingText    strings.Builder
	streamingTok     int
	completionEngine *completion.CompletionEngine
}

func newAppModel(state *bootstrap.State, args []string, theme styles.Theme) appModel {
	prompt := components.NewPromptInput(theme)
	prompt.Focused = true
	prompt.Model = state.Config.Model()
	prompt.PermMode = state.Config.PermissionMode()

	spinner := components.NewSpinner(theme)

	var cmdInfos []completion.CommandInfo
	for _, cmd := range state.Commands {
		cmdInfos = append(cmdInfos, completion.CommandInfo{
			Name:        cmd.Name,
			Aliases:     cmd.Aliases,
			Description: cmd.Description,
		})
	}

	permMode := "default"
	if state.Config.PermissionMode() != "" {
		permMode = state.Config.PermissionMode()
	}

	m := appModel{
		theme:            theme,
		bootstrap:        state,
		messageList:      &components.MessageList{Theme: theme, Height: 20},
		prompt:           prompt,
		spinner:          spinner,
		messages:         make([]types.Message, 0),
		permissionMode:   permMode,
		width:            80,
		height:           24,
		completionEngine: completion.NewEngine(cmdInfos),
	}

	if len(args) > 0 {
		input := strings.Join(args, " ")
		m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
			Type:    "user",
			Content: input,
			Theme:   theme,
		})
	}

	return m
}

func (m appModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

type spinnerTickMsg struct{}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 5 {
			m.messageList.Height = m.height - 5
		}
		return m, nil

	case spinnerTickMsg:
		if m.streaming {
			m.spinner.Tick()
		}
		return m, tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg {
			return spinnerTickMsg{}
		})

	case engineResultMsg:
		return m.handleEngineResult(msg)

	case engineTokenMsg:
		m.messageList.AppendToAssistant(msg.Token)
		m.streamingTok++
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.prompt.Buffer == "" {
				return m, tea.Quit
			}
		case "tab":
			if m.prompt.ShowSuggestions && len(m.prompt.Suggestions) > 0 {
				m.prompt.AcceptSuggestion()
			} else if m.prompt.GhostText != nil {
				m.prompt.AcceptGhostText()
			}
			return m, nil
		case "enter":
			return m.processInput()
		case "up":
			if m.prompt.ShowSuggestions {
				m.prompt.PrevSuggestion()
			} else {
				m.prompt.HistoryUp()
			}
			return m, nil
		case "down":
			if m.prompt.ShowSuggestions {
				m.prompt.NextSuggestion()
			} else {
				m.prompt.HistoryDown()
			}
			return m, nil
		case "pgup":
			m.messageList.PageUp()
			return m, nil
		case "pgdown":
			m.messageList.PageDown(9999)
			return m, nil
		case "left":
			m.prompt.MoveLeft()
			return m, nil
		case "right":
			if m.prompt.GhostText != nil {
				m.prompt.AcceptGhostText()
			} else {
				m.prompt.MoveRight()
			}
			return m, nil
		case "home":
			m.prompt.MoveHome()
			return m, nil
		case "end":
			m.prompt.MoveEnd()
			return m, nil
		case "backspace":
			m.prompt.Backspace()
			m.refreshCompletions()
			return m, nil
		case "delete":
			m.prompt.Delete()
			m.refreshCompletions()
			return m, nil
		case "esc":
			if m.prompt.ShowSuggestions {
				m.prompt.DismissSuggestions()
				return m, nil
			}
		case "ctrl+n":
			if m.prompt.ShowSuggestions {
				m.prompt.NextSuggestion()
				return m, nil
			}
		case "ctrl+p":
			if m.prompt.ShowSuggestions {
				m.prompt.PrevSuggestion()
				return m, nil
			}
		default:
			if len(msg.String()) == 1 {
				m.prompt.Insert(rune(msg.String()[0]))
				m.refreshCompletions()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *appModel) refreshCompletions() {
	if m.completionEngine == nil {
		return
	}
	suggestions := m.completionEngine.GetSuggestions(m.prompt.Buffer, m.bootstrap.Cwd)
	m.prompt.UpdateSuggestions(suggestions)
	m.prompt.GhostText = m.completionEngine.GetGhostText(m.prompt.Buffer, m.bootstrap.Cwd)
}

func (m appModel) processInput() (tea.Model, tea.Cmd) {
	input := m.prompt.Submit()
	if input == "" {
		return m, nil
	}

	m.prompt.DismissSuggestions()
	m.prompt.GhostText = nil

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

	m.completionEngine.RecordCommandUsage(cmdName)

	var output strings.Builder
	ctx := &types.CommandContext{
		Cwd:         m.bootstrap.Cwd,
		GetMessages: func() []types.Message { return m.messages },
		SetMessages: func(msgs []types.Message) { m.messages = msgs },
		Debug:       false,
		WriteOutput: func(s string) {
			output.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				output.WriteString("\n")
			}
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

	m.messageList.Messages = append(m.messageList.Messages, components.DisplayMessage{
		Type:    "assistant",
		Content: "",
		Theme:   m.theme,
	})
	m.messageList.ScrollToBottom()

	m.streaming = true
	m.streamingText.Reset()
	m.streamingTok = 0
	m.spinner.Mode = "responding"
	m.spinner.StartTime = time.Now()
	m.spinner.TokenCount = 0

	return m, m.submitQuery(input)
}

type engineResultMsg struct {
	text  string
	err   error
	usage types.Usage
}

type engineTokenMsg struct {
	Token string
}

func (m appModel) submitQuery(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var textBuffer strings.Builder
		var lastUsage types.Usage

		err := m.bootstrap.QueryEngine.SubmitMessage(ctx, input, func(ev engine.Event) {
			switch e := ev.(type) {
			case engine.TextEvent:
				textBuffer.WriteString(e.Token)
			case engine.UsageEvent:
				lastUsage = e.Usage
			}
		})

		return engineResultMsg{
			text:  textBuffer.String(),
			err:   err,
			usage: lastUsage,
		}
	}
}

func (m appModel) handleEngineResult(msg engineResultMsg) (tea.Model, tea.Cmd) {
	m.streaming = false

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

		for i := len(m.messageList.Messages) - 1; i >= 0; i-- {
			if m.messageList.Messages[i].Type == "assistant" && m.messageList.Messages[i].Content == "" {
				m.messageList.Messages[i].Content = msg.text
				break
			}
		}
	} else {
		if len(m.messageList.Messages) > 0 && m.messageList.Messages[len(m.messageList.Messages)-1].Type == "assistant" {
			m.messageList.Messages = m.messageList.Messages[:len(m.messageList.Messages)-1]
		}
	}

	m.usage = msg.usage
	m.messageList.ScrollToBottom()
	return m, nil
}

func (m appModel) View() tea.View {
	if m.height <= 0 {
		return tea.NewView("")
	}

	var content strings.Builder

	msgHeight := m.height - 4
	if msgHeight > 0 {
		m.messageList.Height = msgHeight
		content.WriteString(m.messageList.Render(m.width))
		content.WriteString("\n")
	}

	if m.streaming {
		m.spinner.TokenCount = m.streamingTok
		content.WriteString(m.spinner.Render(m.width))
		content.WriteString("\n")
	}

	content.WriteString(m.prompt.Render(m.width))

	return tea.NewView(content.String())
}
