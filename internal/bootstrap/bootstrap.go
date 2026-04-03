package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/internal/commands"
	"github.com/Lachine1/claude-gode/internal/engine"
	"github.com/Lachine1/claude-gode/internal/services/auth"
	"github.com/Lachine1/claude-gode/internal/services/config"
	toolspkg "github.com/Lachine1/claude-gode/internal/tools"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// State holds the initialized application state
type State struct {
	Config      *config.Config
	Auth        *auth.AuthState
	Cwd         string
	IsGit       bool
	GitRoot     string
	Tools       []types.Tool
	Commands    []types.Command
	QueryEngine *engine.QueryEngine
}

// Initialize performs the bootstrap sequence
func Initialize() (*State, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return nil, err
	}

	authState, err := auth.Initialize(cfg)
	if err != nil {
		return nil, err
	}

	isGit, gitRoot := detectGitRoot(cwd)

	tools := registerTools()

	queryEngine := engine.NewQueryEngine(engine.EngineConfig{
		Cwd:          cwd,
		Tools:        tools,
		Model:        cfg.Model,
		MaxTokens:    cfg.MaxTokens,
		MaxBudgetUSD: 0,
		CustomPrompt: "",
		AppendPrompt: "",
		Debug:        false,
		Verbose:      false,
	})

	allCommands := commands.RegisterAll(queryEngine, cfg, tools, isGit, gitRoot)

	return &State{
		Config:      cfg,
		Auth:        authState,
		Cwd:         cwd,
		IsGit:       isGit,
		GitRoot:     gitRoot,
		Tools:       tools,
		Commands:    allCommands,
		QueryEngine: queryEngine,
	}, nil
}

func detectGitRoot(cwd string) (bool, string) {
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false, ""
		}
		dir = parent
	}
}

func registerTools() []types.Tool {
	return toolspkg.RegisterTools()
}

func registerCommands() []types.Command {
	// Commands are registered here
	return nil
}
