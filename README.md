# Tr —— 极简桌面翻译助手

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v2.0+-blue?style=flat)](https://wails.io/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](https://github.com/yourname/Tr)

> **Tr** 是一款专为日常办公、阅读、学习设计的轻量级桌面翻译工具。它不再是一个命令行的极客玩具，而是面向所有人的智能助手。

![演示截图占位](docs/screenshot.png)

---

## ✨ 核心特性

- **🖱️ 划词即译**：在任何软件（浏览器、PDF、IDE）中选中文字，自动弹出翻译结果，无需复制粘贴。
- **📷 截屏 OCR 翻译**：遇到图片、扫描件、无法复制的文字，一键截图，即刻 OCR 识别并翻译。
- **🎯 无感弹窗**：仿 Spotlight 风格的毛玻璃小窗，用完即走，不打扰你的工作流。
- **⚡ 极速响应**：本地缓存机制 + 轻量级 Go 后端，毫秒级响应，告别卡顿。
- **⌨️ 全盘快捷键**：支持自定义全局快捷键，随时唤出主界面或触发截图。
- **🌐 多引擎聚合**：内置多种翻译源（MyMemory / 百度 / 有道 / DeepL），自由切换。

---

## 🛠️ 技术架构

- **后端框架**：[Wails](https://wails.io/) + **Go** —— 负责系统钩子、全局监听、HTTP 请求、缓存与 OCR 调度。
- **前端界面**：Vue 3 / React + CSS 毛玻璃效果 —— 构建流畅、美观的无边框透明弹窗。
- **跨平台能力**：一套代码同时编译 Windows 和 macOS 原生应用，资源占用极低。

---

## 📥 下载与安装

前往 [Releases](https://github.com/yourname/Tr/releases) 页面下载对应平台的最新安装包。

| 平台 | 安装方式 |
| :--- | :--- |
| **Windows** | 下载 `.exe` 安装包，双击安装即可。 |
| **macOS** | 下载 `.dmg` 文件，拖拽 `Tr.app` 到 Applications 文件夹。 |
| **Linux** | 下载 `AppImage` 文件，赋予执行权限后运行。 |

---

## 🚀 快速使用指南

1.  **启动**：安装后，程序自动在系统托盘后台运行。
2.  **划词翻译**：选中任意文字 -> 弹出翻译结果（默认只需选中，无需按任何键）。
3.  **手动翻译**：按下 `Ctrl+Shift+T` (默认) 呼出主输入框，输入文字按回车翻译。
4.  **截屏翻译**：按下 `Ctrl+Shift+S` (默认)，鼠标拖拽框选屏幕区域，自动识别并翻译。
5.  **设置**：右键点击托盘图标，选择“设置”可切换翻译源和修改快捷键。

---

## 👨‍💻 开发者指南 (本地构建)

如果你想自行编译或参与开发：

1.  **前置条件**：
    -   Go 1.21+ ([下载](https://golang.org/dl/))
    -   Node.js 16+ ([下载](https://nodejs.org/))
    -   Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

2.  **克隆与安装**：
    ```bash
    git clone https://github.com/yourname/Tr.git
    cd Tr
    wails install