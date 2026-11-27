# Tr 产品介绍片 · 分镜与制作说明

## 成片规格

| 项目 | 16:9 横版 | 9:16 竖版 |
|---|---|---|
| 分辨率 | 1920×1080 | 1080×1920 |
| 帧率 | 30 fps | 30 fps |
| 时长 | 60 s | 60 s |
| 编码 | H.264 / yuv420p / CRF 18 | 同左 |
| 音频 | AAC 128 kbps，单声道 | 同左 |
| 字幕 | 中文硬字幕 + 外挂 SRT | 同左 |
| 配音 | 无 | 无 |
| 风格 | 暗色终端科技感，青色强调 | 同左 |

字体：Microsoft YaHei（中文）、Consolas（终端）、Segoe UI Bold（英文标题）。

## 分镜

| 时间 | 场景 | 画面 | 字幕 |
|---|---|---|---|
| 0.0–3.5s | 开场 | 深色网格背景，Tr 标志与光环淡入，版本徽章 | — |
| 3.5–10.5s | CLI | 终端窗口，逐字输入 `tr "Hello, world!"`，输出 `你好，世界！`，下方两个示例命令 | 一条命令，直接翻译 |
| 10.5–18.0s | 箭头语法 | 上方 `中文 → English` 语言胶囊，终端输入 `tr 你好 -> en`，输出 `Hello` | 自动识别语言，指定目标 |
| 18.0–27.0s | TUI | 左右分栏界面，输入 `hello world`，右侧出现 `你好，世界`；底部快捷键 | 左右对照的 TUI |
| 27.0–35.0s | 文件翻译 | `tr -f README.md docs/zh.md`，文件胶囊、进度条、完成勾选 | 整篇文件翻译，保留格式 |
| 35.0–41.0s | 离线 | `tr -o "hello"`，输出 `你好`，`OFFLINE · 0 token` 徽章 | 离线也能用 |
| 41.0–48.0s | 能力总览 | 四张能力卡片：CLI + TUI、文件翻译、多语言界面、跨平台 | 一个工具，覆盖日常翻译 |
| 48.0–54.0s | 安装 | `go install github.com/Qiuxile/Tr@latest`，完成勾选，GitHub 地址 | 安装只要一行 |
| 54.0–60.0s | 结尾 | 标志、`Tr`、slogan、GitHub 地址、Apache-2.0 淡入 | — |

## 声音设计

- 无配音，仅字幕。
- 背景为程序生成的 A 小调氛围铺底：Am–F–C–G，8 小节，低音量。
- 打字场景叠加轻微键盘点击声（CLI、箭头、文件、离线、安装）。
- 场景切换处有很轻的噪声 whoosh。
- 所有声音由 `video/make_audio.py` 用 numpy 合成，无第三方音乐版权问题。

## 制作方式

- `video/render.py` 用 Pillow 逐帧绘制，通过管道直接送入 ffmpeg 编码。
- 不进行屏幕录制，不会出现本机用户名、路径、API key 或真实网络请求。
- 终端内容为演示文案，翻译结果是预设文本。

```bash
# 重新生成音频
python video/make_audio.py

# 渲染两个版本
python video/render.py --all

# 只渲染某一张预览
python video/render.py --still 21.0 --out Output/preview.png
python video/render.py --still 21.0 --portrait --out Output/preview-p.png
```

## 文件

| 文件 | 说明 |
|---|---|
| `Output/tr-intro-60s-16x9.mp4` | 横版成片 |
| `Output/tr-intro-60s-9x16.mp4` | 竖版成片 |
| `Output/tr-intro-cover-16x9.png` | 横版封面 |
| `Output/tr-intro-cover-9x16.png` | 竖版封面 |
| `video/captions.srt` | 外挂中文字幕 |
| `video/render.py` | 渲染脚本 |
| `video/make_audio.py` | 音频生成脚本 |
| `video/storyboard.md` | 本文件 |

## 可调整项

- 文案：修改 `render.py` 中 `caption(...)` 与 `scene_*` 里的字符串。
- 主色：修改 `ACCENT` / `ACCENT_2`。
- 时长：修改 `SCENES` 的时间区间与 `DURATION`，音频同步调整。
- 配音：用 `--no-audio` 或替换 `audio.wav` 后重新编码；字幕已在画面内。
- 封面：用 `--still` 渲染任意时间点作为封面。
