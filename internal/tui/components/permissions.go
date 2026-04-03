package components

import (
	"fmt"
	"strings"

	"github.com/Lachine1/claude-gode/internal/tui/styles"
)

type PermissionResult string

const (
	PermissionAllowOnce   PermissionResult = "allow_once"
	PermissionAllowAlways PermissionResult = "allow_always"
	PermissionDeny        PermissionResult = "deny"
)

type PermissionDialog struct {
	ToolName string
	Action   string
	Selected int
	Theme    styles.Theme
	Callback func(PermissionResult)
}

func NewPermissionDialog(theme styles.Theme, toolName, action string, callback func(PermissionResult)) *PermissionDialog {
	return &PermissionDialog{
		ToolName: toolName,
		Action:   action,
		Selected: 0,
		Theme:    theme,
		Callback: callback,
	}
}

func (p *PermissionDialog) options() []string {
	return []string{
		"Allow once",
		"Allow always",
		"Deny",
	}
}

func (p *PermissionDialog) Select(idx int) {
	if idx >= 0 && idx < len(p.options()) {
		p.Selected = idx
	}
}

func (p *PermissionDialog) MoveUp() {
	if p.Selected > 0 {
		p.Selected--
	}
}

func (p *PermissionDialog) MoveDown() {
	opts := p.options()
	if p.Selected < len(opts)-1 {
		p.Selected++
	}
}

func (p *PermissionDialog) Confirm() {
	if p.Callback == nil {
		return
	}
	switch p.Selected {
	case 0:
		p.Callback(PermissionAllowOnce)
	case 1:
		p.Callback(PermissionAllowAlways)
	case 2:
		p.Callback(PermissionDeny)
	}
}

func (p *PermissionDialog) HandleKey(key string) bool {
	switch key {
	case "1":
		p.Selected = 0
		p.Confirm()
		return true
	case "2":
		p.Selected = 1
		p.Confirm()
		return true
	case "3":
		p.Selected = 2
		p.Confirm()
		return true
	case "up", "k":
		p.MoveUp()
	case "down", "j":
		p.MoveDown()
	case "enter", " ":
		p.Confirm()
		return true
	case "esc":
		p.Selected = 2
		p.Confirm()
		return true
	}
	return false
}

func (p *PermissionDialog) Render(width int) string {
	opts := p.options()

	title := p.Theme.PermissionTitle.Render(fmt.Sprintf("Permission Required"))
	toolInfo := fmt.Sprintf("Tool: %s", p.Theme.PermissionSelected.Render(p.ToolName))
	actionInfo := fmt.Sprintf("Action: %s", p.Theme.PermissionSelected.Render(p.Action))

	var optLines []string
	for i, opt := range opts {
		prefix := "  "
		if i == p.Selected {
			prefix = "▸ "
		}
		num := fmt.Sprintf("[%d]", i+1)
		line := fmt.Sprintf("%s %s %s", prefix, num, opt)
		if i == p.Selected {
			optLines = append(optLines, p.Theme.PermissionSelected.Render(line))
		} else {
			optLines = append(optLines, p.Theme.PermissionOption.Render(line))
		}
	}

	hint := p.Theme.PermissionOption.Render("Press 1/2/3 or arrows to select, Enter to confirm, Esc to deny")

	content := title + "\n" + toolInfo + "\n" + actionInfo + "\n\n" + strings.Join(optLines, "\n") + "\n\n" + hint

	return p.Theme.PermissionDialog.Render(content)
}
