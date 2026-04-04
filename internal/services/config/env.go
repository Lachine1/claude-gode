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
