package tr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTUILayoutAndView makes sure the side-by-side layout renders without
// panicking and keeps both panes on screen.
func TestTUILayoutAndView(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {40, 10}, {200, 60}} {
		m := newTUIModel(DefaultConfig(), nil)
		m.layout(size[0], size[1])
		view := m.View()
		if strings.TrimSpace(view) == "" {
			t.Fatalf("empty view at %dx%d", size[0], size[1])
		}
		if !strings.Contains(view, "Tr") {
			t.Fatalf("header missing at %dx%d", size[0], size[1])
		}
		// Both pane titles carry their language code.
		if !strings.Contains(view, "("+m.cfg.SourceLang+")") ||
			!strings.Contains(view, "("+m.cfg.TargetLang+")") {
			t.Fatalf("language titles missing at %dx%d:\n%s", size[0], size[1], view)
		}
	}
}

// TestTUIFocusRing checks that Tab toggles between the two panes.
func TestTUIFocusRing(t *testing.T) {
	m := newTUIModel(DefaultConfig(), nil)
	if m.focus != tuiFocusInput {
		t.Fatalf("initial focus = %v, want input", m.focus)
	}
	m.moveFocus(1)
	if m.focus != tuiFocusResult {
		t.Fatalf("focus = %v, want result", m.focus)
	}
	m.moveFocus(1)
	if m.focus != tuiFocusInput {
		t.Fatalf("focus = %v, want input (wrap)", m.focus)
	}
	m.moveFocus(-1)
	if m.focus != tuiFocusResult {
		t.Fatalf("focus = %v, want result (reverse)", m.focus)
	}
}

// TestTUICtrlArrowLanguageSwitch checks that the language of the focused pane
// is switched with Ctrl+←/Ctrl+→.
func TestTUICtrlArrowLanguageSwitch(t *testing.T) {
	newModel := func() *tuiModel {
		m := newTUIModel(DefaultConfig(), nil)
		m.cfg.SourceLang = "en"
		m.cfg.TargetLang = "zh"
		return &m
	}

	// Input pane focused → Ctrl+←/→ drives the source language.
	m := newModel()
	if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlRight}); m.cfg.SourceLang != "ja" {
		t.Fatalf("source language = %q after ctrl+right, want ja", m.cfg.SourceLang)
	}
	if m.cfg.TargetLang != "zh" {
		t.Fatalf("target language changed to %q unexpectedly", m.cfg.TargetLang)
	}
	if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft}); m.cfg.SourceLang != "en" {
		t.Fatalf("source language = %q after ctrl+left, want en", m.cfg.SourceLang)
	}

	// Result pane focused → the same keys drive the target language.
	m.moveFocus(1)
	if m.focus != tuiFocusResult {
		t.Fatalf("focus = %v, want result", m.focus)
	}
	if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlRight}); m.cfg.TargetLang != "en" {
		t.Fatalf("target language = %q after ctrl+right, want en", m.cfg.TargetLang)
	}
	if m.cfg.SourceLang != "en" {
		t.Fatalf("source language changed to %q unexpectedly", m.cfg.SourceLang)
	}
	if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft}); m.cfg.TargetLang != "zh" {
		t.Fatalf("target language = %q after ctrl+left, want zh", m.cfg.TargetLang)
	}
}

// TestTUITranslateKeys checks that Ctrl+T starts a translation while Enter
// keeps its normal meaning in the input area (newline).
func TestTUITranslateKeys(t *testing.T) {
	newModel := func() *tuiModel {
		m := newTUIModel(DefaultConfig(), nil)
		m.input.SetValue("hello")
		return &m
	}

	// Ctrl+T starts a translation.
	m := newModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if cmd == nil || !m.busy {
		t.Fatal("Ctrl+T did not start a translation")
	}

	// Enter inserts a newline instead of translating.
	m = newModel()
	if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}); !strings.Contains(m.input.Value(), "\n") {
		t.Fatalf("enter did not insert a newline: %q", m.input.Value())
	}
	if m.busy {
		t.Fatal("enter started a translation")
	}

	// Ctrl+T with an empty input reports the empty hint instead of translating.
	empty := newTUIModel(DefaultConfig(), nil)
	if _, cmd := empty.Update(tea.KeyMsg{Type: tea.KeyCtrlT}); cmd != nil || empty.busy {
		t.Fatal("empty input started a translation")
	}
	if empty.errText == "" {
		t.Fatal("empty input produced no hint")
	}
}

// TestTUICycleLanguage verifies the language picker cycles and wraps.
func TestTUICycleLanguage(t *testing.T) {
	m := newTUIModel(DefaultConfig(), nil)
	m.cfg.SourceLang = "en"
	m.cycleLanguage(true, 1)
	if m.cfg.SourceLang != "ja" { // list order: zh, en, ja, ...
		t.Fatalf("source language = %q, want ja", m.cfg.SourceLang)
	}
	m.cfg.TargetLang = "zh"
	m.cycleLanguage(false, -1)
	if last := commonLangCodes[len(commonLangCodes)-1]; m.cfg.TargetLang != last {
		t.Fatalf("target language wrap = %q, want %q", m.cfg.TargetLang, last)
	}
}

// TestTUIShellMode covers the Esc system shell: panes hide, output shows,
// command history works, and Esc returns to the translator.
func TestTUIShellMode(t *testing.T) {
	m := newTUIModel(DefaultConfig(), nil)
	m.layout(80, 24)

	// Translating view shows both panes.
	view := m.View()
	if !strings.Contains(view, "原文") || !strings.Contains(view, "译文") {
		t.Fatalf("translator panes missing:\n%s", view)
	}

	// Esc opens the shell: panes are hidden, the shell title takes over.
	m.openShell()
	if m.mode != tuiModeShell {
		t.Fatal("shell mode not entered")
	}
	shell := m.View()
	if strings.Contains(shell, "原文 (") || strings.Contains(shell, "译文 (") {
		t.Fatalf("translator panes still visible in shell mode:\n%s", shell)
	}
	if !strings.Contains(shell, T(m.lang, "tui.shell_title")) {
		t.Fatalf("shell title missing:\n%s", shell)
	}

	// Output is appended to the scrollback and rendered.
	m.appendShellOutput("> echo hi\nhi")
	if !strings.Contains(m.shellBuf, "hi") {
		t.Fatalf("output not buffered: %q", m.shellBuf)
	}

	// History recall walks back and forward.
	m.pushShellHistory("echo one")
	m.pushShellHistory("echo two")
	m.shellHistoryPrev()
	if got := m.command.Value(); got != "echo two" {
		t.Fatalf("history prev = %q, want %q", got, "echo two")
	}
	m.shellHistoryPrev()
	if got := m.command.Value(); got != "echo one" {
		t.Fatalf("history prev = %q, want %q", got, "echo one")
	}
	m.shellHistoryNext()
	if got := m.command.Value(); got != "echo two" {
		t.Fatalf("history next = %q, want %q", got, "echo two")
	}
	m.shellHistoryNext()
	if got := m.command.Value(); got != "" {
		t.Fatalf("history next past the end = %q, want empty", got)
	}

	// Esc with a non-empty line clears it; Esc again leaves shell mode.
	m.command.SetValue("ls")
	if _, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc}); m.mode != tuiModeShell || m.command.Value() != "" {
		t.Fatal("esc did not clear the command line first")
	}
	if _, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc}); m.mode != tuiModeNormal {
		t.Fatal("esc did not leave shell mode")
	}
	if view := m.View(); !strings.Contains(view, "原文") {
		t.Fatalf("translator panes not restored:\n%s", view)
	}
}

// TestTUIShellScrollbackCap keeps the scrollback bounded.
func TestTUIShellScrollbackCap(t *testing.T) {
	m := newTUIModel(DefaultConfig(), nil)
	for i := 0; i < shellMaxLines+200; i++ {
		m.appendShellOutput("line")
	}
	if n := strings.Count(m.shellBuf, "\n") + 1; n > shellMaxLines {
		t.Fatalf("scrollback holds %d lines, want <= %d", n, shellMaxLines)
	}
}

// TestTUIClearKeys checks the two clear shortcuts: Ctrl+L wipes the output
// (translation pane / shell scrollback) and Ctrl+U clears the input.
func TestTUIClearKeys(t *testing.T) {
	m := newTUIModel(DefaultConfig(), nil)

	// Ctrl+L clears the translation pane.
	m.result.SetContent("你好,世界")
	m.srcLabel = "在线 AI"
	if view := m.View(); !strings.Contains(view, "你好,世界") {
		t.Fatal("result content missing before clearing")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	if strings.Contains(m.View(), "你好,世界") || m.srcLabel != "" {
		t.Fatal("ctrl+l did not clear the translation pane")
	}

	// Ctrl+U clears the input text.
	m.input.SetValue("hello")
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.input.Value() != "" {
		t.Fatalf("ctrl+u did not clear the input: %q", m.input.Value())
	}

	// Ctrl+L inside the shell clears its scrollback.
	m.openShell()
	m.appendShellOutput("> echo hi\nhi")
	if m.shellBuf == "" {
		t.Fatal("shell output missing before clearing")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	if m.shellBuf != "" || strings.Contains(m.View(), "hi") {
		t.Fatal("ctrl+l did not clear the shell output")
	}
	if m.mode != tuiModeShell {
		t.Fatal("ctrl+l left shell mode")
	}
}

// TestStripANSI makes sure raw command output cannot corrupt the view.
func TestStripANSI(t *testing.T) {
	in := "\x1b[31mred\x1b[0m plain \x1b]0;title\x07end"
	if got := stripANSI(in); got != "red plain end" {
		t.Fatalf("stripANSI = %q", got)
	}
}

// TestTUICommandsNeverWriteConfig is the guard rail for the "session only"
// promise: no TUI interaction may touch the config file.
func TestTUICommandsNeverWriteConfig(t *testing.T) {
	// Redirect the config location into the workspace so a stray write is
	// detected instead of landing in the user profile.
	dir := filepath.Join(".", ".tui-test-appdata")
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	defer os.RemoveAll(dir)

	cfgPath := ConfigPath()
	m := newTUIModel(DefaultConfig(), nil)
	m.cycleLanguage(false, 3)
	m.cycleLanguage(true, -2)
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlRight}) // source language (input focused)
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft})
	m.openShell()
	m.closeShell()

	if _, err := os.Stat(cfgPath); err == nil {
		t.Fatalf("TUI wrote a config file at %s", cfgPath)
	}
}

// TestTranslateTargetOverride documents the contract of the target parameter:
// empty falls back to the configured language, a value wins over it.
func TestTranslateTargetOverride(t *testing.T) {
	cfg := DefaultConfig() // source en, target zh
	cfg.SourceLang = "en"
	cfg.TargetLang = "en" // deliberately "same as source"

	// Offline + dictionary path: en→zh only, so the configured target ("en")
	// cannot be served while an explicit "zh" can.
	opts := Options{Offline: true}

	if _, err := Translate(cfg, opts, nil, "good morning", ""); err == nil {
		t.Fatal("expected the configured target (en) to find no dictionary entry")
	}
	res, err := Translate(cfg, opts, nil, "good morning", "zh")
	if err != nil {
		t.Fatalf("explicit target failed: %v", err)
	}
	if res.Text == "" || res.Source != "dictionary" {
		t.Fatalf("unexpected result: %+v", res)
	}
	// The config value must not have been touched.
	if cfg.TargetLang != "en" {
		t.Fatalf("config target changed to %q", cfg.TargetLang)
	}
}
