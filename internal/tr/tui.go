package tr

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
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

// tuiInputRows is the number of rows the text area occupies.
const tuiInputRows = 6

type tuiFocus int

const (
	tuiFocusInput tuiFocus = iota
	tuiFocusResult
)

// tuiTranslateMsg carries the result of an asynchronous translation.
type tuiTranslateMsg struct {
	text string
	res  TranslateResult
	err  error
}

// tuiModel is the bubbletea model for the interactive translation UI.
type tuiModel struct {
	cfg      Config
	cache    *Cache
	lang     string
	input    textarea.Model
	result   viewport.Model
	focus    tuiFocus
	width    int
	height   int
	busy     bool
	offline  bool // session-level offline override (Ctrl+O)
	errText  string
	lastText string // last submitted source text ("" = no result yet)
}

// newTUIModel builds the initial model.
func newTUIModel(cfg Config, cache *Cache) tuiModel {
	ta := textarea.New()
	ta.Placeholder = T(cfg.UILang, "tui.placeholder")
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(tuiInputRows)

	vp := viewport.New(60, 10)
	vp.KeyMap.PageDown.SetEnabled(true)

	m := tuiModel{
		cfg:    cfg,
		cache:  cache,
		lang:   cfg.UILang,
		input:  ta,
		result: vp,
		focus:  tuiFocusInput,
		width:  80,
		height: 24,
	}
	m.layout(80, 24)
	m.result.SetContent(lipgloss.NewStyle().Foreground(tuiDim).Render(T(cfg.UILang, "app.desc")))
	return m
}

// layout recalculates component sizes from a terminal size.
func (m *tuiModel) layout(w, h int) {
	if w < 30 {
		w = 30
	}
	if h < 12 {
		h = 12
	}
	m.width, m.height = w, h

	innerW := w - 2 // leave room for box borders
	m.input.SetWidth(innerW)
	m.input.SetHeight(tuiInputRows)

	// Total rows: header(1) + input box + result box + footer(1), each
	// separated by a newline; box height = title row + inner rows + borders.
	inputBoxH := tuiInputRows + 3
	resultBoxH := h - 5 - inputBoxH
	if resultBoxH < 6 {
		resultBoxH = 6
	}
	m.result.Width = innerW
	m.result.Height = resultBoxH - 3 // minus title row and borders
	if m.result.Height < 1 {
		m.result.Height = 1
	}
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
		m.lastText = msg.text
		if msg.err != nil {
			m.errText = T(m.lang, "tui.failed", msg.err)
			m.result.SetContent(lipgloss.NewStyle().Foreground(tuiErr).Render(m.errText))
			return m, nil
		}
		m.errText = ""
		m.renderResult(msg)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey processes keyboard input. Control keys are handled first so they
// never collide with text the user is typing.
func (m *tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "ctrl+t":
		if m.busy {
			return m, nil
		}
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			m.errText = T(m.lang, "tui.empty")
			m.result.SetContent(lipgloss.NewStyle().Foreground(tuiErr).Render(m.errText))
			return m, nil
		}
		m.busy = true
		m.errText = ""
		return m, m.translateCmd(text)

	case "ctrl+s": // swap source/target languages (persisted)
		m.cfg.SourceLang, m.cfg.TargetLang = m.cfg.TargetLang, m.cfg.SourceLang
		_ = m.cfg.Save()

	case "ctrl+o": // toggle offline for this session
		m.offline = !m.offline

	case "tab":
		m.swapFocus()

	case "esc":
		switch {
		case m.focus == tuiFocusInput && m.input.Value() != "":
			m.input.Reset()
		case m.focus == tuiFocusResult:
			m.swapFocus()
		}
		return m, nil
	}

	// Route everything else to the focused component.
	var cmd tea.Cmd
	if m.focus == tuiFocusInput {
		m.input, cmd = m.input.Update(msg)
	} else {
		m.result, cmd = m.result.Update(msg)
	}
	return m, cmd
}

// swapFocus moves focus between the input area and the result viewport.
func (m *tuiModel) swapFocus() {
	if m.focus == tuiFocusInput {
		m.focus = tuiFocusResult
		m.input.Blur()
	} else {
		m.focus = tuiFocusInput
		m.input.Focus()
	}
}

// translateCmd starts an asynchronous translation.
func (m *tuiModel) translateCmd(text string) tea.Cmd {
	cfg := m.cfg // snapshot
	offline := m.offline
	cache := m.cache
	return func() tea.Msg {
		res, err := Translate(cfg, Options{Offline: offline}, cache, text)
		return tuiTranslateMsg{text: text, res: res, err: err}
	}
}

// renderResult fills the result viewport with the latest translation.
func (m *tuiModel) renderResult(msg tuiTranslateMsg) {
	var srcLabel string
	switch msg.res.Source {
	case "cache":
		srcLabel = lipgloss.NewStyle().Foreground(tuiWarn).Bold(true).Render(T(m.lang, "tui.src_cache"))
	case "dictionary":
		srcLabel = lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(T(m.lang, "tui.src_dict"))
	default:
		srcLabel = lipgloss.NewStyle().Foreground(tuiOk).Bold(true).Render(T(m.lang, "tui.src_api"))
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(tuiDim).Render(T(m.lang, "tui.overview") + " (" + m.cfg.SourceLang + ")"))
	b.WriteString("\n")
	b.WriteString(msg.text)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(T(m.lang, "tui.result")+" ("+m.cfg.TargetLang+")") + " · " + srcLabel)
	b.WriteString("\n")
	b.WriteString(msg.res.Text)
	m.result.SetContent(b.String())
	m.result.GotoBottom()
}

// box returns a bordered frame; the border color reflects focus state and
// the frame is stretched to the terminal width.
func (m *tuiModel) box(focused bool) lipgloss.Style {
	c := tuiDim
	if focused {
		c = tuiAccent
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c).
		Width(m.width)
}

// sectionTitle renders a section heading inside a box.
func (m *tuiModel) sectionTitle(text string) string {
	return lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(text)
}

// View satisfies tea.Model.
func (m *tuiModel) View() string {
	lang := m.lang

	// Header: name + direction + mode.
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

	inputTitle := T(lang, "tui.overview") + " (" + m.cfg.SourceLang + ")"
	resultTitle := T(lang, "tui.result") + " (" + m.cfg.TargetLang + ")"

	inputBox := m.box(m.focus == tuiFocusInput).
		Render(m.sectionTitle(inputTitle) + "\n" + m.input.View())
	resultBox := m.box(m.focus == tuiFocusResult).
		Render(m.sectionTitle(resultTitle) + "\n" + m.result.View())
	footer := lipgloss.NewStyle().Foreground(tuiDim).Render(T(lang, "tui.keys"))

	return b.String() + "\n" + inputBox + "\n" + resultBox + "\n" + footer
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
