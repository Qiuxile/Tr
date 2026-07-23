# Tr

> 一个简单优雅的终端翻译工具，使用 Go 语言编写。

Tr 是一款轻量级的命令行翻译工具，直接在终端中运行。默认使用免费的 [MyMemory](https://mymemory.translated.net/) 翻译 API，同时也支持兼容 LibreTranslate / DeepLX 格式的自定义 API。内置离线翻译功能，无网络时也能使用。

## 特性

- **快速翻译** — 输入 `tr <文本>` 即可即时获取翻译结果
- **图形界面** — 支持 `tr -gui` 启动 Web 图形界面，在浏览器中使用
- **多语言界面** — 终端界面支持中文/English/日本語三种显示语言
- **多语言支持** — 可自由配置源语言和目标语言
- **灵活的 API 后端** — 默认使用 MyMemory（免费，无需 API Key），也可接入自己的 API
- **离线翻译** — 内置翻译缓存和嵌入式英中词典，断网也能翻译
- **管道输入** — 支持 stdin 管道输入，可与其他命令组合使用
- **简洁的配置** — 通过命令轻松管理各项设置
- **Only Windows** — 目前只支持 Windows，但很快就可以支持Linux，MacOS了
- **Windows 安装包** — 提供 Inno Setup 安装程序，自动注册系统 PATH

## 安装

### 下载二进制文件

从 [Releases](https://github.com/Qiuxile/Tr/releases) 页面下载最新版本，并将其所在目录添加到系统 `PATH` 中。

### Windows 安装程序

运行 `Tr_v1.1.0_Setup.exe` 安装程序，会自动将 `Tr` 添加到系统 PATH。

### 从源码编译

```bash
git clone https://github.com/Qiuxile/Tr.git
cd Tr
go build -o tr .
```

## 使用

```bash
# 翻译文本（默认：英 → 中）
tr "Hello, world!"

# 查看当前配置
tr -config show

# 修改语言设置
tr -config set source_lang en
tr -config set target_lang ja

# 使用自定义 API（LibreTranslate / DeepLX 格式）
tr -config set api_url https://your-api.example.com/translate

# 离线模式（仅使用缓存和内置词典）
tr -o "hello"
tr --offline "good morning"

# 管道输入
echo "hello world" | tr

# 启动图形界面
tr -gui
tr -g           # 简写形式

# 查看版本
tr -version

# 查看关于信息
tr -about

# 获取帮助
tr -help
```

## 离线翻译

Tr 支持离线翻译，通过两层机制实现：

### 1. 翻译缓存

每次在线翻译的结果会自动缓存到本地（`cache.json`，最多 500 条）。再次翻译相同内容时，即使离线也能从缓存中获取结果。缓存采用 LRU 淘汰策略，自动管理空间。

### 2. 内置词典

内置了约 400 个常用英文单词和短语的英→中翻译对照表。当离线且缓存未命中时，会尝试词典翻译：
- **精确匹配**：直接查找完整短语（如 "good morning" → "早上好"）
- **逐词翻译**：对多词输入逐词查找（如 "beautiful flower" → "美丽 花"）

使用 `-o` / `--offline` 标志可强制使用离线模式，跳过网络请求。

### 翻译回退链路

```
在线 API → 本地缓存 → 内置词典 → 报错
```

## 配置

配置文件存放位置：

- **Windows**：`C:\Users\<用户名>\AppData\Roaming\Tr\config.json`

### 默认配置

```json
{
  "source_lang": "en",
  "target_lang": "zh",
  "api_url": "None"
}
```

| 键 | 说明 | 默认值 |
|---|------|--------|
| `source_lang` | 源语言代码 | `en` |
| `target_lang` | 目标语言代码 | `zh` |
| `api_url` | 自定义 API 地址（设为 `None` 则使用 MyMemory） | `None` |

### 子命令别名

| 主命令 | 别名 |
|--------|------|
| `-config` | `--config`、`-c` |
| `-help` | `--help`、`-h` |
| `-version` | `--version`、`-v` |
| `-about` | `--about`、`-a` |
| `-offline` | `--offline`、`-o` |
| `-gui` | `--gui`、`-g` |

## API 兼容性

Tr 内置支持两种翻译后端：

1. **MyMemory**（默认）— 免费翻译 API，无需 API Key
2. **LibreTranslate / DeepLX** 兼容 API — 将 `api_url` 设置为自托管或第三方 API 地址即可

## 许可证

本项目基于 [MIT License](LICENSE) 开源。

Copyright (c) 2026 Surile

## 作者

由 **Surile**（[@Qiuxile](https://github.com/Qiuxile)）开发
