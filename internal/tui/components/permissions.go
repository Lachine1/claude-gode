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
		"Yes",
		"Yes, and don't ask again",
		"Yes, and allow for session",
		"No",
	}
}

func (p *PermissionDialog) Select(idx int) {
	opts := p.options()
	if idx >= 0 && idx < len(opts) {
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
		p.Callback(PermissionAllowAlways)
	case 3:
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
	case "4":
		p.Selected = 3
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
		p.Selected = 3
		p.Confirm()
		return true
	}
	return false
}

// Render matches PermissionPrompt.tsx:
// ╭─ Permission Title ─────────────────────── (top border only)
//
//	Tool: <name>
//	Action: <action>
//
//	❯ 1. Yes
//	  2. Yes, and don't ask again
//	  3. Yes, allow for session
//	  4. No
//
//	Esc to cancel · Tab to amend
func (p *PermissionDialog) Render(width int) string {
	opts := p.options()

	title := p.Theme.PermissionTitle.Render("Tool Request: " + p.ToolName)
	subtitle := p.Theme.PermissionSub.Render(p.Action)

	// Calculate max index width for alignment
	maxIdxWidth := len(fmt.Sprintf("%d", len(opts)))

	var optLines []string
	for i, opt := range opts {
		idxStr := fmt.Sprintf("%*d.", maxIdxWidth, i+1)
		var line string
		if i == p.Selected {
			// ❯ prefix for focused option
			pointer := p.Theme.PermOptionFocus.Render("❯")
			label := p.Theme.PermOptionFocus.Render(opt)
			line = fmt.Sprintf("  %s %s %s", pointer, p.Theme.PermOptionIdx.Render(idxStr), label)
		} else {
			// Space prefix for non-focused
			var arrow string
			if i == 0 {
				arrow = "↓"
			} else if i == len(opts)-1 {
				arrow = "↑"
			} else {
				arrow = " "
			}
			line = fmt.Sprintf("    %s %s %s", p.Theme.PermOptionBlur.Render(arrow), p.Theme.PermOptionIdx.Render(idxStr), opt)
		}
		optLines = append(optLines, line)
	}

	cancel := p.Theme.PermCancel.Render("Esc to cancel · Tab to amend")

	content := title + "\n" + subtitle + "\n\n" + strings.Join(optLines, "\n") + "\n\n" + cancel

	return p.Theme.PermissionBorder.Width(width).Render(content)
}
