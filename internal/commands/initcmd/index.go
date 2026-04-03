package initcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /init command.
func New() types.Command {
	return types.Command{
		Name:        "init",
		Aliases:     []string{},
		Description: "Initialize CLAUDE.md",
		Usage:       "/init",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleInit(ctx, args)
		},
	}
}

func handleInit(ctx *types.CommandContext, args []string) error {
	claudePath := filepath.Join(ctx.Cwd, "CLAUDE.md")

	if _, err := os.Stat(claudePath); err == nil {
		fmt.Println()
		fmt.Println("  Initialize CLAUDE.md")
		fmt.Println("  ═══════════════════════════════════════")
		fmt.Println()
		fmt.Printf("  CLAUDE.md already exists at %s\n", claudePath)
		fmt.Println("  Edit it directly to update project instructions.")
		fmt.Println()
		return nil
	}

	template := `# Project Guidelines

## Project Overview
- 

## Architecture
- 

## Coding Standards
- 

## Testing
- 

## Important Notes
- 
`

	if err := os.WriteFile(claudePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create CLAUDE.md: %w", err)
	}

	fmt.Println()
	fmt.Println("  Initialize CLAUDE.md")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Created CLAUDE.md at %s\n", claudePath)
	fmt.Println("  Edit the file to add project-specific instructions.")
	fmt.Println("  The AI will follow these guidelines when working on this project.")
	fmt.Println()
	return nil
}
