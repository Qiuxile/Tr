package tr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// aiHTTPClient is dedicated to AI requests: language models can take longer
// than plain dictionary APIs, so it gets a more generous timeout.
var aiHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

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
	Model     string      `json:"model"`
	Messages  []aiMessage `json:"messages"`
	MaxTokens int         `json:"max_tokens,omitempty"`
	Stream    bool        `json:"stream"`
	Thinking  *aiThinking `json:"thinking,omitempty"`
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
// the input, and no automatic retries.
func translateWithAI(cfg Config, text, source, target, lang string) (string, error) {
	payload, err := json.Marshal(buildAIRequest(cfg, text, source, target))
	if err != nil {
		return "", fmt.Errorf("%s: %w", T(lang, "ai.parse"), err)
	}

	req, err := http.NewRequest(http.MethodPost, aiEndpoint(cfg.AIBaseURL), bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("%s: %w", T(lang, "ai.request"), err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.APIKey))

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "ai.request"), err)
	}
	defer resp.Body.Close()

	// Cap the read: a translation reply is small, and a runaway response must
	// not be pulled into memory.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "ai.request"), err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", aiErrorMessage(resp.StatusCode, body, lang))
	}

	var out aiChatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("%s: %w", T(lang, "ai.parse"), err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", fmt.Errorf("%s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("%s", T(lang, "ai.empty"))
	}

	translation := cleanAITranslation(out.Choices[0].Message.Content)
	if translation == "" {
		return "", fmt.Errorf("%s", T(lang, "ai.empty"))
	}
	// A reply cut off at max_tokens is an incomplete translation: report it
	// instead of silently returning half a sentence.
	if out.Choices[0].FinishReason == "length" {
		return "", fmt.Errorf("%s", T(lang, "ai.truncated"))
	}
	return translation, nil
}

// aiRequestChars is the largest amount of source text sent in one request.
// Longer input is split at natural boundaries first: this keeps every request
// comfortably inside the model's context window (which is what produces a
// "maximum context length" error) and avoids truncated replies.
const aiRequestChars = 1500

// translateTextViaAI translates one piece of text, splitting long input into
// several requests when needed. The parts are joined with newlines, which is
// only visible for input that was too long to send as a whole anyway.
func translateTextViaAI(cfg Config, text, source, target, lang string) (string, error) {
	if len([]rune(text)) <= aiRequestChars {
		return translateWithAI(cfg, text, source, target, lang)
	}

	chunks := splitLongText(text, aiRequestChars)
	parts := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		if len(chunks) > 1 && isStderrTerminal() {
			fmt.Fprintf(os.Stderr, "\r%s", T(lang, "ai.progress", i+1, len(chunks)))
		}
		part, err := translateWithAI(cfg, chunk, source, target, lang)
		if err != nil {
			if isStderrTerminal() {
				fmt.Fprint(os.Stderr, "\r\033[K")
			}
			return "", err
		}
		parts = append(parts, part)
	}
	if len(chunks) > 1 && isStderrTerminal() {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}
	return strings.Join(parts, "\n"), nil
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
		MaxTokens: estimateMaxTokens(text),
		Stream:    false,
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
func cleanAITranslation(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "```") {
		s = aiFenceRe.ReplaceAllString(s, "")
	}
	s = aiPrefixRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return trimWrappingQuotes(s)
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
// endpoint can be configured as-is; otherwise "/v1/chat/completions" (or just
// "/chat/completions" for a base that already ends in "/v1") is appended.
func aiEndpoint(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = DefaultAIBaseURL
	}
	base = strings.TrimRight(base, "/")
	switch {
	case strings.HasSuffix(base, "/chat/completions"):
		return base
	case strings.HasSuffix(base, "/v1"):
		return base + "/chat/completions"
	default:
		return base + "/v1/chat/completions"
	}
}

// isDeepSeekEndpoint reports whether the backend is DeepSeek, the only one of
// the two known backends that takes a "thinking" switch.
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
	if len(detail) > 200 {
		detail = detail[:200] + "…"
	}

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

// aiLangNames maps language codes to the names a model understands best.
var aiLangNames = map[string]string{
	"zh": "Chinese (Simplified)", "en": "English", "ja": "Japanese",
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
	"sw": "Swahili", "af": "Afrikaans", "fa-IR": "Persian",
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
