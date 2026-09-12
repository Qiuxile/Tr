package tr

import "fmt"

// messages holds the UI string catalog, keyed by language code.
// Supported: "zh" (Chinese), "en" (English), "ja" (Japanese).
var messages = map[string]map[string]string{
	"zh": zhMessages,
	"en": enMessages,
	"ja": jaMessages,
}

// T returns the localized message for the given language and key.
// Falls back to English if the language or key is not found.
// Additional args are passed to fmt.Sprintf for formatting.
func T(lang, key string, args ...interface{}) string {
	msgs, ok := messages[lang]
	if !ok {
		msgs = messages["en"]
	}
	msg, ok := msgs[key]
	if !ok {
		// Fallback to English
		if enMsgs, ok2 := messages["en"]; ok2 {
			if enMsg, ok3 := enMsgs[key]; ok3 {
				msg = enMsg
			}
		}
	}
	if msg == "" {
		msg = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// --- Chinese (zh) messages ------------------------------------------------

var zhMessages = map[string]string{
	// App
	"app.name":    "Tr - 终端翻译器",
	"app.version": "Tr 版本 %s",
	"app.desc":    "一个用 Go 编写的简单终端翻译工具",
	"app.license": "许可证: MIT",
	"app.repo":    "仓库: https://github.com/Qiuxile/Tr",

	// Warnings
	"warn.defaults":          "警告: %v; 使用默认配置",
	"warn.cache_unavailable": "警告: 缓存不可用: %v",
	"warn.unknown_flag":      "警告: 未知选项 %s",
	"warn.no_tty_tui":        "警告: 当前不是交互式终端,无法启动 TUI",
	"warn.arrow_plain":       "提示: 无法将 \"%s\" 识别为目标语言,已把整个输入按普通文本翻译",

	// Errors
	"err.empty_input":      "错误: 没有输入文本。请传入参数、管道输入,或无参数启动 TUI。",
	"err.usage_hint":       "用法: tr [文本]",
	"err.bad_usage":        "参数有误,未执行任何操作。使用 tr -help 查看用法。",
	"err.translate_failed": "错误: %v",
	"err.empty":            "输入文本不能为空",
	"err.no_translation":   "无可用翻译",
	"err.offline":          "离线模式: 缓存和词典中均未找到该文本",
	"err.no_key":           "错误: 未配置 AI 密钥。请运行 tr config set api_key sk-... (或设置环境变量 TR_API_KEY);只想离线使用可加 -o。",
	"err.network":          "网络请求失败",
	"err.file_usage":       "用法: tr -f <文件> [输出路径]",
	"err.file_read":        "读取文件失败: %v",
	"err.file_write":       "写入文件失败: %v",

	// File translation
	"file.progress": "翻译中 %d/%d …",
	"file.partial":  "警告: %d/%d 个片段翻译失败,已保留原文",
	"file.written":  "已写入: %s",

	// Config
	"config.usage_show_set": "用法: tr config show|set <键> <值>   (兼容: tr -config ...)",
	"config.current":        "当前配置:",
	"config.source_lang":    "  source_lang = %s",
	"config.target_lang":    "  target_lang = %s",
	"config.ui_lang":        "  ui_lang     = %s",
	"config.api_key":        "  api_key      = %s",
	"config.ai_base_url":    "  ai_base_url  = %s",
	"config.ai_model":       "  ai_model     = %s",
	"config.ai_thinking":    "  ai_thinking  = %v",
	"config.key_unset":      "(未配置)",
	"config.file":           "  配置文件     = %s",
	"config.usage_set":      "用法: tr config set <键> <值>",
	"config.valid_keys":     "可用键: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "无效的配置项: %s。可用键: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "无效的布尔值: %s (可用: true/false, on/off, 1/0)",
	"config.save_failed":    "保存失败: %v",
	"config.set_ok":         "设置成功: %s = %s",
	"config.unknown_subcmd": "未知子命令: %s。可用: show, set",

	// Help
	"help.title":          "用法:",
	"help.translate":      "  tr [文本]                  翻译文本并输出结果",
	"help.translate_auto": "  tr <文本> -> <语言代码>     自动识别语言翻译 (例: tr 你好 -> en)",
	"help.file":           "  tr -f <文件> [输出]         翻译文件 (省略输出则打印到终端)",
	"help.tui":            "  tr (无参数, 交互终端)      启动交互式终端界面 (TUI)",
	"help.tui_cmd":        "  tr tui                     显式启动 TUI",
	"help.config_cmd":     "  tr config show|set <键> <值>  查看/修改配置 (-config 兼容)",
	"help.offline":        "  tr -o, --offline           强制离线模式 (缓存 + 内置词典)",
	"help.help":           "  tr -help                   显示帮助",
	"help.version":        "  tr -version                显示版本",
	"help.about":          "  tr -about                  显示关于信息",

	"help.flags_title":   "选项:",
	"help.flags_config":  "  -config, --config, -c   查看/修改配置 (等效 tr config ...)",
	"help.flags_help":    "  -help,   --help,   -h   显示帮助",
	"help.flags_version": "  -version,--version, -v  显示版本",
	"help.flags_about":   "  -about,  --about,   -a  显示关于",
	"help.flags_offline": "  -offline,--offline, -o  强制离线模式",

	"help.pipe_title":   "管道输入:",
	"help.pipe_example": "  echo \"hello\" | tr        从标准输入读取并翻译",
	"help.stdin_note":   "  无参数且标准输入是管道时,自动翻译 stdin 内容",

	"help.config_title":       "配置:",
	"help.config_source":      "  source_lang   源语言代码 (默认: en; auto = 自动检测)",
	"help.config_target":      "  target_lang   目标语言代码 (默认: zh)",
	"help.config_ui":          "  ui_lang       界面语言 (默认: zh, 可选: en, ja)",
	"help.config_key":         "  api_key       AI 密钥 (sk-...), 翻译必需",
	"help.config_ai_url":      "  ai_base_url   OpenAI 兼容地址 (默认: https://api.deepseek.com/v1)",
	"help.config_ai_model":    "  ai_model      模型名 (默认: deepseek-flash)",
	"help.config_ai_thinking": "  ai_thinking   思考模式 (默认: false, 关闭更省 token)",
	"help.config_path":        "  配置文件: %s",

	"help.langs_title": "目标语言简写:",
	"help.langs_line1": "  zh=中文(简) en=英语 ja=日语 ko=韩语 ru=俄语 fr=法语 de=德语 es=西班牙语",
	"help.langs_line2": "  it=意大利语 pt=葡萄牙语 ar=阿拉伯语 th=泰语 vi=越南语 hi=印地语 el=希腊语 nl=荷兰语",
	"help.langs_note":  "  也支持别名(英文 / 中文 / 日本語 / korean / russian …)与任意 ISO 639-1 代码",
	"help.langs_once":  "  -> <语言> 只作用于本次执行,不会写入配置文件;持久修改请用 tr config set target_lang",

	"help.backend_title":   "翻译后端:",
	"help.backend_ai":      "  - AI (OpenAI 兼容, 必填 api_key; 默认 DeepSeek)",
	"help.backend_offline": "  - 离线: 本地缓存 + 内置词典 (加 -o 使用)",

	"help.offline_title": "离线功能:",
	"help.offline_cache": "  - 翻译缓存: 命中缓存直接返回, 不消耗任何 token",
	"help.offline_dict":  "  - 内置词典: 约 400 个常用英→中词汇和短语",
	"help.offline_flag":  "  - 使用 -o 选项强制离线模式",

	"help.env_title": "环境变量:",
	"help.env_key":   "  TR_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY 可作为 api_key 的替代",

	// Help (extra sections)
	"help.examples_title": "示例:",
	"help.ex1":            "  tr \"Hello, world!\"              英 → 中",
	"help.ex2":            "  tr 你好 -> en                   自动识别语言 (只作用于本次执行)",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            目标支持语言代码与名称别名",
	"help.ex4":            "  echo \"hello\" | tr               管道输入翻译",
	"help.ex5":            "  tr -f in.txt out/               翻译文件并写入 out/ 目录",
	"help.ex6":            "  tr config set api_key sk-xxx    配置 AI 密钥",
	"help.flags_file":     "  -f, --file <文件> [输出]  翻译文件 (省略输出则打印到终端)",
	"help.tui_title":      "TUI 快捷键:",
	"help.tui_keys1":      "  Ctrl+T 翻译 | Enter 换行 | Ctrl+←/→ 切语言 | Ctrl+L 清屏 | Ctrl+U 清空输入",
	"help.tui_keys2":      "  Ctrl+S 互换语言 | Ctrl+O 离线 | Esc 系统命令行 | Ctrl+Q 退出",
	"help.tui_note":       "  Esc 进入系统命令行:可直接运行 tr config show 等命令,输出整屏显示",
	"help.quick_title":    "首次使用:",
	"help.quick_key":      "  tr config set api_key sk-...     翻译需要 AI 密钥 (未配置时会有提示)",
	"help.more_title":     "更多:",
	"help.more_readme":    "  完整文档见 README.md (含常见问题、故障排查、省 token 说明)",

	// About
	"about.version": "版本: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "创建默认配置",
	"cfg.read":           "读取配置",
	"cfg.parse":          "解析配置",
	"cfg.create_dir":     "创建配置目录",
	"cfg.marshal":        "序列化配置",
	"cfg.write":          "写入配置",

	// AI errors
	"ai.request":    "AI 请求",
	"ai.parse":      "AI 响应解析",
	"ai.empty":      "AI 未返回译文",
	"ai.bad_key":    "AI 密钥无效 (HTTP %d): %s",
	"ai.no_balance": "AI 账户余额不足 (HTTP %d): %s",
	"ai.rate_limit": "AI 请求被限流 (HTTP %d): %s",
	"ai.status":     "AI 返回状态码 %d: %s",

	// TUI
	"tui.placeholder":   "输入要翻译的文本,Ctrl+T 翻译...",
	"tui.translating":   "翻译中...",
	"tui.empty":         "没有可翻译的文本",
	"tui.failed":        "翻译失败: %v",
	"tui.keys":          "Ctrl+T 翻译 | Tab 焦点 | Ctrl+←/→ 切语言 | Ctrl+L 清屏 | Ctrl+U 清空输入 | Ctrl+S 互换 | Ctrl+O 离线 | Esc 系统命令行 | Ctrl+Q 退出",
	"tui.keys_short":    "Ctrl+T 翻译 | Ctrl+←/→ 切语言 | Ctrl+L 清屏 | Esc 系统命令行 | Ctrl+Q 退出",
	"tui.offline":       "离线",
	"tui.online":        "在线",
	"tui.src_api":       "在线 API",
	"tui.src_ai":        "在线 AI",
	"tui.src_cache":     "本地缓存",
	"tui.src_dict":      "内置词典",
	"tui.overview":      "原文",
	"tui.result":        "译文",
	"tui.shell_title":   "系统命令行 (Esc 返回翻译界面)",
	"tui.shell_input":   "输入系统命令,Enter 执行 (如 dir / ls / git status)",
	"tui.shell_banner":  "在此输入系统命令,输出会显示在这里。↑/↓ 翻历史,PgUp/PgDn 滚动,Ctrl+L 清屏,Esc 返回翻译界面。",
	"tui.shell_running": "执行中...",
	"tui.shell_timeout": "命令超时(%d 秒),已终止",
	"tui.shell_exit":    "[退出状态: %v]",
	"tui.hint_source":   "源语言: %s",
	"tui.hint_target":   "目标语言: %s",
	"tui.hint_cleared":  "已清空输入 (Esc 打开系统命令行)",
}

// --- English (en) messages ------------------------------------------------

var enMessages = map[string]string{
	// App
	"app.name":    "Tr - Terminal Translator",
	"app.version": "Tr version %s",
	"app.desc":    "A simple terminal translation tool written in Go.",
	"app.license": "License: MIT",
	"app.repo":    "Repository: https://github.com/Qiuxile/Tr",

	// Warnings
	"warn.defaults":          "Warning: %v; using defaults",
	"warn.cache_unavailable": "Warning: cache unavailable: %v",
	"warn.unknown_flag":      "Warning: unknown flag %s",
	"warn.no_tty_tui":        "Warning: not an interactive terminal; cannot start TUI",
	"warn.arrow_plain":       "Note: \"%s\" is not a recognized target language; translating the whole input as plain text",

	// Errors
	"err.empty_input":      "Error: no input text. Pass text as an argument, pipe it in, or run without arguments for the TUI.",
	"err.usage_hint":       "Usage: tr [text]",
	"err.bad_usage":        "Invalid arguments; nothing was executed. See tr -help for usage.",
	"err.translate_failed": "Error: %v",
	"err.empty":            "input text is empty",
	"err.no_translation":   "no translation available",
	"err.offline":          "offline mode: text not found in cache or dictionary",
	"err.no_key":           "Error: no AI key configured. Run tr config set api_key sk-... (or set TR_API_KEY); use -o for offline-only mode.",
	"err.network":          "network request failed",
	"err.file_usage":       "Usage: tr -f <file> [output-path]",
	"err.file_read":        "Failed to read file: %v",
	"err.file_write":       "Failed to write file: %v",

	// File translation
	"file.progress": "Translating %d/%d …",
	"file.partial":  "Warning: %d/%d blocks failed and kept their original text",
	"file.written":  "Written: %s",

	// Config
	"config.usage_show_set": "Usage: tr config show|set <key> <value>   (alias: tr -config ...)",
	"config.current":        "Current configuration:",
	"config.source_lang":    "  source_lang = %s",
	"config.target_lang":    "  target_lang = %s",
	"config.ui_lang":        "  ui_lang     = %s",
	"config.api_key":        "  api_key     = %s",
	"config.ai_base_url":    "  ai_base_url = %s",
	"config.ai_model":       "  ai_model    = %s",
	"config.ai_thinking":    "  ai_thinking = %v",
	"config.key_unset":      "(not set)",
	"config.file":           "  config file = %s",
	"config.usage_set":      "Usage: tr config set <key> <value>",
	"config.valid_keys":     "Valid keys: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "Invalid key: %s. Valid keys: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "Invalid boolean: %s (use true/false, on/off, 1/0)",
	"config.save_failed":    "Save failed: %v",
	"config.set_ok":         "Set: %s = %s",
	"config.unknown_subcmd": "Unknown subcommand: %s. Use: show, set",

	// Help
	"help.title":          "Usage:",
	"help.translate":      "  tr [text]                  Translate text and print result",
	"help.translate_auto": "  tr <text> -> <lang-code>  Auto-detect language (e.g. tr 你好 -> en)",
	"help.file":           "  tr -f <file> [output]      Translate a file (no output = stdout)",
	"help.tui":            "  tr (no args, tty)          Start the interactive TUI",
	"help.tui_cmd":        "  tr tui                     Explicitly start the TUI",
	"help.config_cmd":     "  tr config show|set <k> <v> View/set config (alias: -config)",
	"help.offline":        "  tr -o, --offline           Force offline mode (cache + dictionary)",
	"help.help":           "  tr -help                   Show this help",
	"help.version":        "  tr -version                Show version",
	"help.about":          "  tr -about                  Show about information",

	"help.flags_title":   "Flags:",
	"help.flags_config":  "  -config, --config, -c   View/set config (same as tr config ...)",
	"help.flags_help":    "  -help,   --help,   -h   Show help",
	"help.flags_version": "  -version,--version, -v  Show version",
	"help.flags_about":   "  -about,  --about,   -a  Show about",
	"help.flags_offline": "  -offline,--offline, -o  Force offline mode",

	"help.pipe_title":   "Piped input:",
	"help.pipe_example": "  echo \"hello\" | tr       Translate from stdin",
	"help.stdin_note":   "  Without arguments, text piped to stdin is translated automatically",

	"help.config_title":       "Configuration:",
	"help.config_source":      "  source_lang   Source language code (default: en; auto = detect)",
	"help.config_target":      "  target_lang   Target language code (default: zh)",
	"help.config_ui":          "  ui_lang       UI language (default: zh, options: en, ja)",
	"help.config_key":         "  api_key       AI key (sk-...); required for translation",
	"help.config_ai_url":      "  ai_base_url   OpenAI-compatible base URL (default: https://api.deepseek.com/v1)",
	"help.config_ai_model":    "  ai_model      Model id (default: deepseek-flash)",
	"help.config_ai_thinking": "  ai_thinking   Reasoning mode (default: false; off saves tokens)",
	"help.config_path":        "  Config file: %s",

	"help.langs_title": "Target language short codes:",
	"help.langs_line1": "  zh=Chinese en=English ja=Japanese ko=Korean ru=Russian fr=French de=German es=Spanish",
	"help.langs_line2": "  it=Italian pt=Portuguese ar=Arabic th=Thai vi=Vietnamese hi=Hindi el=Greek nl=Dutch",
	"help.langs_note":  "  Aliases (英文 / 中文 / 日本語 / korean / russian ...) and any ISO 639-1 code also work",
	"help.langs_once":  "  -> <lang> applies to this invocation only and never writes the config file (use tr config set target_lang to persist)",

	"help.backend_title":   "Backends:",
	"help.backend_ai":      "  - AI (OpenAI-compatible, api_key required; DeepSeek by default)",
	"help.backend_offline": "  - Offline: local cache + embedded dictionary (use -o)",

	"help.offline_title": "Offline features:",
	"help.offline_cache": "  - Translation cache: hits return instantly and cost no tokens",
	"help.offline_dict":  "  - Embedded dictionary: ~400 common EN->ZH words and phrases",
	"help.offline_flag":  "  - Use -o flag to force offline mode",

	"help.env_title": "Environment:",
	"help.env_key":   "  TR_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY can replace api_key",

	// Help (extra sections)
	"help.examples_title": "Examples:",
	"help.ex1":            "  tr \"Hello, world!\"              English → Chinese",
	"help.ex2":            "  tr 你好 -> en                   auto-detect, this run only",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            language codes and aliases both work",
	"help.ex4":            "  echo \"hello\" | tr               translate stdin",
	"help.ex5":            "  tr -f in.txt out/               translate a file into the out/ directory",
	"help.ex6":            "  tr config set api_key sk-xxx    configure the AI key",
	"help.flags_file":     "  -f, --file <file> [out]  Translate a file (stdout when out is omitted)",
	"help.tui_title":      "TUI keys:",
	"help.tui_keys1":      "  Ctrl+T translate | Enter newline | Ctrl+←/→ switch language | Ctrl+L clear screen | Ctrl+U clear input",
	"help.tui_keys2":      "  Ctrl+S swap languages | Ctrl+O offline | Esc system shell | Ctrl+Q quit",
	"help.tui_note":       "  Esc opens the system shell: run dir / ls / git there, output fills the screen",
	"help.quick_title":    "First run:",
	"help.quick_key":      "  tr config set api_key sk-...     translation needs an AI key (you get a hint otherwise)",
	"help.more_title":     "More:",
	"help.more_readme":    "  Full documentation lives in README.md (FAQ, troubleshooting, token thrift)",

	// About
	"about.version": "Version: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "create default config",
	"cfg.read":           "read config",
	"cfg.parse":          "parse config",
	"cfg.create_dir":     "create config dir",
	"cfg.marshal":        "marshal config",
	"cfg.write":          "write config",

	// AI errors
	"ai.request":    "AI request",
	"ai.parse":      "AI response parse",
	"ai.empty":      "AI returned no translation",
	"ai.bad_key":    "AI key rejected (HTTP %d): %s",
	"ai.no_balance": "AI account has insufficient balance (HTTP %d): %s",
	"ai.rate_limit": "AI request rate limited (HTTP %d): %s",
	"ai.status":     "AI returned status %d: %s",

	// TUI
	"tui.placeholder":   "Enter text to translate, then press Ctrl+T...",
	"tui.translating":   "Translating...",
	"tui.empty":         "Nothing to translate",
	"tui.failed":        "Translation failed: %v",
	"tui.keys":          "Ctrl+T translate | Tab focus | Ctrl+←/→ switch language | Ctrl+L clear screen | Ctrl+U clear input | Ctrl+S swap | Ctrl+O offline | Esc shell | Ctrl+Q quit",
	"tui.keys_short":    "Ctrl+T translate | Ctrl+←/→ switch language | Ctrl+L clear screen | Esc shell | Ctrl+Q quit",
	"tui.offline":       "offline",
	"tui.online":        "online",
	"tui.src_api":       "online API",
	"tui.src_ai":        "online AI",
	"tui.src_cache":     "local cache",
	"tui.src_dict":      "embedded dict",
	"tui.overview":      "Source",
	"tui.result":        "Translation",
	"tui.shell_title":   "System shell (Esc returns to the translator)",
	"tui.shell_input":   "Type a system command, Enter runs it (dir / ls / git status ...)",
	"tui.shell_banner":  "Run system commands here; their output shows below. ↑/↓ history, PgUp/PgDn scroll, Ctrl+L clears the screen, Esc returns to the translator.",
	"tui.shell_running": "running...",
	"tui.shell_timeout": "Command timed out after %d s and was killed",
	"tui.shell_exit":    "[exit status: %v]",
	"tui.hint_source":   "Source language: %s",
	"tui.hint_target":   "Target language: %s",
	"tui.hint_cleared":  "Input cleared (Esc opens the system shell)",
}

// --- Japanese (ja) messages -----------------------------------------------

var jaMessages = map[string]string{
	// App
	"app.name":    "Tr - ターミナル翻訳ツール",
	"app.version": "Tr バージョン %s",
	"app.desc":    "Go で書かれたシンプルなターミナル翻訳ツール",
	"app.license": "ライセンス: MIT",
	"app.repo":    "リポジトリ: https://github.com/Qiuxile/Tr",

	// Warnings
	"warn.defaults":          "警告: %v; デフォルト設定を使用します",
	"warn.cache_unavailable": "警告: キャッシュが利用できません: %v",
	"warn.unknown_flag":      "警告: 不明なオプション %s",
	"warn.no_tty_tui":        "警告: 対話型ターミナルではないため TUI を起動できません",
	"warn.arrow_plain":       "注意: \"%s\" をターゲット言語として認識できません。入力全体をプレーンテキストとして翻訳します",

	// Errors
	"err.empty_input":      "エラー: 入力テキストがありません。引数またはパイプで渡すか、引数なしで TUI を起動してください。",
	"err.usage_hint":       "使い方: tr [テキスト]",
	"err.bad_usage":        "引数が不正です。何も実行しませんでした。tr -help で使い方を確認してください。",
	"err.translate_failed": "エラー: %v",
	"err.empty":            "入力テキストが空です",
	"err.no_translation":   "翻訳が見つかりません",
	"err.offline":          "オフラインモード: キャッシュと辞書にテキストが見つかりません",
	"err.no_key":           "エラー: AI キーが未設定です。tr config set api_key sk-... を実行するか、環境変数 TR_API_KEY を設定してください。オフラインのみの場合は -o を付けてください。",
	"err.network":          "ネットワークリクエストに失敗しました",
	"err.file_usage":       "使い方: tr -f <ファイル> [出力パス]",
	"err.file_read":        "ファイルの読み取りに失敗しました: %v",
	"err.file_write":       "ファイルの書き込みに失敗しました: %v",

	// File translation
	"file.progress": "翻訳中 %d/%d …",
	"file.partial":  "警告: %d/%d のブロックが失敗し、原文を保持しました",
	"file.written":  "書き込み完了: %s",

	// Config
	"config.usage_show_set": "使い方: tr config show|set <キー> <値>   (別名: tr -config ...)",
	"config.current":        "現在の設定:",
	"config.source_lang":    "  source_lang = %s",
	"config.target_lang":    "  target_lang = %s",
	"config.ui_lang":        "  ui_lang     = %s",
	"config.api_key":        "  api_key      = %s",
	"config.ai_base_url":    "  ai_base_url  = %s",
	"config.ai_model":       "  ai_model     = %s",
	"config.ai_thinking":    "  ai_thinking  = %v",
	"config.key_unset":      "(未設定)",
	"config.file":           "  設定ファイル = %s",
	"config.usage_set":      "使い方: tr config set <キー> <値>",
	"config.valid_keys":     "有効なキー: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "無効なキー: %s。有効なキー: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "無効な真偽値: %s (true/false, on/off, 1/0 を使用)",
	"config.save_failed":    "保存に失敗しました: %v",
	"config.set_ok":         "設定しました: %s = %s",
	"config.unknown_subcmd": "不明なサブコマンド: %s。使用可能: show, set",

	// Help
	"help.title":          "使い方:",
	"help.translate":      "  tr [テキスト]                テキストを翻訳して結果を表示",
	"help.translate_auto": "  tr <テキスト> -> <言語コード>  言語を自動判定して翻訳 (例: tr 你好 -> en)",
	"help.file":           "  tr -f <ファイル> [出力]      ファイルを翻訳 (出力省略時は標準出力)",
	"help.tui":            "  tr (引数なし, tty)           対話型 TUI を起動",
	"help.tui_cmd":        "  tr tui                       明示的に TUI を起動",
	"help.config_cmd":     "  tr config show|set <キー> <値>  設定の表示/変更 (-config 互換)",
	"help.offline":        "  tr -o, --offline             オフラインモードを強制 (キャッシュ+辞書)",
	"help.help":           "  tr -help                     ヘルプを表示",
	"help.version":        "  tr -version                  バージョンを表示",
	"help.about":          "  tr -about                    アプリ情報を表示",

	"help.flags_title":   "オプション:",
	"help.flags_config":  "  -config, --config, -c   設定の表示/変更 (tr config ... と同等)",
	"help.flags_help":    "  -help,   --help,   -h   ヘルプを表示",
	"help.flags_version": "  -version,--version, -v  バージョンを表示",
	"help.flags_about":   "  -about,  --about,   -a  アプリ情報を表示",
	"help.flags_offline": "  -offline,--offline, -o  オフラインモードを強制",

	"help.pipe_title":   "パイプ入力:",
	"help.pipe_example": "  echo \"hello\" | tr        標準入力から翻訳",
	"help.stdin_note":   "  引数なしで stdin にパイプされたテキストは自動的に翻訳されます",

	"help.config_title":       "設定:",
	"help.config_source":      "  source_lang   ソース言語コード (デフォルト: en; auto = 自動判定)",
	"help.config_target":      "  target_lang   ターゲット言語コード (デフォルト: zh)",
	"help.config_ui":          "  ui_lang       UI言語 (デフォルト: zh, オプション: en, ja)",
	"help.config_key":         "  api_key       AI キー (sk-...)、翻訳に必須",
	"help.config_ai_url":      "  ai_base_url   OpenAI 互換 URL (デフォルト: https://api.deepseek.com/v1)",
	"help.config_ai_model":    "  ai_model      モデル名 (デフォルト: deepseek-flash)",
	"help.config_ai_thinking": "  ai_thinking   思考モード (デフォルト: false、オフでトークン節約)",
	"help.config_path":        "  設定ファイル: %s",

	"help.langs_title": "ターゲット言語の短縮コード:",
	"help.langs_line1": "  zh=中国語(簡体) en=英語 ja=日本語 ko=韓国語 ru=ロシア語 fr=フランス語 de=ドイツ語 es=スペイン語",
	"help.langs_line2": "  it=イタリア語 pt=ポルトガル語 ar=アラビア語 th=タイ語 vi=ベトナム語 hi=ヒンディー語 el=ギリシャ語 nl=オランダ語",
	"help.langs_note":  "  別名(英文 / 中文 / 日本語 / korean / russian …)と任意の ISO 639-1 コードも使用できます",
	"help.langs_once":  "  -> <言語> は今回の実行にのみ有効で、設定ファイルは変更しません(永続化は tr config set target_lang)",

	"help.backend_title":   "バックエンド:",
	"help.backend_ai":      "  - AI (OpenAI 互換, api_key 必須; 既定は DeepSeek)",
	"help.backend_offline": "  - オフライン: ローカルキャッシュ + 内蔵辞書 (-o で使用)",

	"help.offline_title": "オフライン機能:",
	"help.offline_cache": "  - 翻訳キャッシュ: ヒット時は即時返却、トークンを消費しません",
	"help.offline_dict":  "  - 内蔵辞書: 約400の英→中単語・フレーズ",
	"help.offline_flag":  "  - -o オプションでオフラインモードを強制",

	"help.env_title": "環境変数:",
	"help.env_key":   "  TR_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY を api_key の代わりに使用可能",

	// Help (extra sections)
	"help.examples_title": "例:",
	"help.ex1":            "  tr \"Hello, world!\"              英語 → 中国語",
	"help.ex2":            "  tr 你好 -> en                   言語を自動判定 (今回のみ有効)",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            コードと別名の両方に対応",
	"help.ex4":            "  echo \"hello\" | tr               標準入力から翻訳",
	"help.ex5":            "  tr -f in.txt out/               ファイルを out/ ディレクトリへ翻訳",
	"help.ex6":            "  tr config set api_key sk-xxx    AI キーを設定",
	"help.flags_file":     "  -f, --file <ファイル> [出力]  ファイル翻訳 (出力省略時は標準出力)",
	"help.tui_title":      "TUI キー:",
	"help.tui_keys1":      "  Ctrl+T 翻訳 | Enter 改行 | Ctrl+←/→ 言語切替 | Ctrl+L 画面クリア | Ctrl+U 入力クリア",
	"help.tui_keys2":      "  Ctrl+S 言語入替 | Ctrl+O オフライン | Esc システムコマンドライン | Ctrl+Q 終了",
	"help.tui_note":       "  Esc でシステムコマンドライン:dir / ls / git などを実行でき、出力は全画面に表示",
	"help.quick_title":    "初回設定:",
	"help.quick_key":      "  tr config set api_key sk-...     翻訳には AI キーが必要です (未設定時は案内が出ます)",
	"help.more_title":     "詳細:",
	"help.more_readme":    "  詳しいドキュメントは README.md にあります (FAQ・トラブルシューティング・トークン節約)",

	// About
	"about.version": "バージョン: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "デフォルト設定の作成",
	"cfg.read":           "設定の読み取り",
	"cfg.parse":          "設定の解析",
	"cfg.create_dir":     "設定ディレクトリの作成",
	"cfg.marshal":        "設定のシリアル化",
	"cfg.write":          "設定の書き込み",

	// AI errors
	"ai.request":    "AI リクエスト",
	"ai.parse":      "AI 応答の解析",
	"ai.empty":      "AI が翻訳を返しませんでした",
	"ai.bad_key":    "AI キーが拒否されました (HTTP %d): %s",
	"ai.no_balance": "AI アカウントの残高が不足しています (HTTP %d): %s",
	"ai.rate_limit": "AI リクエストがレート制限されました (HTTP %d): %s",
	"ai.status":     "AI がステータス %d を返しました: %s",

	// TUI
	"tui.placeholder":   "翻訳するテキストを入力し、Ctrl+T で翻訳...",
	"tui.translating":   "翻訳中...",
	"tui.empty":         "翻訳するテキストがありません",
	"tui.failed":        "翻訳に失敗しました: %v",
	"tui.keys":          "Ctrl+T 翻訳 | Tab フォーカス | Ctrl+←/→ 言語切替 | Ctrl+L 画面クリア | Ctrl+U 入力をクリア | Ctrl+S 入替 | Ctrl+O オフライン | Esc システムコマンドライン | Ctrl+Q 終了",
	"tui.keys_short":    "Ctrl+T 翻訳 | Ctrl+←/→ 言語切替 | Ctrl+L 画面クリア | Esc システムコマンドライン | Ctrl+Q 終了",
	"tui.offline":       "オフライン",
	"tui.online":        "オンライン",
	"tui.src_api":       "オンライン API",
	"tui.src_ai":        "オンライン AI",
	"tui.src_cache":     "ローカルキャッシュ",
	"tui.src_dict":      "内蔵辞書",
	"tui.overview":      "原文",
	"tui.result":        "訳文",
	"tui.shell_title":   "システムコマンドライン (Esc で翻訳画面に戻る)",
	"tui.shell_input":   "システムコマンドを入力し Enter で実行 (dir / ls / git status など)",
	"tui.shell_banner":  "ここでシステムコマンドを実行でき、出力は下に表示されます。↑/↓ 履歴、PgUp/PgDn スクロール、Ctrl+L で画面クリア、Esc で翻訳画面に戻ります。",
	"tui.shell_running": "実行中...",
	"tui.shell_timeout": "コマンドがタイムアウトしました (%d 秒)",
	"tui.shell_exit":    "[終了ステータス: %v]",
	"tui.hint_source":   "ソース言語: %s",
	"tui.hint_target":   "ターゲット言語: %s",
	"tui.hint_cleared":  "入力をクリアしました (Esc でシステムコマンドライン)",
}
