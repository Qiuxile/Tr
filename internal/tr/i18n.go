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
	"err.translate_failed": "错误: %v",
	"err.empty":            "输入文本不能为空",
	"err.no_translation":   "无可用翻译",
	"err.offline":          "离线模式: 缓存和词典中均未找到该文本",
	"err.network":          "网络请求失败",

	// Config
	"config.usage_show_set":   "用法: tr config show|set <键> <值>   (兼容: tr -config ...)",
	"config.current":          "当前配置:",
	"config.source_lang":      "  source_lang = %s",
	"config.target_lang":      "  target_lang = %s",
	"config.api_url":          "  api_url     = %s",
	"config.ui_lang":          "  ui_lang     = %s",
	"config.backend_mymemory": "  后端         = MyMemory (免费)",
	"config.backend_custom":   "  后端         = 自定义 API",
	"config.file":             "  配置文件     = %s",
	"config.usage_set":        "用法: tr config set <键> <值>",
	"config.valid_keys":       "可用键: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":      "无效的配置项: %s。可用键: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":      "保存失败: %v",
	"config.set_ok":           "设置成功: %s = %s",
	"config.unknown_subcmd":   "未知子命令: %s。可用: show, set",

	// Help
	"help.title":          "用法:",
	"help.translate":      "  tr [文本]                  翻译文本并输出结果",
	"help.translate_auto": "  tr <文本> -> <语言代码>     自动识别语言翻译 (例: tr 你好 -> en)",
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

	"help.config_title":  "配置:",
	"help.config_source": "  source_lang   源语言代码 (默认: en; auto = 自动检测)",
	"help.config_target": "  target_lang   目标语言代码 (默认: zh)",
	"help.config_api":    "  api_url       自定义 API 地址 (默认: None = MyMemory)",
	"help.config_ui":     "  ui_lang       界面语言 (默认: zh, 可选: en, ja)",
	"help.config_path":   "  配置文件: %s",

	"help.backend_title": "支持的后端:",
	"help.backend_mm":    "  - MyMemory (免费, 默认)",
	"help.backend_libre": "  - LibreTranslate 兼容 API",
	"help.backend_deepl": "  - DeepLX 兼容 API",

	"help.offline_title": "离线功能:",
	"help.offline_cache": "  - 翻译缓存: 将 API 结果保存在本地",
	"help.offline_dict":  "  - 内置词典: 约 400 个常用英→中词汇和短语",
	"help.offline_flag":  "  - 使用 -o 选项强制离线模式",

	// About
	"about.version": "版本: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "创建默认配置",
	"cfg.read":           "读取配置",
	"cfg.parse":          "解析配置",
	"cfg.create_dir":     "创建配置目录",
	"cfg.marshal":        "序列化配置",
	"cfg.write":          "写入配置",

	// API errors
	"api.mm_request":  "MyMemory 请求",
	"api.mm_status":   "MyMemory 返回状态码 %d",
	"api.mm_read":     "MyMemory 读取",
	"api.mm_parse":    "MyMemory 解析",
	"api.mm_empty":    "MyMemory: 文本 %q 的翻译结果为空",
	"api.gen_request": "通用 API 请求",
	"api.gen_status":  "通用 API 返回状态码 %d",
	"api.gen_read":    "通用 API 读取",
	"api.gen_format":  "通用 API: 无法识别的响应格式: %s",
	"api.no_result":   "无可用翻译: API、缓存和词典均无结果",

	// TUI
	"tui.placeholder": "输入要翻译的文本,按 Ctrl+T 翻译...",
	"tui.translating": "翻译中...",
	"tui.empty":       "没有可翻译的文本",
	"tui.failed":      "翻译失败: %v",
	"tui.keys":        "Ctrl+T 翻译 | Tab 切换焦点 | Ctrl+S 互换语言 | Ctrl+O 离线 | Esc 清空 | Ctrl+Q 退出",
	"tui.offline":     "离线",
	"tui.online":      "在线",
	"tui.src_api":     "在线 API",
	"tui.src_cache":   "本地缓存",
	"tui.src_dict":    "内置词典",
	"tui.overview":    "原文",
	"tui.result":      "译文",
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
	"err.translate_failed": "Error: %v",
	"err.empty":            "input text is empty",
	"err.no_translation":   "no translation available",
	"err.offline":          "offline mode: text not found in cache or dictionary",
	"err.network":          "network request failed",

	// Config
	"config.usage_show_set":   "Usage: tr config show|set <key> <value>   (alias: tr -config ...)",
	"config.current":          "Current configuration:",
	"config.source_lang":      "  source_lang = %s",
	"config.target_lang":      "  target_lang = %s",
	"config.api_url":          "  api_url     = %s",
	"config.ui_lang":          "  ui_lang     = %s",
	"config.backend_mymemory": "  backend     = MyMemory (free)",
	"config.backend_custom":   "  backend     = Custom API",
	"config.file":             "  config file = %s",
	"config.usage_set":        "Usage: tr config set <key> <value>",
	"config.valid_keys":       "Valid keys: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":      "Invalid key: %s. Valid keys: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":      "Save failed: %v",
	"config.set_ok":           "Set: %s = %s",
	"config.unknown_subcmd":   "Unknown subcommand: %s. Use: show, set",

	// Help
	"help.title":          "Usage:",
	"help.translate":      "  tr [text]                  Translate text and print result",
	"help.translate_auto": "  tr <text> -> <lang-code>  Auto-detect language (e.g. tr 你好 -> en)",
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

	"help.config_title":  "Configuration:",
	"help.config_source": "  source_lang   Source language code (default: en; auto = detect)",
	"help.config_target": "  target_lang   Target language code (default: zh)",
	"help.config_api":    "  api_url       Custom API endpoint (default: None = MyMemory)",
	"help.config_ui":     "  ui_lang       UI language (default: zh, options: en, ja)",
	"help.config_path":   "  Config file: %s",

	"help.backend_title": "Supported backends:",
	"help.backend_mm":    "  - MyMemory (free, default)",
	"help.backend_libre": "  - LibreTranslate-compatible API",
	"help.backend_deepl": "  - DeepLX-compatible API",

	"help.offline_title": "Offline features:",
	"help.offline_cache": "  - Translation cache: stores previous API results locally",
	"help.offline_dict":  "  - Embedded dictionary: ~400 common EN->ZH words and phrases",
	"help.offline_flag":  "  - Use -o flag to force offline mode",

	// About
	"about.version": "Version: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "create default config",
	"cfg.read":           "read config",
	"cfg.parse":          "parse config",
	"cfg.create_dir":     "create config dir",
	"cfg.marshal":        "marshal config",
	"cfg.write":          "write config",

	// API errors
	"api.mm_request":  "MyMemory request",
	"api.mm_status":   "MyMemory returned status %d",
	"api.mm_read":     "MyMemory read",
	"api.mm_parse":    "MyMemory parse",
	"api.mm_empty":    "MyMemory: empty result for %q",
	"api.gen_request": "generic API request",
	"api.gen_status":  "generic API returned status %d",
	"api.gen_read":    "generic API read",
	"api.gen_format":  "generic API: unrecognized response format: %s",
	"api.no_result":   "no translation available: no result from API, cache, or dictionary",

	// TUI
	"tui.placeholder": "Enter text to translate, then press Ctrl+T...",
	"tui.translating": "Translating...",
	"tui.empty":       "Nothing to translate",
	"tui.failed":      "Translation failed: %v",
	"tui.keys":        "Ctrl+T translate | Tab switch focus | Ctrl+S swap langs | Ctrl+O offline | Esc clear | Ctrl+Q quit",
	"tui.offline":     "offline",
	"tui.online":      "online",
	"tui.src_api":     "online API",
	"tui.src_cache":   "local cache",
	"tui.src_dict":    "embedded dict",
	"tui.overview":    "Source",
	"tui.result":      "Translation",
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
	"err.translate_failed": "エラー: %v",
	"err.empty":            "入力テキストが空です",
	"err.no_translation":   "翻訳が見つかりません",
	"err.offline":          "オフラインモード: キャッシュと辞書にテキストが見つかりません",
	"err.network":          "ネットワークリクエストに失敗しました",

	// Config
	"config.usage_show_set":   "使い方: tr config show|set <キー> <値>   (別名: tr -config ...)",
	"config.current":          "現在の設定:",
	"config.source_lang":      "  source_lang = %s",
	"config.target_lang":      "  target_lang = %s",
	"config.api_url":          "  api_url     = %s",
	"config.ui_lang":          "  ui_lang     = %s",
	"config.backend_mymemory": "  バックエンド = MyMemory (無料)",
	"config.backend_custom":   "  バックエンド = カスタム API",
	"config.file":             "  設定ファイル = %s",
	"config.usage_set":        "使い方: tr config set <キー> <値>",
	"config.valid_keys":       "有効なキー: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":      "無効なキー: %s。有効なキー: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":      "保存に失敗しました: %v",
	"config.set_ok":           "設定しました: %s = %s",
	"config.unknown_subcmd":   "不明なサブコマンド: %s。使用可能: show, set",

	// Help
	"help.title":          "使い方:",
	"help.translate":      "  tr [テキスト]                テキストを翻訳して結果を表示",
	"help.translate_auto": "  tr <テキスト> -> <言語コード>  言語を自動判定して翻訳 (例: tr 你好 -> en)",
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

	"help.config_title":  "設定:",
	"help.config_source": "  source_lang   ソース言語コード (デフォルト: en; auto = 自動判定)",
	"help.config_target": "  target_lang   ターゲット言語コード (デフォルト: zh)",
	"help.config_api":    "  api_url       カスタム API エンドポイント (デフォルト: None = MyMemory)",
	"help.config_ui":     "  ui_lang       UI言語 (デフォルト: zh, オプション: en, ja)",
	"help.config_path":   "  設定ファイル: %s",

	"help.backend_title": "対応バックエンド:",
	"help.backend_mm":    "  - MyMemory (無料, デフォルト)",
	"help.backend_libre": "  - LibreTranslate 互換 API",
	"help.backend_deepl": "  - DeepLX 互換 API",

	"help.offline_title": "オフライン機能:",
	"help.offline_cache": "  - 翻訳キャッシュ: API結果をローカルに保存",
	"help.offline_dict":  "  - 内蔵辞書: 約400の英→中単語・フレーズ",
	"help.offline_flag":  "  - -o オプションでオフラインモードを強制",

	// About
	"about.version": "バージョン: %s",

	// Config errors (internal, wrapped)
	"cfg.create_default": "デフォルト設定の作成",
	"cfg.read":           "設定の読み取り",
	"cfg.parse":          "設定の解析",
	"cfg.create_dir":     "設定ディレクトリの作成",
	"cfg.marshal":        "設定のシリアル化",
	"cfg.write":          "設定の書き込み",

	// API errors
	"api.mm_request":  "MyMemory リクエスト",
	"api.mm_status":   "MyMemory がステータス %d を返しました",
	"api.mm_read":     "MyMemory 読み取り",
	"api.mm_parse":    "MyMemory 解析",
	"api.mm_empty":    "MyMemory: %q の結果が空です",
	"api.gen_request": "汎用 API リクエスト",
	"api.gen_status":  "汎用 API がステータス %d を返しました",
	"api.gen_read":    "汎用 API 読み取り",
	"api.gen_format":  "汎用 API: 認識できないレスポンス形式: %s",
	"api.no_result":   "翻訳が見つかりません: API、キャッシュ、辞書に結果がありません",

	// TUI
	"tui.placeholder": "翻訳するテキストを入力し、Ctrl+T で翻訳...",
	"tui.translating": "翻訳中...",
	"tui.empty":       "翻訳するテキストがありません",
	"tui.failed":      "翻訳に失敗しました: %v",
	"tui.keys":        "Ctrl+T 翻訳 | Tab フォーカス切替 | Ctrl+S 言語入替 | Ctrl+O オフライン | Esc クリア | Ctrl+Q 終了",
	"tui.offline":     "オフライン",
	"tui.online":      "オンライン",
	"tui.src_api":     "オンライン API",
	"tui.src_cache":   "ローカルキャッシュ",
	"tui.src_dict":    "内蔵辞書",
	"tui.overview":    "原文",
	"tui.result":      "訳文",
}
