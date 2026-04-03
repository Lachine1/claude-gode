package help

import (
	"fmt"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /help command.
func New() types.Command {
	return types.Command{
		Name:        "help",
		Aliases:     []string{"h", "?"},
		Description: "Show help for all slash commands",
		Usage:       "/help [command]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			if len(args) > 0 {
				return showCommandHelp(args[0])
			}
			return showGeneralHelp()
		},
	}
}

func showGeneralHelp() error {
	fmt.Println()
	fmt.Println("  Available Slash Commands")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Println("  Conversation")
	fmt.Println("    /help, /?, /h              Show this help")
	fmt.Println("    /clear                     Clear conversation history")
	fmt.Println("    /compact                   Trigger context compaction")
	fmt.Println("    /continue                  Continue last session")
	fmt.Println()
	fmt.Println("  Configuration")
	fmt.Println("    /settings                  View/edit settings")
	fmt.Println("    /config                    Show current config")
	fmt.Println("    /models                    List and switch models")
	fmt.Println("    /permission-mode           Change permission mode")
	fmt.Println("    /memory                    View/edit memory (MEMORY.md)")
	fmt.Println()
	fmt.Println("  Information")
	fmt.Println("    /commands                  List available commands")
	fmt.Println("    /tools                     List available tools")
	fmt.Println("    /skills                    List available skills")
	fmt.Println("    /agents                    List/manage agents")
	fmt.Println("    /mcp                       MCP server management")
	fmt.Println("    /usage                     Show token usage and cost")
	fmt.Println("    /version                   Show version")
	fmt.Println()
	fmt.Println("  Workflow")
	fmt.Println("    /plan                      Enter plan mode")
	fmt.Println("    /review                    Review changes")
	fmt.Println("    /commit                    Commit changes with git")
	fmt.Println("    /init                      Initialize CLAUDE.md")
	fmt.Println()
	fmt.Println("  Type /help <command> for more information on a specific command.")
	fmt.Println()
	return nil
}

func showCommandHelp(name string) error {
	builtins := map[string]struct {
		desc  string
		usage string
	}{
		"help":            {"Show help for all slash commands", "/help [command]"},
		"clear":           {"Clear conversation history", "/clear"},
		"compact":         {"Trigger context compaction", "/compact"},
		"continue":        {"Continue last session", "/continue"},
		"settings":        {"View/edit settings", "/settings [key] [value]"},
		"config":          {"Show current config", "/config"},
		"models":          {"List and switch models", "/models [model-name]"},
		"permission-mode": {"Change permission mode", "/permission-mode [mode]"},
		"memory":          {"View/edit memory (MEMORY.md)", "/memory [show|edit|clear]"},
		"commands":        {"List available commands", "/commands"},
		"tools":           {"List available tools", "/tools"},
		"skills":          {"List available skills", "/skills"},
		"agents":          {"List/manage agents", "/agents"},
		"mcp":             {"MCP server management", "/mcp [list|add|remove]"},
		"usage":           {"Show token usage and cost", "/usage"},
		"version":         {"Show version", "/version"},
		"plan":            {"Enter plan mode", "/plan"},
		"review":          {"Review changes", "/review"},
		"commit":          {"Commit changes with git", "/commit [message]"},
		"init":            {"Initialize CLAUDE.md", "/init"},
	}

	info, ok := builtins[name]
	if !ok {
		return fmt.Errorf("unknown command: /%s", name)
	}

	fmt.Println()
	fmt.Printf("  /%s\n", name)
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Description: %s\n", info.desc)
	fmt.Printf("  Usage:       %s\n", info.usage)
	fmt.Println()
	return nil
}
