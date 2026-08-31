<template>
  <div class="result-list">
    <div v-if="results.length === 0" class="empty-tip">暂无识别结果</div>
    <div v-else class="flex-column gap-12">
      <div v-for="(r, idx) in results" :key="idx" class="result-item">
        <div class="badges">
          <el-tag :type="r.type === 'qrcode' ? 'success' : 'primary'" size="small" effect="dark">
            {{ r.type === 'qrcode' ? '二维码' : '条形码' }}
          </el-tag>
          <el-tag v-if="r.format" type="info" size="small" style="margin-left: 6px">
            {{ r.format }}
          </el-tag>
          <span v-if="itemTime(r)" class="item-time">{{ itemTime(r) }}</span>
        </div>
        <div class="code-text flex-1">{{ r.text }}</div>
        <el-button-group>
          <el-button size="small" type="primary" @click="copy(r.text)">复制</el-button>
          <el-button v-if="isUrl(r.text)" size="small" @click="openUrl(r.text)">打开</el-button>
          <el-button size="small" title="把内容送回生成页" @click="regenerate(r.text)">去生成</el-button>
        </el-button-group>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { OpenBarcodeURL, WriteBarcodeClipboardText } from '../wailsjs/go/app/App.js'
import { sendToGenerate } from '../store'

defineProps({
  results: { type: Array, default: () => [] },
})

const copy = async (text) => {
  try {
    await WriteBarcodeClipboardText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (e) {
    ElMessage.error(e?.message || e || '复制失败')
  }
}

const isUrl = (text) => /^https?:\/\//i.test(text || '')

// Wails 窗口里 window.open 不可靠，走系统浏览器打开
const openUrl = async (text) => {
  try {
    await OpenBarcodeURL(text)
  } catch (e) {
    ElMessage.error(e?.message || e || '打开链接失败')
  }
}

const regenerate = (text) => {
  sendToGenerate(text)
}

const itemTime = (r) => {
  if (!r || !r.time) return ''
  const d = new Date(r.time)
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style scoped>
.result-list {
  padding: 4px 0;
}

.badges {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.item-time {
  margin-left: 8px;
  color: #9ca3af;
  font-size: 11px;
  white-space: nowrap;
}

.result-item {
  align-items: flex-start;
}

.code-text {
  padding: 0 6px;
  color: #374151;
}
</style>
