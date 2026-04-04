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
