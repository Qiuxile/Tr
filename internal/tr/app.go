package tr

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Version is the application version string.
const Version = "1.1.0"

// Mode represents the CLI dispatch mode.
type Mode int

const (
	ModeTranslate Mode = iota
	ModeHelp
	ModeVersion
	ModeAbout
	ModeConfig
	ModeGUI
)

// CLIOptions captures the parsed CLI flags and mode.
type CLIOptions struct {
	Mode       Mode
	Offline    bool
	ConfigArgs []string
}

// Run is the application entry point. It parses CLI arguments,
// loads config and cache, then dispatches to the appropriate handler.
func Run(args []string, assetsFS embed.FS) {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.defaults", err))
		cfg = DefaultConfig()
	}

	opts, remaining := parseArgs(cfg.UILang, args[1:]) // skip program name

	switch opts.Mode {
	case ModeHelp:
		printHelp(cfg.UILang)
		return
	case ModeVersion:
		printVersion(cfg.UILang)
		return
	case ModeAbout:
		printAbout(cfg.UILang)
		return
	case ModeConfig:
		runConfigCommand(cfg, opts.ConfigArgs)
		return
	case ModeGUI:
		if err := StartGUI(cfg, assetsFS); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", T(cfg.UILang, "err.translate_failed"), err)
			os.Exit(1)
		}
		return
	}

	// ModeTranslate: get text from arguments or stdin
	text := strings.TrimSpace(strings.Join(remaining, " "))
	if text == "" {
		// Try stdin (pipe mode)
		stat, statErr := os.Stdin.Stat()
		if statErr == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			stdinData, readErr := io.ReadAll(os.Stdin)
			if readErr == nil && len(stdinData) > 0 {
				text = strings.TrimSpace(string(stdinData))
			}
		}
	}
	if text == "" {
		// No translation text provided — launch the GUI
		if err := StartGUI(cfg, assetsFS); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", T(cfg.UILang, "err.translate_failed"), err)
			os.Exit(1)
		}
		return
	}

	// Load cache for translation
	cache, cacheErr := NewCache()
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.cache_unavailable", cacheErr))
	}

	transOpts := Options{Offline: opts.Offline}
	result, transErr := Translate(cfg, transOpts, cache, text)
	if transErr != nil {
		printLocalizedError(cfg.UILang, transErr)
		os.Exit(1)
	}
	fmt.Println(result.Text)
}

// printLocalizedError prints a user-friendly error message based on the error sentinel.
func printLocalizedError(lang string, err error) {
	switch {
	case errors.Is(err, ErrEmptyInput):
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.empty"))
	case errors.Is(err, ErrOfflineNoCache):
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.offline"))
	case errors.Is(err, ErrNetworkFailed):
		fmt.Fprintf(os.Stderr, "%s: %v\n", T(lang, "err.network"), err)
	case errors.Is(err, ErrNoTranslation):
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.no_translation"))
	default:
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.translate_failed", err))
	}
}

// parseArgs parses CLI arguments into CLIOptions and remaining text arguments.
func parseArgs(lang string, args []string) (CLIOptions, []string) {
	opts := CLIOptions{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-config", "--config", "-c":
			opts.Mode = ModeConfig
			opts.ConfigArgs = args[i+1:]
			return opts, nil // consume all remaining args
		case "-help", "--help", "-h":
			opts.Mode = ModeHelp
			return opts, nil
		case "-version", "--version", "-v":
			opts.Mode = ModeVersion
			return opts, nil
		case "-about", "--about", "-a":
			opts.Mode = ModeAbout
			return opts, nil
		case "-gui", "--gui", "-g":
			opts.Mode = ModeGUI
			return opts, remaining
		case "-offline", "--offline", "-o":
			opts.Offline = true
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.unknown_flag", args[i]))
				continue
			}
			remaining = append(remaining, args[i])
		}
	}
	return opts, remaining
}

// runConfigCommand handles the -config subcommand.
func runConfigCommand(cfg Config, args []string) {
	lang := cfg.UILang

	if len(args) == 0 {
		fmt.Println(T(lang, "config.usage_show_set"))
		return
	}
	switch args[0] {
	case "show":
		fmt.Println(T(lang, "config.current"))
		fmt.Printf(T(lang, "config.source_lang")+"\n", cfg.SourceLang)
		fmt.Printf(T(lang, "config.target_lang")+"\n", cfg.TargetLang)
		fmt.Printf(T(lang, "config.api_url")+"\n", cfg.ApiURL)
		fmt.Printf(T(lang, "config.ui_lang")+"\n", cfg.UILang)
		if cfg.ApiURL == "None" {
			fmt.Println(T(lang, "config.backend_mymemory"))
		} else {
			fmt.Println(T(lang, "config.backend_custom"))
		}
		fmt.Printf(T(lang, "config.file")+"\n", ConfigPath())
	case "set":
		if len(args) != 3 {
			fmt.Println(T(lang, "config.usage_set"))
			fmt.Println(T(lang, "config.valid_keys"))
			return
		}
		key, val := args[1], args[2]
		switch key {
		case "source_lang":
			cfg.SourceLang = val
		case "target_lang":
			cfg.TargetLang = val
		case "api_url":
			cfg.ApiURL = val
		case "ui_lang":
			cfg.UILang = val
		default:
			fmt.Printf(T(lang, "config.invalid_key")+"\n", key)
			return
		}
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", T(lang, "config.save_failed", err))
		} else {
			fmt.Printf(T(lang, "config.set_ok")+"\n", key, val)
		}
	default:
		fmt.Printf(T(lang, "config.unknown_subcmd")+"\n", args[0])
	}
}

func printHelp(lang string) {
	fmt.Printf(T(lang, "app.name")+"\n\n") // title line not in template

	fmt.Println(T(lang, "help.title"))
	fmt.Println(T(lang, "help.translate"))
	fmt.Println(T(lang, "help.config_show"))
	fmt.Println(T(lang, "help.config_set"))
	fmt.Println(T(lang, "help.help"))
	fmt.Println(T(lang, "help.version"))
	fmt.Println(T(lang, "help.about"))
	fmt.Println(T(lang, "help.offline"))
	fmt.Println(T(lang, "help.gui"))
	fmt.Println()
	fmt.Println(T(lang, "help.flags_title"))
	fmt.Println(T(lang, "help.flags_config"))
	fmt.Println(T(lang, "help.flags_help"))
	fmt.Println(T(lang, "help.flags_version"))
	fmt.Println(T(lang, "help.flags_about"))
	fmt.Println(T(lang, "help.flags_offline"))
	fmt.Println(T(lang, "help.flags_gui"))
	fmt.Println()
	fmt.Println(T(lang, "help.pipe_title"))
	fmt.Println(T(lang, "help.pipe_example"))
	fmt.Println()
	fmt.Println(T(lang, "help.config_title"))
	fmt.Println(T(lang, "help.config_source"))
	fmt.Println(T(lang, "help.config_target"))
	fmt.Println(T(lang, "help.config_api"))
	fmt.Println(T(lang, "help.config_ui"))
	fmt.Printf(T(lang, "help.config_path")+"\n", ConfigPath())
	fmt.Println()
	fmt.Println(T(lang, "help.backend_title"))
	fmt.Println(T(lang, "help.backend_mm"))
	fmt.Println(T(lang, "help.backend_libre"))
	fmt.Println(T(lang, "help.backend_deepl"))
	fmt.Println()
	fmt.Println(T(lang, "help.offline_title"))
	fmt.Println(T(lang, "help.offline_cache"))
	fmt.Println(T(lang, "help.offline_dict"))
	fmt.Println(T(lang, "help.offline_flag"))
}

func printVersion(lang string) {
	fmt.Printf(T(lang, "app.version")+"\n", Version)
}

func printAbout(lang string) {
	fmt.Println(T(lang, "app.name"))
	fmt.Printf(T(lang, "about.version")+"\n", Version)
	fmt.Println(T(lang, "app.desc"))
	fmt.Println(T(lang, "app.license"))
	fmt.Println(T(lang, "app.repo"))
}
