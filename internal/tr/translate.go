package tr

import (
	"errors"
	"fmt"
	"strings"
)

// Error sentinels for the translation pipeline.
var (
	ErrEmptyInput     = errors.New("input text is empty")
	ErrNoTranslation  = errors.New("no translation available")
	ErrOfflineNoCache = errors.New("offline mode: text not found in cache or dictionary")
	ErrNetworkFailed  = errors.New("network request failed")
	ErrAIFailed       = errors.New("ai translation failed")
	ErrNoAIKey        = errors.New("no ai key configured")
)

// Options carries per-invocation runtime flags (not persisted to config).
type Options struct {
	Offline bool   // force offline mode; skip AI calls
	Source  string // override source language; "auto" = detect from text
}

// TranslateResult holds the completed translation and metadata about its source.
type TranslateResult struct {
	Text   string // translated text
	Source string // "ai", "cache", or "dictionary"
}

// Translate is the top-level translation pipeline.
//
// Order: cache -> AI -> error. Offline mode (-o) skips the network entirely and
// serves from the cache and the built-in dictionary instead.
//
// target is the per-call target language:
//
//   - empty  → the configured target_lang is used (the config file is read);
//   - set    → that value is used as-is and the configured target is not
//     consulted at all, so a "-> <lang>" invocation neither reads nor
//     writes the configured target.
//
// opts.Source works the same way for the source language, with "auto"
// detecting the language from the text itself.
//
// The cache is consulted first on purpose: a repeated translation must never
// cost a request, and with an AI backend that means it must never cost tokens.
func Translate(cfg Config, opts Options, cache *Cache, text, target string) (TranslateResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return TranslateResult{}, ErrEmptyInput
	}

	source := cfg.SourceLang
	if opts.Source != "" {
		source = opts.Source
	}
	if source == "auto" {
		source = DetectLang(text)
	}
	if target == "" {
		target = cfg.TargetLang
	}

	normalized := normalizeCacheKey(text)
	lang := cfg.UILang

	// Step 1: cache — free, instant, and saves tokens for repeated text.
	if cache != nil {
		if cached, ok := cache.Get(source, target, normalized); ok {
			return TranslateResult{Text: cached, Source: "cache"}, nil
		}
	}

	if !opts.Offline {
		// Step 2: AI. Translation goes through the AI backend only; without a
		// key the user gets an actionable error instead of a silent downgrade
		// to a different engine.
		if !cfg.HasAI() {
			return TranslateResult{}, ErrNoAIKey
		}
		result, err := translateTextViaAI(cfg, text, source, target, lang)
		if err != nil {
			return TranslateResult{}, fmt.Errorf("%w: %v", ErrAIFailed, err)
		}
		if cache != nil {
			cache.Put(source, target, normalized, result)
		}
		return TranslateResult{Text: result, Source: "ai"}, nil
	}

	// Step 3 (offline only): built-in dictionary (en-zh, word level).
	if source == "en" && target == "zh" {
		if dictResult, ok := DictLookup(text); ok {
			return TranslateResult{Text: dictResult, Source: "dictionary"}, nil
		}
	}
	return TranslateResult{}, ErrOfflineNoCache
}

// normalizeCacheKey lowercases and collapses whitespace for consistent cache keys.
func normalizeCacheKey(s string) string {
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}

// errorDetail strips the internal sentinel prefix from a pipeline error so the
// remaining, already localized message can be shown on its own.
func errorDetail(err error) string {
	s := err.Error()
	for _, sentinel := range []error{ErrAIFailed, ErrNetworkFailed, ErrNoTranslation} {
		if prefix := sentinel.Error() + ": "; strings.HasPrefix(s, prefix) {
			return strings.TrimPrefix(s, prefix)
		}
	}
	return s
}
