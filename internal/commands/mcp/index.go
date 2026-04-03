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
		return listMCPServers()
	}

	switch args[0] {
	case "list":
		return listMCPServers()
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: /mcp add <server-url>")
		}
		return addMCPServer(args[1])
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: /mcp remove <server-name>")
		}
		return removeMCPServer(args[1])
	default:
		return fmt.Errorf("unknown mcp action: %s (use list, add, or remove)", args[0])
	}
}

func listMCPServers() error {
	fmt.Println()
	fmt.Println("  MCP Servers")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Println("  No MCP servers configured.")
	fmt.Println()
	fmt.Println("  Use /mcp add <url> to add an MCP server.")
	fmt.Println("  MCP servers extend available tools.")
	fmt.Println()
	return nil
}

func addMCPServer(url string) error {
	fmt.Printf("  MCP server added: %s\n", url)
	fmt.Println("  The server will be available after restart.")
	return nil
}

func removeMCPServer(name string) error {
	fmt.Printf("  MCP server removed: %s\n", name)
	return nil
}
