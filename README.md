# Tr

> 一个简单、快速的终端 AI 翻译工具。**CLI + TUI** 双形态，纯 Go，无桌面/Web 依赖。
>
> 翻译只走 AI 后端：默认使用智谱 BigModel（GLM-5.3-Flash）的 OpenAI 兼容接口，填入默认 key 后开箱即用；也可以切换到 DeepSeek、OpenAI 或任意兼容服务。

```console
$ tr "Hello, world!"
你好，世界！

$ tr 你好 -> en
Hello

$ tr
╭ 原文 (en) ─────────╮ ╭ 译文 (zh) ── · 在线 AI ╮
│ hello world        │ │ 你好，世界              │
╰────────────────────╯ ╰────────────────────────╯
```

---

## 目录

- [快速开始](#快速开始)
- [特性](#特性)
- [安装](#安装)
- [CLI 使用](#cli-使用)
- [默认 AI 接口](#默认-ai-接口)
- [切换到自己的服务商](#切换到自己的服务商)
- [翻译链路](#翻译链路)
- [文件翻译](#文件翻译)
- [TUI 使用](#tui-使用)
- [离线翻译](#离线翻译)
- [配置](#配置)
- [语言与自动检测](#语言与自动检测)
- [实现说明](#实现说明)
- [开发与测试](#开发与测试)
- [打包与发布](#打包与发布)
- [常见问题](#常见问题)
- [安全与隐私](#安全与隐私)
- [历史](#历史)
- [许可证](#许可证)

---

## 快速开始

### 1. 编译

```bash
# 需要 Go 1.26+
go build -trimpath -ldflags "-s -w" -o tr .
```

Windows 下产物为 `tr.exe`。也可以使用仓库里的 `Output/Tr.exe`，或用 `setup.iss` 生成安装包。

### 2. 直接翻译

```bash
tr "Hello, world!"        # 英 → 中
tr 你好 -> en              # 自动识别源语言，译成英文，仅本次生效
tr                        # 无参数：进入交互式 TUI
```

默认后端是智谱 BigModel 的 OpenAI 兼容接口（`glm-5.3-flash`），需要 API key。key 可以来自内置默认 key、`config.json` 或环境变量；详见[默认 AI 接口](#默认-ai-接口)。

### 3. 配置 key（源码构建时必填）

发布版可以内置默认 key，源码构建默认不带 key，需要先配置：

```bash
tr config set api_key sk-xxxxxxxxxxxxxxxx
```

默认地址已经是智谱 BigModel：

```text
https://open.bigmodel.cn/api/paas/v4
glm-5.3-flash
```

如果要换服务商：

```bash
tr config set ai_base_url https://api.deepseek.com/v1
tr config set ai_model deepseek-flash
tr config set api_key sk-xxxxxxxxxxxxxxxx
```

或者用环境变量，不把密钥写进配置文件：

```bash
TR_API_KEY=sk-xxx tr "hello world"
```

优先级：环境变量 > 配置文件 > 内置默认 key。内置默认 key 只用于开箱即用，建议发布时通过 ldflags 注入，而不是提交明文。

---

## 特性

- **纯 AI 翻译** — 翻译统一走 OpenAI 兼容接口，默认智谱 BigModel（GLM-5.3-Flash）。
- **开箱即用** — 内置默认 key（发布时注入）或用户 key 均可；无 key 时可用 `-o` 离线。
- **CLI 友好** — 输出干净，适合管道、脚本和 `$(...)`；支持 stdin。
- **TUI** — 左右布局、多行输入、译文滚动、语言切换、内建系统命令行。
- **文件翻译** — 按行分块、保留缩进/BOM/CRLF，可输出到终端、文件或目录。
- **省 token** — 缓存优先、单轮极简提示词、关闭思考模式、限制输出长度、按块缓存。
- **可靠的缓存** — 真正 LRU、按模型/服务商隔离、原子写盘、可手动清理。
- **离线可用** — 缓存 + 内置英中词典，完全不联网。
- **多语言界面** — 中文 / English / 日本語。
- **跨平台** — Windows / Linux / macOS，纯 Go 依赖。
- **可脚本化** — 明确的退出码、`--` 结束选项解析、`-o` 离线模式。

---

## 安装

### 从源码

```bash
git clone https://github.com/Qiuxile/Tr.git
cd Tr
go build -trimpath -ldflags "-s -w" -o tr .
```

### `go install`

发布 tag 包含当前模块路径后，可以：

```bash
go install github.com/Qiuxile/Tr@latest
```

### Windows 安装包

仓库里的 `setup.iss` 是 Inno Setup 脚本：

```powershell
go build -trimpath -ldflags "-s -w -X github.com/Qiuxile/Tr/internal/tr.Version=2.2.0" -o Output\Tr.exe .
iscc setup.iss
```

产物为 `Output/Tr_v2.2.0_Setup.exe`，会把 `tr.exe` 安装到 `Program Files\Tr`，并按任务选择加入 `PATH`。

### 加入 PATH

把 `tr` / `tr.exe` 所在目录加入系统 `PATH`，即可在任意位置运行。

---

## CLI 使用

```bash
# 文本翻译（默认英 → 中）
tr "Hello, world!"

# 自动识别源语言 + 指定目标语言
tr 你好 -> en
tr "Hello world" -> 中文
tr こんにちは -> zh
tr hello -> nl                 # 任何 ISO 639-1 代码都可以尝试

# 箭头语法是位置式的：内容 = 第一个参数，语言 = 箭头后的 token
tr hello world -> zh           # 只翻译 "hello"
tr "hello world" -> zh         # 翻译整句
tr "你好"->en                  # 粘连写法也支持

# 翻译文件
tr -f README.md                # 打印到终端
tr -f docs/en.md docs/zh.md    # 写入文件
tr -f docs/en.md docs/zh/      # 写入目录（自动创建）
tr -f docs/en.md -> ja         # 文件也支持箭头语法

# 管道
echo "hello world" | tr
echo "hello" | tr -> ja
tr -> ja < input.txt

# 离线模式
tr -o "hello"
tr -o -f docs/en.md

# 配置
tr config show
tr config set target_lang ja
tr config path

# 缓存
tr cache stats
tr cache clear

# 帮助 / 版本 / 关于
tr -help
tr -version
tr -about
```

### 命令速查

| 命令 | 说明 |
|---|---|
| `tr <文本>` | 翻译文本并输出到 stdout |
| `tr <文本> -> <语言>` | 自动识别源语言，译成指定语言；只作用于本次执行 |
| `tr` | 无参数且 stdin/stdout 都是终端 → TUI；stdin 是管道 → 翻译管道内容 |
| `tr tui` | 显式启动 TUI |
| `tr -f <文件> [输出]` | 翻译文件；省略输出打印到终端，给出路径则写入文件/目录 |
| `tr -o` / `--offline` | 强制离线（缓存 + 内置词典，不联网） |
| `tr config show\|set\|path` | 查看/修改配置、打印配置文件路径 |
| `tr cache stats\|clear` | 查看/清空翻译缓存 |
| `tr -help` / `-version` / `-about` | 帮助 / 版本 / 关于 |

选项可以出现在任意位置，所以 `tr -f -o file.txt`、`tr -o -f file.txt` 都可以。

### 退出码

| 退出码 | 含义 |
|---|---|
| `0` | 成功 |
| `1` | 运行时错误：翻译失败、网络/API 失败、文件读写失败、缓存/配置写入失败 |
| `2` | 用法错误：未知选项、缺少参数、非法配置值 |

文件翻译如果**部分块失败**，tr 会把失败块保留为原文、写入结果，然后以退出码 `1` 结束；这样脚本能发现“并非整篇都翻译成功”。

### 管道与 `--`

```bash
echo "hello" | tr
echo "hello" | tr -> zh
```

`tr` 无参数时会读取 stdin；如果 stdin 和 stdout 都是终端，则进入 TUI。

`config`、`cache`、`tui` 是保留字。想翻译这些词本身，用 `--` 结束选项解析：

```bash
tr -- config          # 翻译单词 "config"
tr -- tui             # 翻译单词 "tui"
tr -- -o              # 翻译字符串 "-o"
```

`--` 之后的参数不再被当作选项或子命令。

---

## 默认 AI 接口

默认配置：

```text
ai_base_url = https://open.bigmodel.cn/api/paas/v4
ai_model    = glm-5.3-flash
api_key     = 空（源码构建；发布版可由 DefaultAPIKey 注入）
```

这是智谱 BigModel 的 OpenAI 兼容接口。API key 可以通过内置默认 key、`config.json` 或环境变量提供；没有 key 时翻译会明确提示，不会静默降级到其他引擎。

### 模型与参数

- `glm-5.3-flash` 是当前低成本模型；也可以 `tr config set ai_model glm-5.3` 换旗舰。
- GLM-5.3 / 5.3-Flash 强制开启思考，tr 会发送 `reasoning_effort: low`，尽量减少翻译时的推理开销。
- 单次 HTTP 超时 45 秒；429 / 5xx / 网络错误最多 2 次退避重试（`Retry-After` 会被遵守，最多等 10 秒）。
- 免费/低价额度、模型可用性以智谱控制台为准；tr 不保证第三方服务可用性。
- **不要向第三方接口发送密钥、隐私数据或机密文档。** 文本会离开本机。

### 内置默认 key

[`DefaultAPIKey`](internal/tr/config.go) 是一个可注入的变量：

```bash
go build -ldflags \
  "-s -w -X github.com/Qiuxile/Tr/internal/tr.DefaultAPIKey=sk-..." \
  -o tr .
```

优先级：

```text
TR_API_KEY / ZHIPU_API_KEY / GLM_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY
  > config.json 中的 api_key
  > DefaultAPIKey
```

如果三者都没有，默认 BigModel 地址会报“未配置 AI 密钥”；本地/ Pollinations 等免密钥地址除外。

内置 key 不会写回 `config.json`；`tr config show` 会掩码显示。

### 示例

```bash
tr "Hello, world!"
# 你好，世界！

tr 你好 -> en
# Hello
```

---

## 切换到自己的服务商

tr 只要求服务实现 OpenAI 兼容的 `POST /chat/completions`（或完整 endpoint）。常见配置：

### DeepSeek

```bash
tr config set ai_base_url https://api.deepseek.com/v1
tr config set ai_model deepseek-flash
tr config set api_key sk-xxxxxxxx
```

DeepSeek 还支持思考模式（默认关闭，开启更费 token）：

```bash
tr config set ai_thinking true
```

### OpenAI

```bash
tr config set ai_base_url https://api.openai.com/v1
tr config set ai_model gpt-4o-mini
tr config set api_key sk-xxxxxxxx
```

### 本地模型（Ollama 等）

```bash
tr config set ai_base_url http://127.0.0.1:11434/v1
tr config set ai_model qwen2.5:7b
tr config set api_key ""
```

`localhost` / `127.0.0.1` / `::1` 被视为免密钥地址，适用于本地推理服务。

### 完整 endpoint

`ai_base_url` 可以填 base，也可以直接填完整 endpoint：

```text
https://api.example.com/v1/chat/completions
https://text.pollinations.ai/openai
```

tr 会按以下规则补全：

- 以 `/chat/completions` 结尾 → 原样使用
- 以 `/openai` 结尾 → 原样使用（Pollinations）
- 以 `/v1`、`/v4` 等版本段结尾 → 追加 `/chat/completions`
- 其他 → 追加 `/v1/chat/completions`

### 密钥与环境变量

支持以下环境变量，**优先级从高到低**：

```text
TR_API_KEY
ZHIPU_API_KEY
GLM_API_KEY
DEEPSEEK_API_KEY
OPENAI_API_KEY
```

环境变量会覆盖配置文件里的 `api_key`；环境变量中的密钥**不会被写回** `config.json`。

```bash
TR_API_KEY=sk-xxx tr "hello"
TR_API_KEY=sk-xxx tr config set target_lang ja   # 不会把 key 写进文件
```

也可以交互式输入，不进入 shell history：

```bash
tr config set api_key
# 或
tr config set api_key -
```

---

## 翻译链路

### 在线

```text
缓存命中 → 返回（零 token）
   ↓ 未命中
AI 后端（默认智谱 GLM-5.3-Flash / 可换 DeepSeek、OpenAI、本地模型）
   ↓ 成功
写入缓存 → 返回
   ↓ 失败
报错
```

### 离线（`-o` / `--offline`）

```text
缓存命中 → 返回
   ↓ 未命中
内置英中词典 → 返回
   ↓ 未命中
报错
```

离线模式完全不联网、不消耗 token。

### 缓存

- 缓存 key 由 **后端作用域**（`ai_base_url` + `ai_model`）、源语言、目标语言和**精确原文**的 SHA-256 组成。
- 切换模型或服务商不会错误复用旧译文。
- 最多 500 条；超过后按最近使用时间淘汰最旧的 20%。
- 单条原文或译文超过 32 KiB 不缓存；长文档按 1,500 字符分块缓存。
- 写盘使用临时文件 + rename，进程中断不会留下半截 JSON。
- 每次 CLI 结束、文件翻译结束、TUI 退出时统一 flush。
- `tr cache stats` 查看条目数，`tr cache clear` 清空。

### 长文本

原文超过约 **1,500 字符**时自动分块：

- 在段落、换行、句号等自然边界切分；
- 每块单独请求、单独缓存；
- 拼接时**不额外插入分隔符**，保留原文的标点和换行；
- CLI 会在 stderr 显示 `长文本分块翻译中 3/5 …` 的进度（仅交互终端）。

### 省 token 策略

| 措施 | 说明 |
|---|---|
| 缓存优先 | 命中直接返回，零请求 |
| 关闭思考模式 | 默认 `ai_thinking=false`，避免为一句翻译消耗思维链 token |
| 极简单轮提示词 | 只有一条 system 提示，无示例、无多轮上下文 |
| 限制输出长度 | 按输入字符数估算 `max_tokens`（64–4096） |
| `temperature=0` | 翻译任务尽量确定性；部分思考模型会忽略该参数 |
| 不自动放大请求 | 文件翻译遇到 401/429/网络错误直接停止，不逐行重试 |
| 结果后处理 | 去掉模型加的引号、`Translation:` 前缀、代码围栏和控制字符 |
| 分块缓存 | 改一段只重译一段 |

### 错误处理

API 错误会尽量给出可操作的信息：

| HTTP | 提示 |
|---|---|
| 401 / 403 | 密钥无效，或密钥与服务商不匹配 |
| 402 | 账户余额不足或免费额度受限 |
| 429 | 请求被限流；可稍后重试或换服务商 |
| 5xx | 服务端错误，自动退避重试 |
| 其他 | 状态码 + 服务端返回的详情 |

网络错误、超时、响应解析失败、输出被截断也都有单独提示。

---

## 文件翻译

```bash
tr -f <文件>              # 输出到 stdout
tr -f <文件> <输出文件>    # 写入文件
tr -f <文件> <输出目录>/   # 写入目录，沿用源文件名；目录会自动创建
tr -f <文件> -> ja        # 指定目标语言
```

处理规则：

- 按行分块，每块最多约 1,500 字符；块内行结构保持不变。
- **缩进和行首/行尾空白会保留**（不会被 `TrimSpace` 吃掉）。
- 文件末尾的空行不会进入请求，避免“模型少返回一行 → 整块白做 → 逐行重试”。
- 空行、CRLF、UTF-8 BOM 会原样保留。
- 如果模型返回的行数对不上，或直接回显原文，会降级为逐行翻译。
- 401 / 429 / 网络 / 空输入等致命错误会**立即中止**，不会把一次失败放大成 N 次逐行请求。
- 非 UTF-8 文件会报错；请先转换为 UTF-8。
- 单文件上限 16 MiB；二进制文件请勿使用 `-f`。
- 输出文件使用临时文件 + rename 原子写入。

部分块失败时：失败块保留原文，其余块写入结果，stderr 提示，退出码 `1`。

---

## TUI 使用

在交互式终端直接运行 `tr`，或显式：

```bash
tr tui
```

界面为左右布局：左侧原文，右侧译文。顶部显示语言对、在线/离线、翻译状态。

### 快捷键

| 按键 | 功能 |
|---|---|
| `Ctrl+T` | 翻译输入框内容 |
| `Enter` | 输入框内换行（支持多行） |
| `Ctrl+←` / `Ctrl+→` | 切换当前焦点那一侧的语言：输入框=源语言，译文区=目标语言 |
| `Tab` / `Shift+Tab` | 焦点在输入框 ↔ 译文区之间切换 |
| `↑` / `↓` / `PgUp` / `PgDn` | 译文区聚焦时滚动译文 |
| `Esc` | 打开系统命令行 |
| `Ctrl+L` | 清空译文区 / 系统命令行里清空整屏 |
| `Ctrl+U` | 清空输入框 |
| `Ctrl+S` | 互换源/目标语言（仅当前会话，不写配置） |
| `Ctrl+O` | 临时切换在线 / 离线 |
| `Ctrl+Q` / `Ctrl+C` | 退出（系统命令行执行中时，`Ctrl+C` 先取消命令） |

语言切换是会话级的，不会写配置文件。需要持久修改请用 `tr config set target_lang`。

### 系统命令行（`Esc`）

按 `Esc` 进入真正的系统 shell：翻译面板隐藏，整屏显示命令输出，底部是一行命令输入。

| 操作 | 功能 |
|---|---|
| 输入 + `Enter` | 在当前目录执行命令（stdout/stderr 合并显示） |
| `↑` / `↓` | 翻阅历史命令 |
| `PgUp` / `PgDn` / `Home` / `End` | 滚动输出（保留最近 1,000 行） |
| `Ctrl+L` | 清空整屏输出（历史保留） |
| `Esc` | 输入非空时清空；输入为空时返回翻译界面 |
| `Ctrl+C` | 执行中取消当前命令；空闲时退出 tr |

保护措施：

- 单条命令 30 秒超时；
- 单条命令输出最多保留 1 MiB，超出会截断并提示；
- 同一时间只允许执行一条命令；
- 命令通过 `cmd /c`（Windows）或 `sh -c`（Unix）执行；
- Windows 下先切到 UTF-8 代码页，常见中文输出不会乱码；
- 命令输出会剥掉 ANSI 转义，避免破坏界面。

---

## 离线翻译

```bash
tr -o "hello"
tr -o -f docs/en.md
```

离线模式使用两层机制：

1. **翻译缓存**：在线翻译过的内容会落盘，离线时可直接命中。
2. **内置英中词典**：约 460 个常用英文词/短语，支持精确短语和逐词翻译。

词典是“应急”级别：

- 只支持英 → 中；
- 逐词翻译会保留标点、大小写和代码标识符，但无法处理复杂句式；
- 未命中的词保留原文；
- 一个词都没命中时才算失败，不会把原文原样当成“成功翻译”。

---

## 配置

### 配置文件位置

| 系统 | 路径 |
|---|---|
| Windows | `%APPDATA%\Tr\config.json` |
| Linux | `~/.config/Tr/config.json` |
| macOS | `~/Library/Application Support/Tr/config.json` |

缓存文件与配置同目录：`cache.json`。

配置目录权限 `0700`，配置文件 `0600`（Unix；Windows 使用用户目录 ACL）。

### 配置键

| 键 | 说明 | 默认值 |
|---|---|---|
| `source_lang` | 源语言代码；`auto` = 自动检测 | `en` |
| `target_lang` | 目标语言代码 | `zh` |
| `ui_lang` | 界面语言：`zh` / `en` / `ja` | `zh` |
| `api_key` | AI 密钥；可用内置默认 key 或环境变量 | 空 |
| `ai_base_url` | OpenAI 兼容地址或完整 endpoint | `https://open.bigmodel.cn/api/paas/v4` |
| `ai_model` | 模型名 | `glm-5.3-flash` |
| `ai_thinking` | 是否开启 DeepSeek 思考模式 | `false` |

### 修改配置

```bash
tr config show
tr config set target_lang ja
tr config set source_lang auto
tr config set ui_lang en
tr config set ai_base_url https://api.deepseek.com/v1
tr config set ai_model deepseek-flash
tr config set api_key sk-xxxxxxxx
tr config path
```

配置值会做校验：

- `source_lang`：`auto` 或合法语言代码/别名
- `target_lang`：合法语言代码/别名
- `ui_lang`：只接受 `zh` / `en` / `ja`
- `ai_base_url`：必须是 `http(s)` URL
- `ai_model`：不能为空
- `ai_thinking`：`true/false`、`on/off`、`1/0`、`开/关`

无效值会直接报错，不会写进配置文件。

### 兼容旧写法

```bash
tr -config show
tr -config set target_lang ja
```

---

## 语言与自动检测

### 目标语言

| 代码 | 语言 | 代码 | 语言 | 代码 | 语言 | 代码 | 语言 |
|---|---|---|---|---|---|---|---|
| `zh` | 中文（简） | `zh-hant` | 中文（繁） | `en` | 英语 | `ja` | 日语 |
| `ko` | 韩语 | `ru` | 俄语 | `fr` | 法语 | `de` | 德语 |
| `es` | 西班牙语 | `it` | 意大利语 | `pt` | 葡萄牙语 | `ar` | 阿拉伯语 |
| `th` | 泰语 | `vi` | 越南语 | `hi` | 印地语 | `el` | 希腊语 |
| `nl` | 荷兰语 | `pl` | 波兰语 | `tr` | 土耳其语 | `uk` | 乌克兰语 |
| `id` | 印尼语 | | | | | | |

还接受常见名称/别名：

```text
中文、英文、日本語、한국어、korean、russian、Deutsch、español、繁體 …
```

任意 ISO 639-1 代码都可以尝试（例如 `tr hello -> nl`），模型是否支持取决于后端。

### 自动检测

`-> <语言>` 或 `source_lang=auto` 时，tr 按 Unicode 文字系统检测：

- 含假名 → 日语
- 汉字 → 中文
- 谚文 → 韩语
- 西里尔 → 俄语
- 阿拉伯、希腊、泰、天城文 → 对应语言
- 纯拉丁字母/数字/标点 → 英语

拉丁语系细分（法语/德语/西班牙语等）目前仍按英语处理；需要时请显式设置 `source_lang`。

---

## 实现说明

### 目录结构

```text
main.go                    进程入口，调用 tr.Run(os.Args)
internal/tr/
  app.go                   CLI 分发、参数解析、config/cache 子命令、帮助输出
  ai.go                    OpenAI 兼容后端、请求构造、重试、分块、清洗
  translate.go             翻译链路、Options、错误分类、target 语义
  cache.go                 LRU 缓存、模型作用域、原子落盘
  file.go                  文件翻译、行对齐、BOM/CRLF、原子输出
  config.go                配置读写、环境变量、权限、校验、默认 key
  lang.go                  语言检测、代码/别名归一化、TUI 语言列表
  dict.go / dict_en_zh.go  内置英中词典（离线模式）
  tui.go                   bubbletea TUI、系统命令行
  i18n.go                  中/英/日文案
  fsutil.go                原子文件写入
  *_test.go                单元测试
```

### 数据流

```text
Run(args)
  ├─ LoadConfig()            只读文件，不创建
  ├─ parseArgs()             两遍式：先标志，再位置参数
  ├─ config/cache 子命令     直接返回
  ├─ applyEnvAIKey()         环境变量覆盖
  ├─ applyDefaultAIKey()     内置默认 key（如果未配置其他 key）
  ├─ Translate()             缓存 → AI / 离线词典
  └─ print result            不额外加空行
```

### 参数解析

- 选项可以出现在任意位置；
- `--` 之后全部按文本处理；
- `-f` 只消费位置参数，不会把 `-o` 当成文件名；
- `config show|set|path`、`cache clear|stats` 只在“看起来像子命令”时生效；
- `tui` 仅在它是唯一位置参数时作为命令；翻译单词 `tui` 用 `tr -- tui`。

### 并发与性能

- 单次翻译的 AI 请求是顺序的；长文本按块顺序请求，便于进度和缓存。
- 默认 BigModel 不额外限速；Pollinations 等免费接口会做请求间隔。
- TUI 的翻译和 shell 命令分别在 bubbletea 的异步命令中执行，不阻塞渲染。
- 缓存写盘只在退出/结束时发生，避免每翻译一行重写整个 JSON。
- TUI shell 输出有 1 MiB 上限，防止一条 `dir /s` 把内存吃满。

### 错误处理

- 翻译链路使用带 sentinel 的错误类型，`errors.Is` 可区分“无密钥 / 网络失败 / AI 失败 / 离线未命中”；
- 错误信息在产生处本地化，避免 CLI 输出里混入英文 sentinel；
- 文件翻译把“致命错误”和“结构不匹配”分开：前者中止，后者逐行降级。

### 依赖

```text
bubbletea / bubbles / lipgloss   终端 UI
golang.org/x/term                终端检测与隐藏输入
```

没有 WebView、Node.js、GUI 框架或数据库。

---

## 开发与测试

### 构建

```bash
go build -trimpath -ldflags "-s -w" -o tr .
```

带版本信息与内置默认 key：

```bash
go build -trimpath -ldflags \
  "-s -w -X github.com/Qiuxile/Tr/internal/tr.Version=2.2.0 -X github.com/Qiuxile/Tr/internal/tr.DefaultAPIKey=sk-..." \
  -o tr .
```

`Version` 和 `DefaultAPIKey` 都是变量，所以可以这样注入；`tr -version`、`-about` 和安装包版本应保持一致。

### 测试

```bash
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

测试覆盖：

- `app`：箭头语法、参数解析、`-f` 顺序、`--`、保留字
- `ai`：endpoint 归一化、请求形状、温度、错误分类、429 重试、清洗
- `cache`：LRU、模型隔离、大小上限、损坏备份、版本迁移
- `config`：默认值、私密权限、环境变量不落盘、校验、损坏备份
- `file`：尾空行、缩进、BOM/CRLF、致命错误中止
- `lang` / `dict`：检测、别名、繁体、词典行为
- `tui`：布局、焦点、快捷键、系统命令行、清屏、不写配置

### 交叉编译

```bash
GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o tr-linux .
GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o tr-macos .
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o tr.exe .
```

### 调试 AI 接口

```bash
tr -version
tr config show
tr cache stats
```

如果接口超时/429/402，先检查 key、余额和额度，或切到自己的服务商。想确认请求形状，可以用本地 mock 服务替换 `ai_base_url` 后观察请求体。

---

## 打包与发布

### Windows 安装包

1. 构建：

```powershell
go build -trimpath -ldflags "-s -w -X github.com/Qiuxile/Tr/internal/tr.Version=2.2.0" -o Output\Tr.exe .
```

2. 用 Inno Setup 编译：

```powershell
iscc setup.iss
```

`setup.iss` 已包含：

- UTF-8 BOM（保证中文任务描述不乱码）
- 稳定 `AppId`（升级/卸载识别）
- 发布者、支持 URL、版本资源
- 桌面/开始菜单/PATH 可选任务
- x64 架构限制
- 卸载时从 PATH 移除安装目录

### 发布检查清单

- [ ] `go test -race ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] `gofmt -l .` 无输出
- [ ] `tr -version`、`setup.iss` 的 `AppVersion`、git tag 一致
- [ ] `README` 的默认 API / 配置表已更新
- [ ] `LICENSE` 与 `app.license` 都是 Apache-2.0
- [ ] Windows 安装包用 `iscc` 实际编译过
- [ ] 默认接口和自带 key 两条路径都手测过
- [ ] 不把 `Output/*.exe`、缓存、密钥提交进仓库

---

## 常见问题

**第一次运行很慢 / 429 / 超时**

默认 BigModel 接口的额度/限流以控制台为准；tr 会做最多 2 次退避重试。仍失败时：

1. 等几分钟再试；
2. 配置自己的服务商（见[切换到自己的服务商](#切换到自己的服务商)）；
3. 或先用 `-o` 查缓存/词典。

**提示“未配置 AI 密钥”**

说明 `ai_base_url` 是非免费、非 localhost 的地址，且没有配置 key。运行：

```bash
tr config set api_key sk-xxxxxxxx
# 或
TR_API_KEY=sk-xxxx tr "..."
```

**HTTP 401 / 402**

- 401/403：key 错误，或 key 与 `ai_base_url` 不属于同一家服务商。
- 402：账户余额不足或免费额度受限。

**HTTP 429**

请求太密集。文件翻译和长文本会分块请求，更容易触发限流；可以等待、减少文件大小，或换带额度的服务商。

**翻译结果和原文一样**

很可能是命中了缓存（缓存是刻意设计）。用 `tr cache clear` 清空，或修改原文。切换模型后不会命中旧缓存，因为缓存按 `ai_base_url + ai_model` 隔离。若 AI 直接回显原文，tr 会报错而不是把原文当译文返回。

**文件翻译部分内容没变**

失败块会保留原文并以退出码 `1` 结束；stderr 会有提示。检查密钥/网络/限流，或拆分文件。

**文件报“不是有效的 UTF-8”**

tr 只处理 UTF-8 文本。请先转码（例如 `iconv -f gbk -t utf-8`），再翻译。

**`tr config` 不显示配置**

现在 `tr config` 会被当作普通文本翻译；查看配置请用 `tr config show`。翻译单词 “config” 用 `tr -- config`。

**`tr tui` 和翻译单词 “tui”**

`tr tui` 是显式启动 TUI；翻译单词请用 `tr -- tui`。

**TUI 快捷键没反应**

终端无法区分 `Enter` 和 `Ctrl+Enter`，所以翻译键是 `Ctrl+T`；某些终端会占用 `Ctrl+←/→` 或 Alt 组合键。建议使用 Windows Terminal、iTerm2、Kitty、WezTerm 等现代终端。

**Windows 系统命令行中文乱码**

tr 会先执行 `chcp 65001`。如果个别程序仍输出 GBK，需要在该程序自己的参数里指定编码。

**配置/缓存在哪，怎么清**

- Windows：`%APPDATA%\Tr\`
- Linux：`~/.config/Tr/`
- macOS：`~/Library/Application Support/Tr/`
- `config.json` = 配置；`cache.json` = 缓存
- `tr cache clear` 清空缓存；删除文件也可以

**能不能当库用 / 有没有 HTTP API**

目前只提供 CLI 与 TUI。脚本里直接调用 CLI 即可；stdin/stdout 和退出码都适合脚本。

---

## 安全与隐私

- **默认 AI 接口会把文本发送到第三方（智谱 BigModel）。** 不要翻译密钥、密码、个人隐私、未公开代码或机密文档。
- 使用自己的服务商时，数据同样会发送到该服务商；请阅读其隐私政策。
- `api_key` 存在本地配置文件（Unix `0600`）或环境变量中；`tr config show` 会掩码。
- 环境变量和内置默认 key 不会写回配置文件。
- `tr config set api_key` 不带值时会隐藏输入，避免进入 shell history。
- `ai_base_url` 若是 `http://`（非 localhost），密钥会以明文经过网络；建议使用 HTTPS。
- 缓存文件包含原文和译文，可能有敏感内容；共享机器上请定期 `tr cache clear`。
- TUI 系统命令行会执行你输入的系统命令，请只运行可信命令。
- 内置 `DefaultAPIKey` 会编进可执行文件，能被提取；它的用途是开箱即用，不适合存放高价值长期密钥。

---

## 历史

- **v2.2.0** — 翻译统一为 AI 后端；默认智谱 BigModel GLM-5.3-Flash，支持内置默认 key；新增 `cache` 子命令、隐藏密钥输入、环境变量覆盖；修复文件翻译缩进/尾空行/请求放大、CLI 多余空行、`-f -o` 解析、缓存 LRU/模型隔离/原子写、TUI 进度污染与 shell 输出上限；模块路径改为 `github.com/Qiuxile/Tr`；补充大量测试与安装包元数据。
- **v2.1.0** — 箭头语法、长文本分块、AI 后端与省 token 策略完善。
- **v2.0.0** — 移除 Wails 桌面壳与 Vue 前端，回归 CLI + TUI；新增 `-f`、`->`、中/英/日界面。
- **v1.x** — 早期 GUI / 多后端版本。

---

## 许可证

本项目基于 [Apache License 2.0](LICENSE) 开源。

Copyright (c) 2026 Surile
