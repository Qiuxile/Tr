# Tr

> 一个简单优雅的终端翻译工具,使用 Go 编写。**CLI + TUI** 双形态,无桌面依赖。

Tr 是一款轻量级命令行翻译工具。翻译时使用免费的 [MyMemory](https://mymemory.translated.net/) 翻译 API,也支持兼容 LibreTranslate / DeepLX 格式的自定义 API。内置离线翻译功能,无网络时也能使用。

## 特性

- **CLI 模式** — `tr <文本>` 即刻输出翻译结果,便于脚本与管道
- **自动识别语言** — `tr 你好 -> en`:自动判断原文语言并翻译成指定语言
- **TUI 模式** — 无参数直接进入交互式终端界面;支持多行输入、结果滚动、方向互换、离线开关
- **多语言界面** — 中/English/日本語 三种显示语言
- **多语言翻译** — 自由配置源语言与目标语言
- **灵活的 API 后端** — 默认 MyMemory(免费,无需 Key),可接入自建 API
- **离线翻译** — LRU 翻译缓存 + 内置英中词典,断网也能翻译
- **管道输入** — 支持 stdin 管道,可与其他命令组合
- **简洁配置** — `tr config show|set` 管理设置(兼容旧版 `-config` 写法)
- **跨平台** — 纯 Go 标准依赖,Windows / Linux / macOS 均可编译
- **零前端** — 无 WebView、无 Node.js、无 GUI 框架,二进制体积 ~10MB

## 安装

### 从源码编译

```bash
go build -o tr .
```

> 需要 Go 1.26+。依赖仅 bubbletea(交互界面)及其组件。

## CLI 使用

```bash
# 翻译文本(默认:英 → 中)
tr "Hello, world!"

# 自动识别语言并翻译到指定语言(箭头语法,目标支持语言代码与常用名称)
tr 你好 -> en            # 中文 → 英文
tr "Hello world" -> 中文 # 英文 → 中文(名称别名)
tr こんにちは -> zh       # 日文 → 中文
tr "Bonjour" -> ja       # 注意:拉丁字母文本默认判定为英文
tr "bonjour"->zh         # 粘连写法也支持(目标必须是可识别语言)

# 无参数且标准输入是管道时,自动翻译 stdin
echo "hello world" | tr

# 查看/修改配置
tr config show
tr config set source_lang en
tr config set target_lang ja
# 兼容旧版写法: tr -config show / tr -config set target_lang ja

# 使用自定义 API(LibreTranslate / DeepLX 格式)
tr config set api_url https://your-api.example.com/translate

# 离线模式(仅缓存 + 内置词典)
tr -o "hello"
tr --offline "good morning"

# 查看版本 / 关于 / 帮助
tr -version
tr -about
tr -help
```

## TUI 使用

在**交互式终端**里直接运行 `tr`(无参数),或显式 `tr tui`:

```
┌────────────────────────────────────────────┐
│ Tr  en → zh  ●在线                         │
│ ╭ 原文 (en) ────────────────────────────╮  │
│ │ hello world                           │  │
│ ╰───────────────────────────────────────╯  │
│ ╭ 译文 (zh) ────────────────────────────╮  │
│ │ 你好,世界                              │  │
│ │                                    ↑  │  │
│ ╰───────────────────────────────────────╯  │
│ Ctrl+T 翻译 | Tab 切换焦点 | ...           │
└────────────────────────────────────────────┘
```

| 按键 | 功能 |
|------|------|
| `Ctrl+T` | 翻译输入框中的文本 |
| `Tab` | 在输入框 / 结果区之间切换焦点(结果区支持方向键、PgUp/PgDn 滚动) |
| `Ctrl+S` | 互换源语言与目标语言(会写入配置) |
| `Ctrl+O` | 临时切换 在线 / 离线 模式 |
| `Esc` | 清空输入框 |
| `Ctrl+Q` / `Ctrl+C` | 退出 |

TUI 中翻译同样走「在线 API → 缓存 → 内置词典」链路,结果会标注来源。

## 离线翻译

Tr 支持离线翻译,通过两层机制实现:

1. **翻译缓存**:在线翻译结果自动缓存到本地(`cache.json`,最多 500 条,LRU 淘汰),再次翻译相同内容时即使离线也能命中。
2. **内置词典**:约 400 个常用英文单词/短语的英→中对照表,支持精确短语匹配与逐词翻译。

使用 `-o` / `--offline`(或 TUI 内 `Ctrl+O`)可强制离线。

```
在线 API → 本地缓存 → 内置词典 → 报错
```

## 配置

配置文件位置(依 OS 惯例):

- **Windows**:`%APPDATA%\Tr\config.json`
- **Linux**:`~/.config/Tr/config.json`
- **macOS**:`~/Library/Application Support/Tr/config.json`

| 键 | 说明 | 默认值 |
|---|------|--------|
| `source_lang` | 源语言代码(`auto` = 按文本自动检测) | `en` |
| `target_lang` | 目标语言代码 | `zh` |
| `api_url` | 自定义 API 地址(`None` = MyMemory) | `None` |
| `ui_lang` | 界面语言(`zh` / `en` / `ja`) | `zh` |

### 语言自动检测

`-> <语言>` 会把源语言设为 `auto`,按文本书写系统检测:中文(CJK 汉字)、日文(含假名)、韩文(谚文)、俄文(西里尔)、阿拉伯文、希腊文、泰文、印地文等;纯拉丁字母文本一律按英语处理(如需法/德/西等源语言,请 `tr config set source_lang fr` 后使用普通语法)。

目标语言部分接受 ISO 639-1 代码(如 `en`/`zh`/`ja`/`ko`…)以及常见名称(如 `中文`、`英文`、`日本語`、`korean`、`russian`…),也可设置 `source_lang = auto` 让普通翻译也自动识别原文。

## API 兼容性

1. **MyMemory**(默认)— 免费翻译 API,无需 Key
2. **LibreTranslate / DeepLX 兼容 API** — `api_url` 指向自建或第三方地址即可

## 项目结构

```
main.go            入口
internal/tr/
  app.go           CLI 分发与参数解析(含 -> 箭头语法)
  lang.go          语言自动检测与语言代码/名称归一化
  tui.go           bubbletea 交互界面
  translate.go     翻译链路(API → 缓存 → 词典)
  mymemory.go      MyMemory 后端
  generic.go       LibreTranslate / DeepLX 后端
  cache.go         LRU 缓存
  config.go        配置读写
  dict*.go         内置英中词典
  i18n.go          中/英/日 文案
```

## 历史

v2.0.0 起,项目由 Wails 桌面版回归终端形态:移除 Web 前端与桌面打包,新增 TUI。旧桌面代码见 Git 历史(commit `a04d846` 之前)。

## 许可证

本项目基于 [MIT License](LICENSE) 开源。

Copyright (c) 2026 Surile
