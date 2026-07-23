<template>
  <div id="app" class="window-root">
    <!-- 极简拖拽栏 -->
    <div class="drag-bar" data-wails-drag>
      <span class="app-icon">🔍</span>
      <span class="app-title">Tr</span>
      <button class="close-btn" @click="closeWindow">✕</button>
    </div>

    <!-- 主体 -->
    <div class="main-content">
      <!-- 输入框 -->
      <textarea
        ref="inputRef"
        v-model="sourceText"
        class="input-area"
        placeholder="输入或划词翻译..."
        @keydown.ctrl.enter="handleTranslate"
      ></textarea>

      <!-- 工具栏（紧凑） -->
      <div class="toolbar">
        <button class="action-btn primary" @click="handleTranslate">翻译</button>
        <button class="action-btn" @click="clearAll">清空</button>
        <select v-model="targetLang" class="lang-select">
          <option value="zh">中文</option>
          <option value="en">英文</option>
          <option value="ja">日语</option>
        </select>
      </div>

      <!-- 结果区 -->
      <div v-if="resultText" class="result-area">
        <div class="result-content">{{ resultText }}</div>
        <button class="copy-btn" @click="copyResult">📋</button>
      </div>
      <div v-else class="empty-hint">✨ 等待翻译</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { Translate } from '../wailsjs/go/main/App'

const sourceText = ref('')
const resultText = ref('')
const targetLang = ref('zh')
const inputRef = ref<HTMLTextAreaElement | null>(null)

const handleTranslate = async () => {
  if (!sourceText.value.trim()) return
  try {
    const result = await Translate(sourceText.value, targetLang.value)
    resultText.value = result
  } catch {
    resultText.value = '翻译出错'
  }
}

const clearAll = () => {
  sourceText.value = ''
  resultText.value = ''
  nextTick(() => inputRef.value?.focus())
}

const copyResult = async () => {
  if (resultText.value) {
    await navigator.clipboard.writeText(resultText.value)
  }
}

const closeWindow = () => {
  // window.wails.Window.Hide()
}

onMounted(() => {
  nextTick(() => inputRef.value?.focus())
})
</script>

<style scoped>
/* ===== 全新外观 ===== */
.window-root {
  width: 100vw;
  height: 100vh;
  background: rgba(20, 22, 30, 0.65);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: #eaeef2;
  font-family: -apple-system, 'Segoe UI', Roboto, sans-serif;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
}

/* 拖拽栏 – 更矮，更不显眼 */
.drag-bar {
  height: 28px;
  min-height: 28px;
  display: flex;
  align-items: center;
  padding: 0 14px;
  background: rgba(255, 255, 255, 0.02);
  -webkit-app-region: drag;
  user-select: none;
  gap: 6px;
}
.app-icon {
  font-size: 13px;
  opacity: 0.7;
}
.app-title {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.5px;
  opacity: 0.4;
  text-transform: uppercase;
}
.close-btn {
  -webkit-app-region: no-drag;
  background: rgba(255, 70, 70, 0.1);
  border: none;
  color: rgba(255, 255, 255, 0.5);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  font-size: 11px;
  cursor: pointer;
  transition: 0.2s;
  line-height: 1;
}
.close-btn:hover {
  background: rgba(255, 70, 70, 0.6);
  color: #fff;
}

/* 主体 – 更舒展的间距 */
.main-content {
  flex: 1;
  padding: 12px 16px 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 输入框 – 圆润，字体舒服 */
.input-area {
  flex: 1;
  min-height: 60px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  padding: 10px 14px;
  color: #f0f4f8;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
  outline: none;
  transition: 0.2s;
  font-family: inherit;
}
.input-area:focus {
  border-color: rgba(100, 200, 255, 0.25);
  background: rgba(255, 255, 255, 0.06);
  box-shadow: 0 0 0 3px rgba(100, 200, 255, 0.03);
}
.input-area::placeholder {
  color: rgba(255, 255, 255, 0.2);
  font-size: 13px;
}

/* 工具栏 – 紧凑，视觉平衡 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: none;
  color: #cbd5e1;
  padding: 4px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: 0.2s;
  backdrop-filter: blur(4px);
  height: 28px;
}
.action-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}
.action-btn.primary {
  background: rgba(100, 200, 255, 0.15);
  color: #a0d8ff;
}
.action-btn.primary:hover {
  background: rgba(100, 200, 255, 0.25);
}
.lang-select {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  color: #cbd5e1;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 12px;
  cursor: pointer;
  outline: none;
  height: 28px;
  margin-left: auto;
}
.lang-select option {
  background: #1e222b;
  color: #eaeef2;
}

/* 结果区 – 柔和卡片 */
.result-area {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 14px;
  padding: 10px 14px;
  border-left: 3px solid rgba(100, 200, 255, 0.2);
  animation: fadeIn 0.2s ease;
  position: relative;
  min-height: 40px;
}
.result-content {
  font-size: 14px;
  line-height: 1.6;
  color: #f0f4f8;
  white-space: pre-wrap;
  word-break: break-word;
  padding-right: 30px;
}
.copy-btn {
  position: absolute;
  top: 8px;
  right: 10px;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.25);
  font-size: 16px;
  cursor: pointer;
  transition: 0.2s;
  padding: 0 4px;
}
.copy-btn:hover {
  color: rgba(255, 255, 255, 0.7);
}

.empty-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.15;
  font-size: 13px;
  height: 40px;
  letter-spacing: 0.3px;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>