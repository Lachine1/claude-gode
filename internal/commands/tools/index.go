package tools

import (
	"fmt"
	"strings"

	"github.com/Lachine1/claude-gode/internal/tools"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// Command returns the /tools command
func Command() types.Command {
	return types.Command{
		Name:        "tools",
		Aliases:     []string{},
		Description: "List available tools",
		Usage:       "/tools [tool-name]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleTools(ctx, args)
		},
	}
}

func handleTools(ctx *types.CommandContext, args []string) error {
	toolList := tools.RegisterTools()

	if len(args) > 0 {
		return showToolDetail(args[0], toolList)
	}
	return listTools(toolList)
}

func listTools(toolList []types.Tool) error {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  Available Tools\n")
	sb.WriteString("  ═══════════════════════════════════════\n\n")

	for _, tool := range toolList {
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", tool.Name(), truncateDescription(tool.Description(), 60)))
	}

	sb.WriteString(fmt.Sprintf("\n  Total: %d tools\n", len(toolList)))
	sb.WriteString("\n")
	sb.WriteString("  Use /tools <tool-name> for detailed information.\n")
	sb.WriteString("\n")

	fmt.Println(sb.String())
	return nil
}

func showToolDetail(name string, toolList []types.Tool) error {
	for _, tool := range toolList {
		if tool.Name() == name {
			var sb strings.Builder
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  Tool: %s\n", tool.Name()))
			sb.WriteString("  ═══════════════════════════════════════\n\n")
			sb.WriteString(fmt.Sprintf("  Description:\n  %s\n\n", tool.Description()))

			schema := tool.JSONSchema()
			if props, ok := schema["properties"].(map[string]interface{}); ok {
				sb.WriteString("  Parameters:\n")
				for propName, propDef := range props {
					if propMap, ok := propDef.(map[string]interface{}); ok {
						propType := propMap["type"]
						propDesc := propMap["description"]
						sb.WriteString(fmt.Sprintf("    %-15s %s\n", propName, propType))
						if propDesc != nil {
							sb.WriteString(fmt.Sprintf("                  %s\n", propDesc))
						}
					}
				}
			}

			sb.WriteString("\n")
			fmt.Println(sb.String())
			return nil
		}
	}

	return fmt.Errorf("unknown tool: %s", name)
}

func truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen-3] + "..."
}
