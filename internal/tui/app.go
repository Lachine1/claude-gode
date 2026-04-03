package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Lachine1/claude-gode/internal/bootstrap"
)

// Run starts the terminal user interface
func Run(ctx context.Context, state *bootstrap.State, args []string) error {
	p := tea.NewProgram(newAppModel(state, args), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

type appModel struct {
	state *bootstrap.State
	args  []string
	input string
}

func newAppModel(state *bootstrap.State, args []string) appModel {
	return appModel{
		state: state,
		args:  args,
	}
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			// Process input
			return m, nil
		default:
			m.input += msg.String()
			return m, nil
		}
	}
	return m, nil
}

func (m appModel) View() tea.View {
	s := "\n  Claude Code (Go)\n\n"
	s += "  Type your message and press Enter to start.\n"
	s += "  Press Ctrl+C or q to quit.\n\n"
	s += "  > " + m.input
	s += "\n"
	return tea.NewView(s)
}
