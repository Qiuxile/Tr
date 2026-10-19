# Tr 开发路线 (v2.2.0+ CLI/TUI)

> 项目为纯终端形态：Go CLI + bubbletea TUI，翻译统一走 AI 后端，无 GUI/Web 依赖。

## 已完成

### v2.0.0
- [x] 移除 Wails 桌面壳、Vue 前端与浏览器 GUI
- [x] CLI + TUI、管道、`-config`、`-o`、`-help`
- [x] 中/英/日 界面文案

### v2.1.0
- [x] `->` 箭头语法：自动识别源语言并翻译到指定语言
- [x] AI 后端：OpenAI 兼容接口 + 省 token 策略
- [x] 文件翻译：`tr -f <文件> [输出]`
- [x] 长文本分块；`finish_reason=length` 检测
- [x] TUI 系统命令行、清屏、语言切换、回归测试

### v2.2.0
- [x] 翻译统一为 AI 后端，移除独立翻译后端
- [x] 默认 Pollinations 免费 AI（`openai-fast`），无需 key 即可试用
- [x] 支持内置默认 key（`DefaultAPIKey`，ldflags 注入）与环境变量覆盖
- [x] 默认 key / 环境变量 key 不写回 `config.json`
- [x] 缓存重写：真正 LRU、模型/服务商隔离、原子写盘、大小上限
- [x] `tr cache clear|stats`
- [x] 文件翻译修复：保留缩进、尾空行不参与请求、致命错误立即中止、部分失败 exit 1
- [x] CLI 修复：去掉多余空行、`-f -o` 顺序、`--` 结束解析、保留字可翻译
- [x] 配置：目录/文件权限、语言/URL 校验、隐藏输入、`config path`
- [x] TUI：stdout 终端检测、长文本进度不污染 alt-screen、shell 输出上限/取消/单命令
- [x] 模块路径改为 `github.com/Qiuxile/Tr`，版本可 ldflags 注入
- [x] 安装包：UTF-8 BOM、AppId、发布者/版本元数据、x64 限制
- [x] 测试：AI/缓存/配置/文件/参数/语言/词典

## 待办想法

- [ ] 文件翻译并行化（受 AI 速率限制约束；需先做限速器）
- [ ] TUI 显示 token 用量（读取响应 `usage` 字段）
- [ ] 拉丁语系细分：英文/法文/德文等基于拉丁字母文本的自动识别
- [ ] TUI 增加会话历史浏览（↑/↓ 翻阅已翻译条目）与鼠标滚轮
- [ ] 词典扩充：英→日、中→英 词条，提升离线可用性
- [ ] Linux / macOS CI 构建与发布（GitHub Actions）
- [ ] Windows 系统命令行进程树终止（`taskkill /T`）
- [ ] 可选流式输出 / 自定义 HTTP headers（Azure OpenAI 等）
- [ ] 可复制译文到剪贴板的 TUI 快捷键
- [ ] 按服务商自适应限速（Pollinations 30 秒 / 其他后端不限）
