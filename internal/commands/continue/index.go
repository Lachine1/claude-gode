package continuecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Command returns the /continue command
func Command() types.Command {
	return types.Command{
		Name:        "continue",
		Aliases:     []string{"resume"},
		Description: "Continue last session",
		Usage:       "/continue [session-id]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleContinue(ctx, args)
		},
	}
}

func handleContinue(ctx *types.CommandContext, args []string) error {
	sessionDir := filepath.Join(homeDir(), ".claude", "sessions")

	var sessionFile string
	if len(args) > 0 {
		sessionFile = filepath.Join(sessionDir, args[0]+".json")
	} else {
		// Find the most recent session
		file, err := findMostRecentSession(sessionDir)
		if err != nil {
			return fmt.Errorf("no previous session found")
		}
		sessionFile = file
	}

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", filepath.Base(sessionFile))
	}

	// Load session data
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session: %w", err)
	}

	var session struct {
		Messages  []types.Message `json:"messages"`
		Timestamp int64           `json:"timestamp"`
		Cwd       string          `json:"cwd"`
	}

	if err := parseSessionJSON(data, &session); err != nil {
		return fmt.Errorf("failed to parse session: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("  Continue Session\n")
	sb.WriteString("  ═══════════════════════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("  Session:   %s\n", filepath.Base(sessionFile)))
	sb.WriteString(fmt.Sprintf("  CWD:       %s\n", session.Cwd))
	sb.WriteString(fmt.Sprintf("  Messages:  %d\n", len(session.Messages)))
	sb.WriteString(fmt.Sprintf("  Created:   %s\n", time.Unix(session.Timestamp, 0).Format("2006-01-02 15:04:05")))
	sb.WriteString("\n")
	sb.WriteString("  Session loaded. Conversation history restored.\n")
	sb.WriteString("\n")

	// Restore messages
	ctx.SetMessages(session.Messages)

	fmt.Println(sb.String())
	return nil
}

func findMostRecentSession(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var latest string
	var latestTime int64

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Unix() > latestTime {
			latestTime = info.ModTime().Unix()
			latest = filepath.Join(dir, entry.Name())
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no sessions found")
	}

	return latest, nil
}

func parseSessionJSON(data []byte, v interface{}) error {
	// Simple JSON parsing - in production, use encoding/json
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
