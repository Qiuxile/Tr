package tr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func withTempConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AIBaseURL != DefaultAIBaseURL || cfg.AIModel != DefaultAIModel {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.HasAIKey() {
		t.Fatal("default config should not embed a key in source")
	}
	if cfg.AIReady() {
		t.Fatal("BigModel default should require a key until one is configured")
	}
}

func TestLoadConfigDoesNotCreateFile(t *testing.T) {
	withTempConfig(t)
	if _, err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if _, err := os.Stat(ConfigPath()); !os.IsNotExist(err) {
		t.Fatalf("LoadConfig created a file: %v", err)
	}
}

func TestConfigSavePermissions(t *testing.T) {
	withTempConfig(t)
	cfg := DefaultConfig()
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("config mode = %o, want 600", got)
		}
		dirInfo, _ := os.Stat(ConfigDir())
		if got := dirInfo.Mode().Perm(); got != 0o700 {
			t.Fatalf("config dir mode = %o, want 700", got)
		}
	}
}

func TestEnvKeyIsNotPersisted(t *testing.T) {
	withTempConfig(t)
	t.Setenv("TR_API_KEY", "sk-env-secret")
	cfg := applyEnvAIKey(DefaultConfig())
	if !cfg.HasAIKey() || !cfg.apiKeyFromEnv {
		t.Fatal("environment key was not resolved")
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatal(err)
	}
	if onDisk["api_key"] != "" {
		t.Fatalf("environment key leaked to disk: %v", onDisk["api_key"])
	}
}

func TestEnvKeyOverridesFileKey(t *testing.T) {
	withTempConfig(t)
	cfg := DefaultConfig()
	cfg.APIKey = "sk-file"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TR_API_KEY", "sk-env")
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	loaded = applyEnvAIKey(loaded)
	if loaded.APIKey != "sk-env" || !loaded.apiKeyFromEnv {
		t.Fatalf("env did not override file: %q (from env %v)", loaded.APIKey, loaded.apiKeyFromEnv)
	}
}

func TestDefaultKeyNotPersisted(t *testing.T) {
	withTempConfig(t)
	old := DefaultAPIKey
	DefaultAPIKey = "sk-built-in"
	defer func() { DefaultAPIKey = old }()

	cfg := applyDefaultAIKey(DefaultConfig())
	if !cfg.HasAIKey() || !cfg.apiKeyFromDefault {
		t.Fatal("built-in key was not applied")
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatal(err)
	}
	if onDisk["api_key"] != "" {
		t.Fatalf("built-in key leaked to disk: %v", onDisk["api_key"])
	}
}

func TestConfigOverwrite(t *testing.T) {
	withTempConfig(t)
	cfg := DefaultConfig()
	cfg.TargetLang = "ja"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	cfg.TargetLang = "en"
	if err := cfg.Save(); err != nil {
		t.Fatalf("second save: %v", err)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TargetLang != "en" {
		t.Fatalf("target = %q", loaded.TargetLang)
	}
}

func TestConfigCorruptFileIsBackedUp(t *testing.T) {
	withTempConfig(t)
	if err := os.MkdirAll(filepath.Dir(ConfigPath()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ConfigPath(), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected parse error")
	}
	if _, err := os.Stat(ConfigPath() + ".bak"); err != nil {
		t.Fatalf("corrupt config was not backed up: %v", err)
	}
}

func TestValidateBaseURL(t *testing.T) {
	valid := []string{
		"https://text.pollinations.ai/openai",
		"http://127.0.0.1:11434/v1",
		"https://api.deepseek.com/v1",
	}
	for _, raw := range valid {
		if err := validateBaseURL(raw); err != nil {
			t.Errorf("validateBaseURL(%q) = %v", raw, err)
		}
	}
	for _, raw := range []string{"", "ftp://example.com", "example.com", "https://"} {
		if err := validateBaseURL(raw); err == nil {
			t.Errorf("validateBaseURL(%q) unexpectedly succeeded", raw)
		}
	}
}

func TestNormalizeLangValue(t *testing.T) {
	cases := []struct {
		in     string
		auto   bool
		want   string
		wantOK bool
	}{
		{"en", false, "en", true},
		{"中文", false, "zh", true},
		{"zh-TW", false, "zh-hant", true},
		{"auto", true, "auto", true},
		{"auto", false, "", false},
		{"not-a-language", false, "", false},
	}
	for _, tc := range cases {
		got, ok := normalizeLangValue(tc.in, tc.auto)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("normalizeLangValue(%q, %v) = %q, %v; want %q, %v", tc.in, tc.auto, got, ok, tc.want, tc.wantOK)
		}
	}
}
