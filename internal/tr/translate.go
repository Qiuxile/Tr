package tr

import (
	"context"
	"errors"
	"strings"
)

// Error sentinels for the translation pipeline.
var (
	ErrEmptyInput     = errors.New("input text is empty")
	ErrOfflineNoCache = errors.New("offline mode: text not found in cache or dictionary")
	ErrNetworkFailed  = errors.New("network request failed")
	ErrAIFailed       = errors.New("ai translation failed")
	ErrNoAIKey        = errors.New("no ai key configured")
)

// pipelineError carries an already-localized message while still allowing
// errors.Is to match the pipeline sentinel and the underlying cause.
type pipelineError struct {
	kind  error
	msg   string
	cause error
}

func (e *pipelineError) Error() string { return e.msg }

func (e *pipelineError) Unwrap() []error {
	if e.cause != nil {
		return []error{e.kind, e.cause}
	}
	return []error{e.kind}
}

func wrapPipeline(kind error, msg string, cause error) error {
	return &pipelineError{kind: kind, msg: msg, cause: cause}
}

// Options carries per-invocation runtime flags (not persisted to config).
type Options struct {
	Offline            bool                  // force offline mode; skip AI calls
	Source             string                // override source language; "auto" = detect from text
	Ctx                context.Context       // request cancellation; nil = background
	Progress           func(done, total int) // long-text progress callback; nil = silent
	PreserveWhitespace bool                  // keep leading/trailing spaces (file translation)
}

func (o Options) context() context.Context {
	if o.Ctx != nil {
		return o.Ctx
	}
	return context.Background()
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
//   - empty  -> the configured target_lang is used (the config file is read);
//   - set    -> that value is used as-is and the configured target is not
//     consulted at all, so a "-> <lang>" invocation neither reads nor
//     writes the configured target.
//
// opts.Source works the same way for the source language, with "auto"
// detecting the language from the text itself.
//
// The cache is consulted first on purpose: a repeated translation must never
// cost a request, and with an AI backend that means it must never cost tokens.
//
// The source text is never trimmed: file translation depends on indentation
// and leading/trailing whitespace being preserved.
func Translate(cfg Config, opts Options, cache *Cache, text, target string) (TranslateResult, error) {
	if strings.TrimSpace(text) == "" {
		return TranslateResult{}, ErrEmptyInput
	}

	source := cfg.SourceLang
	if opts.Source != "" {
		source = opts.Source
	}
	if source == "auto" {
		source = DetectLang(text)
	}
	source = canonicalLangCode(source)

	if target == "" {
		target = cfg.TargetLang
	}
	target = canonicalLangCode(target)

	scope := cfg.CacheScope()
	lang := cfg.UILang

	// Step 1: cache -- free, instant, and saves tokens for repeated text.
	if cache != nil {
		if cached, ok := cache.Get(scope, source, target, text); ok {
			return TranslateResult{Text: cached, Source: "cache"}, nil
		}
	}

	if !opts.Offline {
		if !cfg.AIReady() {
			return TranslateResult{}, ErrNoAIKey
		}
		result, cached, err := translateTextViaAI(cfg, cache, scope, text, source, target, lang, opts)
		if err == nil && source != target && strings.TrimSpace(result) == strings.TrimSpace(text) {
			err = wrapPipeline(ErrAIFailed, T(lang, "ai.echo"), nil)
		}
		if err != nil {
			return TranslateResult{}, err
		}
		// Short results are cached as a whole; long documents are cached per
		// chunk inside translateTextViaAI.
		if cache != nil && len([]rune(text)) <= aiRequestChars {
			cache.Put(scope, source, target, text, result)
		}
		origin := "ai"
		if cached {
			origin = "cache"
		}
		return TranslateResult{Text: result, Source: origin}, nil
	}

	// Step 3 (offline only): built-in dictionary (en-zh, word level).
	if source == "en" && target == "zh" {
		if dictResult, ok := DictLookup(text); ok {
			return TranslateResult{Text: dictResult, Source: "dictionary"}, nil
		}
	}
	return TranslateResult{}, ErrOfflineNoCache
}

// canonicalLangCode normalizes aliases to ISO 639-1 codes, leaving custom
// codes intact for backends that understand them.
func canonicalLangCode(code string) string {
	code = strings.TrimSpace(code)
	if c, ok := normalizeLangArg(code); ok {
		return c
	}
	return strings.ToLower(code)
}

// isFatalTranslationError reports whether retrying smaller pieces cannot help.
func isFatalTranslationError(err error) bool {
	return errors.Is(err, ErrNoAIKey) ||
		errors.Is(err, ErrAIFailed) ||
		errors.Is(err, ErrNetworkFailed) ||
		errors.Is(err, ErrEmptyInput)
}
