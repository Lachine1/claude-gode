package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /memory command.
func New() types.Command {
	return types.Command{
		Name:        "memory",
		Aliases:     []string{},
		Description: "View/edit memory (MEMORY.md)",
		Usage:       "/memory [show|edit|clear]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			action := "show"
			if len(args) > 0 {
				action = args[0]
			}
			return handleMemory(ctx, action)
		},
	}
}

func handleMemory(ctx *types.CommandContext, action string) error {
	memoryPath := filepath.Join(ctx.Cwd, "MEMORY.md")

	switch action {
	case "show":
		return showMemory(ctx, memoryPath)
	case "edit":
		return editMemory(ctx, memoryPath)
	case "clear":
		return clearMemory(ctx, memoryPath)
	default:
		return fmt.Errorf("unknown memory action: %s (use show, edit, or clear)", action)
	}
}

func showMemory(ctx *types.CommandContext, path string) error {
	w := ctx.WriteOutput
	w("")
	w("  Memory (MEMORY.md)")
	w("  ═══════════════════════════════════════")
	w("")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			w("  No MEMORY.md found. Create one with /memory edit")
		} else {
			w("  Error reading memory: " + err.Error())
		}
		w("")
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	maxLines := 50
	for i, line := range lines {
		if i < maxLines {
			w("  " + line)
		} else {
			w(fmt.Sprintf("  ... (%d more lines)", len(lines)-i))
			break
		}
	}
	w("")
	return nil
}

func editMemory(ctx *types.CommandContext, path string) error {
	w := ctx.WriteOutput
	if _, err := os.Stat(path); os.IsNotExist(err) {
		template := `# Project Memory

## Key Decisions
- 

## Important Context
- 

## Active Work
- 
`
		if err := os.WriteFile(path, []byte(template), 0644); err != nil {
			return fmt.Errorf("failed to create MEMORY.md: %w", err)
		}
		w("  Created new MEMORY.md at " + path)
		w("  Edit the file to add project-specific memory and context.")
		return nil
	}

	w("  MEMORY.md already exists at " + path)
	w("  Edit the file directly to update memory.")
	return nil
}

func clearMemory(ctx *types.CommandContext, path string) error {
	w := ctx.WriteOutput
	if _, err := os.Stat(path); os.IsNotExist(err) {
		w("  No MEMORY.md found. Nothing to clear.")
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to clear memory: %w", err)
	}

	w("  MEMORY.md has been deleted.")
	return nil
}
