package config

import (
	"os"
	"strings"
)

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
	// API key
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		s.Env["ANTHROPIC_API_KEY"] = v
	}
	if v := os.Getenv("ANTHROPIC_AUTH_TOKEN"); v != "" {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		s.Env["ANTHROPIC_AUTH_TOKEN"] = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		if s.Env == nil {
			s.Env = make(map[string]string)
		}
		s.Env["OPENAI_API_KEY"] = v
	}

	// Model
	if v := os.Getenv("CLAUDE_CODE_MODEL"); v != "" {
		s.Model = v
	}
	if v := os.Getenv("ANTHROPIC_MODEL"); v != "" {
		s.Model = v
	}
	if v := os.Getenv("MODEL"); v != "" {
		s.Model = v
	}

	// Default model overrides (these ALWAYS override s.Model)
	if v := os.Getenv("ANTHROPIC_DEFAULT_OPUS_MODEL"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["default_opus_model"] = v
		s.Model = v
	}
	if v := os.Getenv("ANTHROPIC_DEFAULT_SONNET_MODEL"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["default_sonnet_model"] = v
		s.Model = v
	}
	if v := os.Getenv("ANTHROPIC_DEFAULT_HAIKU_MODEL"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["default_haiku_model"] = v
		s.Model = v
	}

	// API base URL
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["anthropic_base_url"] = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["anthropic_base_url"] = v
	}

	// API version
	if v := os.Getenv("ANTHROPIC_API_VERSION"); v != "" {
		if s.Raw == nil {
			s.Raw = make(map[string]interface{})
		}
		s.Raw["anthropic_api_version"] = v
	}

	// Shell
	if v := os.Getenv("CLAUDE_CODE_DEFAULT_SHELL"); v != "" {
		s.DefaultShell = v
	}
	if v := os.Getenv("SHELL"); v != "" && s.DefaultShell == "" {
		s.DefaultShell = v
	}

	// Theme
	if v := os.Getenv("CLAUDE_CODE_THEME"); v != "" {
		s.Theme = v
	}

	// Permission mode
	if v := os.Getenv("CLAUDE_CODE_PERMISSION_MODE"); v != "" {
		s.PermissionMode = v
	}

	// Feature flags
	if v := os.Getenv("CLAUDE_CODE_FAST_MODE"); v != "" {
		s.FastMode = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_ALWAYS_THINKING"); v != "" {
		s.AlwaysThinkingEnabled = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_AUTO_MEMORY_ENABLED"); v != "" {
		s.AutoMemoryEnabled = v == "true" || v == "1" || v == "yes"
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
	if v := os.Getenv("CLAUDE_CODE_SHOW_WELCOME_BANNER"); v != "" {
		s.ShowWelcomeBanner = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_SHOW_THINKING_SUMMARIES"); v != "" {
		s.ShowThinkingSummaries = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_PROMPT_SUGGESTION_ENABLED"); v != "" {
		s.PromptSuggestionEnabled = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_DISABLE_ALL_HOOKS"); v != "" {
		s.DisableAllHooks = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("CLAUDE_CODE_SPINNER_TIPS_ENABLED"); v != "" {
		s.SpinnerTipsEnabled = v == "true" || v == "1" || v == "yes"
	}

	// Effort level
	if v := os.Getenv("CLAUDE_CODE_EFFORT_LEVEL"); v != "" {
		s.EffortLevel = v
	}

	// Output style
	if v := os.Getenv("CLAUDE_CODE_OUTPUT_STYLE"); v != "" {
		s.OutputStyle = v
	}

	// Language
	if v := os.Getenv("CLAUDE_CODE_LANGUAGE"); v != "" {
		s.Language = v
	}

	// Auto memory
	if v := os.Getenv("CLAUDE_CODE_AUTO_MEMORY_DIRECTORY"); v != "" {
		s.AutoMemoryDirectory = v
	}
}
