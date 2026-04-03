package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// Event represents a telemetry event
type Event struct {
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Tracker handles telemetry and analytics
type Tracker struct {
	disabled bool
	events   chan Event
	wg       sync.WaitGroup
	done     chan struct{}
	mu       sync.Mutex
}

// NewTracker creates a new analytics tracker
func NewTracker(disabled bool) *Tracker {
	t := &Tracker{
		disabled: disabled,
		events:   make(chan Event, 100),
		done:     make(chan struct{}),
	}

	if !disabled {
		t.wg.Add(1)
		go t.processEvents()
	}

	return t
}

// Track records a custom event
func (t *Tracker) Track(name string, properties map[string]interface{}) {
	if t.disabled {
		return
	}

	event := Event{
		Name:       name,
		Timestamp:  time.Now(),
		Properties: properties,
	}

	select {
	case t.events <- event:
	default:
	}
}

// TrackToolUse records a tool execution
func (t *Tracker) TrackToolUse(toolName string, duration time.Duration, success bool) {
	t.Track("tool_use", map[string]interface{}{
		"tool_name": toolName,
		"duration":  duration.Milliseconds(),
		"success":   success,
	})
}

// TrackCommand records a command execution
func (t *Tracker) TrackCommand(command string, duration time.Duration) {
	t.Track("command", map[string]interface{}{
		"command":  command,
		"duration": duration.Milliseconds(),
	})
}

// TrackSessionStart records the start of a session
func (t *Tracker) TrackSessionStart(model string) {
	t.Track("session_start", map[string]interface{}{
		"model": model,
	})
}

// TrackSessionEnd records the end of a session
func (t *Tracker) TrackSessionEnd(duration time.Duration, totalCost float64, usage types.Usage) {
	t.Track("session_end", map[string]interface{}{
		"duration":      duration.Milliseconds(),
		"total_cost":    totalCost,
		"input_tokens":  usage.InputTokens,
		"output_tokens": usage.OutputTokens,
		"cache_read":    usage.CacheRead,
		"cache_write":   usage.CacheWrite,
	})
}

// Close stops the tracker and flushes remaining events
func (t *Tracker) Close() {
	if t.disabled {
		return
	}

	close(t.events)
	t.wg.Wait()
}

func (t *Tracker) processEvents() {
	defer t.wg.Done()

	for event := range t.events {
		if err := t.persistEvent(event); err != nil {
			fmt.Fprintf(os.Stderr, "[analytics] failed to persist event: %v\n", err)
		}
	}
}

func (t *Tracker) persistEvent(event Event) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	dir := filepath.Join(homeDir(), ".claude", "analytics")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create analytics directory: %w", err)
	}

	date := event.Timestamp.Format("2006-01-02")
	path := filepath.Join(dir, fmt.Sprintf("events-%s.jsonl", date))

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open event file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
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
