package completion

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	shellHistoryCache     []string
	shellHistoryCacheTime time.Time
	shellHistoryMutex     sync.Mutex
	shellHistoryTTL       = 60 * time.Second
)

// GetShellHistoryCompletion returns a ghost text suggestion from shell history
func GetShellHistoryCompletion(input string) *GhostText {
	if input == "" {
		return nil
	}

	history := getShellHistory()
	inputRunes := []rune(input)
	for _, cmd := range history {
		if strings.HasPrefix(cmd, input) && cmd != input {
			cmdRunes := []rune(cmd)
			if len(cmdRunes) <= len(inputRunes) {
				continue
			}
			suffix := string(cmdRunes[len(inputRunes):])
			return &GhostText{
				Text:           suffix,
				FullCommand:    cmd,
				InsertPosition: len(inputRunes),
			}
		}
	}
	return nil
}

func getShellHistory() []string {
	shellHistoryMutex.Lock()
	defer shellHistoryMutex.Unlock()

	if time.Since(shellHistoryCacheTime) < shellHistoryTTL && shellHistoryCache != nil {
		return shellHistoryCache
	}

	shellHistoryCache = loadShellHistory()
	shellHistoryCacheTime = time.Now()
	return shellHistoryCache
}

func loadShellHistory() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".bash_history"),
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".history"),
	}

	for _, path := range candidates {
		if entries := readHistoryFile(path); len(entries) > 0 {
			return entries
		}
	}

	return nil
}

func readHistoryFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	seen := make(map[string]bool)
	var commands []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ": ") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				line = strings.TrimSpace(parts[2])
			}
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !seen[line] {
			seen[line] = true
			commands = append(commands, line)
		}
	}

	if len(commands) > 50 {
		commands = commands[len(commands)-50:]
	}

	for i := len(commands)/2 - 1; i >= 0; i-- {
		opp := len(commands) - 1 - i
		commands[i], commands[opp] = commands[opp], commands[i]
	}

	return commands
}
