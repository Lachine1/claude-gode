package config

// DefaultSettings returns Settings with sensible defaults
func DefaultSettings() *Settings {
	return &Settings{
		Model:                   "claude-sonnet-4-20250514",
		PermissionMode:          "default",
		RespectGitignore:        true,
		ShowWelcomeBanner:       true,
		SpinnerTipsEnabled:      true,
		PromptSuggestionEnabled: true,
		Permissions: Permissions{
			DefaultMode: "default",
		},
		Env:   make(map[string]string),
		Hooks: make(map[string]Hook),
	}
}

// DefaultGlobalConfig returns GlobalConfig with sensible defaults
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Theme:                      "default",
		RespectGitignore:           true,
		PermissionExplainerEnabled: true,
		TodoFeatureEnabled:         true,
		ShowExpandedTodos:          true,
		AutoConnectIde:             true,
	}
}
