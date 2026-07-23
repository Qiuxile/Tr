package main

import "context"

type App struct{}

func NewApp() *App {
    return &App{}
}

func (a *App) Startup(ctx context.Context) {
    // 初始化（如有需要）
}

// 供前端调用的翻译方法
func (a *App) Translate(text string, targetLang string) string {
    // 暂时返回模拟结果，后续接入真实 API
    return "翻译结果: " + text + " (目标语言: " + targetLang + ")"
}