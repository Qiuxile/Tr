package tr

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSplitBlocksTrailingBlank(t *testing.T) {
	lines := strings.Split("hello\nworld\n\n", "\n") // hello, world, "", ""
	blocks := splitBlocks(lines, 1500)
	if len(blocks) != 1 {
		t.Fatalf("blocks = %+v, want 1", blocks)
	}
	if blocks[0].start != 0 || blocks[0].end != 2 {
		t.Fatalf("block = %+v, want [0,2)", blocks[0])
	}
}

func TestSplitBlocksKeepsInnerBlanks(t *testing.T) {
	lines := []string{"one", "", "two", "", ""}
	blocks := splitBlocks(lines, 1500)
	if len(blocks) != 1 || blocks[0].end != 3 {
		t.Fatalf("blocks = %+v", blocks)
	}
}

func writeTempFile(t *testing.T, content string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.txt")
	out := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(in, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return in, out
}

func TestTranslateFileTrailingBlankUsesOneRequest(t *testing.T) {
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

	in, out := writeTempFile(t, "hello\nworld\n\n")
	cfg := testAIConfig(srv.URL)
	if err := TranslateFile(cfg, Options{}, nil, in, out, ""); err != nil {
		t.Fatalf("TranslateFile: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("requests = %d, want 1 (no per-line fallback)", got)
	}
	data, _ := os.ReadFile(out)
	if string(data) != "HELLO\nWORLD\n\n" {
		t.Fatalf("output = %q", string(data))
	}
}

func TestTranslateFilePreservesIndentation(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req aiChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		got = req.Messages[len(req.Messages)-1].Content
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{
				"message":       map[string]string{"content": strings.ToUpper(req.Messages[len(req.Messages)-1].Content)},
				"finish_reason": "stop",
			}},
		})
	}))
	defer srv.Close()

	in, out := writeTempFile(t, "  hello\n")
	cfg := testAIConfig(srv.URL)
	if err := TranslateFile(cfg, Options{}, nil, in, out, ""); err != nil {
		t.Fatalf("TranslateFile: %v", err)
	}
	if got != "  hello" {
		t.Fatalf("request = %q, want leading spaces preserved", got)
	}
	data, _ := os.ReadFile(out)
	if string(data) != "  HELLO\n" {
		t.Fatalf("output = %q", string(data))
	}
}

func TestTranslateFileAbortsOnFatalError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"message":"bad key"}}`)
	}))
	defer srv.Close()

	in, out := writeTempFile(t, "one\ntwo\nthree\n")
	cfg := testAIConfig(srv.URL)
	err := TranslateFile(cfg, Options{}, nil, in, out, "")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("requests = %d, want 1 (must not fall back per line)", got)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatal("a failed fatal block should not write an output file")
	}
}

func TestTranslateFileRejectsInvalidUTF8(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(in, []byte{0xff, 0xfe, 0x00}, 0o600); err != nil {
		t.Fatal(err)
	}
	err := TranslateFile(DefaultConfig(), Options{Offline: true}, nil, in, "", "")
	if err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("error = %v", err)
	}
}

func TestTranslateFilePreservesBOMAndCRLF(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req aiChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{
				"message":       map[string]string{"content": strings.ToUpper(req.Messages[len(req.Messages)-1].Content)},
				"finish_reason": "stop",
			}},
		})
	}))
	defer srv.Close()

	in, out := writeTempFile(t, "\ufeffhello\r\nworld\r\n")
	cfg := testAIConfig(srv.URL)
	if err := TranslateFile(cfg, Options{}, nil, in, out, ""); err != nil {
		t.Fatalf("TranslateFile: %v", err)
	}
	data, _ := os.ReadFile(out)
	want := "\ufeffHELLO\r\nWORLD\r\n"
	if string(data) != want {
		t.Fatalf("output = %q, want %q", string(data), want)
	}
}
