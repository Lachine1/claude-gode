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
	w := ctx.WriteOutput

	if _, err := os.Stat(claudePath); err == nil {
		w("")
		w("  Initialize CLAUDE.md")
		w("  ═══════════════════════════════════════")
		w("")
		w("  CLAUDE.md already exists at " + claudePath)
		w("  Edit it directly to update project instructions.")
		w("")
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

	w("")
	w("  Initialize CLAUDE.md")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Created CLAUDE.md at " + claudePath)
	w("  Edit the file to add project-specific instructions.")
	w("  The AI will follow these guidelines when working on this project.")
	w("")
	return nil
}
