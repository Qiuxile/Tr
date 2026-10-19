package tr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// Version is the application version. It is a variable so release builds can
// inject the tag:
//
//	go build -ldflags "-X github.com/Qiuxile/Tr/internal/tr.Version=2.2.0"
var Version = "2.2.0"

// Mode represents the CLI dispatch mode.
type Mode int

const (
	ModeTranslate Mode = iota
	ModeHelp
	ModeVersion
	ModeAbout
	ModeConfig
	ModeCache
	ModeFile
	ModeTUI
)

// CLIOptions captures the parsed CLI flags and mode.
type CLIOptions struct {
	Mode       Mode
	Offline    bool
	ConfigArgs []string
	CacheArgs  []string
	FileIn     string // -f: source file
	FileOut    string // -f: output file or directory (optional)
	ParseError bool   // an unknown flag or malformed argument was seen
}

const maxStdinBytes = 16 << 20 // 16 MiB

// Run is the application entry point. It parses CLI arguments, loads config
// and cache, then dispatches to the appropriate handler. It returns a process
// exit code.
func Run(args []string) int {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.defaults", err))
		cfg = DefaultConfig()
	}

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
		// Config commands operate on the file only. Applying an environment
		// key here would write that secret back to disk on an unrelated set.
		return runConfigCommand(cfg, opts.ConfigArgs)
	case ModeCache:
		return runCacheCommand(cfg, opts.CacheArgs)
	}

	// From here on a runtime key may come from the environment or the
	// built-in default.
	cfg = applyEnvAIKey(cfg)
	cfg = applyDefaultAIKey(cfg)

	switch opts.Mode {
	case ModeFile:
		return runFileMode(cfg, opts, remaining)
	case ModeTUI:
		if !isStdinTTY() || !isStdoutTTY() {
			fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.no_tty_tui"))
			return 1
		}
		return runTUI(cfg)
	}

	// ModeTranslate: recognize the arrow syntax "tr <text> -> <lang>",
	// which auto-detects the source language and translates to <lang>.
	text, targetCode, hasArrow := splitArrow(cfg.UILang, remaining)

	if strings.TrimSpace(text) == "" && !isStdinTTY() {
		// Pipe mode: take the text from stdin.
		stdinData, readErr := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes+1))
		if readErr == nil && len(stdinData) > 0 {
			if len(stdinData) > maxStdinBytes {
				fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "err.stdin_too_large", maxStdinBytes/1024/1024))
				return 1
			}
			text = strings.TrimRight(string(stdinData), "\r\n")
		}
	}
	if strings.TrimSpace(text) == "" {
		// Only a bare `tr` (no arguments at all) opens the TUI. Arguments
		// that carried no translatable text are a usage error instead of a
		// surprise interactive session.
		if len(args) <= 1 && isStdinTTY() && isStdoutTTY() {
			return runTUI(cfg)
		}
		fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.empty_input"))
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, T(cfg.UILang, "err.usage_hint"))
			return 2
		}
		return 1
	}

	cache, cacheErr := NewCache()
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.cache_unavailable", cacheErr))
	}
	if cache != nil {
		defer func() {
			if err := cache.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.cache_save", err))
			}
		}()
	}

	transOpts := Options{
		Offline:  opts.Offline,
		Ctx:      context.Background(),
		Progress: stderrProgress(cfg.UILang),
	}
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
	fmt.Print(result.Text)
	if !strings.HasSuffix(result.Text, "\n") {
		fmt.Println()
	}
	return 0
}

// splitArrow applies the arrow syntax to the collected arguments.
//
// Arrow mode is strict and positional:
//
//	tr <content> -> <lang> [ignored...]
//
// The translation input is exactly the first argument (quote it if it contains
// spaces), the target language is the token following the arrow, and every
// other argument is ignored. A glued form inside the first argument
// ("你好->en") is accepted too, as is a leading arrow for piped input
// ("-> en" with the text on stdin).
//
// Without an arrow the whole argument list is joined, so `tr hello world`
// still translates both words.
//
// The clause is a one-shot override for this single invocation: it only affects
// the Options handed to the pipeline and never writes the config file.
//
// If an arrow is present but the target cannot be recognized as a language, the
// whole line is treated as plain text (with a notice), so ordinary text
// containing "->" is never mangled.
func splitArrow(lang string, args []string) (text string, targetCode string, ok bool) {
	if len(args) == 0 {
		return "", "", false
	}
	plain := strings.TrimSpace(strings.Join(args, " "))

	// Arrow inside the first argument: "你好->en" / "你好→en".
	for _, arrow := range []string{"->", "→"} {
		idx := strings.Index(args[0], arrow)
		if idx <= 0 {
			continue
		}
		code := strings.TrimSpace(args[0][idx+len(arrow):])
		if c, recognized := normalizeLangArg(code); recognized {
			return strings.TrimSpace(args[0][:idx]), c, true
		}
	}

	// Arrow as its own argument (possibly glued to the language: "->en").
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !isArrowArgument(arg) {
			continue
		}
		code := ""
		switch {
		case arg == "->" || arg == "→":
			if i+1 < len(args) {
				code = args[i+1]
			}
		case strings.HasPrefix(arg, "->"):
			code = arg[2:]
		default:
			code = strings.TrimPrefix(arg, "→")
		}

		code = strings.TrimSpace(code)
		c, recognized := normalizeLangArg(code)
		if !recognized {
			fmt.Fprintln(os.Stderr, T(lang, "warn.arrow_plain", code))
			return plain, "", false
		}
		if i == 0 {
			// "tr -> en" with the text coming from stdin.
			return "", c, true
		}
		// The content is the first argument; the rest is ignored.
		return strings.TrimSpace(args[0]), c, true
	}
	return plain, "", false
}

// isArrowArgument reports whether an argument carries the arrow clause.
func isArrowArgument(arg string) bool {
	return arg == "->" || arg == "→" ||
		strings.HasPrefix(arg, "->") || strings.HasPrefix(arg, "→")
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
	if cache != nil {
		defer func() {
			if err := cache.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.cache_save", err))
			}
		}()
	}

	transOpts := Options{Offline: opts.Offline, Ctx: context.Background(), PreserveWhitespace: true}
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

func isStdinTTY() bool  { return term.IsTerminal(int(os.Stdin.Fd())) }
func isStdoutTTY() bool { return term.IsTerminal(int(os.Stdout.Fd())) }

// printLocalizedError prints a user-friendly error message based on the error sentinel.
func printLocalizedError(lang string, err error) {
	switch {
	case errors.Is(err, ErrEmptyInput):
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.empty"))
	case errors.Is(err, ErrOfflineNoCache):
		fmt.Fprintf(os.Stderr, "%s\n", T(lang, "err.offline"))
	case errors.Is(err, ErrNoAIKey):
		fmt.Fprintln(os.Stderr, T(lang, "err.no_key"))
	default:
		// AI/network/file errors already carry a localized, actionable
		// message (bad key, no balance, rate limit, ...).
		fmt.Fprintln(os.Stderr, err)
	}
}

// parseArgs parses CLI arguments into CLIOptions and remaining text arguments.
// Flags may appear anywhere; a bare "--" ends flag parsing so reserved words
// and leading dashes become translatable text.
func parseArgs(lang string, args []string) (CLIOptions, []string) {
	opts := CLIOptions{}
	var positionals []string
	fileMode := false
	literal := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !literal && arg == "--" {
			literal = true
			continue
		}

		if !literal && strings.HasPrefix(arg, "-") && !isArrowArgument(arg) {
			switch arg {
			case "-help", "--help", "-h":
				opts.Mode = ModeHelp
			case "-version", "--version", "-v":
				opts.Mode = ModeVersion
			case "-about", "--about", "-a":
				opts.Mode = ModeAbout
			case "-offline", "--offline", "-o":
				opts.Offline = true
			case "-f", "--file":
				opts.Mode = ModeFile
				fileMode = true
			case "-config", "--config", "-c":
				opts.Mode = ModeConfig
				opts.ConfigArgs = args[i+1:]
				return opts, nil
			default:
				fmt.Fprintf(os.Stderr, "%s\n", T(lang, "warn.unknown_flag", arg))
				opts.ParseError = true
			}
			continue
		}

		positionals = append(positionals, arg)
	}

	if opts.Mode == ModeHelp || opts.Mode == ModeVersion || opts.Mode == ModeAbout || opts.Mode == ModeTUI {
		return opts, nil
	}

	// Reserved-word subcommands are recognized only when they look like one.
	if !fileMode && !literal && len(positionals) >= 2 {
		switch positionals[0] {
		case "config":
			if positionals[1] == "show" || positionals[1] == "set" || positionals[1] == "path" {
				opts.Mode = ModeConfig
				opts.ConfigArgs = positionals[1:]
				return opts, nil
			}
		case "cache":
			if positionals[1] == "clear" || positionals[1] == "stats" {
				opts.Mode = ModeCache
				opts.CacheArgs = positionals[1:]
				return opts, nil
			}
		}
	}
	if !fileMode && !literal && len(positionals) == 1 && positionals[0] == "tui" {
		opts.Mode = ModeTUI
		return opts, nil
	}

	if fileMode {
		opts.Mode = ModeFile
		if len(positionals) > 0 {
			opts.FileIn = positionals[0]
			rest := positionals[1:]
			if len(rest) > 0 && !isArrowArgument(rest[0]) {
				opts.FileOut = rest[0]
				rest = rest[1:]
			}
			if len(rest) > 0 && !isArrowArgument(rest[0]) {
				fmt.Fprintln(os.Stderr, T(lang, "err.file_usage"))
				opts.ParseError = true
				return opts, nil
			}
			return opts, rest
		}
	}

	return opts, positionals
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
	case "path":
		fmt.Println(ConfigPath())
		return 0
	case "set":
		if len(args) == 2 && args[1] == "api_key" {
			secret, err := readSecret(T(lang, "config.api_key_prompt"))
			if err != nil {
				fmt.Fprintln(os.Stderr, T(lang, "config.api_key_read_failed", err))
				return 1
			}
			cfg.APIKey = strings.TrimSpace(secret)
			cfg.apiKeyFromEnv = false
			cfg.apiKeyFromDefault = false
			if err := cfg.Save(); err != nil {
				fmt.Fprintln(os.Stderr, T(lang, "config.save_failed", err))
				return 1
			}
			fmt.Println(T(lang, "config.set_ok", "api_key", MaskKey(cfg.APIKey)))
			return 0
		}
		if len(args) != 3 {
			fmt.Println(T(lang, "config.usage_set"))
			fmt.Println(T(lang, "config.valid_keys"))
			return 2
		}
		key, val := args[1], args[2]
		if key == "api_key" && val == "-" {
			secret, err := readSecret(T(lang, "config.api_key_prompt"))
			if err != nil {
				fmt.Fprintln(os.Stderr, T(lang, "config.api_key_read_failed", err))
				return 1
			}
			val = secret
		}
		switch key {
		case "source_lang":
			code, ok := normalizeLangValue(val, true)
			if !ok {
				fmt.Println(T(lang, "config.invalid_lang", val))
				return 2
			}
			cfg.SourceLang = code
		case "target_lang":
			code, ok := normalizeLangValue(val, false)
			if !ok {
				fmt.Println(T(lang, "config.invalid_lang", val))
				return 2
			}
			cfg.TargetLang = code
		case "ui_lang":
			val = strings.ToLower(strings.TrimSpace(val))
			switch val {
			case "zh", "en", "ja":
				cfg.UILang = val
			default:
				fmt.Println(T(lang, "config.invalid_ui_lang", val))
				return 2
			}
		case "api_key":
			cfg.APIKey = strings.TrimSpace(val)
			cfg.apiKeyFromEnv = false
			cfg.apiKeyFromDefault = false
		case "ai_base_url":
			val = strings.TrimSpace(val)
			if err := validateBaseURL(val); err != nil {
				fmt.Println(T(lang, "config.invalid_url", val, err))
				return 2
			}
			cfg.AIBaseURL = val
		case "ai_model":
			val = strings.TrimSpace(val)
			if val == "" {
				fmt.Println(T(lang, "config.invalid_model"))
				return 2
			}
			cfg.AIModel = val
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
		switch key {
		case "api_key":
			shown = MaskKey(val)
		case "source_lang":
			shown = cfg.SourceLang
		case "target_lang":
			shown = cfg.TargetLang
		case "ui_lang":
			shown = cfg.UILang
		}
		fmt.Println(T(lang, "config.set_ok", key, shown))
		return 0
	default:
		fmt.Println(T(lang, "config.unknown_subcmd", args[0]))
		return 2
	}
}

// runCacheCommand handles "tr cache clear|stats".
func runCacheCommand(cfg Config, args []string) int {
	lang := cfg.UILang
	if len(args) == 0 {
		fmt.Println(T(lang, "cache.usage"))
		return 2
	}
	cache, err := NewCache()
	if err != nil {
		fmt.Fprintln(os.Stderr, T(lang, "warn.cache_unavailable", err))
	}

	switch args[0] {
	case "clear":
		cache.Clear()
		if err := cache.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, T(lang, "warn.cache_save", err))
			return 1
		}
		fmt.Println(T(lang, "cache.cleared"))
		return 0
	case "stats":
		fmt.Println(T(lang, "cache.stats", cache.Stats(), filepath.Join(ConfigDir(), cacheFileName)))
		return 0
	default:
		fmt.Println(T(lang, "cache.usage"))
		return 2
	}
}

func readSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	defer fmt.Fprintln(os.Stderr)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := term.ReadPassword(int(os.Stdin.Fd()))
		return string(data), err
	}
	data, err := io.ReadAll(io.LimitReader(os.Stdin, 4096))
	return strings.TrimSpace(string(data)), err
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
	fmt.Println(T(lang, "help.cache_cmd"))
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
	fmt.Println(T(lang, "help.flags_dashdash"))
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
	fmt.Println(T(lang, "help.arrow_args"))
	fmt.Println(T(lang, "help.long_text"))
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

// keyOrUnset renders the configured AI key (masked), the environment key
// source, or a placeholder when none is set.
func keyOrUnset(cfg Config, lang string) string {
	if env := envAIKeyName(); env != "" {
		return T(lang, "config.key_env", env)
	}
	if cfg.HasAIKey() {
		return MaskKey(cfg.APIKey)
	}
	if key := strings.TrimSpace(DefaultAPIKey); key != "" {
		return T(lang, "config.key_default", MaskKey(key))
	}
	if isKeylessEndpoint(cfg.AIBaseURL) {
		return T(lang, "config.key_free")
	}
	return T(lang, "config.key_unset")
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
