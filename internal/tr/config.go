package tr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the persistent application configuration.
type Config struct {
	SourceLang string `json:"source_lang"` // e.g. "en"
	TargetLang string `json:"target_lang"` // e.g. "zh"
	UILang     string `json:"ui_lang"`     // UI display language: "zh", "en", "ja" (default "zh")

	APIKey     string `json:"api_key"`     // provider key; a default can be embedded at build time
	AIBaseURL  string `json:"ai_base_url"` // OpenAI-compatible base URL or full endpoint
	AIModel    string `json:"ai_model"`    // model id, e.g. "openai-fast"
	AIThinking bool   `json:"ai_thinking"` // enable the model's reasoning mode (costs extra tokens; off by default)

	// apiKeyFromEnv / apiKeyFromDefault record where APIKey came from. Keys
	// that were not written by the user are never persisted to config.json.
	apiKeyFromEnv     bool
	apiKeyFromDefault bool
}

// DefaultAPIKey is an optional built-in key. It is intentionally a variable so
// release builds can inject it with:
//
//	go build -ldflags "-X github.com/Qiuxile/Tr/internal/tr.DefaultAPIKey=sk-..."
//
// Leave it empty to rely on the keyless default endpoint or a user-supplied key.
var DefaultAPIKey = ""

// DefaultConfig returns a Config with standard defaults.
//
// Translation always goes through an OpenAI-compatible AI backend. The default
// endpoint is Zhipu BigModel (GLM); a user-provided or built-in key is required
// unless a keyless endpoint (e.g. a local model) is configured.
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

const (
	// DefaultAIBaseURL is the OpenAI-compatible endpoint used by default.
	// Zhipu BigModel exposes chat completions at /api/paas/v4/chat/completions.
	DefaultAIBaseURL = "https://open.bigmodel.cn/api/paas/v4"

	// DefaultAIModel is the current low-cost GLM model. Users can switch to
	// glm-5.3, glm-4.7, glm-4.6, etc. with tr config set ai_model.
	DefaultAIModel = "glm-5.3-flash"
)

// HasAIKey reports whether an AI key is configured or embedded.
func (c Config) HasAIKey() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

// AIReady reports whether translation can run: either a key is configured or
// the endpoint is a known keyless service.
func (c Config) AIReady() bool {
	return c.HasAIKey() || isKeylessEndpoint(c.AIBaseURL)
}

// CacheScope identifies the backend configuration a cached translation belongs
// to. Switching model or provider must not silently reuse old translations.
func (c Config) CacheScope() string {
	model := strings.TrimSpace(c.AIModel)
	if model == "" {
		model = DefaultAIModel
	}
	base := strings.TrimSpace(c.AIBaseURL)
	if base == "" {
		base = DefaultAIBaseURL
	}
	return base + "|" + model
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
		// Prefer a per-user fallback over the current directory, which would
		// pollute whatever project the user happens to be in.
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			return filepath.Join(home, ".tr")
		}
		return ".tr"
	}
	return filepath.Join(dir, "Tr")
}

// ConfigPath returns the full path to config.json.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// LoadConfig reads the config file. A missing file is not an error and does
// not create anything: defaults are returned and the file is only written when
// the user changes a setting. A corrupt file is backed up before defaults are
// used so the user's settings are not silently destroyed.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		backup := ConfigPath() + ".bak"
		_ = os.Rename(ConfigPath(), backup)
		return DefaultConfig(), fmt.Errorf("parse config (backed up to %s): %w", backup, err)
	}
	cfg.normalize()
	return cfg, nil
}

// normalize fills empty values with defaults so an older or hand-edited file
// cannot produce an unusable runtime configuration.
func (c *Config) normalize() {
	c.SourceLang = strings.TrimSpace(c.SourceLang)
	c.TargetLang = strings.TrimSpace(c.TargetLang)
	c.UILang = strings.TrimSpace(strings.ToLower(c.UILang))
	c.AIBaseURL = strings.TrimSpace(c.AIBaseURL)
	c.AIModel = strings.TrimSpace(c.AIModel)

	if c.SourceLang == "" {
		c.SourceLang = "en"
	}
	if c.TargetLang == "" {
		c.TargetLang = "zh"
	}
	switch c.UILang {
	case "zh", "en", "ja":
	default:
		c.UILang = "zh"
	}
	if c.AIBaseURL == "" {
		c.AIBaseURL = DefaultAIBaseURL
	}
	if c.AIModel == "" {
		c.AIModel = DefaultAIModel
	}
}

// applyEnvAIKey resolves the AI key from the environment. Environment values
// intentionally override the file so CI and shell-local secrets can replace a
// stale key without editing config.json.
//
// Checked in order: TR_API_KEY, DEEPSEEK_API_KEY, OPENAI_API_KEY.
func applyEnvAIKey(cfg Config) Config {
	for _, name := range []string{"TR_API_KEY", "ZHIPU_API_KEY", "GLM_API_KEY", "DEEPSEEK_API_KEY", "OPENAI_API_KEY"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			cfg.APIKey = v
			cfg.apiKeyFromEnv = true
			return cfg
		}
	}
	return cfg
}

// applyDefaultAIKey fills in the built-in key when neither the config file nor
// the environment supplied one. A user key always wins over the built-in key.
func applyDefaultAIKey(cfg Config) Config {
	if cfg.HasAIKey() {
		return cfg
	}
	if key := strings.TrimSpace(DefaultAPIKey); key != "" {
		cfg.APIKey = key
		cfg.apiKeyFromDefault = true
	}
	return cfg
}

// envAIKeyName reports which environment variable supplied the key, if any.
func envAIKeyName() string {
	for _, name := range []string{"TR_API_KEY", "ZHIPU_API_KEY", "GLM_API_KEY", "DEEPSEEK_API_KEY", "OPENAI_API_KEY"} {
		if strings.TrimSpace(os.Getenv(name)) != "" {
			return name
		}
	}
	return ""
}

// Save writes the config to the config file, creating directories as needed.
// Secrets that came from the environment or from the built-in default are
// never persisted.
func (c Config) Save() error {
	out := c
	if out.apiKeyFromEnv || out.apiKeyFromDefault {
		out.APIKey = ""
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := writeFileAtomic(ConfigPath(), data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// validateBaseURL accepts an http(s) URL with a host.
func validateBaseURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}

// normalizeLangValue canonicalizes a language code or alias. "auto" is only
// accepted for the source language.
func normalizeLangValue(value string, allowAuto bool) (string, bool) {
	if allowAuto && strings.EqualFold(strings.TrimSpace(value), "auto") {
		return "auto", true
	}
	return normalizeLangArg(value)
}
