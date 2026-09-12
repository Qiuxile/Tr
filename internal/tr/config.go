package tr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the persistent application configuration.
type Config struct {
	SourceLang string `json:"source_lang"` // e.g. "en"
	TargetLang string `json:"target_lang"` // e.g. "zh"
	UILang     string `json:"ui_lang"`     // UI display language: "zh", "en", "ja" (default "zh")

	APIKey     string `json:"api_key"`     // AI key ("sk-..."); required for translation
	AIBaseURL  string `json:"ai_base_url"` // OpenAI-compatible base URL
	AIModel    string `json:"ai_model"`    // model id, e.g. "deepseek-flash"
	AIThinking bool   `json:"ai_thinking"` // enable the model's reasoning mode (costs extra tokens; off by default)
}

// DefaultConfig returns a Config with standard defaults (en -> zh, DeepSeek AI
// backend without a key yet, zh UI).
func DefaultConfig() Config {
	return Config{
		SourceLang: "en",
		TargetLang: "zh",
		UILang:     "zh",
		AIBaseURL:  DefaultAIBaseURL,
		AIModel:    DefaultAIModel,
		AIThinking: false,
	}
}

// DefaultAIBaseURL is the OpenAI-compatible endpoint used by default.
const DefaultAIBaseURL = "https://api.deepseek.com/v1"

// DefaultAIModel is the model used by default (fast and cheap, good enough for
// translation; no reasoning pass needed).
const DefaultAIModel = "deepseek-flash"

// HasAI reports whether an AI key is configured.
func (c Config) HasAI() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

// MaskKey renders a secret for display, keeping only a short prefix/suffix.
func MaskKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if len(key) <= 10 {
		return "****"
	}
	return key[:6] + "…" + key[len(key)-4:]
}

// ConfigDir returns the OS-appropriate configuration directory for Tr.
//
//	Windows: %APPDATA%\Tr
//	Linux:   ~/.config/Tr
//	macOS:   ~/Library/Application Support/Tr
func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".tr"
	}
	return filepath.Join(dir, "Tr")
}

// ConfigPath returns the full path to config.json.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// LoadConfig reads the config file. If the file does not exist, it creates one
// with default values and returns the default.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			if saveErr := cfg.Save(); saveErr != nil {
				return cfg, fmt.Errorf("create default config: %w", saveErr)
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// applyEnvAIKey fills in the AI key from the environment when the config file
// has none, so a key can stay out of the config file entirely.
//
// Checked in order: TR_API_KEY, DEEPSEEK_API_KEY, OPENAI_API_KEY.
func applyEnvAIKey(cfg Config) Config {
	if cfg.HasAI() {
		return cfg
	}
	for _, name := range []string{"TR_API_KEY", "DEEPSEEK_API_KEY", "OPENAI_API_KEY"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			cfg.APIKey = v
			return cfg
		}
	}
	return cfg
}

// Save writes the config to the config file, creating directories as needed.
func (c Config) Save() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
