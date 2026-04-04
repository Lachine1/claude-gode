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
	Settings    *config.Settings
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

	settings, err := config.LoadSettings(cwd)
	if err != nil {
		return nil, err
	}
	cfg := &config.Config{Settings: settings}

	authState, err := auth.Initialize(cfg)
	if err != nil {
		return nil, err
	}

	isGit, gitRoot := detectGitRoot(cwd)

	tools := registerTools()

	// Get API key: env var takes priority, then settings.Env, then auth state
	apiKey := authState.APIKey
	if settings.APIKey() != "" {
		apiKey = settings.APIKey()
	}

	// Get base URL from settings or env
	baseURL := ""
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		baseURL = v
	} else if v := os.Getenv("BASE_URL"); v != "" {
		baseURL = v
	} else if settings.Raw != nil {
		if v, ok := settings.Raw["anthropic_base_url"]; ok {
			if s, ok := v.(string); ok {
				baseURL = s
			}
		}
	}

	// Get model with correct priority order:
	// 1. CLAUDE_CODE_MODEL env (highest)
	// 2. ANTHROPIC_MODEL env
	// 3. MODEL env
	// 4. ANTHROPIC_DEFAULT_SONNET_MODEL env
	// 5. settings.json "model" field
	// 6. settings.json default_sonnet_model
	// 7. hardcoded default
	model := "claude-sonnet-4-20250514" // fallback

	// Start with settings.json model
	if settings.Model != "" {
		model = settings.Model
	}

	// Override with env vars (highest priority first)
	if v := os.Getenv("CLAUDE_CODE_MODEL"); v != "" {
		model = v
	} else if v := os.Getenv("ANTHROPIC_MODEL"); v != "" {
		model = v
	} else if v := os.Getenv("MODEL"); v != "" {
		model = v
	} else if v := os.Getenv("ANTHROPIC_DEFAULT_SONNET_MODEL"); v != "" {
		model = v
	} else if settings.Raw != nil {
		if v, ok := settings.Raw["default_sonnet_model"]; ok {
			if s, ok := v.(string); ok {
				model = s
			}
		}
	}

	queryEngine := engine.NewQueryEngine(engine.EngineConfig{
		Cwd:          cwd,
		Tools:        tools,
		Model:        model,
		MaxTokens:    settings.MaxTokens(),
		MaxBudgetUSD: 0,
		CustomPrompt: "",
		AppendPrompt: "",
		Debug:        false,
		Verbose:      false,
		APIKey:       apiKey,
		BaseURL:      baseURL,
	})

	allCommands := commands.RegisterAll(queryEngine, cfg, tools, isGit, gitRoot)

	return &State{
		Config:      cfg,
		Settings:    settings,
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
