package tr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI color palette.
var (
	tuiAccent = lipgloss.Color("39")  // bright cyan
	tuiDim    = lipgloss.Color("240") // muted gray
	tuiOk     = lipgloss.Color("42")  // green
	tuiWarn   = lipgloss.Color("220") // yellow
	tuiErr    = lipgloss.Color("196") // red
)

// tuiFocus is the currently focused pane. The language pair is not a focus
// stop: it is changed directly with the Alt+arrow shortcuts.
type tuiFocus int

const (
	tuiFocusInput tuiFocus = iota
	tuiFocusResult
)

// tuiFocusCount is the number of focus stops (used for wrapping).
const tuiFocusCount = 2

// tuiMode separates the translator view from the system shell opened with Esc.
type tuiMode int

const (
	tuiModeNormal tuiMode = iota
	tuiModeShell
)

// shell settings.
const (
	// shellRunTimeout bounds one command so the UI can never hang on it.
	shellRunTimeout = 30 * time.Second
	// shellMaxLines caps the retained scrollback.
	shellMaxLines = 1000
)

// ansiRe matches colour/OSC escape sequences produced by shell commands.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b[=>]`)

// tuiTranslateMsg carries the result of an asynchronous translation.
type tuiTranslateMsg struct {
	text string
	res  TranslateResult
	err  error
}

// tuiShellDoneMsg carries the result of an asynchronous shell command.
type tuiShellDoneMsg struct {
	output   string
	err      error
	timedOut bool
}

// tuiModel is the bubbletea model for the interactive translation UI.
type tuiModel struct {
	cfg     Config
	cache   *Cache
	lang    string
	input   textarea.Model  // left pane: source text
	result  viewport.Model  // right pane: translation
	command textinput.Model // shell command line (Esc)
	shell   viewport.Model  // shell output, fills the screen in shell mode
	focus   tuiFocus
	mode    tuiMode
	width   int
	height  int

	leftWidth  int
	rightWidth int

	busy     bool
	offline  bool   // session-level offline override (Ctrl+O)
	errText  string // last error, shown in the result pane
	hint     string // transient status message (language switch)
	srcLabel string // translation source badge for the result title

	shellBuf     string   // scrollback text
	shellHist    []string // command history
	shellHistIdx int
	shellRunning bool
}

// newTUIModel builds the initial model.
func newTUIModel(cfg Config, cache *Cache) tuiModel {
	ta := textarea.New()
	ta.Placeholder = T(cfg.UILang, "tui.placeholder")
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.SetWidth(40)
	ta.SetHeight(10)

	ci := textinput.New()
	ci.Prompt = "> "
	ci.Placeholder = T(cfg.UILang, "tui.shell_input")
	ci.CharLimit = 512
	ci.PromptStyle = lipgloss.NewStyle().Foreground(tuiAccent).Bold(true)
	ci.TextStyle = lipgloss.NewStyle()

	vp := viewport.New(40, 10)
	vp.KeyMap.PageDown.SetEnabled(true)

	sp := viewport.New(80, 20)
	sp.KeyMap.PageDown.SetEnabled(true)

	m := tuiModel{
		cfg:          cfg,
		cache:        cache,
		lang:         cfg.UILang,
		input:        ta,
		result:       vp,
		command:      ci,
		shell:        sp,
		focus:        tuiFocusInput,
		width:        80,
		height:       24,
		shellHistIdx: 0,
	}
	m.layout(80, 24)
	m.result.SetContent(lipgloss.NewStyle().Foreground(tuiDim).Render(T(cfg.UILang, "app.desc")))
	m.shell.SetContent(lipgloss.NewStyle().Foreground(tuiDim).Render(T(cfg.UILang, "tui.shell_banner")))
	return m
}

// layout recalculates component sizes for a terminal size. The translator view
// puts the two panes side by side; the shell view uses the full width.
func (m *tuiModel) layout(w, h int) {
	if w < 40 {
		w = 40
	}
	if h < 10 {
		h = 10
	}
	m.width, m.height = w, h

	left := (w - 1) / 2
	m.leftWidth, m.rightWidth = left, w-1-left

	// Translator rows: header(1) + box + footer(1); each box is a title row,
	// the inner rows, and two border rows.
	inner := h - 5
	if inner < 3 {
		inner = 3
	}
	m.input.SetWidth(m.leftWidth - 2)
	m.input.SetHeight(inner)
	m.result.Width = m.rightWidth - 2
	m.result.Height = inner

	// Shell rows: title(1) + output + command line(1).
	shellRows := h - 3
	if shellRows < 3 {
		shellRows = 3
	}
	m.shell.Width = w - 2
	m.shell.Height = shellRows
	m.command.Width = w - 4
}

// Init satisfies tea.Model.
func (m tuiModel) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout(msg.Width, msg.Height)
		return m, nil

	case tuiTranslateMsg:
		m.busy = false
		if msg.err != nil {
			m.errText = T(m.lang, "tui.failed", errorDetail(msg.err))
			m.srcLabel = ""
			m.result.SetContent(lipgloss.NewStyle().Foreground(tuiErr).Render(m.errText))
			return m, nil
		}
		m.errText = ""
		m.srcLabel = tuiSourceLabel(m.lang, msg.res.Source)
		m.result.SetContent(msg.res.Text)
		m.result.GotoBottom()
		return m, nil

	case tuiShellDoneMsg:
		m.shellRunning = false
		m.appendShellOutput(shellResultText(m.lang, msg))
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey processes keyboard input. Control keys are handled first so they
// never collide with text the user is typing.
func (m *tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Shell mode owns the keyboard until it is dismissed with Esc.
	if m.mode == tuiModeShell {
		switch key {
		case "esc":
			if m.command.Value() != "" {
				m.command.SetValue("")
				return m, nil
			}
			m.closeShell()
			return m, nil
		case "enter":
			line := strings.TrimSpace(m.command.Value())
			if line == "" {
				return m, nil
			}
			m.command.SetValue("")
			m.pushShellHistory(line)
			m.appendShellOutput(lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render("> ") + line)
			m.shellRunning = true
			return m, m.shellCmd(line)
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+l": // clear the terminal output, like a shell's Ctrl+L
			m.clearShellOutput()
			return m, nil
		case "up":
			m.shellHistoryPrev()
			return m, nil
		case "down":
			m.shellHistoryNext()
			return m, nil
		}
		// Scrolling the scrollback is the only other key the output pane owns.
		if key == "pgup" || key == "pgdown" || key == "home" || key == "end" {
			var cmd tea.Cmd
			m.shell, cmd = m.shell.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.command, cmd = m.command.Update(msg)
		return m, cmd
	}

	switch key {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc": // open the system shell
		m.openShell()
		return m, nil

	case "ctrl+t": // translate
		return m.startTranslate()

	case "ctrl+l": // clear the terminal output (the translation pane)
		m.clearResult()
		return m, nil

	case "ctrl+u": // clear the input text
		m.input.Reset()
		m.hint = T(m.lang, "tui.hint_cleared")
		return m, nil

	case "ctrl+s": // swap source/target for this session only — never writes the config
		m.cfg.SourceLang, m.cfg.TargetLang = m.cfg.TargetLang, m.cfg.SourceLang
		m.hint = fmt.Sprintf("%s → %s", m.cfg.SourceLang, m.cfg.TargetLang)
		return m, nil

	case "ctrl+o": // toggle offline for this session
		m.offline = !m.offline
		return m, nil

	case "tab":
		m.moveFocus(1)
		return m, nil

	case "shift+tab":
		m.moveFocus(-1)
		return m, nil

	case "ctrl+left": // switch the language of the focused pane
		m.cycleFocusedLanguage(-1)
		return m, nil

	case "ctrl+right":
		m.cycleFocusedLanguage(1)
		return m, nil
	}

	// Route everything else to the focused pane.
	var cmd tea.Cmd
	switch m.focus {
	case tuiFocusInput:
		m.input, cmd = m.input.Update(msg)
	case tuiFocusResult:
		m.result, cmd = m.result.Update(msg)
	}
	return m, cmd
}

// startTranslate kicks off a translation of the current input.
func (m *tuiModel) startTranslate() (tea.Model, tea.Cmd) {
	if m.busy {
		return m, nil
	}
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		m.errText = T(m.lang, "tui.empty")
		m.srcLabel = ""
		m.result.SetContent(lipgloss.NewStyle().Foreground(tuiErr).Render(m.errText))
		return m, nil
	}
	m.busy = true
	m.errText = ""
	return m, m.translateCmd(text)
}

// moveFocus advances the focus ring: input → result → input.
func (m *tuiModel) moveFocus(delta int) {
	m.focus = tuiFocus((int(m.focus) + delta + tuiFocusCount) % tuiFocusCount)
	if m.focus == tuiFocusInput {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
}

// cycleFocusedLanguage steps the language of the focused pane: the source
// language while the input pane has focus, the target language for the result
// pane. Bound to Ctrl+←/Ctrl+→.
func (m *tuiModel) cycleFocusedLanguage(delta int) {
	m.cycleLanguage(m.focus == tuiFocusInput, delta)
}

// cycleLanguage steps one side of the language pair through the picker list.
// The change is session-scoped: the config file is never written.
func (m *tuiModel) cycleLanguage(isSource bool, delta int) {
	if isSource {
		m.cfg.SourceLang = nextCommonLang(m.cfg.SourceLang, delta)
		m.hint = T(m.lang, "tui.hint_source", m.cfg.SourceLang)
		return
	}
	m.cfg.TargetLang = nextCommonLang(m.cfg.TargetLang, delta)
	m.hint = T(m.lang, "tui.hint_target", m.cfg.TargetLang)
}

// clearShellOutput wipes the system shell's scrollback (Ctrl+L).
func (m *tuiModel) clearShellOutput() {
	m.shellBuf = ""
	m.shell.SetContent("")
}

// clearResult wipes the translation pane (Ctrl+L in the translator view).
func (m *tuiModel) clearResult() {
	m.result.SetContent("")
	m.srcLabel = ""
	m.errText = ""
}

// openShell hides the translator panes and shows the system shell.
func (m *tuiModel) openShell() {
	m.mode = tuiModeShell
	m.input.Blur()
	m.command.SetValue("")
	m.command.Focus()
	m.shell.GotoBottom()
}

// closeShell returns to the translator view.
func (m *tuiModel) closeShell() {
	m.mode = tuiModeNormal
	m.command.Blur()
	m.command.SetValue("")
	if m.focus == tuiFocusInput {
		m.input.Focus()
	}
}

// shellCmd runs one command line in the background.
func (m *tuiModel) shellCmd(line string) tea.Cmd {
	return func() tea.Msg {
		output, err, timedOut := runShellLine(line)
		return tuiShellDoneMsg{output: output, err: err, timedOut: timedOut}
	}
}

// runShellLine executes a command through the platform shell and returns its
// combined output. The command is killed after shellRunTimeout.
func runShellLine(line string) (string, error, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), shellRunTimeout)
	defer cancel()

	cmd := newShellCommand(ctx, line)
	out, err := cmd.CombinedOutput()
	text := normalizeShellOutput(string(out))

	if ctx.Err() == context.DeadlineExceeded {
		return text, nil, true
	}
	return text, err, false
}

// normalizeShellOutput makes raw command output safe to render: escape
// sequences are stripped, CRLF collapses to LF, and stray carriage returns
// (progress bars overwriting a line) are dropped.
func normalizeShellOutput(s string) string {
	s = stripANSI(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "")
}

// newShellCommand builds the platform shell invocation. On Windows the console
// code page is switched to UTF-8 first, otherwise built-in commands print in
// the OEM code page and the output would be unreadable.
func newShellCommand(ctx context.Context, line string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", "chcp 65001>nul 2>&1 && "+line)
	}
	return exec.CommandContext(ctx, "sh", "-c", line)
}

// stripANSI removes escape sequences so raw command output cannot corrupt the
// rendered view.
func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// shellResultText renders the outcome of a finished command.
func shellResultText(lang string, msg tuiShellDoneMsg) string {
	var b strings.Builder
	if strings.TrimSpace(msg.output) != "" {
		b.WriteString(strings.TrimRight(msg.output, "\n"))
		b.WriteString("\n")
	}
	switch {
	case msg.timedOut:
		b.WriteString(lipgloss.NewStyle().Foreground(tuiErr).
			Render(T(lang, "tui.shell_timeout", int(shellRunTimeout.Seconds()))) + "\n")
	case msg.err != nil:
		b.WriteString(lipgloss.NewStyle().Foreground(tuiErr).
			Render(T(lang, "tui.shell_exit", msg.err.Error())) + "\n")
	}
	return b.String()
}

// appendShellOutput appends text to the scrollback, trimming the oldest lines
// beyond shellMaxLines.
func (m *tuiModel) appendShellOutput(text string) {
	if text == "" {
		return
	}
	combined := m.shellBuf
	if combined != "" && !strings.HasSuffix(combined, "\n") {
		combined += "\n"
	}
	combined += text

	if lines := strings.Split(combined, "\n"); len(lines) > shellMaxLines {
		combined = strings.Join(lines[len(lines)-shellMaxLines:], "\n")
	}
	m.shellBuf = combined
	m.shell.SetContent(combined)
	m.shell.GotoBottom()
}

// pushShellHistory records a command line for ↑/↓ recall.
func (m *tuiModel) pushShellHistory(line string) {
	if line == "" {
		return
	}
	if n := len(m.shellHist); n == 0 || m.shellHist[n-1] != line {
		m.shellHist = append(m.shellHist, line)
	}
	m.shellHistIdx = len(m.shellHist)
}

// shellHistoryPrev recalls the previous command line.
func (m *tuiModel) shellHistoryPrev() {
	if len(m.shellHist) == 0 {
		return
	}
	if m.shellHistIdx > 0 {
		m.shellHistIdx--
	}
	m.command.SetValue(m.shellHist[m.shellHistIdx])
	m.command.CursorEnd()
}

// shellHistoryNext recalls the next command line (or clears the input at the end).
func (m *tuiModel) shellHistoryNext() {
	if len(m.shellHist) == 0 {
		return
	}
	if m.shellHistIdx < len(m.shellHist)-1 {
		m.shellHistIdx++
		m.command.SetValue(m.shellHist[m.shellHistIdx])
	} else {
		m.shellHistIdx = len(m.shellHist)
		m.command.SetValue("")
	}
	m.command.CursorEnd()
}

// translateCmd starts an asynchronous translation. The target language is left
// empty so the session-language pair is used.
func (m *tuiModel) translateCmd(text string) tea.Cmd {
	cfg := m.cfg // snapshot
	offline := m.offline
	cache := m.cache
	return func() tea.Msg {
		res, err := Translate(cfg, Options{Offline: offline}, cache, text, "")
		return tuiTranslateMsg{text: text, res: res, err: err}
	}
}

// tuiSourceLabel renders the badge describing where a translation came from.
func tuiSourceLabel(lang, source string) string {
	switch source {
	case "cache":
		return lipgloss.NewStyle().Foreground(tuiWarn).Bold(true).Render(T(lang, "tui.src_cache"))
	case "dictionary":
		return lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(T(lang, "tui.src_dict"))
	case "ai":
		return lipgloss.NewStyle().Foreground(tuiOk).Bold(true).Render(T(lang, "tui.src_ai"))
	default:
		return lipgloss.NewStyle().Foreground(tuiOk).Bold(true).Render(T(lang, "tui.src_api"))
	}
}

// langTitle renders a pane title with its language code.
func (m *tuiModel) langTitle(label, code string) string {
	return lipgloss.NewStyle().Foreground(tuiDim).Render(fmt.Sprintf(" %s (%s)", label, code))
}

// box returns a bordered frame of the given width; the border colour reflects
// the focus state.
func (m *tuiModel) box(focused bool, width int) lipgloss.Style {
	c := tuiDim
	if focused {
		c = tuiAccent
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c).
		Width(width - 2) // Width excludes the border
}

// View satisfies tea.Model.
func (m *tuiModel) View() string {
	lang := m.lang
	if m.mode == tuiModeShell {
		return m.shellView()
	}

	// Header: name + language pair + mode + hint.
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(tuiAccent).Render("Tr"))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(tuiDim).Render(
		fmt.Sprintf("%s → %s", m.cfg.SourceLang, m.cfg.TargetLang)))
	if m.offline {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(tuiWarn).Bold(true).Render(T(lang, "tui.offline")))
	} else {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(tuiOk).Bold(true).Render(T(lang, "tui.online")))
	}
	if m.busy {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(tuiWarn).Render(T(lang, "tui.translating")))
	}
	if m.hint != "" {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(tuiAccent).Render(m.hint))
	}

	leftTitle := m.langTitle(T(lang, "tui.overview"), m.cfg.SourceLang)
	rightTitle := m.langTitle(T(lang, "tui.result"), m.cfg.TargetLang)
	if m.srcLabel != "" {
		rightTitle += " · " + m.srcLabel
	}

	leftBox := m.box(m.focus == tuiFocusInput, m.leftWidth).Render(leftTitle + "\n" + m.input.View())
	rightBox := m.box(m.focus == tuiFocusResult, m.rightWidth).Render(rightTitle + "\n" + m.result.View())
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// The full shortcut list is too wide for an 80-column terminal, so narrow
	// windows get the short variant instead of a wrapped footer.
	keys := T(lang, "tui.keys")
	if m.width < 110 {
		keys = T(lang, "tui.keys_short")
	}
	footer := lipgloss.NewStyle().Foreground(tuiDim).Render(keys)

	return b.String() + "\n" + body + "\n" + footer
}

// shellView renders the full-screen system shell: the translator panes are
// hidden and the command output takes their place.
func (m *tuiModel) shellView() string {
	lang := m.lang
	title := lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(T(lang, "tui.shell_title"))
	if m.shellRunning {
		title += "  " + lipgloss.NewStyle().Foreground(tuiWarn).Render(T(lang, "tui.shell_running"))
	}
	return title + "\n" + m.shell.View() + "\n" + m.command.View()
}

// runTUI launches the interactive terminal UI. Returns a process exit code.
func runTUI(cfg Config) int {
	cache, cacheErr := NewCache()
	if cacheErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "warn.cache_unavailable", cacheErr))
	}
	m := newTUIModel(cfg, cache)
	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "err.translate_failed", err))
		return 1
	}
	return 0
}
