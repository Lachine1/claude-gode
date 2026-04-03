package engine

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	defaultCacheSize  = 100
	fileUnchangedStub = "File unchanged, no need to re-read."
	cacheTTL          = 5 * time.Minute
)

// FileEntry represents a cached file state
type FileEntry struct {
	Path       string
	Content    string
	ModTime    time.Time
	Size       int64
	Hash       string
	LastAccess time.Time
	ReadCount  int
}

// FileStateCache is an LRU-like cache for file contents
type FileStateCache struct {
	mu      sync.RWMutex
	entries map[string]*FileEntry
	order   []string
	maxSize int
}

// NewFileStateCache creates a new file state cache
func NewFileStateCache() *FileStateCache {
	return &FileStateCache{
		entries: make(map[string]*FileEntry),
		order:   make([]string, 0),
		maxSize: defaultCacheSize,
	}
}

// Get retrieves a cached file entry. Returns nil if not found or expired.
func (c *FileStateCache) Get(path string) *FileEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[path]
	if !ok {
		return nil
	}

	if time.Since(entry.LastAccess) > cacheTTL {
		return nil
	}

	return entry
}

// Set adds or updates a file entry in the cache
func (c *FileStateCache) Set(path string, content string, modTime time.Time, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := computeHash(content)

	if existing, ok := c.entries[path]; ok {
		existing.Content = content
		existing.ModTime = modTime
		existing.Size = size
		existing.Hash = hash
		existing.LastAccess = time.Now()
		existing.ReadCount++
		c.moveToFront(path)
		return
	}

	entry := &FileEntry{
		Path:       path,
		Content:    content,
		ModTime:    modTime,
		Size:       size,
		Hash:       hash,
		LastAccess: time.Now(),
		ReadCount:  1,
	}

	c.entries[path] = entry
	c.order = append([]string{path}, c.order...)

	if len(c.order) > c.maxSize {
		oldest := c.order[len(c.order)-1]
		c.order = c.order[:len(c.order)-1]
		delete(c.entries, oldest)
	}
}

// IsUnchanged checks if a file on disk matches the cached version
func (c *FileStateCache) IsUnchanged(path string) bool {
	entry := c.Get(path)
	if entry == nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.ModTime().Equal(entry.ModTime) && info.Size() == entry.Size
}

// GetUnchangedStub returns the FILE_UNCHANGED_STUB for unchanged files
func (c *FileStateCache) GetUnchangedStub(path string) string {
	if c.IsUnchanged(path) {
		return fileUnchangedStub
	}
	return ""
}

// Invalidate removes a file from the cache
func (c *FileStateCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(path)
}

// InvalidatePrefix removes all entries matching a prefix
func (c *FileStateCache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []string
	for path := range c.entries {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			toRemove = append(toRemove, path)
		}
	}
	for _, path := range toRemove {
		c.remove(path)
	}
}

// Clear removes all entries from the cache
func (c *FileStateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*FileEntry)
	c.order = make([]string, 0)
}

// Size returns the number of entries in the cache
func (c *FileStateCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// remove removes an entry (caller must hold write lock)
func (c *FileStateCache) remove(path string) {
	delete(c.entries, path)
	for i, p := range c.order {
		if p == path {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// moveToFront moves an entry to the front of the LRU order (caller must hold write lock)
func (c *FileStateCache) moveToFront(path string) {
	for i, p := range c.order {
		if p == path {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append([]string{path}, c.order...)
			break
		}
	}
}

func computeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:8])
}
