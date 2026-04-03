package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Command returns the /init command
func Command() types.Command {
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
		return fmt.Errorf("CLAUDE.md already exists at %s", claudePath)
	}

	template := generateTemplate(ctx)

	if err := os.WriteFile(claudePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create CLAUDE.md: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  Initialize CLAUDE.md\n")
	sb.WriteString("  ═══════════════════════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("  Created CLAUDE.md at %s\n\n", claudePath))
	sb.WriteString("  This file contains project-specific instructions\n")
	sb.WriteString("  that guide the assistant's behavior.\n\n")
	sb.WriteString("  Edit it to customize:\n")
	sb.WriteString("  • Code style and conventions\n")
	sb.WriteString("  • Project structure and architecture\n")
	sb.WriteString("  • Testing requirements\n")
	sb.WriteString("  • Tool preferences\n")
	sb.WriteString("  • Any other project-specific guidance\n")
	sb.WriteString("\n")

	fmt.Println(sb.String())
	return nil
}

func generateTemplate(ctx *types.CommandContext) string {
	dirName := filepath.Base(ctx.Cwd)

	return fmt.Sprintf(`# %s

## Project Overview
<!-- Describe what this project does and its purpose -->

## Architecture
<!-- High-level architecture overview -->

## Code Style
<!-- Coding conventions, naming patterns, etc. -->

## Testing
<!-- Testing requirements and conventions -->

## Development Workflow
<!-- How to build, test, and run the project -->

## Important Notes
<!-- Any other important context for the assistant -->
`, dirName)
}
