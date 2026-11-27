package tr

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// configProvider is a preset shown by the guided configuration wizard.
type configProvider struct {
	id      string
	name    string
	descZh  string
	descEn  string
	baseURL string
	model   string
	local   bool
}

var configWizardProviders = []configProvider{
	{
		id:      "zhipu",
		name:    "智谱 BigModel (GLM)",
		descZh:  "默认 · GLM-5.3-Flash · OpenAI 兼容",
		descEn:  "default · GLM-5.3-Flash · OpenAI-compatible",
		baseURL: DefaultAIBaseURL,
		model:   DefaultAIModel,
	},
	{
		id:      "deepseek",
		name:    "DeepSeek",
		descZh:  "deepseek-chat / deepseek-reasoner",
		descEn:  "deepseek-chat / deepseek-reasoner",
		baseURL: "https://api.deepseek.com/v1",
		model:   "deepseek-chat",
	},
	{
		id:      "openai",
		name:    "OpenAI",
		descZh:  "gpt-4o-mini 等 OpenAI Chat Completions 模型",
		descEn:  "gpt-4o-mini and other Chat Completions models",
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4o-mini",
	},
	{
		id:      "local",
		name:    "本地模型 / Local",
		descZh:  "Ollama · LM Studio · vLLM（通常不需要 key）",
		descEn:  "Ollama · LM Studio · vLLM (usually no key)",
		baseURL: "http://127.0.0.1:11434/v1",
		model:   "qwen2.5:7b",
		local:   true,
	},
	{
		id:     "custom",
		name:   "自定义 / Custom",
		descZh: "任意 OpenAI 兼容服务；保留当前地址与模型",
		descEn: "any OpenAI-compatible service; keep current URL/model",
	},
}

func wizardText(lang, zh, en string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "zh") {
		return zh
	}
	// Japanese falls back to English for this compact settings wizard; the
	// rest of the application remains localized.
	return en
}

type configWizardModel struct {
	cfg      Config
	lang     string
	step     int // 0 provider, 1 connection, 2 preferences, 3 confirm, 4 saved
	provider int
	focus    int
	width    int
	height   int

	base   textinput.Model
	model  textinput.Model
	key    textinput.Model
	source textinput.Model
	target textinput.Model
	ui     textinput.Model

	thinking   bool
	keySet     bool
	keyCleared bool
	envKey     string
	err        string
	savedPath  string
}

func newWizardInput(value string, width int) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 0
	ti.Width = width
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(tuiDim)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(tuiAccent)
	ti.SetValue(value)
	return ti
}

func detectConfigProvider(cfg Config) int {
	base := strings.ToLower(strings.TrimSpace(cfg.AIBaseURL))
	switch {
	case strings.Contains(base, "bigmodel.cn"):
		return 0
	case strings.Contains(base, "deepseek"):
		return 1
	case strings.Contains(base, "openai.com"):
		return 2
	case strings.Contains(base, "localhost"), strings.Contains(base, "127.0.0.1"),
		strings.Contains(base, "::1"), strings.Contains(base, "0.0.0.0"):
		return 3
	default:
		return 4
	}
}

func newConfigWizard(cfg Config) *configWizardModel {
	lang := cfg.UILang
	if lang == "" {
		lang = "zh"
	}
	m := &configWizardModel{
		cfg:      cfg,
		lang:     lang,
		provider: detectConfigProvider(cfg),
		width:    90,
		height:   28,
		thinking: cfg.AIThinking,
		keySet:   cfg.HasAIKey(),
		envKey:   envAIKeyName(),
	}
	m.base = newWizardInput(cfg.AIBaseURL, 72)
	m.model = newWizardInput(cfg.AIModel, 48)
	m.key = newWizardInput("", 56)
	m.key.EchoMode = textinput.EchoPassword
	m.key.Placeholder = m.keyPlaceholder()
	m.source = newWizardInput(cfg.SourceLang, 18)
	m.target = newWizardInput(cfg.TargetLang, 18)
	m.ui = newWizardInput(cfg.UILang, 10)
	m.layoutInputs()
	return m
}

func (m *configWizardModel) keyPlaceholder() string {
	switch {
	case m.keySet:
		return wizardText(m.lang, "已配置，留空保持不变", "configured; leave blank to keep")
	case m.envKey != "":
		return wizardText(m.lang, "环境变量已设置", "set via environment")
	case strings.TrimSpace(DefaultAPIKey) != "":
		return wizardText(m.lang, "内置默认 key 已设置", "built-in default key")
	default:
		return "sk-..."
	}
}

func (m *configWizardModel) layoutInputs() {
	w := m.width - 22
	if w < 36 {
		w = 36
	}
	if w > 90 {
		w = 90
	}
	m.base.Width = w
	m.model.Width = w - 10
	m.key.Width = w - 8
	m.source.Width = 18
	m.target.Width = 18
	m.ui.Width = 8
}

func (m *configWizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *configWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layoutInputs()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *configWizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	m.err = ""

	if key == "ctrl+c" || key == "ctrl+q" {
		return m, tea.Quit
	}
	if m.step == 4 {
		if key == "enter" || key == "esc" || key == "q" {
			return m, tea.Quit
		}
		return m, nil
	}

	switch key {
	case "ctrl+s":
		if m.step == 0 {
			return m, nil
		}
		m.save()
		return m, nil
	case "esc":
		if m.step == 0 {
			return m, tea.Quit
		}
		m.step--
		m.focus = 0
		m.refreshFocus()
		return m, nil
	case "tab":
		m.cycleFocus(1)
		return m, nil
	case "shift+tab":
		m.cycleFocus(-1)
		return m, nil
	case "enter":
		m.advance()
		return m, nil
	}

	if m.step == 0 {
		switch key {
		case "up", "k":
			m.provider = (m.provider - 1 + len(configWizardProviders)) % len(configWizardProviders)
		case "down", "j":
			m.provider = (m.provider + 1) % len(configWizardProviders)
		}
		return m, nil
	}

	if m.step == 1 && key == "ctrl+t" {
		m.thinking = !m.thinking
		return m, nil
	}
	if m.step == 1 && key == "ctrl+k" && m.focus == 2 {
		m.key.SetValue("")
		m.cfg.APIKey = ""
		m.keySet = false
		m.keyCleared = true
		m.key.Placeholder = wizardText(m.lang, "已清除，保存后移除", "cleared; removed on save")
		return m, nil
	}
	if m.step == 2 && m.focus == 2 && (key == "left" || key == "right") {
		m.cycleUILang(key == "right")
		return m, nil
	}

	// Route to the focused field.
	var cmd tea.Cmd
	switch m.step {
	case 1:
		switch m.focus {
		case 0:
			m.base, cmd = m.base.Update(msg)
		case 1:
			m.model, cmd = m.model.Update(msg)
		case 2:
			m.key, cmd = m.key.Update(msg)
			if strings.TrimSpace(m.key.Value()) != "" {
				m.keyCleared = false
			}
		}
	case 2:
		switch m.focus {
		case 0:
			m.source, cmd = m.source.Update(msg)
		case 1:
			m.target, cmd = m.target.Update(msg)
		case 2:
			m.ui, cmd = m.ui.Update(msg)
		}
	}
	return m, cmd
}

func (m *configWizardModel) cycleFocus(delta int) {
	if m.step != 1 && m.step != 2 {
		return
	}
	m.focus = (m.focus + delta + 3) % 3
	m.refreshFocus()
}

func (m *configWizardModel) refreshFocus() {
	for _, ti := range []*textinput.Model{&m.base, &m.model, &m.key, &m.source, &m.target, &m.ui} {
		ti.Blur()
	}
	if m.step == 1 {
		switch m.focus {
		case 0:
			m.base.Focus()
		case 1:
			m.model.Focus()
		case 2:
			m.key.Focus()
		}
	}
	if m.step == 2 {
		switch m.focus {
		case 0:
			m.source.Focus()
		case 1:
			m.target.Focus()
		case 2:
			m.ui.Focus()
		}
	}
}

func (m *configWizardModel) cycleUILang(forward bool) {
	order := []string{"zh", "en", "ja"}
	idx := 0
	for i, v := range order {
		if v == strings.ToLower(strings.TrimSpace(m.ui.Value())) {
			idx = i
			break
		}
	}
	if forward {
		idx = (idx + 1) % len(order)
	} else {
		idx = (idx - 1 + len(order)) % len(order)
	}
	m.ui.SetValue(order[idx])
	m.lang = order[idx]
	m.key.Placeholder = m.keyPlaceholder()
}

func (m *configWizardModel) advance() {
	switch m.step {
	case 0:
		m.applyProvider()
		m.step = 1
		m.focus = 0
		m.refreshFocus()
	case 1:
		if m.focus < 2 {
			m.focus++
			m.refreshFocus()
			return
		}
		if err := m.validateConnection(); err != "" {
			m.err = err
			return
		}
		m.step = 2
		m.focus = 0
		m.refreshFocus()
	case 2:
		if m.focus < 2 {
			m.focus++
			m.refreshFocus()
			return
		}
		if err := m.validateLanguages(); err != "" {
			m.err = err
			return
		}
		m.step = 3
	case 3:
		m.save()
	}
}

func (m *configWizardModel) applyProvider() {
	p := configWizardProviders[m.provider]
	if p.id != "custom" {
		m.base.SetValue(p.baseURL)
		m.model.SetValue(p.model)
	}
	current := configWizardProviders[detectConfigProvider(m.cfg)]
	if p.id != "custom" && p.id != current.id {
		// Do not silently reuse a key belonging to a different provider.
		m.cfg.APIKey = ""
		m.key.SetValue("")
		m.keySet = false
		m.keyCleared = true
		m.key.Placeholder = wizardText(m.lang, "sk-...", "sk-...")
	}
	if p.local {
		m.thinking = false
		m.key.Placeholder = wizardText(m.lang, "本地服务可留空", "optional for local")
	}
}

func (m *configWizardModel) validateConnection() string {
	base := strings.TrimSpace(m.base.Value())
	if base == "" {
		return wizardText(m.lang, "API Base URL 不能为空", "API Base URL must not be empty")
	}
	if err := validateBaseURL(base); err != nil {
		return fmt.Sprintf(wizardText(m.lang, "API Base URL 无效: %v", "invalid API Base URL: %v"), err)
	}
	model := strings.TrimSpace(m.model.Value())
	if model == "" {
		return wizardText(m.lang, "模型不能为空", "model must not be empty")
	}
	if v := strings.TrimSpace(m.key.Value()); v != "" {
		m.cfg.APIKey = v
		m.cfg.apiKeyFromEnv = false
		m.cfg.apiKeyFromDefault = false
		m.keySet = true
	}
	if m.keyCleared && strings.TrimSpace(m.key.Value()) == "" {
		m.cfg.APIKey = ""
		m.keySet = false
	}
	if !m.cfg.HasAIKey() && !isKeylessEndpoint(base) && m.envKey == "" && strings.TrimSpace(DefaultAPIKey) == "" {
		return wizardText(m.lang, "当前服务商需要 API Key", "this provider requires an API key")
	}
	m.cfg.AIBaseURL = base
	m.cfg.AIModel = model
	return ""
}

func (m *configWizardModel) validateLanguages() string {
	src, ok := normalizeLangValue(m.source.Value(), true)
	if !ok {
		return fmt.Sprintf(wizardText(m.lang, "源语言无效: %s", "invalid source language: %s"), strings.TrimSpace(m.source.Value()))
	}
	tgt, ok := normalizeLangValue(m.target.Value(), false)
	if !ok {
		return fmt.Sprintf(wizardText(m.lang, "目标语言无效: %s", "invalid target language: %s"), strings.TrimSpace(m.target.Value()))
	}
	ui := strings.ToLower(strings.TrimSpace(m.ui.Value()))
	switch ui {
	case "zh", "en", "ja":
	default:
		return fmt.Sprintf(wizardText(m.lang, "界面语言无效: %s（可用 zh/en/ja）", "invalid UI language: %s (use zh/en/ja)"), ui)
	}
	m.cfg.SourceLang = src
	m.cfg.TargetLang = tgt
	m.cfg.UILang = ui
	m.lang = ui
	m.key.Placeholder = m.keyPlaceholder()
	return ""
}

func (m *configWizardModel) save() {
	if err := m.validateConnection(); err != "" {
		m.err = err
		return
	}
	if err := m.validateLanguages(); err != "" {
		m.err = err
		return
	}
	m.cfg.AIThinking = m.thinking
	if err := m.cfg.Save(); err != nil {
		m.err = fmt.Sprintf(wizardText(m.lang, "保存失败: %v", "save failed: %v"), err)
		return
	}
	m.savedPath = ConfigPath()
	m.step = 4
}

func (m *configWizardModel) View() string {
	lang := m.lang
	title := lipgloss.NewStyle().Bold(true).Foreground(tuiAccent).Render("Tr")
	subtitle := lipgloss.NewStyle().Foreground(tuiDim).Render(
		wizardText(lang, "配置向导  ·  AI 与语言设置", "Configuration wizard  ·  AI & language"))
	stepNames := []string{
		wizardText(lang, "服务商", "Provider"),
		wizardText(lang, "连接", "Connection"),
		wizardText(lang, "偏好", "Preferences"),
		wizardText(lang, "确认", "Confirm"),
	}
	if m.step < 4 {
		subtitle = lipgloss.NewStyle().Foreground(tuiDim).Render(
			fmt.Sprintf("%s  ·  %s %d/4", subtitle, stepNames[m.step], m.step+1))
	} else {
		subtitle = lipgloss.NewStyle().Foreground(tuiOk).Render(wizardText(lang, "配置已保存", "configuration saved"))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		title,
		"  ",
		subtitle,
	)
	progress := m.progressBar()

	var body string
	switch m.step {
	case 0:
		body = m.viewProviders()
	case 1:
		body = m.viewConnection()
	case 2:
		body = m.viewPreferences()
	case 3:
		body = m.viewConfirm()
	case 4:
		body = m.viewSaved()
	}

	footer := m.footer()
	main := lipgloss.JoinVertical(lipgloss.Left, header, progress, "", body)
	if m.err != "" {
		main += "\n\n" + lipgloss.NewStyle().Foreground(tuiErr).Render("  "+m.err)
	}
	main += "\n\n" + lipgloss.NewStyle().Foreground(tuiDim).Render(footer)

	box := lipgloss.NewStyle().
		Width(maxInt(60, m.width-8)).
		Padding(0, 2)
	return box.Render(main)
}

func (m *configWizardModel) progressBar() string {
	var b strings.Builder
	for i := 0; i < 4; i++ {
		if i < m.step+1 && m.step < 4 {
			b.WriteString(lipgloss.NewStyle().Foreground(tuiAccent).Render("━━"))
		} else if m.step == 4 {
			b.WriteString(lipgloss.NewStyle().Foreground(tuiOk).Render("━━"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("━━"))
		}
	}
	return b.String()
}

func (m *configWizardModel) sectionTitle(s string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(s)
}

func (m *configWizardModel) viewProviders() string {
	lang := m.lang
	lines := []string{
		m.sectionTitle(wizardText(lang, "选择 AI 服务商", "Choose an AI provider")),
		lipgloss.NewStyle().Foreground(tuiDim).Render(
			wizardText(lang, "选择预设会填充地址和模型；按 Enter 继续。", "Presets fill the URL and model; press Enter to continue.")),
		"",
	}
	for i, p := range configWizardProviders {
		marker := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if i == m.provider {
			marker = "▸ "
			style = style.Foreground(tuiAccent).Bold(true)
		}
		name := style.Render(p.name)
		current := ""
		if i == detectConfigProvider(m.cfg) {
			current = "  " + lipgloss.NewStyle().Foreground(tuiOk).Render(wizardText(lang, "当前", "current"))
		}
		lines = append(lines, fmt.Sprintf("%s%s%s", marker, name, current))
		lines = append(lines, "    "+lipgloss.NewStyle().Foreground(tuiDim).Render(wizardText(lang, p.descZh, p.descEn)))
	}
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) field(label, hint string, input textinput.Model) string {
	lines := []string{
		lipgloss.NewStyle().Foreground(tuiDim).Render(label),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(tuiDim).
			Padding(0, 1).
			Render(input.View()),
	}
	if hint != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(tuiDim).Render("  "+hint))
	}
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) viewConnection() string {
	lang := m.lang
	lines := []string{
		m.sectionTitle(wizardText(lang, "连接信息", "Connection")),
		"",
		m.field("API Base URL", wizardText(lang, "OpenAI 兼容地址，可填 base 或完整 endpoint", "OpenAI-compatible base URL or full endpoint"), m.base),
		"",
		m.field(wizardText(lang, "模型 (Model)", "Model"), wizardText(lang, "例如 glm-5.3-flash / deepseek-chat / gpt-4o-mini", "e.g. glm-5.3-flash / deepseek-chat / gpt-4o-mini"), m.model),
		"",
		m.field("API Key", "", m.key),
	}
	if m.keySet || m.envKey != "" || strings.TrimSpace(DefaultAPIKey) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(tuiDim).Render(
			wizardText(lang, "输入新 key 会覆盖；Ctrl+K 清除配置文件中的 key。",
				"typing a new key overwrites it; Ctrl+K removes the key from config.json.")))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(tuiDim).Render(
			wizardText(lang, "本地模型可留空；云端服务通常必须填写。",
				"leave blank for local models; cloud providers usually require a key.")))
	}
	if m.envKey != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(tuiWarn).Render(
			fmt.Sprintf(wizardText(lang, "环境变量 %s 已设置，会优先于配置文件。", "environment variable %s is set and overrides the file."), m.envKey)))
	}
	lines = append(lines,
		"",
		fmt.Sprintf("%s  %s   %s",
			lipgloss.NewStyle().Foreground(tuiDim).Render(wizardText(lang, "思考模式", "Thinking")),
			lipgloss.NewStyle().Foreground(tuiAccent).Bold(true).Render(boolLabel(lang, m.thinking)),
			lipgloss.NewStyle().Foreground(tuiDim).Render(wizardText(lang, "Ctrl+T 切换（仅 DeepSeek 生效）", "Ctrl+T toggle (DeepSeek only)")),
		),
	)
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) viewPreferences() string {
	lang := m.lang
	lines := []string{
		m.sectionTitle(wizardText(lang, "翻译偏好", "Translation preferences")),
		"",
		m.field(wizardText(lang, "源语言 (source_lang)", "Source language (source_lang)"),
			wizardText(lang, "en / zh / ja / auto", "en / zh / ja / auto"), m.source),
		m.field(wizardText(lang, "目标语言 (target_lang)", "Target language (target_lang)"),
			wizardText(lang, "zh / en / ja / zh-hant ...", "zh / en / ja / zh-hant ..."), m.target),
		m.field(wizardText(lang, "界面语言 (ui_lang)", "UI language (ui_lang)"),
			wizardText(lang, "输入 zh/en/ja，或按 ←/→ 切换", "type zh/en/ja, or use ←/→"), m.ui),
		"",
		lipgloss.NewStyle().Foreground(tuiDim).Render(
			wizardText(lang, "语言支持代码和别名（如 中文、English、日本語）。",
				"codes and aliases are accepted (中文 / English / 日本語).")),
	}
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) viewConfirm() string {
	lang := m.lang
	provider := configWizardProviders[m.provider]
	key := wizardText(lang, "未设置", "not set")
	if v := strings.TrimSpace(m.key.Value()); v != "" {
		key = MaskKey(v)
	} else if m.cfg.HasAIKey() {
		key = MaskKey(m.cfg.APIKey)
	} else if m.envKey != "" {
		key = wizardText(lang, "来自环境变量 "+m.envKey, "from env "+m.envKey)
	} else if strings.TrimSpace(DefaultAPIKey) != "" {
		key = wizardText(lang, "内置默认 key", "built-in default key")
	}
	configLabel := wizardText(lang, "配置文件", "Config file")
	rows := [][2]string{
		{wizardText(lang, "服务商", "Provider"), provider.name},
		{"API Base URL", m.base.Value()},
		{wizardText(lang, "模型", "Model"), m.model.Value()},
		{"API Key", key},
		{wizardText(lang, "语言", "Languages"), fmt.Sprintf("%s → %s", m.source.Value(), m.target.Value())},
		{wizardText(lang, "界面语言", "UI language"), m.ui.Value()},
		{wizardText(lang, "思考模式", "Thinking"), boolLabel(lang, m.thinking)},
		{configLabel, ConfigPath()},
	}
	lines := []string{
		m.sectionTitle(wizardText(lang, "确认并保存", "Confirm and save")),
		"",
	}
	labelStyle := lipgloss.NewStyle().Foreground(tuiDim).Width(18)
	for _, r := range rows {
		if r[0] == configLabel {
			lines = append(lines, "  "+lipgloss.NewStyle().Foreground(tuiDim).Render(r[0]))
			lines = append(lines, "    "+r[1])
			continue
		}
		lines = append(lines, "  "+labelStyle.Render(r[0])+" "+r[1])
	}
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) viewSaved() string {
	lang := m.lang
	path := m.savedPath
	if path == "" {
		path = ConfigPath()
	}
	check := lipgloss.NewStyle().Bold(true).Foreground(tuiOk).Render("✓")
	lines := []string{
		"",
		"  " + check + "  " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(
			wizardText(lang, "配置已保存", "Configuration saved")),
		"",
		"  " + lipgloss.NewStyle().Foreground(tuiDim).Render(wizardText(lang, "写入", "written to")),
		"  " + lipgloss.NewStyle().Foreground(tuiAccent).Render(path),
		"",
		"  " + lipgloss.NewStyle().Foreground(tuiDim).Render(
			wizardText(lang, "现在可以运行 tr \"Hello, world!\" 开始翻译。", "Run tr \"Hello, world!\" to start translating.")),
	}
	return strings.Join(lines, "\n")
}

func (m *configWizardModel) footer() string {
	lang := m.lang
	switch m.step {
	case 0:
		return wizardText(lang, "↑/↓ 选择  ·  Enter 继续  ·  Ctrl+Q 退出",
			"↑/↓ select  ·  Enter continue  ·  Ctrl+Q quit")
	case 3:
		return wizardText(lang, "Enter 保存  ·  Esc 返回  ·  Ctrl+Q 取消",
			"Enter save  ·  Esc back  ·  Ctrl+Q cancel")
	case 4:
		return wizardText(lang, "Enter / Esc 退出", "Enter / Esc exit")
	default:
		return wizardText(lang, "Tab 切换  ·  Enter 继续  ·  Esc 返回  ·  Ctrl+S 保存  ·  Ctrl+Q 退出",
			"Tab switch  ·  Enter continue  ·  Esc back  ·  Ctrl+S save  ·  Ctrl+Q quit")
	}
}

func boolLabel(lang string, v bool) string {
	if v {
		return wizardText(lang, "开启", "on")
	}
	return wizardText(lang, "关闭", "off")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// runConfigWizard launches the guided configuration TUI.
func runConfigWizard(cfg Config) int {
	m := newConfigWizard(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", T(cfg.UILang, "err.tui_start", err))
		return 1
	}
	return 0
}
