package tr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// aiHTTPClient is dedicated to AI requests: language models can take longer
// than plain dictionary APIs, so it gets a more generous timeout.
var aiHTTPClient = &http.Client{
	Timeout: 45 * time.Second,
}

// Pollinations' anonymous free tier allows roughly one reply every 30
// seconds. Pacing client-side keeps a long file translation from turning into
// a series of 402/429 responses.
const keylessMinInterval = 30 * time.Second

var (
	keylessMu   sync.Mutex
	keylessLast time.Time
)

// Retry policy for transient failures. Translation is idempotent, so a bounded
// retry is safe; authentication and balance errors are never retried.
const (
	aiMaxRetries = 2
	aiRetryBase  = 500 * time.Millisecond
)

// aiMessage is a single chat message.
type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiThinking toggles the reasoning mode on backends that support it.
type aiThinking struct {
	Type string `json:"type"` // "enabled" | "disabled"
}

// aiChatRequest is an OpenAI-compatible chat completion request. Only the
// fields we actually need are sent, which keeps every request small.
type aiChatRequest struct {
	Model           string      `json:"model"`
	Messages        []aiMessage `json:"messages"`
	MaxTokens       int         `json:"max_tokens,omitempty"`
	Temperature     float64     `json:"temperature"`
	Stream          bool        `json:"stream"`
	Thinking        *aiThinking `json:"thinking,omitempty"`
	ReasoningEffort string      `json:"reasoning_effort,omitempty"`
}

// aiChatResponse is the subset of the response we care about.
type aiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Prefixes/models like to add even when told to output only the translation.
var (
	aiFenceRe  = regexp.MustCompile("(?s)^```[a-zA-Z0-9_-]*\\s*|\\s*```$")
	aiPrefixRe = regexp.MustCompile(`(?i)^\s*(translation|translated text|译文|翻译结果|翻译)\s*[:：]\s*`)
)

// translateWithAI translates text through an OpenAI-compatible chat API.
//
// Token thrift: one request per translation with a one-line system prompt, no
// few-shot examples, no conversation history, a reply-length cap derived from
// the input, and no reasoning unless explicitly enabled.
func translateWithAI(ctx context.Context, cfg Config, text, source, target, lang string, preserveWhitespace bool) (string, error) {
	payload, err := json.Marshal(buildAIRequest(cfg, text, source, target))
	if err != nil {
		return "", wrapPipeline(ErrAIFailed, fmt.Sprintf("%s: %v", T(lang, "ai.parse"), err), err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	for attempt := 0; attempt <= aiMaxRetries; attempt++ {
		if isPollinationsURL(cfg.AIBaseURL) {
			if err := paceKeyless(ctx); err != nil {
				return "", wrapPipeline(ErrNetworkFailed, T(lang, "err.network"), err)
			}
		}
		status, header, body, err := postChat(ctx, cfg, payload)
		if err != nil {
			if ctx.Err() != nil {
				return "", wrapPipeline(ErrNetworkFailed, T(lang, "err.network"), err)
			}
			lastErr = err
			if attempt < aiMaxRetries {
				if sleepErr := sleepContext(ctx, retryDelay(attempt+1)); sleepErr != nil {
					return "", wrapPipeline(ErrNetworkFailed, T(lang, "err.network"), sleepErr)
				}
				continue
			}
			return "", wrapPipeline(ErrNetworkFailed,
				fmt.Sprintf("%s: %v", T(lang, "err.network"), err), err)
		}

		if status == http.StatusOK {
			return parseAIChatResponse(body, lang, preserveWhitespace)
		}

		msg := aiErrorMessage(status, body, lang)
		if status == http.StatusPaymentRequired && isPollinationsURL(cfg.AIBaseURL) {
			msg = T(lang, "ai.free_limit")
		}
		if status == http.StatusTooManyRequests && isKeylessEndpoint(cfg.AIBaseURL) {
			msg += " " + T(lang, "ai.free_rate_limit")
		}
		lastErr = fmt.Errorf("%s", msg)
		if retryableStatus(status) && attempt < aiMaxRetries {
			delay := retryAfter(header)
			if delay <= 0 {
				delay = retryDelay(attempt + 1)
			}
			if sleepErr := sleepContext(ctx, delay); sleepErr != nil {
				return "", wrapPipeline(ErrNetworkFailed, T(lang, "err.network"), sleepErr)
			}
			continue
		}
		return "", wrapPipeline(ErrAIFailed, msg, lastErr)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("%s", T(lang, "ai.empty"))
	}
	return "", wrapPipeline(ErrAIFailed, lastErr.Error(), lastErr)
}

// postChat performs one HTTP request and returns the status, headers and body.
func postChat(ctx context.Context, cfg Config, payload []byte) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, aiEndpoint(cfg.AIBaseURL), bytes.NewReader(payload))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return 0, resp.Header, nil, err
	}
	return resp.StatusCode, resp.Header, body, nil
}

func parseAIChatResponse(body []byte, lang string, preserveWhitespace bool) (string, error) {
	var out aiChatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", wrapPipeline(ErrAIFailed, fmt.Sprintf("%s: %v", T(lang, "ai.parse"), err), err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", wrapPipeline(ErrAIFailed, out.Error.Message, nil)
	}
	if len(out.Choices) == 0 {
		return "", wrapPipeline(ErrAIFailed, T(lang, "ai.empty"), nil)
	}

	translation := cleanAITranslation(out.Choices[0].Message.Content, preserveWhitespace)
	if translation == "" {
		return "", wrapPipeline(ErrAIFailed, T(lang, "ai.empty"), nil)
	}
	// A reply cut off at max_tokens is an incomplete translation: report it
	// instead of silently returning half a sentence.
	if out.Choices[0].FinishReason == "length" {
		return "", wrapPipeline(ErrAIFailed, T(lang, "ai.truncated"), nil)
	}
	return translation, nil
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func retryDelay(attempt int) time.Duration {
	return time.Duration(attempt) * aiRetryBase
}

// retryAfter honours a Retry-After header (seconds or HTTP date) and caps it
// so a misbehaving server cannot stall the CLI indefinitely.
func retryAfter(header http.Header) time.Duration {
	if header == nil {
		return 0
	}
	v := strings.TrimSpace(header.Get("Retry-After"))
	if v == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
		d := time.Duration(seconds) * time.Second
		if d > 10*time.Second {
			d = 10 * time.Second
		}
		return d
	}
	if when, err := http.ParseTime(v); err == nil {
		d := time.Until(when)
		if d < 0 {
			return 0
		}
		if d > 10*time.Second {
			d = 10 * time.Second
		}
		return d
	}
	return 0
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// paceKeyless enforces a minimum interval between anonymous Pollinations
// requests. The free tier queues aggressively; without pacing a file
// translation would spend most of its time in 45-second timeouts.
func paceKeyless(ctx context.Context) error {
	keylessMu.Lock()
	wait := keylessMinInterval - time.Since(keylessLast)
	keylessMu.Unlock()
	if wait > 0 {
		if err := sleepContext(ctx, wait); err != nil {
			return err
		}
	}
	keylessMu.Lock()
	keylessLast = time.Now()
	keylessMu.Unlock()
	return nil
}

// aiRequestChars is the largest amount of source text sent in one request.
// Longer input is split at natural boundaries first: this keeps every request
// comfortably inside the model's context window (which is what produces a
// "maximum context length" error) and avoids truncated replies.
const aiRequestChars = 1500

// translateTextViaAI translates one piece of text, splitting long input into
// several requests when needed. Chunks are cached individually so editing one
// paragraph does not force every chunk to be paid for again. Chunks are joined
// without adding separators: splitLongText keeps the original punctuation and
// newlines.
//
// The second return value reports whether every chunk came from the cache.
func translateTextViaAI(cfg Config, cache *Cache, scope, text, source, target, lang string, opts Options) (string, bool, error) {
	if len([]rune(text)) <= aiRequestChars {
		result, err := translateWithAI(opts.context(), cfg, text, source, target, lang, opts.PreserveWhitespace)
		return result, false, err
	}

	chunks := splitLongText(text, aiRequestChars)
	parts := make([]string, len(chunks))
	allCached := cache != nil

	for i, chunk := range chunks {
		if opts.Progress != nil {
			opts.Progress(i+1, len(chunks))
		}
		if cache != nil {
			if cached, ok := cache.Get(scope, source, target, chunk); ok {
				parts[i] = cached
				continue
			}
		}
		allCached = false
		part, err := translateWithAI(opts.context(), cfg, chunk, source, target, lang, opts.PreserveWhitespace)
		if err != nil {
			return "", false, err
		}
		parts[i] = part
		if cache != nil {
			cache.Put(scope, source, target, chunk, part)
		}
	}
	return strings.Join(parts, ""), allCached, nil
}

// splitLongText breaks long text into chunks no larger than maxChars, cutting
// at paragraph, line and sentence boundaries before falling back to a hard cut.
func splitLongText(text string, maxChars int) []string {
	if maxChars < 1 {
		maxChars = 1
	}
	var chunks []string
	var cur strings.Builder
	curLen := 0

	flush := func() {
		if cur.Len() > 0 {
			chunks = append(chunks, cur.String())
			cur.Reset()
			curLen = 0
		}
	}

	for _, unit := range splitTextUnits(text) {
		unitLen := len([]rune(unit))
		if unitLen > maxChars {
			flush()
			for _, piece := range hardSplitRunes(unit, maxChars) {
				chunks = append(chunks, piece)
			}
			continue
		}
		if curLen > 0 && curLen+unitLen > maxChars {
			flush()
		}
		cur.WriteString(unit)
		curLen += unitLen
	}
	flush()
	return chunks
}

// splitTextUnits splits text into sentence-sized units, keeping separators.
func splitTextUnits(s string) []string {
	var units []string
	start := 0
	for i, r := range s {
		if !isTextBoundary(r) {
			continue
		}
		units = append(units, s[start:i+len(string(r))])
		start = i + len(string(r))
	}
	if start < len(s) {
		units = append(units, s[start:])
	}
	return units
}

// isTextBoundary reports whether a rune is a natural place to break a long text.
func isTextBoundary(r rune) bool {
	switch r {
	case '\n', '.', '!', '?', ';', ':':
		return true
	// Full-width CJK punctuation: 。 ! ? ; :
	case 0x3002, 0xFF01, 0xFF1F, 0xFF1B, 0xFF1A:
		return true
	}
	return false
}

// hardSplitRunes cuts a string into fixed-size rune chunks.
func hardSplitRunes(s string, maxChars int) []string {
	runes := []rune(s)
	var out []string
	for len(runes) > 0 {
		n := maxChars
		if n > len(runes) {
			n = len(runes)
		}
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}

// buildAIRequest assembles a minimal chat request for one translation.
func buildAIRequest(cfg Config, text, source, target string) aiChatRequest {
	model := strings.TrimSpace(cfg.AIModel)
	if model == "" {
		model = DefaultAIModel
	}
	req := aiChatRequest{
		Model: model,
		Messages: []aiMessage{
			{Role: "system", Content: aiSystemPrompt(source, target)},
			{Role: "user", Content: text},
		},
		MaxTokens:   estimateMaxTokens(text),
		Temperature: 0,
		Stream:      false,
	}
	// Reasoning models spend a large amount of tokens thinking about what is
	// a lookup task. Say so explicitly for the backend that supports it.
	if isDeepSeekEndpoint(cfg.AIBaseURL) {
		if cfg.AIThinking {
			req.Thinking = &aiThinking{Type: "enabled"}
		} else {
			req.Thinking = &aiThinking{Type: "disabled"}
		}
	}
	// Zhipu GLM-5.3/5.3-Flash force thinking on; reasoning_effort=low keeps the
	// per-translation reasoning budget small.
	if isZhipuEndpoint(cfg.AIBaseURL) && strings.HasPrefix(strings.ToLower(model), "glm-5.") {
		req.ReasoningEffort = "low"
	}
	return req
}

// aiSystemPrompt is deliberately one short sentence: no examples, no style
// guide, no persona. Every extra word is paid for on every request.
func aiSystemPrompt(source, target string) string {
	tgt := langDisplayName(target)
	if src := langDisplayName(source); src != "" {
		return fmt.Sprintf("Translate from %s into %s. Reply with the translation only.", src, tgt)
	}
	return fmt.Sprintf("Translate the text into %s. Reply with the translation only.", tgt)
}

// estimateMaxTokens caps reply length so a chatty model cannot burn tokens on
// explanations. CJK costs roughly one token per character and Latin text far
// less, so twice the rune count plus a margin is generous for a translation.
func estimateMaxTokens(text string) int {
	n := len([]rune(text))*2 + 32
	if n < 64 {
		n = 64
	}
	if n > 4096 {
		n = 4096
	}
	return n
}

// cleanAITranslation removes formatting the model may have added despite the
// instruction, so the cached value is the bare translation.
func cleanAITranslation(s string, preserveWhitespace bool) string {
	s = stripANSI(s)
	s = stripControlChars(s)
	if preserveWhitespace {
		// Keep indentation for file translation; only drop model-added blank
		// lines at the very edges.
		s = strings.Trim(s, "\r\n")
	} else {
		s = strings.TrimSpace(s)
	}
	if strings.Contains(s, "```") {
		s = aiFenceRe.ReplaceAllString(s, "")
	}
	s = aiPrefixRe.ReplaceAllString(s, "")
	if !preserveWhitespace {
		s = strings.TrimSpace(s)
	}
	return trimWrappingQuotes(s)
}

// stripControlChars removes escape sequences and control bytes that could
// corrupt the terminal. Newlines and tabs are preserved.
func stripControlChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\n', '\t', '\r':
			b.WriteRune(r)
		default:
			if r >= 0x20 && r != 0x7f {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// trimWrappingQuotes drops a symmetric pair of quotes around the whole text,
// which chat models add frequently. Quotes that are part of the translation
// (unbalanced, or repeated inside) are preserved.
func trimWrappingQuotes(s string) string {
	pairs := [][2]string{
		{`"`, `"`}, {`“`, `”`}, {`'`, `'`}, {`「`, `」`}, {`『`, `』`}, {`《`, `》`},
	}
	for _, p := range pairs {
		if len(s) > len(p[0])+len(p[1]) &&
			strings.HasPrefix(s, p[0]) && strings.HasSuffix(s, p[1]) {
			inner := s[len(p[0]) : len(s)-len(p[1])]
			if !strings.Contains(inner, p[1]) {
				return strings.TrimSpace(inner)
			}
		}
	}
	return s
}

// aiEndpoint normalizes a base URL into a chat completions endpoint. A full
// endpoint can be configured as-is; otherwise the appropriate suffix is
// appended. Versioned bases such as /v1 or /api/paas/v4 are handled generically.
func aiEndpoint(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = DefaultAIBaseURL
	}
	base = strings.TrimRight(base, "/")
	switch {
	case strings.HasSuffix(base, "/chat/completions"):
		return base
	case strings.HasSuffix(base, "/openai"):
		return base
	case isVersionedBase(base):
		return base + "/chat/completions"
	}
	if isPollinationsURL(base) {
		return base + "/openai"
	}
	return base + "/v1/chat/completions"
}

// isVersionedBase reports whether the URL path ends in a version segment such
// as /v1, /v4, or /api/paas/v4.
func isVersionedBase(base string) bool {
	u, err := url.Parse(base)
	if err != nil {
		return false
	}
	path := strings.TrimRight(u.Path, "/")
	idx := strings.LastIndexByte(path, '/')
	if idx < 0 {
		return false
	}
	last := path[idx+1:]
	if len(last) < 2 || last[0] != 'v' {
		return false
	}
	for _, r := range last[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isPollinationsURL(base string) bool {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return strings.Contains(strings.ToLower(base), "pollinations")
	}
	return strings.Contains(strings.ToLower(u.Hostname()), "pollinations.ai")
}

// isZhipuEndpoint reports whether the backend is Zhipu BigModel/GLM.
func isZhipuEndpoint(base string) bool {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return strings.Contains(strings.ToLower(base), "bigmodel.cn")
	}
	return strings.Contains(strings.ToLower(u.Hostname()), "bigmodel.cn")
}

// isKeylessEndpoint reports whether the endpoint works without an API key.
// Local model servers and the Pollinations playground are supported.
func isKeylessEndpoint(base string) bool {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return strings.Contains(strings.ToLower(base), "pollinations")
	}
	host := strings.ToLower(u.Hostname())
	return strings.Contains(host, "pollinations.ai") ||
		host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// isDeepSeekEndpoint reports whether the backend is DeepSeek, the only known
// backend that takes a "thinking" switch.
func isDeepSeekEndpoint(base string) bool {
	base = strings.TrimSpace(base)
	if base == "" {
		base = DefaultAIBaseURL
	}
	return strings.Contains(strings.ToLower(base), "deepseek")
}

// aiErrorMessage turns an HTTP error response into a localized, actionable
// message.
func aiErrorMessage(status int, body []byte, lang string) string {
	detail := ""
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil {
		detail = strings.TrimSpace(e.Error.Message)
	}
	if detail == "" {
		detail = strings.TrimSpace(string(body))
	}
	detail = truncateRunes(detail, 200)

	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return T(lang, "ai.bad_key", status, detail)
	case http.StatusPaymentRequired:
		return T(lang, "ai.no_balance", status, detail)
	case http.StatusTooManyRequests:
		return T(lang, "ai.rate_limit", status, detail)
	default:
		return T(lang, "ai.status", status, detail)
	}
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

// aiLangNames maps language codes to the names a model understands best.
var aiLangNames = map[string]string{
	"zh": "Chinese (Simplified)", "zh-hant": "Chinese (Traditional)",
	"en": "English", "ja": "Japanese",
	"ko": "Korean", "ru": "Russian", "fr": "French", "de": "German",
	"es": "Spanish", "it": "Italian", "pt": "Portuguese", "ar": "Arabic",
	"th": "Thai", "hi": "Hindi", "vi": "Vietnamese", "el": "Greek",
	"nl": "Dutch", "pl": "Polish", "tr": "Turkish", "uk": "Ukrainian",
	"id": "Indonesian", "ms": "Malay", "he": "Hebrew", "sv": "Swedish",
	"da": "Danish", "fi": "Finnish", "no": "Norwegian", "cs": "Czech",
	"ro": "Romanian", "hu": "Hungarian", "bg": "Bulgarian", "fa": "Persian",
	"bn": "Bengali", "ta": "Tamil", "ur": "Urdu", "sr": "Serbian",
	"hr": "Croatian", "sk": "Slovak", "sl": "Slovenian", "lt": "Lithuanian",
	"lv": "Latvian", "et": "Estonian", "ca": "Catalan", "tl": "Filipino",
	"sw": "Swahili", "af": "Afrikaans",
}

// langDisplayName returns a human readable language name for prompts, falling
// back to the raw code so custom backends still get something workable.
func langDisplayName(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" || code == "auto" {
		return ""
	}
	if name, ok := aiLangNames[code]; ok {
		return name
	}
	return code
}

// stderrProgress prints chunk progress only when stderr is a terminal. The TUI
// passes a nil callback instead: writing to stderr would corrupt its alt-screen.
func stderrProgress(lang string) func(done, total int) {
	if !isStderrTerminal() {
		return nil
	}
	return func(done, total int) {
		fmt.Fprintf(os.Stderr, "\r%s", T(lang, "ai.progress", done, total))
		if done == total {
			fmt.Fprint(os.Stderr, "\r\033[K")
		}
	}
}
