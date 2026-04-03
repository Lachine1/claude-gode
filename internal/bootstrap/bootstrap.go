package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/Lachine1/claude-gode/internal/services/auth"
	"github.com/Lachine1/claude-gode/internal/services/config"
	"github.com/Lachine1/claude-gode/pkg/types"
)

// State holds the initialized application state
type State struct {
	Config   *config.Config
	Auth     *auth.AuthState
	Cwd      string
	IsGit    bool
	GitRoot  string
	Tools    []types.Tool
	Commands []types.Command
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

	return &State{
		Config:   cfg,
		Auth:     authState,
		Cwd:      cwd,
		IsGit:    isGit,
		GitRoot:  gitRoot,
		Tools:    registerTools(),
		Commands: registerCommands(),
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
	// Tools are registered here
	return nil
}

func registerCommands() []types.Command {
	// Commands are registered here
	return nil
}
