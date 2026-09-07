package tr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// Version is the application version string.
const Version = "2.0.0"

// Mode represents the CLI dispatch mode.
type Mode int

const (
	ModeTranslate Mode = iota
	ModeHelp
	ModeVersion
	ModeAbout
	ModeConfig
	ModeTUI
)

// CLIOptions captures the parsed CLI flags and mode.
type CLIOptions struct {
	Mode       Mode
	Offline    bool
	ConfigArgs []string
}

// Run is the application entry point. It parses CLI arguments,
// loads config and cache, then dispatches to the appropriate handler.
// It returns a process exit code.
func Run(args []string) int {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.defaults", err))
		cfg = DefaultConfig()
	}

	opts, remaining := parseArgs(cfg.UILang, args[1:]) // skip program name

	switch opts.Mode {
	case ModeHelp:
		printHelp(cfg.UILang)
		return 0
	case ModeVersion:
		printVersion(cfg.UILang)
		return 0
	case ModeAbout:
		printAbout(cfg.UILang)
		return 0
	case ModeConfig:
		return runConfigCommand(cfg, opts.ConfigArgs)
	case ModeTUI:
		if !isTerminal() {
			fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.no_tty_tui"))
			return 1
		}
		return runTUI(cfg)
	}

	// ModeTranslate: recognize the arrow syntax "tr <text> -> <lang>",
	// which auto-detects the source language and translates to <lang>.
	// If the target cannot be recognized as a language, the whole line is
	// treated as plain text (with a notice), so ordinary text containing
	// "->" is never mangled.
	joined := strings.TrimSpace(strings.Join(remaining, " "))
	text := joined
	arrowLang := ""
	hasArrow := false
	if m := arrowSpaced.FindStringSubmatch(joined); m != nil {
		text, arrowLang, hasArrow = m[1], m[2], true
	} else if m := arrowGlued.FindStringSubmatch(joined); m != nil {
		text, arrowLang, hasArrow = m[1], m[2], true
	}
	targetCode := ""
	if hasArrow {
		code, ok := normalizeLangArg(arrowLang)
		if !ok {
			fmt.Fprintln(os.Stderr, T(cfg.UILang, "warn.arrow_plain", arrowLang))
			text = joined
			hasArrow = false
		} else {
			targetCode = code
		}
	}

	text = strings.TrimSpace(text)
	if text == "" && isTerminal() {
		// Interactive terminal without arguments — launch the TUI.
		return runTUI(cfg)
	}
	if text == "" {
		// Try stdin (pipe mode)
		stdinData, readErr := io.ReadAll(os.Stdin)
		if readErr == nil && len(stdinData) > 0 {
			text = strings.TrimSpace(string(stdinData))
		}
	}
	if text == "" {
		fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.empty_input"))
		return 1
	}

	// Load cache for translation
	cache, cacheErr := NewCache()
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.cache_unavailable", cacheErr))
	}

	transOpts := Options{Offline: opts.Offline}
	if hasArrow {
		transOpts.Source = "auto"
		transOpts.Target = targetCode
	}
	result, transErr := Translate(cfg, transOpts, cache, text)
	if transErr != nil {
		printLocalizedError(cfg.UILang, transErr)
		return 1
	}
	fmt.Println(result.Text)
	return 0
}

// Arrow-syntax matchers for "tr <text> -> <lang>".
// arrowSpaced requires whitespace around the arrow ("你好 -> en").
// arrowGlued accepts the no-space form ("你好->en") and lines that begin
// with the arrow (for piped input: "-> en").
var (
	arrowSpaced = regexp.MustCompile(`^(.*?)\s+(?:->|→)\s*([^\s]+)\s*$`)
	arrowGlued  = regexp.MustCompile(`^(.*?)(?:->|→)([^\s]+)\s*$`)
)

// isTerminal reports whether stdin is an interactive terminal (tty).
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
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
// Both subcommand style ("tr config show") and legacy flag style ("tr -config show")
// are accepted.
func parseArgs(lang string, args []string) (CLIOptions, []string) {
	opts := CLIOptions{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "config", "-config", "--config", "-c":
			opts.Mode = ModeConfig
			opts.ConfigArgs = args[i+1:]
			return opts, nil // consume all remaining args
		case "tui", "-tui", "--tui":
			opts.Mode = ModeTUI
			return opts, nil
		case "-help", "--help", "-h":
			opts.Mode = ModeHelp
			return opts, nil
		case "-version", "--version", "-v":
			opts.Mode = ModeVersion
			return opts, nil
		case "-about", "--about", "-a":
			opts.Mode = ModeAbout
			return opts, nil
		case "-offline", "--offline", "-o":
			opts.Offline = true
		default:
			if strings.HasPrefix(args[i], "-") {
				// The arrow syntax starts with "-" ("->" / "->en"): keep it.
				if strings.HasPrefix(args[i], "->") {
					remaining = append(remaining, args[i])
					continue
				}
				fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.unknown_flag", args[i]))
				continue
			}
			remaining = append(remaining, args[i])
		}
	}
	return opts, remaining
}

// runConfigCommand handles the config subcommand.
func runConfigCommand(cfg Config, args []string) int {
	lang := cfg.UILang

	if len(args) == 0 {
		fmt.Println(T(lang, "config.usage_show_set"))
		return 0
	}
	switch args[0] {
	case "show":
		fmt.Println(T(lang, "config.current"))
		fmt.Println(T(lang, "config.source_lang", cfg.SourceLang))
		fmt.Println(T(lang, "config.target_lang", cfg.TargetLang))
		fmt.Println(T(lang, "config.api_url", cfg.ApiURL))
		fmt.Println(T(lang, "config.ui_lang", cfg.UILang))
		if cfg.ApiURL == "None" {
			fmt.Println(T(lang, "config.backend_mymemory"))
		} else {
			fmt.Println(T(lang, "config.backend_custom"))
		}
		fmt.Println(T(lang, "config.file", ConfigPath()))
		return 0
	case "set":
		if len(args) != 3 {
			fmt.Println(T(lang, "config.usage_set"))
			fmt.Println(T(lang, "config.valid_keys"))
			return 0
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
			fmt.Println(T(lang, "config.invalid_key", key))
			return 0
		}
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, T(lang, "config.save_failed", err))
			return 1
		}
		fmt.Println(T(lang, "config.set_ok", key, val))
		return 0
	default:
		fmt.Println(T(lang, "config.unknown_subcmd", args[0]))
		return 0
	}
}

func printHelp(lang string) {
	fmt.Println(T(lang, "app.name")) // title line
	fmt.Println()

	fmt.Println(T(lang, "help.title"))
	fmt.Println(T(lang, "help.translate"))
	fmt.Println(T(lang, "help.translate_auto"))
	fmt.Println(T(lang, "help.tui"))
	fmt.Println(T(lang, "help.tui_cmd"))
	fmt.Println(T(lang, "help.config_cmd"))
	fmt.Println(T(lang, "help.offline"))
	fmt.Println(T(lang, "help.help"))
	fmt.Println(T(lang, "help.version"))
	fmt.Println(T(lang, "help.about"))
	fmt.Println()
	fmt.Println(T(lang, "help.flags_title"))
	fmt.Println(T(lang, "help.flags_config"))
	fmt.Println(T(lang, "help.flags_help"))
	fmt.Println(T(lang, "help.flags_version"))
	fmt.Println(T(lang, "help.flags_about"))
	fmt.Println(T(lang, "help.flags_offline"))
	fmt.Println()
	fmt.Println(T(lang, "help.pipe_title"))
	fmt.Println(T(lang, "help.pipe_example"))
	fmt.Println(T(lang, "help.stdin_note"))
	fmt.Println()
	fmt.Println(T(lang, "help.config_title"))
	fmt.Println(T(lang, "help.config_source"))
	fmt.Println(T(lang, "help.config_target"))
	fmt.Println(T(lang, "help.config_api"))
	fmt.Println(T(lang, "help.config_ui"))
	fmt.Println(T(lang, "help.config_path", ConfigPath()))
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
