package tr

import "testing"

func TestDetectLang(t *testing.T) {
	cases := map[string]string{
		"hello":     "en",
		"你好":        "zh",
		"こんにちは":     "ja",
		"漢字":        "zh",
		"안녕하세요":     "ko",
		"привет":    "ru",
		"مرحبا":     "ar",
		"γειά":      "el",
		"สวัสดี":    "th",
		"नमस्ते":    "hi",
		"12345 !!!": "en",
		"hello 你好":  "zh",
	}
	for text, want := range cases {
		if got := DetectLang(text); got != want {
			t.Errorf("DetectLang(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestNormalizeLangArgTraditionalChinese(t *testing.T) {
	for _, in := range []string{"zh-TW", "zh-HK", "zh-hant", "繁體"} {
		if got, ok := normalizeLangArg(in); !ok || got != "zh-hant" {
			t.Errorf("normalizeLangArg(%q) = %q, %v; want zh-hant", in, got, ok)
		}
	}
}

func TestDictLookup(t *testing.T) {
	if got, ok := DictLookup("hello"); !ok || got == "" {
		t.Fatalf("hello not found: %q, %v", got, ok)
	}
	if got, ok := DictLookup("good morning"); !ok || got == "" {
		t.Fatalf("good morning not found: %q, %v", got, ok)
	}
	if _, ok := DictLookup("zzz qqq vvv"); ok {
		t.Fatal("unknown phrase reported as translated")
	}
}

func TestDictLookupPreservesCodeAndPunctuation(t *testing.T) {
	got, ok := DictLookup("CLI and API: hello!")
	if !ok {
		t.Fatal("expected a partial translation")
	}
	if got != "CLI and API: 你好!" {
		t.Fatalf("DictLookup = %q", got)
	}
}
