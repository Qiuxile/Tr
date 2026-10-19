package tr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c, err := LoadCache(path)
	if err != nil {
		t.Fatalf("LoadCache: %v", err)
	}
	c.Put("scope", "en", "zh", "hello", "你好")
	if err := c.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	reloaded, err := LoadCache(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got, ok := reloaded.Get("scope", "en", "zh", "hello"); !ok || got != "你好" {
		t.Fatalf("Get = %q, %v", got, ok)
	}
}

func TestCacheScopeIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c, _ := LoadCache(path)
	c.Put("model-a", "en", "zh", "hello", "你好")
	if _, ok := c.Get("model-b", "en", "zh", "hello"); ok {
		t.Fatal("cache leaked an entry across model scopes")
	}
}

func TestCacheRejectsOversizedEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c, _ := LoadCache(path)
	big := make([]byte, cacheMaxEntryBytes+1)
	for i := range big {
		big[i] = 'a'
	}
	c.Put("scope", "en", "zh", string(big), "x")
	if n := c.Stats(); n != 0 {
		t.Fatalf("oversized entry was cached: %d", n)
	}
}

func TestCacheGetUpdatesLRU(t *testing.T) {
	c := &Cache{entries: make(map[string]CacheEntry)}
	base := time.Now().Add(-time.Hour)
	for i := 0; i < cacheMaxSize; i++ {
		text := string(rune('a' + i%26))
		text += time.Duration(i).String()
		c.entries[cacheKey("scope", "en", "zh", text)] = CacheEntry{
			TranslatedText: text,
			Timestamp:      base.Add(time.Duration(i) * time.Second),
		}
	}
	// Touch the oldest entry, then insert one more to trigger eviction.
	oldestKey := cacheKey("scope", "en", "zh", "a0s")
	if _, ok := c.Get("scope", "en", "zh", "a0s"); !ok {
		t.Fatal("seed entry missing")
	}
	c.Put("scope", "en", "zh", "new", "new")
	if _, ok := c.Get("scope", "en", "zh", "a0s"); !ok {
		t.Fatal("recently used entry was evicted")
	}
	if _, ok := c.entries[oldestKey]; !ok {
		t.Fatal("entry key disappeared")
	}
}

func TestCacheCorruptFileIsBackedUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCache(path); err == nil {
		t.Fatal("expected a parse error")
	}
	if _, err := os.Stat(path + ".corrupt"); err != nil {
		t.Fatalf("corrupt file was not backed up: %v", err)
	}
}

func TestCacheVersionMismatchStartsFresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	data, _ := json.Marshal(map[string]any{"version": 1, "entries": map[string]any{}})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCache(path)
	if err != nil {
		t.Fatalf("LoadCache: %v", err)
	}
	if c.Stats() != 0 {
		t.Fatalf("old cache was not discarded: %d", c.Stats())
	}
	if _, err := os.Stat(path + ".old"); err != nil {
		t.Fatalf("old cache was not preserved: %v", err)
	}
}

func TestCacheClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c, _ := LoadCache(path)
	c.Put("scope", "en", "zh", "hello", "你好")
	c.Clear()
	if c.Stats() != 0 {
		t.Fatal("Clear did not empty the cache")
	}
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := LoadCache(path)
	if reloaded.Stats() != 0 {
		t.Fatal("cleared cache reappeared after reload")
	}
}
