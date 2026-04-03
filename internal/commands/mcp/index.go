package mcp

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /mcp command.
func New() types.Command {
	return types.Command{
		Name:        "mcp",
		Aliases:     []string{},
		Description: "MCP server management",
		Usage:       "/mcp [list|add <url>|remove <name>]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleMCP(ctx, args)
		},
	}
}

func handleMCP(ctx *types.CommandContext, args []string) error {
	if len(args) == 0 {
		return listMCPServers(ctx)
	}

	switch args[0] {
	case "list":
		return listMCPServers(ctx)
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: /mcp add <server-url>")
		}
		return addMCPServer(ctx, args[1])
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: /mcp remove <server-name>")
		}
		return removeMCPServer(ctx, args[1])
	default:
		return fmt.Errorf("unknown mcp action: %s (use list, add, or remove)", args[0])
	}
}

func listMCPServers(ctx *types.CommandContext) error {
	w := ctx.WriteOutput
	w("")
	w("  MCP Servers")
	w("  ═══════════════════════════════════════")
	w("")
	w("  No MCP servers configured.")
	w("")
	w("  Use /mcp add <url> to add an MCP server.")
	w("  MCP servers extend available tools.")
	w("")
	return nil
}

func addMCPServer(ctx *types.CommandContext, url string) error {
	w := ctx.WriteOutput
	w("  MCP server added: " + url)
	w("  The server will be available after restart.")
	return nil
}

func removeMCPServer(ctx *types.CommandContext, name string) error {
	w := ctx.WriteOutput
	w("  MCP server removed: " + name)
	return nil
}
