package tr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	cacheFileName = "cache.json"
	cacheVersion  = 3
	cacheMaxSize  = 500

	// cacheMaxEntryBytes keeps one huge document from dominating cache.json.
	// Long documents are cached per chunk instead of as a single entry.
	cacheMaxEntryBytes = 32 * 1024
)

// CacheEntry stores a single cached translation.
type CacheEntry struct {
	TranslatedText string    `json:"translated_text"`
	Timestamp      time.Time `json:"timestamp"`
}

// Cache is an in-memory + on-disk translation cache.
//
// Entries are evicted by least-recent use: Get updates the timestamp, Put adds
// the entry, and the oldest 20% is removed once the limit is exceeded. Disk
// writes are deferred until Flush so translating a large file does not rewrite
// the whole file once per line.
type Cache struct {
	mu       sync.Mutex
	entries  map[string]CacheEntry
	filePath string
	dirty    bool
}

// NewCache loads or creates a cache at the default path.
func NewCache() (*Cache, error) {
	return LoadCache(filepath.Join(ConfigDir(), cacheFileName))
}

// LoadCache reads a cache file from disk. A missing file yields an empty
// cache. A corrupt file is renamed to .corrupt and reported, so the user can
// inspect it instead of losing it silently.
func LoadCache(path string) (*Cache, error) {
	c := &Cache{
		entries:  make(map[string]CacheEntry),
		filePath: path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, fmt.Errorf("read cache: %w", err)
	}
	var disk struct {
		Version int                   `json:"version"`
		Entries map[string]CacheEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &disk); err != nil {
		backup := path + ".corrupt"
		_ = os.Rename(path, backup)
		return c, fmt.Errorf("parse cache (backed up to %s): %w", backup, err)
	}
	if disk.Version != cacheVersion {
		// Old key formats are not compatible; start clean but keep the file
		// intact for reference.
		_ = os.Rename(path, path+".old")
		return c, nil
	}
	if disk.Entries != nil {
		c.entries = disk.Entries
	}
	return c, nil
}

// cacheKey builds a bounded key from the backend scope and exact source text.
// The text is hashed so keys stay short even for file blocks; the exact string
// (including case and whitespace) is part of the hash to avoid returning a
// translation produced for different formatting.
func cacheKey(scope, source, target, text string) string {
	sum := sha256.Sum256([]byte(text))
	return scope + "\x1f" + source + "\x1f" + target + "\x1f" + hex.EncodeToString(sum[:])
}

// Get retrieves a cached translation and marks it as recently used.
func (c *Cache) Get(scope, source, target, text string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := cacheKey(scope, source, target, text)
	entry, ok := c.entries[key]
	if !ok {
		return "", false
	}
	entry.Timestamp = time.Now()
	c.entries[key] = entry
	c.dirty = true
	return entry.TranslatedText, true
}

// Put stores a translation in the cache. Oversized entries are skipped: a
// long document is cached chunk by chunk instead.
func (c *Cache) Put(scope, source, target, text, translatedText string) {
	if len(text) > cacheMaxEntryBytes || len(translatedText) > cacheMaxEntryBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(scope, source, target, text)
	c.entries[key] = CacheEntry{
		TranslatedText: translatedText,
		Timestamp:      time.Now(),
	}
	if len(c.entries) > cacheMaxSize {
		c.evictOldest(cacheMaxSize / 5)
	}
	c.dirty = true
}

// evictOldest removes the n least recently used entries.
func (c *Cache) evictOldest(n int) {
	if n <= 0 || len(c.entries) == 0 {
		return
	}
	type keyTime struct {
		key string
		t   time.Time
	}
	all := make([]keyTime, 0, len(c.entries))
	for k, v := range c.entries {
		all = append(all, keyTime{k, v.Timestamp})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].t.Before(all[j].t) })
	if n > len(all) {
		n = len(all)
	}
	for _, kv := range all[:n] {
		delete(c.entries, kv.key)
	}
}

// Flush writes the cache to disk if it changed. The write is atomic.
func (c *Cache) Flush() error {
	c.mu.Lock()
	if !c.dirty {
		c.mu.Unlock()
		return nil
	}
	disk := struct {
		Version int                   `json:"version"`
		Entries map[string]CacheEntry `json:"entries"`
	}{
		Version: cacheVersion,
		Entries: c.entries,
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.dirty = false
	c.mu.Unlock()

	if err := writeFileAtomic(c.filePath, data, 0o600); err != nil {
		c.mu.Lock()
		c.dirty = true
		c.mu.Unlock()
		return err
	}
	return nil
}

// Clear removes every cached translation.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]CacheEntry)
	c.dirty = true
	c.mu.Unlock()
}

// Stats returns the number of cached translations.
func (c *Cache) Stats() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}
