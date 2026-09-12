package tr

import (
	"strings"
	"testing"
)

// TestSplitArrowStrictArguments pins down the arrow syntax: the translation
// input is the first argument, the language follows the arrow, and anything
// else on the command line is ignored.
func TestSplitArrowStrictArguments(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantText string
		wantLang string
		wantOK   bool
	}{
		{"spaced", []string{"你好", "->", "en"}, "你好", "en", true},
		{"quoted content", []string{"hello world", "->", "zh"}, "hello world", "zh", true},
		{"extra args ignored", []string{"你好", "->", "en", "extra", "junk"}, "你好", "en", true},
		{"only first arg is content", []string{"hello", "world", "->", "en"}, "hello", "en", true},
		{"glued in first arg", []string{"你好->en"}, "你好", "en", true},
		{"glued with extra args", []string{"你好->en", "junk"}, "你好", "en", true},
		{"unicode arrow glued", []string{"你好→ja"}, "你好", "ja", true},
		{"prefix arrow for stdin", []string{"->", "en"}, "", "en", true},
		{"prefix arrow glued", []string{"->en"}, "", "en", true},
		{"alias language", []string{"hello", "->", "中文"}, "hello", "zh", true},
		{"no arrow joins all args", []string{"hello", "world"}, "hello world", "", false},
		{"plain text keeps arrows", []string{"a->b", "hello"}, "a->b hello", "", false},
		{"unusable arrow falls back", []string{"hello", "->", "world"}, "hello -> world", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text, code, ok := splitArrow("en", tc.args)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (text=%q code=%q)", ok, tc.wantOK, text, code)
			}
			if text != tc.wantText {
				t.Fatalf("text = %q, want %q", text, tc.wantText)
			}
			if code != tc.wantLang {
				t.Fatalf("lang = %q, want %q", code, tc.wantLang)
			}
		})
	}
}

// TestSplitLongText checks the chunker used to keep requests inside the model
// context window.
func TestSplitLongText(t *testing.T) {
	// Short text stays a single chunk.
	if got := splitLongText("hello", 100); len(got) != 1 || got[0] != "hello" {
		t.Fatalf("short text = %q", got)
	}

	// Sentence boundaries are preferred, and no chunk exceeds the limit.
	text := strings.Repeat("This is a sentence. ", 60) // ~1200 chars
	chunks := splitLongText(text, 100)
	if len(chunks) < 2 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if n := len([]rune(c)); n > 100 {
			t.Fatalf("chunk %d has %d runes", i, n)
		}
	}
	if joined := strings.Join(chunks, ""); joined != text {
		t.Fatalf("chunks do not reassemble the input (%d vs %d bytes)", len(joined), len(text))
	}

	// A single oversized unit is hard-split.
	long := strings.Repeat("字", 250) // no boundaries at all
	chunks = splitLongText(long, 100)
	if len(chunks) != 3 {
		t.Fatalf("hard split produced %d chunks, want 3", len(chunks))
	}
	for _, c := range chunks {
		if n := len([]rune(c)); n > 100 {
			t.Fatalf("hard-split chunk has %d runes", n)
		}
	}
	if joined := strings.Join(chunks, ""); joined != long {
		t.Fatal("hard-split chunks do not reassemble the input")
	}

	// Paragraph structure is preserved in the reassembled text.
	para := strings.Repeat("段一。", 40) + "\n\n" + strings.Repeat("段二。", 40)
	chunks = splitLongText(para, 80)
	if joined := strings.Join(chunks, ""); joined != para {
		t.Fatal("paragraph text was not preserved")
	}
}
