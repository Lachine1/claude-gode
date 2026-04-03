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
		return showMemory(memoryPath)
	case "edit":
		return editMemory(memoryPath)
	case "clear":
		return clearMemory(memoryPath)
	default:
		return fmt.Errorf("unknown memory action: %s (use show, edit, or clear)", action)
	}
}

func showMemory(path string) error {
	fmt.Println()
	fmt.Println("  Memory (MEMORY.md)")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("  No MEMORY.md found. Create one with /memory edit")
		} else {
			fmt.Printf("  Error reading memory: %v\n", err)
		}
		fmt.Println()
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	maxLines := 50
	for i, line := range lines {
		if i < maxLines {
			fmt.Printf("  %s\n", line)
		} else {
			fmt.Printf("  ... (%d more lines)\n", len(lines)-i)
			break
		}
	}
	fmt.Println()
	return nil
}

func editMemory(path string) error {
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
		fmt.Printf("  Created new MEMORY.md at %s\n", path)
		fmt.Println("  Edit the file to add project-specific memory and context.")
		return nil
	}

	fmt.Printf("  MEMORY.md already exists at %s\n", path)
	fmt.Println("  Edit the file directly to update memory.")
	return nil
}

func clearMemory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("  No MEMORY.md found. Nothing to clear.")
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to clear memory: %w", err)
	}

	fmt.Println("  MEMORY.md has been deleted.")
	return nil
}
