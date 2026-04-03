package help

import (
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
				return showCommandHelp(ctx, args[0])
			}
			return showGeneralHelp(ctx)
		},
	}
}

func showGeneralHelp(ctx *types.CommandContext) error {
	w := ctx.WriteOutput
	w("")
	w("  Available Slash Commands")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Conversation")
	w("    /help, /?, /h              Show this help")
	w("    /clear                     Clear conversation history")
	w("    /compact                   Trigger context compaction")
	w("    /continue                  Continue last session")
	w("")
	w("  Configuration")
	w("    /settings                  View/edit settings")
	w("    /config                    Show current config")
	w("    /models                    List and switch models")
	w("    /permission-mode           Change permission mode")
	w("    /memory                    View/edit memory (MEMORY.md)")
	w("")
	w("  Information")
	w("    /commands                  List available commands")
	w("    /tools                     List available tools")
	w("    /skills                    List available skills")
	w("    /agents                    List/manage agents")
	w("    /mcp                       MCP server management")
	w("    /usage                     Show token usage and cost")
	w("    /version                   Show version")
	w("")
	w("  Workflow")
	w("    /plan                      Enter plan mode")
	w("    /review                    Review changes")
	w("    /commit                    Commit changes with git")
	w("    /init                      Initialize CLAUDE.md")
	w("")
	w("  Type /help <command> for more information on a specific command.")
	w("")
	return nil
}

func showCommandHelp(ctx *types.CommandContext, name string) error {
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
		return &types.CommandError{Message: "unknown command: /" + name}
	}

	w := ctx.WriteOutput
	w("")
	w("  /" + name)
	w("  ═══════════════════════════════════════")
	w("")
	w("  Description: " + info.desc)
	w("  Usage:       " + info.usage)
	w("")
	return nil
}
