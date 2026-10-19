package tr

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testAIConfig(baseURL string) Config {
	cfg := DefaultConfig()
	cfg.AIBaseURL = baseURL
	cfg.AIModel = "test-model"
	cfg.APIKey = "sk-test"
	return cfg
}

func TestAIEndpoint(t *testing.T) {
	cases := map[string]string{
		DefaultAIBaseURL:                          DefaultAIBaseURL + "/chat/completions",
		"https://text.pollinations.ai":            "https://text.pollinations.ai/openai",
		"https://api.deepseek.com/v1":             "https://api.deepseek.com/v1/chat/completions",
		"https://api.deepseek.com/v1/":            "https://api.deepseek.com/v1/chat/completions",
		"https://api.openai.com":                  "https://api.openai.com/v1/chat/completions",
		"https://example.com/v1/chat/completions": "https://example.com/v1/chat/completions",
	}
	for in, want := range cases {
		if got := aiEndpoint(in); got != want {
			t.Errorf("aiEndpoint(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsKeylessEndpoint(t *testing.T) {
	if isKeylessEndpoint(DefaultAIBaseURL) {
		t.Fatal("BigModel requires a key")
	}
	if !isKeylessEndpoint("https://text.pollinations.ai/openai") {
		t.Fatal("Pollinations endpoint should be keyless")
	}
	if !isKeylessEndpoint("http://127.0.0.1:1234/v1") {
		t.Fatal("localhost endpoint should be keyless")
	}
	if isKeylessEndpoint("https://api.deepseek.com/v1") {
		t.Fatal("DeepSeek requires a key")
	}
}

func TestDefaultConfigRequiresKey(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HasAIKey() {
		t.Fatal("source default must not embed a key")
	}
	if cfg.AIReady() {
		t.Fatal("BigModel default should require a key")
	}
	if DefaultAIBaseURL != "https://open.bigmodel.cn/api/paas/v4" {
		t.Fatalf("unexpected default endpoint %q", DefaultAIBaseURL)
	}
	if DefaultAIModel != "glm-5.3-flash" {
		t.Fatalf("unexpected default model %q", DefaultAIModel)
	}
}

func TestBuildAIRequest(t *testing.T) {
	req := buildAIRequest(DefaultConfig(), "hello", "en", "zh")
	if req.Model != DefaultAIModel {
		t.Fatalf("model = %q", req.Model)
	}
	if req.Temperature != 0 {
		t.Fatalf("temperature = %v, want 0", req.Temperature)
	}
	if len(req.Messages) != 2 || req.Messages[0].Role != "system" {
		t.Fatalf("unexpected messages: %+v", req.Messages)
	}
	if !strings.Contains(req.Messages[0].Content, "English") ||
		!strings.Contains(req.Messages[0].Content, "Chinese") {
		t.Fatalf("system prompt lost languages: %q", req.Messages[0].Content)
	}
	if req.Thinking != nil {
		t.Fatal("thinking must not be sent to the default provider")
	}
	if req.ReasoningEffort != "low" {
		t.Fatalf("reasoning_effort = %q, want low for GLM-5.3", req.ReasoningEffort)
	}

	ds := DefaultConfig()
	ds.AIBaseURL = "https://api.deepseek.com/v1"
	req = buildAIRequest(ds, "hello", "en", "zh")
	if req.Thinking == nil || req.Thinking.Type != "disabled" {
		t.Fatalf("DeepSeek thinking = %+v, want disabled", req.Thinking)
	}
	ds.AIThinking = true
	req = buildAIRequest(ds, "hello", "en", "zh")
	if req.Thinking == nil || req.Thinking.Type != "enabled" {
		t.Fatalf("DeepSeek thinking = %+v, want enabled", req.Thinking)
	}
}

func TestTranslateWithAIRequestShape(t *testing.T) {
	var got aiChatRequest
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"content":"你好"},"finish_reason":"stop"}]}`)
	}))
	defer srv.Close()

	cfg := testAIConfig(srv.URL)
	out, err := translateWithAI(context.Background(), cfg, "Hello", "en", "zh", "en", false)
	if err != nil {
		t.Fatalf("translateWithAI: %v", err)
	}
	if out != "你好" {
		t.Fatalf("result = %q", out)
	}
	if auth != "Bearer sk-test" {
		t.Fatalf("authorization = %q", auth)
	}
	if got.Temperature != 0 {
		t.Fatalf("temperature = %v", got.Temperature)
	}
	if len(got.Messages) != 2 || got.Messages[1].Content != "Hello" {
		t.Fatalf("messages = %+v", got.Messages)
	}
}

func TestTranslatePreservesWhitespace(t *testing.T) {
	var got aiChatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		io.WriteString(w, `{"choices":[{"message":{"content":"OK"},"finish_reason":"stop"}]}`)
	}))
	defer srv.Close()

	cfg := testAIConfig(srv.URL)
	if _, err := Translate(cfg, Options{}, nil, "  hello  ", "zh"); err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if got.Messages[1].Content != "  hello  " {
		t.Fatalf("request lost whitespace: %q", got.Messages[1].Content)
	}
}

func TestLongTextKeepsChunksTogether(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		var req aiChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		content := strings.ToUpper(req.Messages[len(req.Messages)-1].Content)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{
				"message":       map[string]string{"content": content},
				"finish_reason": "stop",
			}},
		})
	}))
	defer srv.Close()

	cfg := testAIConfig(srv.URL)
	text := strings.Repeat("a", 1600)
	res, err := Translate(cfg, Options{}, nil, text, "zh")
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if want := strings.ToUpper(text); res.Text != want {
		t.Fatalf("long result = %q..., want %q...", res.Text[:40], want[:40])
	}
	if strings.Contains(res.Text, "\n") {
		t.Fatal("chunks were joined with an extra newline")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("requests = %d, want 2 (1500 + 100 chars)", got)
	}
}

func TestAIErrorsAreLocalizedAndClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"message":"bad key"}}`)
	}))
	defer srv.Close()

	cfg := testAIConfig(srv.URL)
	_, err := translateWithAI(context.Background(), cfg, "Hello", "en", "zh", "en", false)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, ErrAIFailed) {
		t.Fatalf("error is not ErrAIFailed: %v", err)
	}
	if strings.Contains(err.Error(), ErrAIFailed.Error()) {
		t.Fatalf("sentinel leaked into user message: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "AI key rejected") {
		t.Fatalf("unexpected message: %q", err.Error())
	}
}

func TestRetryOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{"error":{"message":"slow down"}}`)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer srv.Close()

	cfg := testAIConfig(srv.URL)
	start := time.Now()
	out, err := translateWithAI(context.Background(), cfg, "Hello", "en", "zh", "en", false)
	if err != nil {
		t.Fatalf("translateWithAI: %v", err)
	}
	if out != "ok" {
		t.Fatalf("result = %q", out)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
	if time.Since(start) < time.Second {
		t.Fatal("retries did not back off")
	}
}

func TestCleanAITranslationStripsControlAndFormatting(t *testing.T) {
	in := "\x1b[31m```text\nTranslation: \"你好\"\n```\x1b[0m"
	if got := cleanAITranslation(in, false); got != "你好" {
		t.Fatalf("cleanAITranslation = %q", got)
	}
}
