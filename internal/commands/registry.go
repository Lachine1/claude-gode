package commands

import (
	"github.com/Lachine1/claude-gode/internal/commands/agents"
	"github.com/Lachine1/claude-gode/internal/commands/clear"
	cmdlist "github.com/Lachine1/claude-gode/internal/commands/commands"
	"github.com/Lachine1/claude-gode/internal/commands/commit"
	"github.com/Lachine1/claude-gode/internal/commands/compact"
	"github.com/Lachine1/claude-gode/internal/commands/configcmd"
	"github.com/Lachine1/claude-gode/internal/commands/continuecmd"
	"github.com/Lachine1/claude-gode/internal/commands/help"
	"github.com/Lachine1/claude-gode/internal/commands/initcmd"
	"github.com/Lachine1/claude-gode/internal/commands/mcp"
	"github.com/Lachine1/claude-gode/internal/commands/memory"
	"github.com/Lachine1/claude-gode/internal/commands/models"
	"github.com/Lachine1/claude-gode/internal/commands/permissionmode"
	"github.com/Lachine1/claude-gode/internal/commands/plan"
	"github.com/Lachine1/claude-gode/internal/commands/review"
	"github.com/Lachine1/claude-gode/internal/commands/settings"
	"github.com/Lachine1/claude-gode/internal/commands/skills"
	toolscmd "github.com/Lachine1/claude-gode/internal/commands/toolscmd"
	"github.com/Lachine1/claude-gode/internal/commands/usage"
	"github.com/Lachine1/claude-gode/internal/commands/version"
	"github.com/Lachine1/claude-gode/internal/engine"
	svcconfig "github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// RegisterAll creates and returns all registered slash commands.
func RegisterAll(eng *engine.QueryEngine, cfg *svcconfig.Config, toolList []types.Tool, isGit bool, gitRoot string) []types.Command {
	allCmds := []types.Command{
		help.New(),
		version.New(),
		clear.New(),
		usage.New(eng),
		compact.New(eng),
		models.New(cfg),
		settings.New(cfg),
		configcmd.New(cfg),
		memory.New(),
		plan.New(),
		review.New(isGit, gitRoot),
		commit.New(isGit, gitRoot),
		skills.New(),
		agents.New(),
		mcp.New(),
		permissionmode.New(cfg),
		continuecmd.New(),
		initcmd.New(),
	}

	listCmd := cmdlist.New(allCmds)
	toolsCmd := toolscmd.New(toolList)

	allCmds = append(allCmds, listCmd, toolsCmd)
	return allCmds
}

// FindCommand looks up a command by name or alias.
func FindCommand(name string, cmds []types.Command) *types.Command {
	for i := range cmds {
		if cmds[i].Name == name {
			return &cmds[i]
		}
		for _, alias := range cmds[i].Aliases {
			if alias == name {
				return &cmds[i]
			}
		}
	}
	return nil
}
