package tr

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigWizardProviderPreset(t *testing.T) {
	m := newConfigWizard(DefaultConfig())
	if m.provider != 0 || m.base.Value() != DefaultAIBaseURL {
		t.Fatalf("default provider = %d, base = %q", m.provider, m.base.Value())
	}
	m.provider = 1
	m.applyProvider()
	if m.base.Value() != "https://api.deepseek.com/v1" {
		t.Fatalf("DeepSeek base = %q", m.base.Value())
	}
	if m.model.Value() != "deepseek-chat" {
		t.Fatalf("DeepSeek model = %q", m.model.Value())
	}
	m.provider = 3
	m.applyProvider()
	if m.base.Value() != "http://127.0.0.1:11434/v1" {
		t.Fatalf("local base = %q", m.base.Value())
	}
	if !isKeylessEndpoint(m.base.Value()) {
		t.Fatal("local provider should be keyless")
	}
}

func TestConfigWizardValidation(t *testing.T) {
	oldDefault := DefaultAPIKey
	DefaultAPIKey = ""
	defer func() { DefaultAPIKey = oldDefault }()

	m := newConfigWizard(DefaultConfig())
	m.base.SetValue("http://127.0.0.1:11434/v1")
	m.model.SetValue("qwen2.5:7b")
	if err := m.validateConnection(); err != "" {
		t.Fatalf("local validation: %s", err)
	}

	m.base.SetValue("ftp://example.com")
	if err := m.validateConnection(); err == "" {
		t.Fatal("invalid URL accepted")
	}

	m.base.SetValue("https://api.example.com/v1")
	m.model.SetValue("some-model")
	m.cfg.APIKey = ""
	m.key.SetValue("")
	if err := m.validateConnection(); err == "" {
		t.Fatal("missing key accepted for non-keyless endpoint")
	}
}

func TestConfigWizardSave(t *testing.T) {
	withTempConfig(t)
	m := newConfigWizard(DefaultConfig())
	m.base.SetValue("https://api.deepseek.com/v1")
	m.model.SetValue("deepseek-chat")
	m.key.SetValue("sk-wizard-test")
	m.source.SetValue("auto")
	m.target.SetValue("ja")
	m.ui.SetValue("en")
	m.save()
	if m.err != "" {
		t.Fatalf("save error: %s", m.err)
	}
	if m.step != 4 {
		t.Fatalf("step = %d, want saved", m.step)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.AIBaseURL != "https://api.deepseek.com/v1" ||
		cfg.AIModel != "deepseek-chat" ||
		cfg.APIKey != "sk-wizard-test" ||
		cfg.SourceLang != "auto" ||
		cfg.TargetLang != "ja" ||
		cfg.UILang != "en" {
		t.Fatalf("saved config = %+v", cfg)
	}
}

func TestConfigWizardPreservesExistingKey(t *testing.T) {
	withTempConfig(t)
	cfg := DefaultConfig()
	cfg.APIKey = "sk-existing"
	m := newConfigWizard(cfg)
	m.save()
	if m.err != "" {
		t.Fatalf("save error: %s", m.err)
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.APIKey != "sk-existing" {
		t.Fatalf("existing key was not preserved: %q", got.APIKey)
	}
}

func TestConfigWizardLanguageValidation(t *testing.T) {
	m := newConfigWizard(DefaultConfig())
	m.source.SetValue("not-a-language")
	if err := m.validateLanguages(); err == "" {
		t.Fatal("invalid source language accepted")
	}
	m.source.SetValue("en")
	m.target.SetValue("zh")
	m.ui.SetValue("xx")
	if err := m.validateLanguages(); err == "" {
		t.Fatal("invalid UI language accepted")
	}
}

func TestConfigWizardProviderSwitchClearsKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "sk-zhipu"
	m := newConfigWizard(cfg)
	m.provider = 1 // DeepSeek
	m.applyProvider()
	if m.cfg.APIKey != "" || m.key.Value() != "" {
		t.Fatalf("provider switch kept old key: cfg=%q input=%q", m.cfg.APIKey, m.key.Value())
	}
}

func TestConfigWizardClearKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "sk-existing"
	m := newConfigWizard(cfg)
	m.step = 1
	m.focus = 2
	m.refreshFocus()
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlK})
	if m.key.Value() != "" || m.cfg.APIKey != "" {
		t.Fatalf("Ctrl+K did not clear key: input=%q cfg=%q", m.key.Value(), m.cfg.APIKey)
	}
}

func TestConfigWizardView(t *testing.T) {
	for _, lang := range []string{"zh", "en"} {
		cfg := DefaultConfig()
		cfg.UILang = lang
		m := newConfigWizard(cfg)
		m.width, m.height = 100, 30
		view := m.View()
		if strings.TrimSpace(view) == "" || !strings.Contains(view, "Tr") {
			t.Fatalf("empty wizard view for %s", lang)
		}
		m.step = 1
		m.refreshFocus()
		if strings.TrimSpace(m.View()) == "" {
			t.Fatalf("empty connection step for %s", lang)
		}
	}
}
