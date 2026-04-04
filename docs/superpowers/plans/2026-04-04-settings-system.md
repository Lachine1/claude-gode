# Settings System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current flat config system with a comprehensive typed Settings system that matches Claude Code's settings architecture, including multi-file merge, env file support, global config, and environment variable overrides.

**Architecture:** The Settings system uses a layered merge approach: defaults → user settings → project settings → local settings → env files → environment variables → global config. Each layer is a JSON file or env file that gets unmarshaled and merged in priority order. The typed Settings struct provides compile-time safety while Raw map preserves unknown keys.

**Tech Stack:** Go 1.25, standard library (encoding/json, os, path/filepath, strconv, strings)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/services/config/defaults.go` | Default values for Settings and GlobalConfig |
| `internal/services/config/settings.go` | Settings struct, LoadSettings, Get/Set/Save |
| `internal/services/config/global.go` | GlobalConfig struct, LoadGlobalConfig, SaveGlobalConfig |
| `internal/services/config/env.go` | LoadEnvFile, ApplyEnv utilities |
| `internal/services/config/config.go` | Updated to wrap Settings, maintain backward compatibility |
| `internal/bootstrap/bootstrap.go` | Updated to use Settings system |
| `internal/tui/app.go` | Updated to use Settings instead of Config methods |

---

### Task 1: Create defaults.go

**Files:**
- Create: `internal/services/config/defaults.go`

- [ ] **Step 1: Create defaults.go with DefaultSettings() and DefaultGlobalConfig()**

```go
package config

// DefaultSettings returns Settings with sensible defaults
func DefaultSettings() *Settings {
	return &Settings{
		Model:                "claude-sonnet-4-20250514",
		PermissionMode:       "default",
		RespectGitignore:     true,
		ShowWelcomeBanner:    true,
		SpinnerTipsEnabled:   true,
		PromptSuggestionEnabled: true,
		Permissions: Permissions{
			DefaultMode: "default",
		},
		Env: make(map[string]string),
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
		SpinnerTipsEnabled:         true,
	}
}
```

---

### Task 2: Create settings.go

**Files:**
- Create: `internal/services/config/settings.go`

- [ ] **Step 1: Create the Settings struct and all sub-structs**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// Settings holds all application settings
type Settings struct {
	Model                   string            `json:"model"`
	AvailableModels         []string          `json:"availableModels"`
	PermissionMode          string            `json:"permissionMode"`
	Permissions             Permissions       `json:"permissions"`
	FileSuggestion          FileSuggestion    `json:"fileSuggestion"`
	RespectGitignore        bool              `json:"respectGitignore"`
	Env                     map[string]string `json:"env"`
	EnableAllProjectMcpServers bool           `json:"enableAllProjectMcpServers"`
	EnabledMcpjsonServers   []string          `json:"enabledMcpjsonServers"`
	DisabledMcpjsonServers  []string          `json:"disabledMcpjsonServers"`
	Hooks                   map[string]Hook   `json:"hooks"`
	DisableAllHooks         bool              `json:"disableAllHooks"`
	DefaultShell            string            `json:"defaultShell"`
	Theme                   string            `json:"theme"`
	OutputStyle             string            `json:"outputStyle"`
	Language                string            `json:"language"`
	ShowWelcomeBanner       bool              `json:"showWelcomeBanner"`
	ShowThinkingSummaries   bool              `json:"showThinkingSummaries"`
	PromptSuggestionEnabled bool              `json:"promptSuggestionEnabled"`
	FastMode                bool              `json:"fastMode"`
	AlwaysThinkingEnabled   bool              `json:"alwaysThinkingEnabled"`
	EffortLevel             string            `json:"effortLevel"`
	AutoMemoryEnabled       bool              `json:"autoMemoryEnabled"`
	AutoMemoryDirectory     string            `json:"autoMemoryDirectory"`
	StatusLine              StatusLine        `json:"statusLine"`
	SpinnerTipsEnabled      bool              `json:"spinnerTipsEnabled"`
	Raw                     map[string]interface{} `json:"-"`
	// Internal fields
	userSettingsPath   string
	projectSettingsPath string
	localSettingsPath  string
}

type Permissions struct {
	Allow                        []string `json:"allow"`
	Deny                         []string `json:"deny"`
	Ask                          []string `json:"ask"`
	DefaultMode                  string   `json:"defaultMode"`
	DisableBypassPermissionsMode bool     `json:"disableBypassPermissionsMode"`
	DisableAutoMode              bool     `json:"disableAutoMode"`
	AdditionalDirectories        []string `json:"additionalDirectories"`
}

type FileSuggestion struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type StatusLine struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Padding int    `json:"padding"`
}
```

- [ ] **Step 2: Add LoadSettings function with full merge logic**

```go
// LoadSettings loads and merges all settings files in priority order
func LoadSettings(cwd string) (*Settings, error) {
	s := DefaultSettings()

	// 1. User settings (~/.claude/settings.json)
	userPath := filepath.Join(homeDir(), ".claude", "settings.json")
	if data, err := os.ReadFile(userPath); err == nil {
		mergeSettings(s, data)
	}
	s.userSettingsPath = userPath

	// 2. Project settings (<cwd>/.claude/settings.json)
	projectPath := filepath.Join(cwd, ".claude", "settings.json")
	if data, err := os.ReadFile(projectPath); err == nil {
		mergeSettings(s, data)
	}
	s.projectSettingsPath = projectPath

	// 3. Local project settings (<cwd>/.claude/settings.local.json)
	localPath := filepath.Join(cwd, ".claude", "settings.local.json")
	if data, err := os.ReadFile(localPath); err == nil {
		mergeSettings(s, data)
	}
	s.localSettingsPath = localPath

	// 4. Load env files and apply to Settings.Env
	userEnvPath := filepath.Join(homeDir(), ".claude", ".env")
	if env, err := LoadEnvFile(userEnvPath); err == nil {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		for k, v := range env {
			s.Env[k] = v
		}
	}

	projectEnvPath := filepath.Join(cwd, ".claude", ".env")
	if env, err := LoadEnvFile(projectEnvPath); err == nil {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		for k, v := range env {
			s.Env[k] = v
		}
	}

	// 5. Environment variables override settings
	applyEnvVars(s)

	// 6. Global config (~/.claude.json)
	if gc, err := LoadGlobalConfig(); err == nil {
		if gc.Theme != "" && s.Theme == "" {
			s.Theme = gc.Theme
		}
		if gc.RespectGitignore {
			s.RespectGitignore = gc.RespectGitignore
		}
	}

	return s, nil
}

func mergeSettings(s *Settings, data []byte) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	// Marshal back to JSON and unmarshal into Settings for typed fields
	// This handles nested structs properly
	data, err := json.Marshal(raw)
	if err != nil {
		return
	}

	var partial Settings
	if err := json.Unmarshal(data, &partial); err != nil {
		return
	}

	// Merge typed fields
	if partial.Model != "" {
		s.Model = partial.Model
	}
	if partial.PermissionMode != "" {
		s.PermissionMode = partial.PermissionMode
	}
	if partial.DefaultShell != "" {
		s.DefaultShell = partial.DefaultShell
	}
	if partial.Theme != "" {
		s.Theme = partial.Theme
	}
	if partial.OutputStyle != "" {
		s.OutputStyle = partial.OutputStyle
	}
	if partial.Language != "" {
		s.Language = partial.Language
	}
	if partial.EffortLevel != "" {
		s.EffortLevel = partial.EffortLevel
	}
	if partial.AutoMemoryDirectory != "" {
		s.AutoMemoryDirectory = partial.AutoMemoryDirectory
	}
	if partial.FileSuggestion.Type != "" {
		s.FileSuggestion.Type = partial.FileSuggestion.Type
	}
	if partial.FileSuggestion.Command != "" {
		s.FileSuggestion.Command = partial.FileSuggestion.Command
	}
	if partial.StatusLine.Type != "" {
		s.StatusLine.Type = partial.StatusLine.Type
	}
	if partial.StatusLine.Command != "" {
		s.StatusLine.Command = partial.StatusLine.Command
	}
	if partial.StatusLine.Padding != 0 {
		s.StatusLine.Padding = partial.StatusLine.Padding
	}

	// Merge Permissions
	if partial.Permissions.DefaultMode != "" {
		s.Permissions.DefaultMode = partial.Permissions.DefaultMode
	}
	if partial.Permissions.Allow != nil {
		s.Permissions.Allow = partial.Permissions.Allow
	}
	if partial.Permissions.Deny != nil {
		s.Permissions.Deny = partial.Permissions.Deny
	}
	if partial.Permissions.Ask != nil {
		s.Permissions.Ask = partial.Permissions.Ask
	}
	if partial.Permissions.AdditionalDirectories != nil {
		s.Permissions.AdditionalDirectories = partial.Permissions.AdditionalDirectories
	}
	s.Permissions.DisableBypassPermissionsMode = s.Permissions.DisableBypassPermissionsMode || partial.Permissions.DisableBypassPermissionsMode
	s.Permissions.DisableAutoMode = s.Permissions.DisableAutoMode || partial.Permissions.DisableAutoMode

	// Merge slices (replace, not append)
	if partial.AvailableModels != nil {
		s.AvailableModels = partial.AvailableModels
	}
	if partial.EnabledMcpjsonServers != nil {
		s.EnabledMcpjsonServers = partial.EnabledMcpjsonServers
	}
	if partial.DisabledMcpjsonServers != nil {
		s.DisabledMcpjsonServers = partial.DisabledMcpjsonServers
	}

	// Merge maps
	if partial.Hooks != nil {
		if s.Hooks == nil {
			s.Hooks = make(map[string]Hook)
		}
		for k, v := range partial.Hooks {
			s.Hooks[k] = v
		}
	}
	if partial.Env != nil {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		for k, v := range partial.Env {
			s.Env[k] = v
		}
	}

	// Merge booleans (only override if explicitly set in JSON)
	// We check the raw map for presence
	if v, ok := raw["respectGitignore"]; ok {
		if b, ok := v.(bool); ok {
			s.RespectGitignore = b
		}
	}
	if v, ok := raw["showWelcomeBanner"]; ok {
		if b, ok := v.(bool); ok {
			s.ShowWelcomeBanner = b
		}
	}
	if v, ok := raw["showThinkingSummaries"]; ok {
		if b, ok := v.(bool); ok {
			s.ShowThinkingSummaries = b
		}
	}
	if v, ok := raw["promptSuggestionEnabled"]; ok {
		if b, ok := v.(bool); ok {
			s.PromptSuggestionEnabled = b
		}
	}
	if v, ok := raw["fastMode"]; ok {
		if b, ok := v.(bool); ok {
			s.FastMode = b
		}
	}
	if v, ok := raw["alwaysThinkingEnabled"]; ok {
		if b, ok := v.(bool); ok {
			s.AlwaysThinkingEnabled = b
		}
	}
	if v, ok := raw["autoMemoryEnabled"]; ok {
		if b, ok := v.(bool); ok {
			s.AutoMemoryEnabled = b
		}
	}
	if v, ok := raw["disableAllHooks"]; ok {
		if b, ok := v.(bool); ok {
			s.DisableAllHooks = b
		}
	}
	if v, ok := raw["enableAllProjectMcpServers"]; ok {
		if b, ok := v.(bool); ok {
			s.EnableAllProjectMcpServers = b
		}
	}
	if v, ok := raw["spinnerTipsEnabled"]; ok {
		if b, ok := v.(bool); ok {
			s.SpinnerTipsEnabled = b
		}
	}

	// Store raw values for unknown keys
	if s.Raw == nil {
		s.Raw = make(map[string]interface{})
	}
	knownKeys := map[string]bool{
		"model": true, "availableModels": true, "permissionMode": true,
		"permissions": true, "fileSuggestion": true, "respectGitignore": true,
		"env": true, "enableAllProjectMcpServers": true, "enabledMcpjsonServers": true,
		"disabledMcpjsonServers": true, "hooks": true, "disableAllHooks": true,
		"defaultShell": true, "theme": true, "outputStyle": true, "language": true,
		"showWelcomeBanner": true, "showThinkingSummaries": true,
		"promptSuggestionEnabled": true, "fastMode": true,
		"alwaysThinkingEnabled": true, "effortLevel": true,
		"autoMemoryEnabled": true, "autoMemoryDirectory": true,
		"statusLine": true, "spinnerTipsEnabled": true,
	}
	for k, v := range raw {
		if !knownKeys[k] {
			s.Raw[k] = v
		}
	}
}
```

- [ ] **Step 3: Add Get, Set, Save, and helper methods**

```go
// Get returns a raw setting value by key
func (s *Settings) Get(key string) interface{} {
	// Check typed fields first
	switch key {
	case "model":
		return s.Model
	case "permissionMode":
		return s.PermissionMode
	case "respectGitignore":
		return s.RespectGitignore
	case "showWelcomeBanner":
		return s.ShowWelcomeBanner
	case "fastMode":
		return s.FastMode
	case "defaultShell":
		return s.DefaultShell
	case "theme":
		return s.Theme
	}
	// Check raw
	if s.Raw != nil {
		if v, ok := s.Raw[key]; ok {
			return v
		}
	}
	return nil
}

// Set sets a runtime override for a setting
func (s *Settings) Set(key string, value interface{}) {
	if s.Raw == nil {
		s.Raw = make(map[string]interface{})
	}
	s.Raw[key] = value
}

// Save saves settings to the user settings file
func (s *Settings) Save() error {
	path := s.userSettingsPath
	if path == "" {
		path = filepath.Join(homeDir(), ".claude", "settings.json")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// APIKey returns the API key from env
func (s *Settings) APIKey() string {
	if s.Env != nil {
		if key, ok := s.Env["ANTHROPIC_API_KEY"]; ok {
			return key
		}
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}

// MaxTokens returns the configured max tokens
func (s *Settings) MaxTokens() int {
	if s.Raw != nil {
		if v, ok := s.Raw["max_tokens"]; ok {
			switch val := v.(type) {
			case float64:
				return int(val)
			case int:
				return val
			case string:
				if i, err := strconv.Atoi(val); err == nil {
					return i
				}
			}
		}
	}
	return 8192
}
```

---

### Task 3: Create global.go

**Files:**
- Create: `internal/services/config/global.go`

- [ ] **Step 1: Create GlobalConfig struct and load/save functions**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GlobalConfig stores user preferences separate from settings.json
type GlobalConfig struct {
	Theme                      string `json:"theme"`
	Verbose                    bool   `json:"verbose"`
	AutoCompactEnabled         bool   `json:"autoCompactEnabled"`
	ShowTurnDuration           bool   `json:"showTurnDuration"`
	RespectGitignore           bool   `json:"respectGitignore"`
	PermissionExplainerEnabled bool   `json:"permissionExplainerEnabled"`
	TodoFeatureEnabled         bool   `json:"todoFeatureEnabled"`
	ShowExpandedTodos          bool   `json:"showExpandedTodos"`
	AutoConnectIde             bool   `json:"autoConnectIde"`
	FileCheckpointingEnabled   bool   `json:"fileCheckpointingEnabled"`
	TerminalProgressBarEnabled bool   `json:"terminalProgressBarEnabled"`
	TaskCompleteNotifEnabled   bool   `json:"taskCompleteNotifEnabled"`
	InputNeededNotifEnabled    bool   `json:"inputNeededNotifEnabled"`
	TeammateMode               bool   `json:"teammateMode"`
	TeammateDefaultModel       string `json:"teammateDefaultModel"`
}

// LoadGlobalConfig loads ~/.claude.json
func LoadGlobalConfig() (*GlobalConfig, error) {
	path := filepath.Join(homeDir(), ".claude.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	gc := DefaultGlobalConfig()
	if err := json.Unmarshal(data, gc); err != nil {
		return gc, nil
	}

	return gc, nil
}

// SaveGlobalConfig saves to ~/.claude.json
func (gc *GlobalConfig) SaveGlobalConfig() error {
	path := filepath.Join(homeDir(), ".claude.json")

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(gc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
```

---

### Task 4: Create env.go

**Files:**
- Create: `internal/services/config/env.go`

- [ ] **Step 1: Create env file parsing and application utilities**

```go
package config

import (
	"os"
	"strings"
)

// LoadEnvFile parses a .env file and returns key-value pairs
func LoadEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = strings.Trim(value, "\"'")
			result[key] = value
		}
	}

	return result, nil
}

// ApplyEnv applies environment variables from a map, only if not already set
func ApplyEnv(env map[string]string) {
	for k, v := range env {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

// applyEnvVars applies environment variable overrides to Settings
func applyEnvVars(s *Settings) {
	if v := os.Getenv("CLAUDE_CODE_MODEL"); v != "" {
		s.Model = v
	}
	if v := os.Getenv("CLAUDE_CODE_DEFAULT_SHELL"); v != "" {
		s.DefaultShell = v
	}
	if v := os.Getenv("CLAUDE_CODE_THEME"); v != "" {
		s.Theme = v
	}
	if v := os.Getenv("CLAUDE_CODE_FAST_MODE"); v != "" {
		s.FastMode = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_ALWAYS_THINKING"); v != "" {
		s.AlwaysThinkingEnabled = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_AUTO_MEMORY_ENABLED"); v != "" {
		s.AutoMemoryEnabled = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_PERMISSION_MODE"); v != "" {
		s.PermissionMode = v
	}
	if v := os.Getenv("CLAUDE_CODE_RESPECT_GITIGNORE"); v != "" {
		s.RespectGitignore = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_VERBOSE"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["verbose"] = v == "true" || v == "1" || v == "yes"
	}
}
```

---

### Task 5: Update config.go

**Files:**
- Modify: `internal/services/config/config.go`

- [ ] **Step 1: Rewrite config.go to wrap Settings and maintain backward compatibility**

Replace the entire file with:

```go
package config

import (
	"os"
	"path/filepath"
)

// Config holds application configuration (wraps Settings for compatibility)
type Config struct {
	Settings *Settings
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Settings: DefaultSettings(),
	}
}

// Load loads configuration from files and environment
func Load(cwd string) (*Config, error) {
	settings, err := LoadSettings(cwd)
	if err != nil {
		return nil, err
	}
	return &Config{Settings: settings}, nil
}

// GetString returns a string setting value with optional default
func (c *Config) GetString(key string, defaultVal string) string {
	if v, ok := c.Settings.Raw[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// GetInt returns an int setting value with optional default
func (c *Config) GetInt(key string, defaultVal int) int {
	if v, ok := c.Settings.Raw[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

// GetBool returns a bool setting value with optional default
func (c *Config) GetBool(key string, defaultVal bool) bool {
	if v, ok := c.Settings.Raw[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true" || val == "1" || val == "yes"
		case float64:
			return val != 0
		}
	}
	return defaultVal
}

// GetEnv returns an environment variable value
func (c *Config) GetEnv(key string) string {
	if c.Settings.Env != nil {
		if v, ok := c.Settings.Env[key]; ok {
			return v
		}
	}
	return os.Getenv(key)
}

// Set sets a setting value
func (c *Config) Set(key string, value any) {
	c.Settings.Set(key, value)
}

// All returns all raw settings
func (c *Config) All() map[string]any {
	return c.Settings.Raw
}

// Model returns the configured model
func (c *Config) Model() string {
	return c.Settings.Model
}

// MaxTokens returns the configured max tokens
func (c *Config) MaxTokens() int {
	return c.Settings.MaxTokens()
}

// PermissionMode returns the configured permission mode
func (c *Config) PermissionMode() string {
	return c.Settings.PermissionMode
}

// APIKey returns the API key from env
func (c *Config) APIKey() string {
	return c.Settings.APIKey()
}

// ShowWelcomeBanner returns whether to show the welcome banner
func (c *Config) ShowWelcomeBanner() bool {
	return c.Settings.ShowWelcomeBanner
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return "."
}
```

Note: Need to add `strconv` import.

---

### Task 6: Update bootstrap.go

**Files:**
- Modify: `internal/bootstrap/bootstrap.go`

- [ ] **Step 1: Update State to expose Settings alongside Config**

The State struct already has `Config *config.Config`. The Config now wraps Settings, so all existing code continues to work. No changes needed to the State struct itself, but we should expose Settings directly for new code:

```go
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
```

- [ ] **Step 2: Update Initialize to populate Settings**

```go
// In Initialize(), after cfg is loaded:
settings, err := config.LoadSettings(cwd)
if err != nil {
    return nil, err
}
cfg := &config.Config{Settings: settings}
```

Full updated Initialize:

```go
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

	queryEngine := engine.NewQueryEngine(engine.EngineConfig{
		Cwd:          cwd,
		Tools:        tools,
		Model:        settings.Model,
		MaxTokens:    settings.MaxTokens(),
		MaxBudgetUSD: 0,
		CustomPrompt: "",
		AppendPrompt: "",
		Debug:        false,
		Verbose:      false,
		APIKey:       authState.APIKey,
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
```

---

### Task 7: Update tui/app.go

**Files:**
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Update newAppModel to use Settings directly**

The current code uses `state.Config.Model()` and `state.Config.PermissionMode()`. These still work via the Config wrapper, but we should also show Settings usage. The existing code is compatible - no changes needed for compilation.

No changes required to tui/app.go since Config wrapper maintains backward compatibility with `.Model()` and `.PermissionMode()` methods.

---

### Task 8: Build and verify

**Files:** All of the above

- [ ] **Step 1: Run go build to verify compilation**

```bash
go build ./...
```

- [ ] **Step 2: Fix any compilation errors**

Common issues to watch for:
- Missing `strconv` import in config.go
- Type mismatches in mergeSettings
- Missing method receivers

- [ ] **Step 3: Run go vet for additional checks**

```bash
go vet ./...
```

---

## Self-Review

### Spec Coverage Checklist

| Requirement | Task |
|-------------|------|
| Settings struct with all fields | Task 2 |
| Permissions, FileSuggestion, Hook, StatusLine sub-structs | Task 2 |
| GlobalConfig struct | Task 3 |
| LoadSettings with merge logic | Task 2 |
| Get/Set/Save methods | Task 2 |
| LoadGlobalConfig/SaveGlobalConfig | Task 3 |
| LoadEnvFile/ApplyEnv | Task 4 |
| Environment variable overrides | Task 4 |
| Default values | Task 1 |
| Merge order (user → project → local → env → env vars → global) | Task 2 |
| Raw access for unknown keys | Task 2 |
| Update config.go wrapper | Task 5 |
| Update bootstrap.go | Task 6 |
| Update tui/app.go | Task 7 |
| Build verification | Task 8 |

### Placeholder Scan
No placeholders found - all code is complete.

### Type Consistency
- Settings struct fields match spec exactly
- Config wrapper delegates to Settings for all methods
- Bootstrap State exposes both Config and Settings
- All method signatures consistent across files
