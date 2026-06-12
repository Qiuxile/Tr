package tr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Error sentinels for the translation pipeline.
var (
	ErrEmptyInput     = errors.New("input text is empty")
	ErrNoTranslation  = errors.New("no translation available")
	ErrOfflineNoCache = errors.New("offline mode: text not found in cache or dictionary")
	ErrNetworkFailed  = errors.New("network request failed")
)

// Options carries per-invocation runtime flags (not persisted to config).
type Options struct {
	Offline bool // force offline mode; skip API calls
}

// TranslateResult holds the completed translation and metadata about its source.
type TranslateResult struct {
	Text   string // translated text
	Source string // "api", "cache", or "dictionary"
}

// httpClient is a shared HTTP client with a sensible timeout.
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// Translate is the top-level translation pipeline.
// It tries: online API -> cache -> dictionary -> error.
func Translate(cfg Config, opts Options, cache *Cache, text string) (TranslateResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return TranslateResult{}, ErrEmptyInput
	}

	normalized := normalizeCacheKey(text)
	lang := cfg.UILang

	// Step 1: Try online API (unless offline mode forced)
	if !opts.Offline && cfg.ApiURL != "None" {
		result, err := translateWithGenericAPI(text, cfg.SourceLang, cfg.TargetLang, cfg.ApiURL, lang)
		if err == nil {
			if cache != nil {
				cache.Put(cfg.SourceLang, cfg.TargetLang, normalized, result)
			}
			return TranslateResult{Text: result, Source: "api"}, nil
		}
	}
	if !opts.Offline && cfg.ApiURL == "None" {
		result, err := translateWithMyMemory(text, cfg.SourceLang, cfg.TargetLang, lang)
		if err == nil {
			if cache != nil {
				cache.Put(cfg.SourceLang, cfg.TargetLang, normalized, result)
			}
			return TranslateResult{Text: result, Source: "api"}, nil
		}
	}

	// Step 2: Try cache
	if cache != nil {
		if cached, ok := cache.Get(cfg.SourceLang, cfg.TargetLang, normalized); ok {
			return TranslateResult{Text: cached, Source: "cache"}, nil
		}
	}

	// Step 3: Try dictionary (en-zh only, word-level)
	if cfg.SourceLang == "en" && cfg.TargetLang == "zh" {
		if dictResult, ok := DictLookup(text); ok {
			return TranslateResult{Text: dictResult, Source: "dictionary"}, nil
		}
	}

	// Step 4: Nothing worked
	if opts.Offline {
		return TranslateResult{}, ErrOfflineNoCache
	}
	return TranslateResult{}, fmt.Errorf("%w: %s", ErrNoTranslation, T(lang, "api.no_result"))
}

// normalizeCacheKey lowercases and collapses whitespace for consistent cache keys.
func normalizeCacheKey(s string) string {
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}
