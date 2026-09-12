# Tr

> 一个简单优雅的终端翻译工具,使用 Go 编写。**CLI + TUI** 双形态,无桌面依赖。

Tr 是一款轻量级命令行翻译工具,**只通过 AI 翻译**(OpenAI 兼容接口,内置 DeepSeek 配置),填一个 `sk-...` 密钥即可使用。翻译结果会本地缓存,重复内容不重复计费;断网或没配密钥时仍可用缓存与内置词典离线翻译。

```
$ tr 你好 -> en
Hello.

$ tr
╭ 原文 (en) ─────────╮ ╭ 译文 (zh) ── · 在线 AI ╮   ← 无参数进入 TUI
```

## 目录

- [快速开始](#快速开始)
- [特性](#特性)
- [CLI 使用](#cli-使用) · [命令速查](#命令速查)
- [AI 翻译](#ai-翻译) · [省 Token 策略](#省-token-策略)
- [TUI 使用](#tui-使用) · [系统命令行](#系统命令行esc)
- [离线翻译](#离线翻译)
- [配置](#配置) · [目标语言简写](#目标语言简写)
- [常见问题](#常见问题)
- [项目结构](#项目结构) · [历史](#历史)

## 快速开始

```bash
# 1) 编译(或使用 Releases 里的安装包)
go build -o tr .

# 2) 配置 AI 密钥(必填,默认走 DeepSeek)
tr config set api_key sk-xxxxxxxxxxxxxxxx

# 3) 开始翻译
tr "Hello, world!"        # 英 → 中
tr 你好 -> en              # 自动识别语言,译成英文
tr                        # 无参数:进入交互式 TUI
```

## 特性

- **AI 翻译** — 翻译只走 AI 后端,支持任意 OpenAI 兼容服务(DeepSeek / OpenAI / GLM / 自建中转)
- **省 token 设计** — 缓存优先、单轮极简提示词、关闭思考模式、限制输出长度,详见[省 Token 策略](#省-token-策略)
- **CLI 模式** — `tr <文本>` 即刻输出翻译结果,便于脚本与管道
- **文件翻译** — `tr -f <文件>` 整篇翻译:可打印到终端,或写入指定文件/目录
- **自动识别语言** — `tr 你好 -> en`:自动判断原文语言并翻译成指定语言
- **TUI 模式** — 左右布局的交互界面(原文/译文并排),支持多行输入、译文滚动、语言切换
- **内建系统命令行** — TUI 里按 `Esc` 直接开一个真终端:执行 `dir`/`ls`/`git` 等命令,输出整屏显示
- **多语言界面** — 中/English/日本語 三种显示语言
- **离线可用** — LRU 翻译缓存 + 内置英中词典,离线也能翻译(不消耗 token)
- **管道输入** — 支持 stdin 管道,可与其他命令组合
- **简洁配置** — `tr config show|set` 管理设置(兼容旧版 `-config` 写法)
- **跨平台** — 纯 Go 标准依赖,Windows / Linux / macOS 均可编译
- **零前端** — 无 WebView、无 Node.js、无 GUI 框架

## 安装

### 从源码编译

```bash
go build -o tr .
```

> 需要 Go 1.26+。运行时依赖仅 bubbletea(交互界面)及其组件。

### Windows 安装包

仓库内的 `setup.iss` 是 Inno Setup 脚本,可把编译产物打包成安装程序(自动把 `tr` 加入 `PATH`),产物输出到 `Output/` 目录。

### 放进 PATH

自定义编译时,把可执行文件所在目录加入系统 `PATH`,即可在任意位置直接使用 `tr`。

## CLI 使用

```bash
# 首次使用:配置 AI 密钥(必需)
tr config set api_key sk-xxxxxxxxxxxxxxxx
tr config set ai_model deepseek-flash                  # 换模型
tr config set ai_base_url https://api.openai.com/v1    # 换服务商

# 翻译文本(默认:英 → 中)
tr "Hello, world!"

# 自动识别语言并翻译到指定语言(箭头语法只作用于本次执行,不写入配置)
# 箭头模式只取「第一个参数」作为内容,其余参数一律忽略;内容含空格请加引号
tr 你好 -> en            # 中文 → 英文
tr "Hello world" -> 中文 # 英文 → 中文(名称别名;内容是一个参数)
tr こんにちは -> zh       # 日文 → 中文
tr "Bonjour" -> ja       # 注意:拉丁字母文本默认判定为英文
tr "bonjour"->zh         # 粘连写法也支持(目标必须是可识别语言)
tr hello -> nl           # 任意 ISO 639-1 代码均可(此处为荷兰语)
tr hello world -> zh     # 只翻译 hello,后面的 world 被忽略
tr "good morning" -> zh extra   # 箭头之后的额外参数同样被忽略
tr -> ja                 # 内容来自 stdin(管道)

# 翻译文件:省略输出路径则打印到终端,给出路径则写入该文件或目录
tr -f README.md
tr -f docs/en.md docs/zh.md
tr -f docs/en.md docs/zh/          # 目录(不存在会自动创建),保留原文件名
tr -f docs/en.md -> ja             # 文件同样支持箭头语法

# 无参数且标准输入是管道时,自动翻译 stdin
echo "hello world" | tr

# 查看/修改配置
tr config show
tr config set source_lang en
tr config set target_lang ja
# 兼容旧版写法: tr -config show / tr -config set target_lang ja

# 离线模式(仅缓存 + 内置词典,不联网、不消耗 token)
tr -o "hello"
tr --offline "good morning"
tr -o -f docs/en.md

# 查看版本 / 关于 / 帮助
tr -version
tr -about
tr -help
```

### 命令速查

| 命令 | 说明 |
|---|---|
| `tr <文本>` | 翻译文本并输出到 stdout |
| `tr <文本> -> <语言>` | 自动识别源语言,译成指定语言(仅本次执行)。**只取第一个参数作为内容**,其余忽略 |
| `tr` | 无参数且是交互终端 → 进入 TUI;是管道 → 翻译 stdin |
| `tr tui` | 显式启动 TUI |
| `tr -f <文件> [输出]` | 翻译文件;省略输出打印到终端,给出路径则写入文件或目录 |
| `tr -o` / `--offline` | 强制离线(缓存 + 内置词典,不联网) |
| `tr config show` | 显示当前配置(密钥做掩码) |
| `tr config set <键> <值>` | 修改配置(兼容 `tr -config set ...`) |
| `tr <文本> -> <语言>` 中的 `<语言>` | ISO 639-1 代码或常见名称,见[目标语言简写](#目标语言简写) |
| `tr -help` / `-version` / `-about` | 帮助 / 版本 / 关于 |

退出码:`0` 成功;`1` 运行时错误(翻译失败、文件读写失败);`2` 参数或用法错误。

## AI 翻译

翻译只有 AI 一个后端——**必须先配置密钥**:

```bash
tr config set api_key sk-xxxxxxxxxxxxxxxx
tr config set ai_model deepseek-flash
```

默认走 DeepSeek(`https://api.deepseek.com/v1`),任意 OpenAI 兼容服务都可以用 `ai_base_url` + `ai_model` 指过去(OpenAI、GLM、Qwen、自建中转等)。

密钥也可以放在环境变量里,不必写进配置文件(优先级:`TR_API_KEY` → `DEEPSEEK_API_KEY` → `OPENAI_API_KEY`):

```bash
TR_API_KEY=sk-xxx tr "hello world"
```

密钥只存在本地配置文件,`tr config show` 与 `config set` 的回显都会做掩码(`sk-tes…abcd`)。

未配置密钥时执行翻译会直接报错并给出指引(而不是悄悄换别的引擎):想离线使用请加 `-o`。

### 翻译链路

```
缓存命中 → AI → 报错          (联网)
缓存命中 → 内置词典 → 报错     (加 -o 离线)
```

### 文件翻译

`tr -f <文件> [输出]` 会把整篇文件交给 AI 翻译,并按需分块:

- **输出为空** → 译文打印到终端
- **输出是路径** → 写入该文件;若是已存在目录或路径以 `/`、`\` 结尾,则按目录处理并沿用源文件名(目录会自动创建)
- 分块大小随后端自适应(AI 1500 字符),块内行结构保持不变;若模型返回的行数对不上(或直接回显原文),自动降级为逐行翻译
- 空行、CRLF 换行、UTF-8 BOM 都会原样保留;某块彻底失败时保留该块原文并在 stderr 提示
- 每个块同样走缓存,重复翻译同一文件不重复计费

### 省 Token 策略

翻译是"输入多少、输出多少"的确定性任务,不值得花推理预算。Tr 为此做了几件事:

| 措施 | 说明 |
|---|---|
| **缓存优先** | 缓存命中直接返回,不发请求,**零 token**;结果落盘复用 |
| **关闭思考模式** | `ai_thinking = false`(默认)会显式告知后端禁用推理,避免为一句翻译消耗成百上千个思维链 token |
| **极简单轮提示词** | 单条 system 提示("Translate from X into Y. Reply with the translation only."),无示例、无风格说明、无多轮上下文 |
| **限制输出长度** | 按输入字符数估算 `max_tokens`(64–4096),模型无法借机长篇解释 |
| **不自动重试** | 失败直接报错,不会因网络抖动重复计费 |
| **结果后处理** | 自动剥掉模型加的引号/`Translation:` 前缀,缓存里存干净的译文 |
| **离线零成本** | `-o` 或 TUI 内 `Ctrl+O` 时完全不联网 |
| **长文本分块** | 原文超过约 1500 字符自动分块翻译,避免超出上下文上限导致的失败与截断 |

如果你确实需要模型思考,`tr config set ai_thinking true` 可重新开启(会明显增加用量)。

## TUI 使用

在**交互式终端**里直接运行 `tr`(无参数),或显式 `tr tui`。界面为**左右布局**:左侧原文,右侧译文:

```
Tr  en → zh  ●在线
╭ 原文 (en) ─────────╮ ╭ 译文 (zh) ── · 在线 AI ╮
│ hello world        │ │ 你好,世界              │
│                    │ │                        │
╰────────────────────╯ ╰────────────────────────╯
Ctrl+T 翻译 | Ctrl+←/→ 切语言 | Ctrl+L 清屏 | Esc 系统命令行 | Ctrl+Q 退出
```

| 按键 | 功能 |
|------|------|
| `Ctrl+T` | 翻译输入框中的文本 |
| `Enter` | 输入框内换行(多行文本) |
| `Ctrl+←` / `Ctrl+→` | 切换**当前焦点那一侧**的语言:焦点在输入框=源语言,焦点在译文区=目标语言(会话内生效,不写配置) |
| `Tab` / `Shift+Tab` | 焦点在 输入框 ↔ 译文区 之间切换 |
| `↑` / `↓` / `PgUp` / `PgDn`(译文区聚焦时) | 滚动译文 |
| `Esc` | 打开系统命令行(见下) |
| `Ctrl+L` | **清空终端输出**:翻译界面清空译文区,系统命令行里清空整屏输出(同 shell 的 `Ctrl+L`) |
| `Ctrl+U` | 清空输入框 |
| `Ctrl+S` | 互换源语言与目标语言(仅当前会话,不写配置文件) |
| `Ctrl+O` | 临时切换 在线 / 离线 模式 |
| `Ctrl+Q` / `Ctrl+C` | 退出 |

> 语言切换跟随焦点:焦点在**输入框**(原文侧)时 `Ctrl+←/→` 改源语言,焦点在译文区时改目标语言,切换只作用于当前会话。

### 系统命令行(Esc)

按 `Esc` 进入**真正的系统命令行**:翻译面板隐藏,整屏用于显示命令输出,底部是一行命令输入(`>` 提示符)。这里执行的就是系统命令本身,没有自定义命令语法:

```
系统命令行 (Esc 返回翻译界面)
> git status
On branch master
nothing to commit, working tree clean

> dir / ls / tr config show ...
```

| 操作 | 功能 |
|------|------|
| 输入 + `Enter` | 在当前目录执行该命令(stdout/stderr 一并显示) |
| `↑` / `↓` | 翻阅历史命令 |
| `PgUp` / `PgDn` / `Home` / `End` | 滚动输出(保留最近 1000 行) |
| `Ctrl+L` | 清空整屏输出(同 shell 的 `Ctrl+L`),命令历史保留 |
| `Esc` | 输入非空时清空;输入为空时返回翻译界面 |
| `Ctrl+C` | 退出 tr |

命令通过系统 shell 执行(`cmd /c` / `sh -c`),30 秒超时保护,ANSI 转义与 Windows `\r` 会被清理以免破坏界面;Windows 下会先切到 UTF-8 代码页,中文输出不会乱码。想改配置直接在里面跑 `tr config set target_lang ja` 即可(等价于退出后用 CLI 执行)。

TUI 中翻译走与 CLI 相同的链路,结果会标注来源:**在线 AI** / 本地缓存 / 内置词典。

## 离线翻译

Tr 支持离线翻译,通过两层机制实现:

1. **翻译缓存**:在线翻译结果自动缓存到本地(`cache.json`,最多 500 条,LRU 淘汰),再次翻译相同内容时即使离线也能命中(也意味着不重复计费)。
2. **内置词典**:约 400 个常用英文单词/短语的英→中对照表,支持精确短语匹配与逐词翻译。

使用 `-o` / `--offline`(或 TUI 内 `Ctrl+O`)可强制离线:完全不联网,只查缓存与词典,不消耗任何 token。

```
缓存 → AI → 报错            (联网)
缓存 → 内置词典 → 报错      (离线 -o)
```

## 配置

配置文件位置(依 OS 惯例):

- **Windows**:`%APPDATA%\Tr\config.json`
- **Linux**:`~/.config/Tr/config.json`
- **macOS**:`~/Library/Application Support/Tr/config.json`

| 键 | 说明 | 默认值 |
|---|------|--------|
| `api_key` | AI 密钥(`sk-...`),翻译必需 | 空(未配置) |
| `ai_base_url` | OpenAI 兼容接口地址 | `https://api.deepseek.com/v1` |
| `ai_model` | 模型名 | `deepseek-flash` |
| `ai_thinking` | 是否启用模型思考模式(开启更费 token) | `false` |
| `source_lang` | 源语言代码(`auto` = 按文本自动检测) | `en` |
| `target_lang` | 目标语言代码 | `zh` |
| `ui_lang` | 界面语言(`zh` / `en` / `ja`) | `zh` |

### 箭头语法(`->`)与语言自动检测

`-> <语言>` 会把源语言设为 `auto`,按文本书写系统检测:中文(CJK 汉字)、日文(含假名)、韩文(谚文)、俄文(西里尔)、阿拉伯文、希腊文、泰文、印地文等;纯拉丁字母文本一律按英语处理(如需法/德/西等源语言,请 `tr config set source_lang fr` 后使用普通语法)。也可设置 `source_lang = auto`,让不带箭头的普通翻译同样自动识别原文。

**`->` 只作用于本次执行**,不会写入配置文件;要持久修改目标语言请用 `tr config set target_lang <代码>`。(TUI 里 `Ctrl+S` 互换方向同样只在当前会话生效。)

**参数规则**:箭头模式是位置式的——内容 = `tr` 后的**第一个参数**,语言 = 箭头后的那一个 token,其余参数全部忽略。所以内容含空格时要加引号;`tr hello world -> zh` 只会翻译 `hello`。不带箭头时,所有参数照旧拼接成一句话翻译。

### 超长文本

单次请求的原文超过约 1500 字符时,tr 会**自动分块翻译再拼接**,不会把超长文本一次性丢给模型——这正是"maximum context length"这类报错和译文被截断的根源。分块点是段落/换行/句号等自然边界,CLI 下会在 stderr 显示 `长文本分块翻译中 3/5 …` 的进度。

整篇文件请优先用 `tr -f <文件>`:它按行对齐分块,能完整保留原文结构与空行。

### 目标语言简写

| 代码 | 语言 | 代码 | 语言 | 代码 | 语言 | 代码 | 语言 |
|---|---|---|---|---|---|---|---|
| `zh` | 中文(简体) | `en` | 英语 | `ja` | 日语 | `ko` | 韩语 |
| `ru` | 俄语 | `fr` | 法语 | `de` | 德语 | `es` | 西班牙语 |
| `it` | 意大利语 | `pt` | 葡萄牙语 | `ar` | 阿拉伯语 | `th` | 泰语 |
| `vi` | 越南语 | `hi` | 印地语 | `el` | 希腊语 | `nl` | 荷兰语 |
| `pl` | 波兰语 | `tr` | 土耳其语 | `uk` | 乌克兰语 | `id` | 印尼语 |

除代码外,也接受常见名称与别名:`中文`、`英文`、`日本語`、`한국어`、`korean`、`russian`、`Deutsch`、`español`…;此外任意 ISO 639-1 代码都能直接使用(例如 `tr "hello" -> nl`),模型只要认识该语言即可翻译。

同样的列表也内建在 `tr -help` 里,忘了代码随时可查。

## 翻译后端

只有 AI 一个后端:

| 后端 | 何时使用 | 说明 |
|---|---|---|
| **AI** | 始终(需配置 `api_key`) | OpenAI 兼容 `POST /v1/chat/completions`;默认 DeepSeek `deepseek-flash`,可换任意服务商 |
| **离线**(缓存 + 内置词典) | 加 `-o` | 不联网时的降级路径,不是独立翻译引擎 |

AI 后端只依赖标准的 Chat Completions 接口,`ai_base_url` 既可以填 base(`.../v1`),也可以直接填完整端点(`.../v1/chat/completions`)。

### 目标语言的传递方式

翻译函数把目标语言作为**显式参数**接收:

| 调用方式 | 传入参数 | 效果 |
|---|---|---|
| `tr "hello"` | 空 | 读取配置里的 `target_lang` |
| `tr "hello" -> ja` | `ja` | 只用 `ja`,**不读**也不写配置中的目标语言 |

所以箭头语法不会因为一次临时调整而残留副作用,`tr config show` 始终显示你配置的那份设置;想改默认目标语言就用 `tr config set target_lang <代码>`。

## 常见问题

**提示"未配置 AI 密钥"** — 翻译必须要有密钥:

```bash
tr config set api_key sk-xxxxxxxx
# 或临时用环境变量,不写进配置文件
TR_API_KEY=sk-xxxxxxxx tr "hello"
```

只想用缓存与内置词典离线工作,加 `-o` 即可(不需要密钥)。

**`AI 密钥无效 (HTTP 401)`** — 密钥写错或已失效;还要确认密钥与 `ai_base_url` 属于同一家服务商(把 DeepSeek 的 key 配到 OpenAI 地址必然 401)。

**`AI 账户余额不足 (HTTP 402)`** — 到服务商控制台充值。

**`AI 请求被限流 (HTTP 429)`** — 请求太密集,稍后重试。文件翻译与长文本是分块请求,大文件更容易触发限流。

**报 `maximum context length` / 译文被截断** — 早期版本会把整段长文本一次性发给模型。现在原文超过约 1500 字符会自动分块翻译,正常情况下不会再出现;若仍看到"输出被长度上限截断"的提示,说明单块译文异常膨胀(例如要求模型解释),可改用 `tr -f <文件>` 逐块处理。

**翻译结果和原文一样 / 内容没变** — 很可能命中了本地缓存(`cache.json`)。缓存是刻意的省 token 设计;想强制重译可以清空缓存文件,或改动原文。

**离线模式报"缓存和词典中均未找到"** — `-o` 只查缓存与内置词典(英→中,约 400 词),没翻译过的句子必然未命中。先联网翻一次即可入库,之后离线可用。

**TUI 里某些快捷键没反应** — 见 [TUI 使用](#tui-使用) 的说明:终端无法区分 `Enter` 与 `Ctrl+Enter`,所以翻译键是 `Ctrl+T`;某些终端会占用 `Alt+组合键`,因此本项目不使用 Alt。

**Windows 下命令输出乱码** — tr 执行命令前会先 `chcp 65001`(UTF-8)。若个别程序仍输出乱码,那通常是该程序自行使用 GBK,可在它自己的参数里指定编码。

**配置文件、缓存在哪,怎么清**

- Windows:`%APPDATA%\Tr\`;Linux:`~/.config/Tr/`;macOS:`~/Library/Application Support/Tr/`
- `config.json` = 配置;`cache.json` = 翻译缓存(最多 500 条,LRU 淘汰,直接删除即清空)

**能不能当库用 / 有没有 HTTP API** — 目前只提供 CLI 与 TUI;脚本里直接调用 CLI 即可,管道输入输出都对脚本友好。

## 项目结构

```
main.go            入口
internal/tr/
  app.go           CLI 分发与参数解析(-> 箭头语法、-f 文件模式、config 子命令、帮助输出)
  lang.go          语言自动检测 + 语言代码/名称归一化 + TUI 语言选择列表
  ai.go            AI 后端(OpenAI 兼容)+ 省 token 策略
  file.go          文件翻译(分块、行对齐、输出到终端/文件/目录)
  tui.go           bubbletea 交互界面(左右布局、Ctrl 快捷键、内建系统命令行)
  tui_test.go      布局 / 焦点 / 快捷键 / 清屏 / 不写配置的回归测试
  translate.go     翻译链路(缓存 → AI;离线时缓存 → 词典)+ target 参数语义
  cache.go         LRU 缓存
  config.go        配置读写 + 环境变量密钥
  dict*.go         内置英中词典
  i18n.go          中/英/日 文案(含 tr -help 的全部输出)
```

运行测试:

```bash
go test ./...
```

## 历史

- **v2.0.0 起回归终端**:移除 Wails 桌面壳、Vue 前端与浏览器 GUI,改为 CLI + TUI,新增 `->` 箭头语法、`-f` 文件翻译与中/英/日三语界面。
- **翻译改为 AI**:移除 MyMemory 与自建 API 后端,翻译只走 AI(OpenAI 兼容,默认 DeepSeek),并针对翻译场景做了省 token 优化;`-o` 时用缓存与内置词典离线降级。
- **TUI 演进**:左右并排布局 → 语言随焦点用 `Ctrl+←/→` 切换 → `Esc` 打开真正的系统命令行(隐藏翻译面板、整屏显示命令输出)→ `Ctrl+L` 清屏、`Ctrl+U` 清空输入。

## 许可证

本项目基于 [Apache 2.0 License](LICENSE) 开源。

Copyright (c) 2026 Surile
