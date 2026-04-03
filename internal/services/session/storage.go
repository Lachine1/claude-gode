package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// SessionInfo holds metadata about a saved session
type SessionInfo struct {
	ID           string      `json:"id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Model        string      `json:"model"`
	MessageCount int         `json:"message_count"`
	Usage        types.Usage `json:"usage"`
}

// SessionData represents the full session data stored on disk
type SessionData struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Model     string          `json:"model"`
	Messages  []types.Message `json:"messages"`
	Usage     types.Usage     `json:"usage"`
}

// Storage handles session persistence
type Storage struct {
	basePath string
}

// NewStorage creates a new session storage
func NewStorage(basePath string) *Storage {
	return &Storage{
		basePath: basePath,
	}
}

// SaveSession persists a session to disk
func (s *Storage) SaveSession(sessionID string, messages []types.Message, usage types.Usage) error {
	if err := os.MkdirAll(s.basePath, 0o700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	sessionPath := filepath.Join(s.basePath, sessionID)
	if err := os.MkdirAll(sessionPath, 0o700); err != nil {
		return fmt.Errorf("failed to create session folder: %w", err)
	}

	data := SessionData{
		ID:        sessionID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  messages,
		Usage:     usage,
	}

	existingPath := filepath.Join(sessionPath, "session.json")
	if existing, err := os.ReadFile(existingPath); err == nil {
		var existingData SessionData
		if err := json.Unmarshal(existing, &existingData); err == nil {
			data.CreatedAt = existingData.CreatedAt
		}
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(existingPath, out, 0o600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession retrieves a session from disk
func (s *Storage) LoadSession(sessionID string) ([]types.Message, types.Usage, error) {
	sessionPath := filepath.Join(s.basePath, sessionID, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.Usage{}, fmt.Errorf("session %q not found", sessionID)
		}
		return nil, types.Usage{}, fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, types.Usage{}, fmt.Errorf("failed to parse session file: %w", err)
	}

	return session.Messages, session.Usage, nil
}

// ListSessions returns metadata for all saved sessions
func (s *Storage) ListSessions() ([]SessionInfo, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []SessionInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(s.basePath, entry.Name(), "session.json")
		data, err := os.ReadFile(sessionPath)
		if err != nil {
			continue
		}

		var session SessionData
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sessions = append(sessions, SessionInfo{
			ID:           session.ID,
			CreatedAt:    session.CreatedAt,
			UpdatedAt:    session.UpdatedAt,
			Model:        session.Model,
			MessageCount: len(session.Messages),
			Usage:        session.Usage,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// DeleteSession removes a saved session
func (s *Storage) DeleteSession(sessionID string) error {
	sessionPath := filepath.Join(s.basePath, sessionID)

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session %q not found", sessionID)
	}

	return os.RemoveAll(sessionPath)
}
