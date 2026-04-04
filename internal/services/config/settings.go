package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// Settings holds all application settings
type Settings struct {
	Model                      string                 `json:"model"`
	AvailableModels            []string               `json:"availableModels"`
	PermissionMode             string                 `json:"permissionMode"`
	Permissions                Permissions            `json:"permissions"`
	FileSuggestion             FileSuggestion         `json:"fileSuggestion"`
	RespectGitignore           bool                   `json:"respectGitignore"`
	Env                        map[string]string      `json:"env"`
	EnableAllProjectMcpServers bool                   `json:"enableAllProjectMcpServers"`
	EnabledMcpjsonServers      []string               `json:"enabledMcpjsonServers"`
	DisabledMcpjsonServers     []string               `json:"disabledMcpjsonServers"`
	Hooks                      map[string]Hook        `json:"hooks"`
	DisableAllHooks            bool                   `json:"disableAllHooks"`
	DefaultShell               string                 `json:"defaultShell"`
	Theme                      string                 `json:"theme"`
	OutputStyle                string                 `json:"outputStyle"`
	Language                   string                 `json:"language"`
	ShowWelcomeBanner          bool                   `json:"showWelcomeBanner"`
	ShowThinkingSummaries      bool                   `json:"showThinkingSummaries"`
	PromptSuggestionEnabled    bool                   `json:"promptSuggestionEnabled"`
	FastMode                   bool                   `json:"fastMode"`
	AlwaysThinkingEnabled      bool                   `json:"alwaysThinkingEnabled"`
	EffortLevel                string                 `json:"effortLevel"`
	AutoMemoryEnabled          bool                   `json:"autoMemoryEnabled"`
	AutoMemoryDirectory        string                 `json:"autoMemoryDirectory"`
	StatusLine                 StatusLine             `json:"statusLine"`
	SpinnerTipsEnabled         bool                   `json:"spinnerTipsEnabled"`
	Raw                        map[string]interface{} `json:"-"`

	userSettingsPath    string
	projectSettingsPath string
	localSettingsPath   string
}

// Permissions configures tool and file access permissions
type Permissions struct {
	Allow                        []string `json:"allow"`
	Deny                         []string `json:"deny"`
	Ask                          []string `json:"ask"`
	DefaultMode                  string   `json:"defaultMode"`
	DisableBypassPermissionsMode bool     `json:"disableBypassPermissionsMode"`
	DisableAutoMode              bool     `json:"disableAutoMode"`
	AdditionalDirectories        []string `json:"additionalDirectories"`
}

// FileSuggestion configures file suggestion behavior
type FileSuggestion struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Hook configures a hook event handler
type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// StatusLine configures the status line display
type StatusLine struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Padding int    `json:"padding"`
}

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
	}

	return s, nil
}

var knownKeys = map[string]bool{
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

func mergeSettings(s *Settings, data []byte) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	if v, ok := raw["model"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.Model = str
		}
	}
	if v, ok := raw["permissionMode"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.PermissionMode = str
		}
	}
	if v, ok := raw["defaultShell"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.DefaultShell = str
		}
	}
	if v, ok := raw["theme"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.Theme = str
		}
	}
	if v, ok := raw["outputStyle"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.OutputStyle = str
		}
	}
	if v, ok := raw["language"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.Language = str
		}
	}
	if v, ok := raw["effortLevel"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.EffortLevel = str
		}
	}
	if v, ok := raw["autoMemoryDirectory"]; ok {
		if str, ok := v.(string); ok && str != "" {
			s.AutoMemoryDirectory = str
		}
	}

	if v, ok := raw["availableModels"]; ok {
		if arr, ok := v.([]interface{}); ok {
			s.AvailableModels = make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					s.AvailableModels = append(s.AvailableModels, str)
				}
			}
		}
	}
	if v, ok := raw["enabledMcpjsonServers"]; ok {
		if arr, ok := v.([]interface{}); ok {
			s.EnabledMcpjsonServers = make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					s.EnabledMcpjsonServers = append(s.EnabledMcpjsonServers, str)
				}
			}
		}
	}
	if v, ok := raw["disabledMcpjsonServers"]; ok {
		if arr, ok := v.([]interface{}); ok {
			s.DisabledMcpjsonServers = make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					s.DisabledMcpjsonServers = append(s.DisabledMcpjsonServers, str)
				}
			}
		}
	}

	if v, ok := raw["permissions"]; ok {
		if perm, ok := v.(map[string]interface{}); ok {
			if val, ok := perm["defaultMode"]; ok {
				if str, ok := val.(string); ok && str != "" {
					s.Permissions.DefaultMode = str
				}
			}
			if val, ok := perm["allow"]; ok {
				if arr, ok := val.([]interface{}); ok {
					s.Permissions.Allow = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							s.Permissions.Allow = append(s.Permissions.Allow, str)
						}
					}
				}
			}
			if val, ok := perm["deny"]; ok {
				if arr, ok := val.([]interface{}); ok {
					s.Permissions.Deny = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							s.Permissions.Deny = append(s.Permissions.Deny, str)
						}
					}
				}
			}
			if val, ok := perm["ask"]; ok {
				if arr, ok := val.([]interface{}); ok {
					s.Permissions.Ask = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							s.Permissions.Ask = append(s.Permissions.Ask, str)
						}
					}
				}
			}
			if val, ok := perm["additionalDirectories"]; ok {
				if arr, ok := val.([]interface{}); ok {
					s.Permissions.AdditionalDirectories = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							s.Permissions.AdditionalDirectories = append(s.Permissions.AdditionalDirectories, str)
						}
					}
				}
			}
			if val, ok := perm["disableBypassPermissionsMode"]; ok {
				if b, ok := val.(bool); ok {
					s.Permissions.DisableBypassPermissionsMode = b
				}
			}
			if val, ok := perm["disableAutoMode"]; ok {
				if b, ok := val.(bool); ok {
					s.Permissions.DisableAutoMode = b
				}
			}
		}
	}

	if v, ok := raw["fileSuggestion"]; ok {
		if fs, ok := v.(map[string]interface{}); ok {
			if val, ok := fs["type"]; ok {
				if str, ok := val.(string); ok && str != "" {
					s.FileSuggestion.Type = str
				}
			}
			if val, ok := fs["command"]; ok {
				if str, ok := val.(string); ok && str != "" {
					s.FileSuggestion.Command = str
				}
			}
		}
	}

	if v, ok := raw["statusLine"]; ok {
		if sl, ok := v.(map[string]interface{}); ok {
			if val, ok := sl["type"]; ok {
				if str, ok := val.(string); ok && str != "" {
					s.StatusLine.Type = str
				}
			}
			if val, ok := sl["command"]; ok {
				if str, ok := val.(string); ok && str != "" {
					s.StatusLine.Command = str
				}
			}
			if val, ok := sl["padding"]; ok {
				switch n := val.(type) {
				case float64:
					s.StatusLine.Padding = int(n)
				case int:
					s.StatusLine.Padding = n
				}
			}
		}
	}

	if v, ok := raw["hooks"]; ok {
		if hooks, ok := v.(map[string]interface{}); ok {
			if s.Hooks == nil {
				s.Hooks = make(map[string]Hook)
			}
			for k, hv := range hooks {
				if hm, ok := hv.(map[string]interface{}); ok {
					h := Hook{}
					if val, ok := hm["type"]; ok {
						if str, ok := val.(string); ok {
							h.Type = str
						}
					}
					if val, ok := hm["command"]; ok {
						if str, ok := val.(string); ok {
							h.Command = str
						}
					}
					s.Hooks[k] = h
				}
			}
		}
	}

	if v, ok := raw["env"]; ok {
		if envMap, ok := v.(map[string]interface{}); ok {
			if s.Env == nil {
				s.Env = make(map[string]string)
			}
			for k, ev := range envMap {
				if str, ok := ev.(string); ok {
					s.Env[k] = str
				}
			}
		}
	}

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
	for k, v := range raw {
		if !knownKeys[k] {
			s.Raw[k] = v
		}
	}
}

// Get returns a raw setting value by key
func (s *Settings) Get(key string) interface{} {
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
