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
	"app.desc":    "一个用 Go 编写的终端 AI 翻译工具",
	"app.license": "许可证: Apache-2.0",
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
	"err.no_key":           "错误: 未配置 AI 密钥。请运行 tr config set api_key sk-... 或设置 TR_API_KEY。",
	"err.network":          "网络请求失败",
	"err.file_usage":       "用法: tr -f <文件> [输出路径]",
	"err.file_read":        "读取文件失败: %v",
	"err.file_write":       "写入文件失败: %v",

	// File translation
	"file.progress": "翻译中 %d/%d …",
	"file.partial":  "警告: %d/%d 个片段翻译失败,已保留原文",
	"file.written":  "已写入: %s",

	// Config
	"config.usage_show_set": "用法: tr config show|set|path ...   (兼容: tr -config ...)",
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
	"config.usage_set":      "用法: tr config set <键> <值>   (api_key 可省略值,交互输入)",
	"config.valid_keys":     "可用键: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "无效的配置项: %s。可用键: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "无效的布尔值: %s (可用: true/false, on/off, 1/0)",
	"config.save_failed":    "保存失败: %v",
	"config.set_ok":         "设置成功: %s = %s",
	"config.unknown_subcmd": "未知子命令: %s。可用: show, set, path",

	// Help
	"help.title":          "用法:",
	"help.translate":      "  tr [文本]                  翻译文本并输出结果",
	"help.translate_auto": "  tr <文本> -> <语言代码>     自动识别语言翻译 (例: tr 你好 -> en)",
	"help.file":           "  tr -f <文件> [输出]         翻译文件 (省略输出则打印到终端)",
	"help.tui":            "  tr (无参数, 交互终端)      启动交互式终端界面 (TUI)",
	"help.tui_cmd":        "  tr tui                     显式启动 TUI",
	"help.config_cmd":     "  tr config show|set|path ...   查看/修改配置 (-config 兼容)",
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
	"help.config_key":         "  api_key       AI 密钥;默认接口可免密钥,自定义服务通常需要",
	"help.config_ai_url":      "  ai_base_url   AI 后端 OpenAI 兼容地址 (默认: https://open.bigmodel.cn/api/paas/v4)",
	"help.config_ai_model":    "  ai_model      AI 后端模型名 (默认: glm-5.3-flash)",
	"help.config_ai_thinking": "  ai_thinking   DeepSeek 思考模式 (默认: false,关闭更省 token)",
	"help.config_path":        "  配置文件: %s",

	"help.langs_title": "目标语言简写:",
	"help.langs_line1": "  zh=中文(简) zh-hant=繁体 en=英语 ja=日语 ko=韩语 ru=俄语 fr=法语 de=德语 es=西班牙语",
	"help.langs_line2": "  it=意大利语 pt=葡萄牙语 ar=阿拉伯语 th=泰语 vi=越南语 hi=印地语 el=希腊语 nl=荷兰语",
	"help.langs_note":  "  也支持别名(英文 / 中文 / 日本語 / korean / russian …)与任意 ISO 639-1 代码",
	"help.langs_once":  "  -> <语言> 只作用于本次执行,不会写入配置文件;持久修改请用 tr config set target_lang",
	"help.arrow_args":  "  箭头模式只取第一个参数作为内容(含空格请加引号),其余参数一律忽略",
	"help.long_text":   "  长文本超过约 1500 字符会自动分块翻译",

	"help.backend_title":   "翻译后端:",
	"help.backend_ai":      "  - AI (OpenAI 兼容;默认智谱 BigModel GLM-5.3-Flash,可换 DeepSeek/OpenAI)",
	"help.backend_offline": "  - 离线: 本地缓存 + 内置词典 (加 -o 使用)",

	"help.offline_title": "离线功能:",
	"help.offline_cache": "  - 翻译缓存: 命中缓存直接返回, 不消耗任何 token",
	"help.offline_dict":  "  - 内置词典: 约 460 个常用英→中词汇和短语(逐词翻译,仅供应急)",
	"help.offline_flag":  "  - 使用 -o 选项强制离线模式",

	"help.env_title": "环境变量:",
	"help.env_key":   "  TR_API_KEY / ZHIPU_API_KEY / GLM_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY 可替代 api_key(环境变量优先)",

	// Help (extra sections)
	"help.examples_title": "示例:",
	"help.ex1":            "  tr \"Hello, world!\"              英 → 中",
	"help.ex2":            "  tr 你好 -> en                   自动识别语言 (只作用于本次执行)",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            目标支持语言代码与名称别名",
	"help.ex4":            "  echo \"hello\" | tr               管道输入翻译",
	"help.ex5":            "  tr -f in.txt out/               翻译文件并写入 out/ 目录",
	"help.ex6":            "  tr config set api_key sk-xxx    可选: 配置自己的 AI 密钥",
	"help.flags_file":     "  -f, --file <文件> [输出]  翻译文件 (省略输出则打印到终端)",
	"help.tui_title":      "TUI 快捷键:",
	"help.tui_keys1":      "  Ctrl+T 翻译 | Enter 换行 | Ctrl+←/→ 切语言 | Ctrl+L 清屏 | Ctrl+U 清空输入",
	"help.tui_keys2":      "  Ctrl+S 互换语言 | Ctrl+O 离线 | Esc 系统命令行 | Ctrl+Q 退出",
	"help.tui_note":       "  Esc 进入系统命令行:可直接运行 tr config show 等命令,输出整屏显示",
	"help.quick_title":    "首次使用:",
	"help.quick_key":      "  发布版可内置默认 key;源码构建请先 tr config set api_key sk-...",
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
	"ai.truncated":  "AI 输出被长度上限截断(译文不完整)。输入过长时请分段翻译,或用 tr -f <文件>",
	"ai.progress":   "长文本分块翻译中 %d/%d …",

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

	// v2.2: free default backend and cache/config helpers
	"err.file_encoding":          "文件不是有效的 UTF-8 编码,请先转换为 UTF-8 再翻译",
	"err.file_too_large":         "文件超过 %d MiB 上限,请拆分后再翻译",
	"err.stdin_too_large":        "标准输入超过 %d MiB 上限",
	"err.tui_start":              "TUI 启动失败: %v",
	"warn.cache_save":            "警告: 缓存写入失败: %v",
	"config.key_free":            "(免费接口,无需密钥)",
	"config.key_env":             "(来自环境变量 %s)",
	"config.api_key_prompt":      "请输入 AI 密钥(输入不回显): ",
	"config.api_key_read_failed": "读取密钥失败: %v",
	"config.invalid_lang":        "无效的语言代码或别名: %s",
	"config.invalid_ui_lang":     "无效的界面语言: %s (可用: zh, en, ja)",
	"config.invalid_url":         "无效的 ai_base_url: %s (%v)",
	"config.invalid_model":       "ai_model 不能为空",
	"help.cache_cmd":             "  tr cache clear|stats       清空/查看翻译缓存",
	"help.flags_dashdash":        "  --                        结束选项解析(用于翻译 config/tui 等保留字)",
	"cache.usage":                "用法: tr cache clear|stats",
	"cache.cleared":              "缓存已清空",
	"cache.stats":                "缓存条目: %d\n缓存文件: %s",
	"ai.free_rate_limit":         "默认免费接口有频率限制,请稍后重试;或在 tr config show 中改用带密钥的服务商。",
	"tui.shell_cancelled":        "命令已取消",
	"tui.shell_truncated":        "输出超过 %d KiB,已截断",
	"err.file_binary":            "文件看起来是二进制文件,已拒绝翻译",
	"ai.echo":                    "AI 返回了原文,可能未执行翻译",
	"ai.free_limit":              "免费 AI 接口限流(约每 30 秒 1 次),请稍后重试或配置自己的服务商",
	"config.key_default":         "(内置默认密钥 %s)",
}

// --- English (en) messages ------------------------------------------------

var enMessages = map[string]string{
	// App
	"app.name":    "Tr - Terminal Translator",
	"app.version": "Tr version %s",
	"app.desc":    "A simple terminal AI translation tool written in Go.",
	"app.license": "License: Apache-2.0",
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
	"err.no_key":           "Error: no AI key configured. Run tr config set api_key sk-... or set TR_API_KEY.",
	"err.network":          "network request failed",
	"err.file_usage":       "Usage: tr -f <file> [output-path]",
	"err.file_read":        "Failed to read file: %v",
	"err.file_write":       "Failed to write file: %v",

	// File translation
	"file.progress": "Translating %d/%d …",
	"file.partial":  "Warning: %d/%d blocks failed and kept their original text",
	"file.written":  "Written: %s",

	// Config
	"config.usage_show_set": "Usage: tr config show|set|path ...   (alias: tr -config ...)",
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
	"config.usage_set":      "Usage: tr config set <key> <value>   (api_key may omit the value for a hidden prompt)",
	"config.valid_keys":     "Valid keys: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "Invalid key: %s. Valid keys: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "Invalid boolean: %s (use true/false, on/off, 1/0)",
	"config.save_failed":    "Save failed: %v",
	"config.set_ok":         "Set: %s = %s",
	"config.unknown_subcmd": "Unknown subcommand: %s. Use: show, set, path",

	// Help
	"help.title":          "Usage:",
	"help.translate":      "  tr [text]                  Translate text and print result",
	"help.translate_auto": "  tr <text> -> <lang-code>  Auto-detect language (e.g. tr 你好 -> en)",
	"help.file":           "  tr -f <file> [output]      Translate a file (no output = stdout)",
	"help.tui":            "  tr (no args, tty)          Start the interactive TUI",
	"help.tui_cmd":        "  tr tui                     Explicitly start the TUI",
	"help.config_cmd":     "  tr config show|set|path ...   View/set config (alias: -config)",
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
	"help.config_key":         "  api_key       AI key; optional for the default keyless endpoint, required by most providers",
	"help.config_ai_url":      "  ai_base_url   AI backend OpenAI-compatible URL (default: https://open.bigmodel.cn/api/paas/v4)",
	"help.config_ai_model":    "  ai_model      AI backend model id (default: glm-5.3-flash)",
	"help.config_ai_thinking": "  ai_thinking   DeepSeek reasoning mode (default: false; off saves tokens)",
	"help.config_path":        "  Config file: %s",

	"help.langs_title": "Target language short codes:",
	"help.langs_line1": "  zh=Chinese(simplified) zh-hant=traditional en=English ja=Japanese ko=Korean ru=Russian fr=French de=German es=Spanish",
	"help.langs_line2": "  it=Italian pt=Portuguese ar=Arabic th=Thai vi=Vietnamese hi=Hindi el=Greek nl=Dutch",
	"help.langs_note":  "  Aliases (英文 / 中文 / 日本語 / korean / russian ...) and any ISO 639-1 code also work",
	"help.langs_once":  "  -> <lang> applies to this invocation only and never writes the config file (use tr config set target_lang to persist)",
	"help.arrow_args":  "  Arrow mode uses the first argument as the content (quote it if it has spaces); other arguments are ignored",
	"help.long_text":   "  Text over ~1500 characters is translated in chunks",

	"help.backend_title":   "Backends:",
	"help.backend_ai":      "  - AI (OpenAI-compatible; Zhipu BigModel GLM-5.3-Flash default, or DeepSeek/OpenAI)",
	"help.backend_offline": "  - Offline: local cache + embedded dictionary (use -o)",

	"help.offline_title": "Offline features:",
	"help.offline_cache": "  - Translation cache: hits return instantly and cost no tokens",
	"help.offline_dict":  "  - Embedded dictionary: ~460 common EN->ZH words and phrases (word-by-word, emergency use)",
	"help.offline_flag":  "  - Use -o flag to force offline mode",

	"help.env_title": "Environment:",
	"help.env_key":   "  TR_API_KEY / ZHIPU_API_KEY / GLM_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY override api_key (environment wins)",

	// Help (extra sections)
	"help.examples_title": "Examples:",
	"help.ex1":            "  tr \"Hello, world!\"              English → Chinese",
	"help.ex2":            "  tr 你好 -> en                   auto-detect, this run only",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            language codes and aliases both work",
	"help.ex4":            "  echo \"hello\" | tr               translate stdin",
	"help.ex5":            "  tr -f in.txt out/               translate a file into the out/ directory",
	"help.ex6":            "  tr config set api_key sk-xxx    optional: use your own AI key",
	"help.flags_file":     "  -f, --file <file> [out]  Translate a file (stdout when out is omitted)",
	"help.tui_title":      "TUI keys:",
	"help.tui_keys1":      "  Ctrl+T translate | Enter newline | Ctrl+←/→ switch language | Ctrl+L clear screen | Ctrl+U clear input",
	"help.tui_keys2":      "  Ctrl+S swap languages | Ctrl+O offline | Esc system shell | Ctrl+Q quit",
	"help.tui_note":       "  Esc opens the system shell: run dir / ls / git there, output fills the screen",
	"help.quick_title":    "First run:",
	"help.quick_key":      "  Release builds can embed a default key; source builds: tr config set api_key sk-...",
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
	"ai.truncated":  "The AI reply was cut off at the length limit (incomplete translation). Split long input, or use tr -f <file>",
	"ai.progress":   "Translating long text in chunks %d/%d …",

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

	// v2.2: free default backend and cache/config helpers
	"err.file_encoding":          "The file is not valid UTF-8; convert it to UTF-8 before translating",
	"err.file_too_large":         "The file exceeds the %d MiB limit; split it before translating",
	"err.stdin_too_large":        "Standard input exceeds the %d MiB limit",
	"err.tui_start":              "Failed to start the TUI: %v",
	"warn.cache_save":            "Warning: failed to write the cache: %v",
	"config.key_free":            "(free endpoint, no key required)",
	"config.key_env":             "(from environment %s)",
	"config.api_key_prompt":      "Enter the AI key (input hidden): ",
	"config.api_key_read_failed": "Failed to read the key: %v",
	"config.invalid_lang":        "Invalid language code or alias: %s",
	"config.invalid_ui_lang":     "Invalid UI language: %s (use: zh, en, ja)",
	"config.invalid_url":         "Invalid ai_base_url: %s (%v)",
	"config.invalid_model":       "ai_model must not be empty",
	"help.cache_cmd":             "  tr cache clear|stats       Clear or inspect the translation cache",
	"help.flags_dashdash":        "  --                        End flag parsing (translate reserved words such as config/tui)",
	"cache.usage":                "Usage: tr cache clear|stats",
	"cache.cleared":              "Cache cleared",
	"cache.stats":                "Cached entries: %d\nCache file: %s",
	"ai.free_rate_limit":         "The free default is rate-limited; retry shortly, or configure a keyed provider via tr config show.",
	"tui.shell_cancelled":        "Command cancelled",
	"tui.shell_truncated":        "Output exceeded %d KiB and was truncated",
	"err.file_binary":            "The file looks binary; refusing to translate it",
	"ai.echo":                    "The AI returned the input unchanged; translation may have failed",
	"ai.free_limit":              "Free AI endpoint is rate-limited (about one request every 30 seconds); retry later or configure your own provider",
	"config.key_default":         "(built-in default key %s)",
}

// --- Japanese (ja) messages -----------------------------------------------

var jaMessages = map[string]string{
	// App
	"app.name":    "Tr - ターミナル翻訳ツール",
	"app.version": "Tr バージョン %s",
	"app.desc":    "Go 製のシンプルなターミナル AI 翻訳ツール",
	"app.license": "ライセンス: Apache-2.0",
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
	"err.no_key":           "エラー: AI キーが設定されていません。tr config set api_key sk-... または TR_API_KEY を設定してください。",
	"err.network":          "ネットワークリクエストに失敗しました",
	"err.file_usage":       "使い方: tr -f <ファイル> [出力パス]",
	"err.file_read":        "ファイルの読み取りに失敗しました: %v",
	"err.file_write":       "ファイルの書き込みに失敗しました: %v",

	// File translation
	"file.progress": "翻訳中 %d/%d …",
	"file.partial":  "警告: %d/%d のブロックが失敗し、原文を保持しました",
	"file.written":  "書き込み完了: %s",

	// Config
	"config.usage_show_set": "使い方: tr config show|set|path ...   (別名: tr -config ...)",
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
	"config.usage_set":      "使い方: tr config set <キー> <値>   (api_key は値を省略して非表示入力できます)",
	"config.valid_keys":     "有効なキー: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_key":    "無効なキー: %s。有効なキー: api_key, ai_model, ai_base_url, ai_thinking, source_lang, target_lang, ui_lang",
	"config.invalid_bool":   "無効な真偽値: %s (true/false, on/off, 1/0 を使用)",
	"config.save_failed":    "保存に失敗しました: %v",
	"config.set_ok":         "設定しました: %s = %s",
	"config.unknown_subcmd": "不明なサブコマンド: %s。使用可能: show, set, path",

	// Help
	"help.title":          "使い方:",
	"help.translate":      "  tr [テキスト]                テキストを翻訳して結果を表示",
	"help.translate_auto": "  tr <テキスト> -> <言語コード>  言語を自動判定して翻訳 (例: tr 你好 -> en)",
	"help.file":           "  tr -f <ファイル> [出力]      ファイルを翻訳 (出力省略時は標準出力)",
	"help.tui":            "  tr (引数なし, tty)           対話型 TUI を起動",
	"help.tui_cmd":        "  tr tui                       明示的に TUI を起動",
	"help.config_cmd":     "  tr config show|set|path ...   設定の表示/変更 (-config 互換)",
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
	"help.config_key":         "  api_key       AI キー。デフォルトのキー不要エンドポイント以外では通常必要",
	"help.config_ai_url":      "  ai_base_url   AI バックエンドの OpenAI 互換 URL (デフォルト: https://open.bigmodel.cn/api/paas/v4)",
	"help.config_ai_model":    "  ai_model      AI バックエンドのモデル名 (デフォルト: glm-5.3-flash)",
	"help.config_ai_thinking": "  ai_thinking   DeepSeek の思考モード (デフォルト: false、オフで token 節約)",
	"help.config_path":        "  設定ファイル: %s",

	"help.langs_title": "ターゲット言語の短縮コード:",
	"help.langs_line1": "  zh=中国語(簡体) zh-hant=繁体 en=英語 ja=日本語 ko=韓国語 ru=ロシア語 fr=フランス語 de=ドイツ語 es=スペイン語",
	"help.langs_line2": "  it=イタリア語 pt=ポルトガル語 ar=アラビア語 th=タイ語 vi=ベトナム語 hi=ヒンディー語 el=ギリシャ語 nl=オランダ語",
	"help.langs_note":  "  別名(英文 / 中文 / 日本語 / korean / russian …)と任意の ISO 639-1 コードも使用できます",
	"help.langs_once":  "  -> <言語> は今回の実行にのみ有効で、設定ファイルは変更しません(永続化は tr config set target_lang)",
	"help.arrow_args":  "  矢印構文は最初の引数のみを内容として扱います(空白を含む場合は引用符で括る)。他の引数は無視されます",
	"help.long_text":   "  長文は約1500文字を超えると分割して翻訳します",

	"help.backend_title":   "バックエンド:",
	"help.backend_ai":      "  - AI (OpenAI 互換;デフォルトは智谱 BigModel GLM-5.3-Flash、DeepSeek/OpenAI も可)",
	"help.backend_offline": "  - オフライン: ローカルキャッシュ + 内蔵辞書 (-o で使用)",

	"help.offline_title": "オフライン機能:",
	"help.offline_cache": "  - 翻訳キャッシュ: ヒット時は即時返却、トークンを消費しません",
	"help.offline_dict":  "  - 内蔵辞書: 約 460 の英→中単語・短句 (逐語訳、緊急用)",
	"help.offline_flag":  "  - -o オプションでオフラインモードを強制",

	"help.env_title": "環境変数:",
	"help.env_key":   "  TR_API_KEY / ZHIPU_API_KEY / GLM_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY は api_key より優先されます",

	// Help (extra sections)
	"help.examples_title": "例:",
	"help.ex1":            "  tr \"Hello, world!\"              英語 → 中国語",
	"help.ex2":            "  tr 你好 -> en                   言語を自動判定 (今回のみ有効)",
	"help.ex3":            "  tr \"Bonjour\" -> 中文            コードと別名の両方に対応",
	"help.ex4":            "  echo \"hello\" | tr               標準入力から翻訳",
	"help.ex5":            "  tr -f in.txt out/               ファイルを out/ ディレクトリへ翻訳",
	"help.ex6":            "  tr config set api_key sk-xxx    任意: 自分の AI キーを設定",
	"help.flags_file":     "  -f, --file <ファイル> [出力]  ファイル翻訳 (出力省略時は標準出力)",
	"help.tui_title":      "TUI キー:",
	"help.tui_keys1":      "  Ctrl+T 翻訳 | Enter 改行 | Ctrl+←/→ 言語切替 | Ctrl+L 画面クリア | Ctrl+U 入力クリア",
	"help.tui_keys2":      "  Ctrl+S 言語入替 | Ctrl+O オフライン | Esc システムコマンドライン | Ctrl+Q 終了",
	"help.tui_note":       "  Esc でシステムコマンドライン:dir / ls / git などを実行でき、出力は全画面に表示",
	"help.quick_title":    "初回設定:",
	"help.quick_key":      "  リリース版はデフォルトキーを内蔵できます。ソースビルドでは tr config set api_key sk-... を実行してください",
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
	"ai.truncated":  "AI の出力が長さ制限で打ち切られました (翻訳が不完全)。長い入力は分割するか tr -f <ファイル> を使用してください",
	"ai.progress":   "長文を分割して翻訳中 %d/%d …",

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

	// v2.2: free default backend and cache/config helpers
	"err.file_encoding":          "ファイルが有効な UTF-8 ではありません。UTF-8 に変換してから翻訳してください",
	"err.file_too_large":         "ファイルが %d MiB の上限を超えています。分割してから翻訳してください",
	"err.stdin_too_large":        "標準入力が %d MiB の上限を超えています",
	"err.tui_start":              "TUI の起動に失敗しました: %v",
	"warn.cache_save":            "警告: キャッシュの書き込みに失敗しました: %v",
	"config.key_free":            "(無料エンドポイント、キー不要)",
	"config.key_env":             "(環境変数 %s から)",
	"config.api_key_prompt":      "AI キーを入力してください (入力は表示されません): ",
	"config.api_key_read_failed": "キーの読み取りに失敗しました: %v",
	"config.invalid_lang":        "無効な言語コードまたは別名: %s",
	"config.invalid_ui_lang":     "無効な UI 言語: %s (zh, en, ja のいずれか)",
	"config.invalid_url":         "無効な ai_base_url: %s (%v)",
	"config.invalid_model":       "ai_model は空にできません",
	"help.cache_cmd":             "  tr cache clear|stats       翻訳キャッシュを消去/確認",
	"help.flags_dashdash":        "  --                        オプション解析を終了 (config/tui などの予約語を翻訳)",
	"cache.usage":                "使い方: tr cache clear|stats",
	"cache.cleared":              "キャッシュを消去しました",
	"cache.stats":                "キャッシュ件数: %d\nキャッシュファイル: %s",
	"ai.free_rate_limit":         "無料デフォルトはレート制限されています。少し待って再試行するか、tr config show でキー付きサービスに切り替えてください。",
	"tui.shell_cancelled":        "コマンドをキャンセルしました",
	"tui.shell_truncated":        "出力が %d KiB を超えたため切り詰めました",
	"err.file_binary":            "バイナリファイルの可能性があるため翻訳を中止しました",
	"ai.echo":                    "AI が原文をそのまま返しました。翻訳に失敗した可能性があります",
	"ai.free_limit":              "無料 AI エンドポイントは約30秒に1回の制限があります。後で再試行するか、独自サービスを設定してください",
	"config.key_default":         "(内蔵デフォルトキー %s)",
}
