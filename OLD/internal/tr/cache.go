package tr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	cacheFileName = "cache.json"
	cacheMaxSize  = 500 // maximum number of cached entries
)

// CacheEntry stores a single cached translation.
type CacheEntry struct {
	TranslatedText string    `json:"translated_text"`
	Timestamp      time.Time `json:"timestamp"`
}

// Cache is an in-memory + on-disk translation cache.
// Thread-safe for concurrent reads/writes via sync.RWMutex.
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]CacheEntry // key = "source:target:normalizedText"
	filePath string
}

// NewCache loads or creates a cache at the default path.
func NewCache() (*Cache, error) {
	path := filepath.Join(ConfigDir(), cacheFileName)
	return LoadCache(path)
}

// LoadCache reads a cache file from disk and returns an initialized Cache.
// If the file is missing or corrupt, returns an empty cache (no error).
func LoadCache(path string) (*Cache, error) {
	c := &Cache{
		entries:  make(map[string]CacheEntry),
		filePath: path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return c, nil // missing or unreadable; start fresh
	}
	var disk struct {
		Version int                    `json:"version"`
		Entries map[string]CacheEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &disk); err != nil {
		return c, nil // corrupt; start fresh
	}
	if disk.Entries != nil {
		c.entries = disk.Entries
	}
	return c, nil
}

// cacheKey builds the lookup key: "source:target:normalizedText".
func cacheKey(source, target, normalizedText string) string {
	return source + ":" + target + ":" + normalizedText
}

// Get retrieves a cached translation. Returns ("", false) on miss.
func (c *Cache) Get(source, target, normalizedText string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[cacheKey(source, target, normalizedText)]
	if !ok {
		return "", false
	}
	return entry.TranslatedText, true
}

// Put stores a translation in the cache. If the cache exceeds cacheMaxSize,
// it evicts the oldest entries. Writes to disk asynchronously.
func (c *Cache) Put(source, target, normalizedText, translatedText string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(source, target, normalizedText)
	c.entries[key] = CacheEntry{
		TranslatedText: translatedText,
		Timestamp:      time.Now(),
	}

	// Evict if over limit
	if len(c.entries) > cacheMaxSize {
		c.evictOldest(cacheMaxSize / 5)
	}

	// Fire-and-forget save to disk
	c.save()
}

// evictOldest removes approximately n oldest entries from the map.
func (c *Cache) evictOldest(n int) {
	if n <= 0 || len(c.entries) == 0 {
		return
	}
	type keyTime struct {
		key string
		t   time.Time
	}
	// Collect all entries and sort by finding the n oldest
	oldest := make([]keyTime, 0, n)
	for k, v := range c.entries {
		if len(oldest) < n {
			oldest = append(oldest, keyTime{k, v.Timestamp})
			continue
		}
		// Find the newest among our "oldest" batch and replace if this is older
		newestIdx := 0
		for i := 1; i < len(oldest); i++ {
			if oldest[i].t.After(oldest[newestIdx].t) {
				newestIdx = i
			}
		}
		if v.Timestamp.Before(oldest[newestIdx].t) {
			oldest[newestIdx] = keyTime{k, v.Timestamp}
		}
	}
	for _, kv := range oldest {
		delete(c.entries, kv.key)
	}
}

// save writes the current in-memory cache to disk.
func (c *Cache) save() {
	disk := struct {
		Version int                    `json:"version"`
		Entries map[string]CacheEntry `json:"entries"`
	}{
		Version: 1,
		Entries: c.entries,
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(c.filePath), 0755)
	_ = os.WriteFile(c.filePath, data, 0644)
}

// Flush forces a write to disk (useful for graceful shutdown).
func (c *Cache) Flush() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	disk := struct {
		Version int                    `json:"version"`
		Entries map[string]CacheEntry `json:"entries"`
	}{
		Version: 1,
		Entries: c.entries,
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(c.filePath), 0755)
	return os.WriteFile(c.filePath, data, 0644)
}
