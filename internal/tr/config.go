package tr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the persistent application configuration.
type Config struct {
	SourceLang string `json:"source_lang"` // e.g. "en"
	TargetLang string `json:"target_lang"` // e.g. "zh"
	ApiURL     string `json:"api_url"`     // "None" means use MyMemory; otherwise a LibreTranslate/DeepLX endpoint
	UILang     string `json:"ui_lang"`     // UI display language: "zh", "en", "ja" (default "zh")
}

// DefaultConfig returns a Config with standard defaults (en -> zh, MyMemory, zh UI).
func DefaultConfig() Config {
	return Config{
		SourceLang: "en",
		TargetLang: "zh",
		ApiURL:     "None",
		UILang:     "zh",
	}
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
