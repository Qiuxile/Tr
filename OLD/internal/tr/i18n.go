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
	"warn.defaults":         "警告: %v; 使用默认配置",
	"warn.cache_unavailable": "警告: 缓存不可用: %v",
	"warn.unknown_flag":     "警告: 未知选项 %s",

	// Errors
	"err.empty_input":     "错误: 没有输入文本。请通过管道传入或作为参数传递。",
	"err.usage_hint":      "用法: tr [文本]",
	"err.translate_failed": "错误: %v",
	"err.empty":           "输入文本不能为空",
	"err.no_translation":  "无可用翻译",
	"err.offline":         "离线模式: 缓存和词典中均未找到该文本",
	"err.network":         "网络请求失败",

	// Config
	"config.usage_show_set":  "用法: tr -config show|set <键> <值>",
	"config.current":         "当前配置:",
	"config.source_lang":     "  source_lang = %s",
	"config.target_lang":     "  target_lang = %s",
	"config.api_url":         "  api_url     = %s",
	"config.ui_lang":         "  ui_lang     = %s",
	"config.backend_mymemory": "  后端         = MyMemory (免费)",
	"config.backend_custom":   "  后端         = 自定义 API",
	"config.file":            "  配置文件     = %s",
	"config.usage_set":       "用法: tr -config set <键> <值>",
	"config.valid_keys":      "可用键: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":     "无效的配置项: %s。可用键: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":     "保存失败: %v",
	"config.set_ok":          "设置成功: %s = %s",
	"config.unknown_subcmd":  "未知子命令: %s。可用: show, set",

	// Help
	"help.title":       "用法:",
	"help.translate":   "  tr [文本]                       翻译文本并输出结果",
	"help.config_show": "  tr -config show                 显示当前配置",
	"help.config_set":  "  tr -config set <键> <值>        修改配置项",
	"help.help":        "  tr -help                        显示此帮助信息",
	"help.version":     "  tr -version                     显示版本",
	"help.about":       "  tr -about                       显示关于信息",
	"help.offline":     "  tr -o, --offline                强制离线模式 (仅缓存+词典)",

	"help.flags_title":     "选项:",
	"help.flags_config":    "  -config, --config, -c   配置管理",
	"help.flags_help":      "  -help,   --help,   -h   显示帮助",
	"help.flags_version":   "  -version,--version, -v  显示版本",
	"help.flags_about":     "  -about,  --about,   -a  显示关于",
	"help.flags_offline":   "  -offline,--offline, -o  强制离线模式",

	"help.pipe_title": "管道输入:",
	"help.pipe_example": "  echo \"hello\" | tr         从标准输入翻译",

	"help.config_title": "配置:",
	"help.config_source": "  source_lang   源语言代码 (默认: en)",
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
	"help.gui":       "  tr -gui, -g                    启动图形界面",
	"help.flags_gui": "  -gui, --gui, -g        启动图形界面",

	"api.mm_request":    "MyMemory 请求",
	"api.mm_status":     "MyMemory 返回状态码 %d",
	"api.mm_read":       "MyMemory 读取",
	"api.mm_parse":      "MyMemory 解析",
	"api.mm_empty":      "MyMemory: 文本 %q 的翻译结果为空",
	"api.gen_request":   "通用 API 请求",
	"api.gen_status":    "通用 API 返回状态码 %d",
	"api.gen_read":      "通用 API 读取",
	"api.gen_format":    "通用 API: 无法识别的响应格式: %s",
	"api.no_result":     "无可用翻译: API、缓存和词典均无结果",

	// GUI
	"gui.translate":         "翻译",
	"gui.input_placeholder": "输入要翻译的文本...",
	"gui.source_lang":       "源语言",
	"gui.target_lang":       "目标语言",
	"gui.offline_mode":      "离线模式",
	"gui.translate_btn":     "翻译",
	"gui.result":            "翻译结果",
	"gui.result_placeholder": "输入文本后点击翻译",
	"gui.settings":          "设置",
	"gui.ui_lang":           "界面语言",
	"gui.save":              "保存",
	"gui.free":              "免费",
	"gui.custom_api":        "自定义 API",
	"gui.ready":             "就绪",
	"gui.empty_input":       "请输入要翻译的文本",
	"gui.translating":       "翻译中...",
	"gui.cache":             "缓存",
	"gui.dict":              "词典",
	"gui.network_error":     "网络错误，请检查网络连接",
	"gui.saved":             "已保存",
	"gui.save_failed":       "保存失败",
	"gui.offline":           "离线模式",
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

	// Errors
	"err.empty_input":      "Error: no input text. Pipe text or pass as argument.",
	"err.usage_hint":       "Usage: tr [text]",
	"err.translate_failed": "Error: %v",
	"err.empty":            "input text is empty",
	"err.no_translation":   "no translation available",
	"err.offline":          "offline mode: text not found in cache or dictionary",
	"err.network":          "network request failed",

	// Config
	"config.usage_show_set":   "Usage: tr -config show|set <key> <value>",
	"config.current":          "Current configuration:",
	"config.source_lang":      "  source_lang = %s",
	"config.target_lang":      "  target_lang = %s",
	"config.api_url":          "  api_url     = %s",
	"config.ui_lang":          "  ui_lang     = %s",
	"config.backend_mymemory": "  backend     = MyMemory (free)",
	"config.backend_custom":   "  backend     = Custom API",
	"config.file":             "  config file = %s",
	"config.usage_set":        "Usage: tr -config set <key> <value>",
	"config.valid_keys":       "Valid keys: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":      "Invalid key: %s. Valid keys: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":      "Save failed: %v",
	"config.set_ok":           "Set: %s = %s",
	"config.unknown_subcmd":   "Unknown subcommand: %s. Use: show, set",

	// Help
	"help.title":       "Usage:",
	"help.translate":   "  tr [text]                      Translate text and print result",
	"help.config_show": "  tr -config show                Show current configuration",
	"help.config_set":  "  tr -config set <key> <value>   Set a configuration value",
	"help.help":        "  tr -help                       Show this help message",
	"help.version":     "  tr -version                    Show version",
	"help.about":       "  tr -about                      Show about information",
	"help.offline":     "  tr -o, --offline               Force offline mode (cache + dictionary only)",

	"help.flags_title":   "Flags:",
	"help.flags_config":  "  -config, --config, -c   Configuration management",
	"help.flags_help":    "  -help,   --help,   -h   Show help",
	"help.flags_version": "  -version,--version, -v  Show version",
	"help.flags_about":   "  -about,  --about,   -a  Show about",
	"help.flags_offline": "  -offline,--offline, -o  Force offline mode",

	"help.pipe_title":   "Piped input:",
	"help.pipe_example": "  echo \"hello\" | tr        Translate from stdin",

	"help.config_title":  "Configuration:",
	"help.config_source": "  source_lang   Source language code (default: en)",
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
	"help.gui":       "  tr -gui, -g                    Launch graphical interface",
	"help.flags_gui": "  -gui, --gui, -g        Launch graphical interface",

	"api.mm_request":   "MyMemory request",
	"api.mm_status":    "MyMemory returned status %d",
	"api.mm_read":      "MyMemory read",
	"api.mm_parse":     "MyMemory parse",
	"api.mm_empty":     "MyMemory: empty result for %q",
	"api.gen_request":  "generic API request",
	"api.gen_status":   "generic API returned status %d",
	"api.gen_read":     "generic API read",
	"api.gen_format":   "generic API: unrecognized response format: %s",
	"api.no_result":    "no translation available: no result from API, cache, or dictionary",

	// GUI
	"gui.translate":          "Translate",
	"gui.input_placeholder":  "Enter text to translate...",
	"gui.source_lang":        "Source Language",
	"gui.target_lang":        "Target Language",
	"gui.offline_mode":       "Offline Mode",
	"gui.translate_btn":      "Translate",
	"gui.result":             "Translation Result",
	"gui.result_placeholder": "Enter text and click Translate",
	"gui.settings":           "Settings",
	"gui.ui_lang":            "UI Language",
	"gui.save":               "Save",
	"gui.free":               "Free",
	"gui.custom_api":         "Custom API",
	"gui.ready":              "Ready",
	"gui.empty_input":        "Please enter text to translate",
	"gui.translating":        "Translating...",
	"gui.cache":              "Cache",
	"gui.dict":               "Dictionary",
	"gui.network_error":      "Network error, please check connection",
	"gui.saved":              "Saved",
	"gui.save_failed":        "Save failed",
	"gui.offline":            "Offline",
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

	// Errors
	"err.empty_input":      "エラー: 入力テキストがありません。パイプまたは引数で渡してください。",
	"err.usage_hint":       "使い方: tr [テキスト]",
	"err.translate_failed": "エラー: %v",
	"err.empty":            "入力テキストが空です",
	"err.no_translation":   "翻訳が見つかりません",
	"err.offline":          "オフラインモード: キャッシュと辞書にテキストが見つかりません",
	"err.network":          "ネットワークリクエストに失敗しました",

	// Config
	"config.usage_show_set":   "使い方: tr -config show|set <キー> <値>",
	"config.current":          "現在の設定:",
	"config.source_lang":      "  source_lang = %s",
	"config.target_lang":      "  target_lang = %s",
	"config.api_url":          "  api_url     = %s",
	"config.ui_lang":          "  ui_lang     = %s",
	"config.backend_mymemory": "  バックエンド = MyMemory (無料)",
	"config.backend_custom":   "  バックエンド = カスタム API",
	"config.file":             "  設定ファイル = %s",
	"config.usage_set":        "使い方: tr -config set <キー> <値>",
	"config.valid_keys":       "有効なキー: source_lang, target_lang, api_url, ui_lang",
	"config.invalid_key":      "無効なキー: %s。有効なキー: source_lang, target_lang, api_url, ui_lang",
	"config.save_failed":      "保存に失敗しました: %v",
	"config.set_ok":           "設定しました: %s = %s",
	"config.unknown_subcmd":   "不明なサブコマンド: %s。使用可能: show, set",

	// Help
	"help.title":       "使い方:",
	"help.translate":   "  tr [テキスト]                    テキストを翻訳して結果を表示",
	"help.config_show": "  tr -config show                 現在の設定を表示",
	"help.config_set":  "  tr -config set <キー> <値>      設定を変更",
	"help.help":        "  tr -help                        このヘルプを表示",
	"help.version":     "  tr -version                     バージョンを表示",
	"help.about":       "  tr -about                       アプリ情報を表示",
	"help.offline":     "  tr -o, --offline                オフラインモードを強制 (キャッシュ+辞書のみ)",

	"help.flags_title":   "オプション:",
	"help.flags_config":  "  -config, --config, -c   設定管理",
	"help.flags_help":    "  -help,   --help,   -h   ヘルプを表示",
	"help.flags_version": "  -version,--version, -v  バージョンを表示",
	"help.flags_about":   "  -about,  --about,   -a  アプリ情報を表示",
	"help.flags_offline": "  -offline,--offline, -o  オフラインモードを強制",

	"help.pipe_title":   "パイプ入力:",
	"help.pipe_example": "  echo \"hello\" | tr        標準入力から翻訳",

	"help.config_title":  "設定:",
	"help.config_source": "  source_lang   ソース言語コード (デフォルト: en)",
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
	"help.gui":       "  tr -gui, -g                    GUIを起動",
	"help.flags_gui": "  -gui, --gui, -g        GUIを起動",

	"api.mm_request":   "MyMemory リクエスト",
	"api.mm_status":    "MyMemory がステータス %d を返しました",
	"api.mm_read":      "MyMemory 読み取り",
	"api.mm_parse":     "MyMemory 解析",
	"api.mm_empty":     "MyMemory: %q の結果が空です",
	"api.gen_request":  "汎用 API リクエスト",
	"api.gen_status":   "汎用 API がステータス %d を返しました",
	"api.gen_read":     "汎用 API 読み取り",
	"api.gen_format":   "汎用 API: 認識できないレスポンス形式: %s",
	"api.no_result":    "翻訳が見つかりません: API、キャッシュ、辞書に結果がありません",

	// GUI
	"gui.translate":          "翻訳",
	"gui.input_placeholder":  "翻訳するテキストを入力...",
	"gui.source_lang":        "ソース言語",
	"gui.target_lang":        "ターゲット言語",
	"gui.offline_mode":       "オフラインモード",
	"gui.translate_btn":      "翻訳",
	"gui.result":             "翻訳結果",
	"gui.result_placeholder": "テキストを入力して翻訳をクリック",
	"gui.settings":           "設定",
	"gui.ui_lang":            "UI言語",
	"gui.save":               "保存",
	"gui.free":               "無料",
	"gui.custom_api":         "カスタムAPI",
	"gui.ready":              "準備完了",
	"gui.empty_input":        "翻訳するテキストを入力してください",
	"gui.translating":        "翻訳中...",
	"gui.cache":              "キャッシュ",
	"gui.dict":               "辞書",
	"gui.network_error":      "ネットワークエラー、接続を確認してください",
	"gui.saved":              "保存しました",
	"gui.save_failed":        "保存に失敗しました",
	"gui.offline":            "オフライン",
}
