package auth

import (
	"os"

	"github.com/Lachine1/claude-gode/internal/services/config"
)

// AuthState holds authentication state
type AuthState struct {
	APIKey     string
	IsLoggedIn bool
	OrgID      string
}

// Initialize sets up authentication
func Initialize(cfg *config.Config) (*AuthState, error) {
	apiKey := cfg.APIKey()
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	return &AuthState{
		APIKey:     apiKey,
		IsLoggedIn: apiKey != "",
	}, nil
}
