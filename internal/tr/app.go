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
	ModeFile
	ModeTUI
)

// CLIOptions captures the parsed CLI flags and mode.
type CLIOptions struct {
	Mode       Mode
	Offline    bool
	ConfigArgs []string
	FileIn     string // -f: source file
	FileOut    string // -f: output file or directory (optional)
	ParseError bool   // an unknown flag or malformed argument was seen
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
	cfg = applyEnvAIKey(cfg)

	opts, remaining := parseArgs(cfg.UILang, args[1:]) // skip program name

	// Bad flags must fail loudly: never fall through to translation or the TUI.
	if opts.ParseError {
		fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.bad_usage"))
		return 2
	}

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
	case ModeFile:
		return runFileMode(cfg, opts, remaining)
	case ModeTUI:
		if !isTerminal() {
			fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.no_tty_tui"))
			return 1
		}
		return runTUI(cfg)
	}

	// ModeTranslate: recognize the arrow syntax "tr <text> -> <lang>",
	// which auto-detects the source language and translates to <lang>.
	text, targetCode, hasArrow := splitArrow(cfg.UILang, remaining)

	text = strings.TrimSpace(text)
	if text == "" && !isTerminal() {
		// Pipe mode: take the text from stdin.
		stdinData, readErr := io.ReadAll(os.Stdin)
		if readErr == nil && len(stdinData) > 0 {
			text = strings.TrimSpace(string(stdinData))
		}
	}
	if text == "" {
		// Only a bare `tr` (no arguments at all) opens the TUI. Arguments
		// that carried no translatable text are a usage error instead of a
		// surprise interactive session.
		if len(args) <= 1 && isTerminal() {
			return runTUI(cfg)
		}
		fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.empty_input"))
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.usage_hint"))
			return 2
		}
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
	}
	// targetCode is empty without an arrow, which tells Translate to use the
	// configured target language; with an arrow the value is used as-is and the
	// config is neither read nor modified.
	result, transErr := Translate(cfg, transOpts, cache, text, targetCode)
	if transErr != nil {
		printLocalizedError(cfg.UILang, transErr)
		return 1
	}
	fmt.Println("")
	fmt.Print(result.Text)
	fmt.Println()
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

// splitArrow applies the arrow syntax to the collected arguments: the returned
// text is the input with the "-> <lang>" clause removed, targetCode the
// canonical target language, and ok whether a clause was accepted.
//
// The clause is a one-shot override for this single invocation: it only affects
// the Options handed to the pipeline and never writes the config file. Changing
// the persistent target language is what `tr config set target_lang` is for.
//
// If the target cannot be recognized as a language, the whole line is treated
// as plain text (with a notice), so ordinary text containing "->" is never
// mangled.
func splitArrow(lang string, args []string) (text string, targetCode string, ok bool) {
	joined := strings.TrimSpace(strings.Join(args, " "))

	match := arrowSpaced.FindStringSubmatch(joined)
	if match == nil {
		match = arrowGlued.FindStringSubmatch(joined)
	}
	if match == nil {
		return joined, "", false
	}
	code, recognized := normalizeLangArg(match[2])
	if !recognized {
		fmt.Fprintln(os.Stderr, T(lang, "warn.arrow_plain", match[2]))
		return joined, "", false
	}
	return match[1], code, true
}

// isPlainArgument reports whether an argument is a value rather than a flag or
// an arrow clause, used when taking the optional output path of -f.
func isPlainArgument(arg string) bool {
	if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "→") {
		return false
	}
	return true
}

// runFileMode translates a file (-f): to stdout when no output path was given,
// otherwise into the target file or directory.
func runFileMode(cfg Config, opts CLIOptions, remaining []string) int {
	lang := cfg.UILang
	if strings.TrimSpace(opts.FileIn) == "" {
		fmt.Fprintln(os.Stderr, T(lang, "err.file_usage"))
		return 2
	}

	cache, cacheErr := NewCache()
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.cache_unavailable", cacheErr))
	}

	transOpts := Options{Offline: opts.Offline}
	// The arrow syntax also works for files: tr -f doc.txt -> en
	target := ""
	if _, code, hasArrow := splitArrow(lang, remaining); hasArrow {
		transOpts.Source = "auto"
		target = code
	}

	if err := TranslateFile(cfg, transOpts, cache, opts.FileIn, opts.FileOut, target); err != nil {
		printLocalizedError(lang, err)
		return 1
	}
	return 0
}

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
	case errors.Is(err, ErrNoAIKey):
		fmt.Fprintln(os.Stderr, T(lang, "err.no_key"))
	case errors.Is(err, ErrAIFailed):
		// The AI error already carries a localized, actionable message
		// (bad key, no balance, rate limit, ...).
		fmt.Fprintln(os.Stderr, errorDetail(err))
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
		case "-f", "--file":
			opts.Mode = ModeFile
			if i+1 < len(args) {
				i++
				opts.FileIn = args[i]
				if i+1 < len(args) && isPlainArgument(args[i+1]) {
					i++
					opts.FileOut = args[i]
				}
			}
		default:
			if strings.HasPrefix(args[i], "-") {
				// The arrow syntax starts with "-" ("->" / "->en"): keep it.
				if strings.HasPrefix(args[i], "->") {
					remaining = append(remaining, args[i])
					continue
				}
				fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.unknown_flag", args[i]))
				opts.ParseError = true
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
		return 2
	}
	switch args[0] {
	case "show":
		fmt.Println(T(lang, "config.current"))
		fmt.Println(T(lang, "config.source_lang", cfg.SourceLang))
		fmt.Println(T(lang, "config.target_lang", cfg.TargetLang))
		fmt.Println(T(lang, "config.ui_lang", cfg.UILang))
		fmt.Println(T(lang, "config.api_key", keyOrUnset(cfg, lang)))
		fmt.Println(T(lang, "config.ai_base_url", cfg.AIBaseURL))
		fmt.Println(T(lang, "config.ai_model", cfg.AIModel))
		fmt.Println(T(lang, "config.ai_thinking", cfg.AIThinking))
		fmt.Println(T(lang, "config.file", ConfigPath()))
		return 0
	case "set":
		if len(args) != 3 {
			fmt.Println(T(lang, "config.usage_set"))
			fmt.Println(T(lang, "config.valid_keys"))
			return 2
		}
		key, val := args[1], args[2]
		switch key {
		case "source_lang":
			cfg.SourceLang = strings.ToLower(val)
		case "target_lang":
			cfg.TargetLang = strings.ToLower(val)
		case "ui_lang":
			cfg.UILang = strings.ToLower(val)
		case "api_key":
			cfg.APIKey = strings.TrimSpace(val)
		case "ai_base_url":
			cfg.AIBaseURL = strings.TrimSpace(val)
		case "ai_model":
			cfg.AIModel = strings.TrimSpace(val)
		case "ai_thinking":
			on, ok := parseBoolArg(val)
			if !ok {
				fmt.Println(T(lang, "config.invalid_bool", val))
				return 2
			}
			cfg.AIThinking = on
		default:
			fmt.Println(T(lang, "config.invalid_key", key))
			return 2
		}
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, T(lang, "config.save_failed", err))
			return 1
		}
		shown := val
		if key == "api_key" {
			shown = MaskKey(val)
		}
		fmt.Println(T(lang, "config.set_ok", key, shown))
		return 0
	default:
		fmt.Println(T(lang, "config.unknown_subcmd", args[0]))
		return 2
	}
}

func printHelp(lang string) {
	fmt.Println(T(lang, "app.name")) // title line
	fmt.Println()

	fmt.Println(T(lang, "help.title"))
	fmt.Println(T(lang, "help.translate"))
	fmt.Println(T(lang, "help.translate_auto"))
	fmt.Println(T(lang, "help.file"))
	fmt.Println(T(lang, "help.tui"))
	fmt.Println(T(lang, "help.tui_cmd"))
	fmt.Println(T(lang, "help.config_cmd"))
	fmt.Println(T(lang, "help.offline"))
	fmt.Println(T(lang, "help.help"))
	fmt.Println(T(lang, "help.version"))
	fmt.Println(T(lang, "help.about"))
	fmt.Println()
	fmt.Println(T(lang, "help.examples_title"))
	fmt.Println(T(lang, "help.ex1"))
	fmt.Println(T(lang, "help.ex2"))
	fmt.Println(T(lang, "help.ex3"))
	fmt.Println(T(lang, "help.ex4"))
	fmt.Println(T(lang, "help.ex5"))
	fmt.Println(T(lang, "help.ex6"))
	fmt.Println()
	fmt.Println(T(lang, "help.flags_title"))
	fmt.Println(T(lang, "help.flags_config"))
	fmt.Println(T(lang, "help.flags_file"))
	fmt.Println(T(lang, "help.flags_offline"))
	fmt.Println(T(lang, "help.flags_help"))
	fmt.Println(T(lang, "help.flags_version"))
	fmt.Println(T(lang, "help.flags_about"))
	fmt.Println()
	fmt.Println(T(lang, "help.pipe_title"))
	fmt.Println(T(lang, "help.pipe_example"))
	fmt.Println(T(lang, "help.stdin_note"))
	fmt.Println()
	fmt.Println(T(lang, "help.tui_title"))
	fmt.Println(T(lang, "help.tui_keys1"))
	fmt.Println(T(lang, "help.tui_keys2"))
	fmt.Println(T(lang, "help.tui_note"))
	fmt.Println()
	fmt.Println(T(lang, "help.quick_title"))
	fmt.Println(T(lang, "help.quick_key"))
	fmt.Println()
	fmt.Println(T(lang, "help.config_title"))
	fmt.Println(T(lang, "help.config_source"))
	fmt.Println(T(lang, "help.config_target"))
	fmt.Println(T(lang, "help.config_ui"))
	fmt.Println(T(lang, "help.config_key"))
	fmt.Println(T(lang, "help.config_ai_url"))
	fmt.Println(T(lang, "help.config_ai_model"))
	fmt.Println(T(lang, "help.config_ai_thinking"))
	fmt.Println(T(lang, "help.config_path", ConfigPath()))
	fmt.Println()
	fmt.Println(T(lang, "help.langs_title"))
	fmt.Println(T(lang, "help.langs_line1"))
	fmt.Println(T(lang, "help.langs_line2"))
	fmt.Println(T(lang, "help.langs_note"))
	fmt.Println(T(lang, "help.langs_once"))
	fmt.Println()
	fmt.Println(T(lang, "help.backend_title"))
	fmt.Println(T(lang, "help.backend_ai"))
	fmt.Println(T(lang, "help.backend_offline"))
	fmt.Println()
	fmt.Println(T(lang, "help.offline_title"))
	fmt.Println(T(lang, "help.offline_cache"))
	fmt.Println(T(lang, "help.offline_dict"))
	fmt.Println(T(lang, "help.offline_flag"))
	fmt.Println()
	fmt.Println(T(lang, "help.env_title"))
	fmt.Println(T(lang, "help.env_key"))
	fmt.Println()
	fmt.Println(T(lang, "help.more_title"))
	fmt.Println(T(lang, "help.more_readme"))
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

// keyOrUnset renders the configured AI key (masked) or a placeholder when none
// is set.
func keyOrUnset(cfg Config, lang string) string {
	if !cfg.HasAI() {
		return T(lang, "config.key_unset")
	}
	return MaskKey(cfg.APIKey)
}

// parseBoolArg accepts the usual spellings for boolean config values.
func parseBoolArg(v string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "on", "yes", "y", "enable", "enabled", "开", "开启":
		return true, true
	case "0", "false", "off", "no", "n", "disable", "disabled", "关", "关闭":
		return false, true
	}
	return false, false
}
