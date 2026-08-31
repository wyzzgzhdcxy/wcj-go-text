<template>
  <el-row :gutter="20" class="decode-panel">
    <el-col :span="15" class="left-col">
      <div class="toolbar card-like">
        <el-button type="primary" :icon="FolderOpened" @click="decodeFromFile" :loading="loading">导入文件</el-button>
        <el-button :icon="DocumentCopy" @click="decodeFromClipboard" :loading="loading">剪贴板图片</el-button>
        <el-button :icon="Crop" @click="captureScreen" :loading="loading">截图框选</el-button>
        <el-button :icon="FullScreen" @click="decodeFullScreen" :loading="loading">整屏识别</el-button>
      </div>

      <div
        class="drop-zone card-like"
        @click="triggerFileInput"
        @drop.prevent="onDrop"
        @dragover.prevent="dragOver = true"
        @dragleave.prevent="dragOver = false"
        :class="{ 'drag-over': dragOver }"
      >
        <input ref="fileInput" type="file" accept="image/*" style="display: none" @change="onFileSelected" />
        <el-icon :size="40" color="#9ca3af"><Picture /></el-icon>
        <div class="drop-text">拖拽图片到这里，或点击选择文件</div>
        <div class="drop-sub">支持 PNG / JPG / GIF / BMP / TIFF / WebP，可同时框选多个码</div>
      </div>

      <div v-if="preview" class="preview-box card-like">
        <div class="preview-header">
          <span>待识别图片</span>
          <el-button size="small" text :icon="Delete" @click="clearPreview">清除</el-button>
        </div>
        <div class="preview-wrap decode-preview">
          <div class="preview-stack">
            <img ref="previewImg" :src="preview" alt="待识别" @load="imgVersion++" />
            <div
              v-for="(box, i) in resultBoxes"
              :key="i"
              class="detect-box"
              :style="{
                left: box.left + '%',
                top: box.top + '%',
                width: box.width + '%',
                height: box.height + '%',
              }"
            >
              <span class="detect-index">{{ i + 1 }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-col>

    <el-col :span="9" class="right-col">
      <div class="result-box card-like">
        <div class="result-header">
          <span>识别结果</span>
          <div class="result-actions">
            <el-button v-if="results.length > 1" size="small" text type="primary" @click="copyAll">
              复制全部
            </el-button>
            <el-tag v-if="results.length" type="success" effect="dark" size="small">{{ results.length }} 条</el-tag>
          </div>
        </div>
        <ResultList :results="results" />
      </div>

      <div class="history-box card-like">
        <div class="history-header">
          <span>历史记录</span>
          <el-button v-if="history.length" size="small" text type="danger" @click="clearHistory">清空</el-button>
        </div>
        <ResultList :results="history" />
      </div>
    </el-col>
  </el-row>

  <CaptureOverlay
    :visible="captureVisible"
    :image="captureImage"
    @confirm="onCaptureConfirm"
    @cancel="captureVisible = false"
    @full-screen="onCaptureFullScreen"
  />
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  FolderOpened,
  DocumentCopy,
  Crop,
  FullScreen,
  Picture,
  Delete,
} from '@element-plus/icons-vue'
import {
  CaptureBarcodeScreen,
  DecodeBarcodeImageData,
  DecodeBarcodeScreenshot,
  ReadBarcodeClipboardImage,
  SelectBarcodeImageFile,
  ReadBarcodeImageFile,
  WriteBarcodeClipboardText,
} from '../wailsjs/go/app/App.js'
import ResultList from './ResultList.vue'
import CaptureOverlay from './CaptureOverlay.vue'

const preview = ref('')
const previewImg = ref(null)
const imgVersion = ref(0)
const results = ref([])
const history = ref([])
const loading = ref(false)
const dragOver = ref(false)
const fileInput = ref(null)

const captureVisible = ref(false)
const captureImage = ref('')

const HISTORY_KEY = 'wcj-barcode:history'

// 把识别结果的定位点换算成预览图上的百分比高亮框
const resultBoxes = computed(() => {
  void imgVersion.value
  if (!preview.value || results.value.length === 0) return []
  const img = previewImg.value
  if (!img || !img.naturalWidth || !img.complete) return []
  const nw = img.naturalWidth
  const nh = img.naturalHeight
  const pad = 2

  return results.value
    .filter((r) => r.points && r.points.length >= 2)
    .map((r) => {
      const xs = r.points.map((p) => p.x)
      const ys = r.points.map((p) => p.y)
      const minX = Math.min(...xs)
      const minY = Math.min(...ys)
      const maxX = Math.max(...xs)
      const maxY = Math.max(...ys)
      const left = Math.max(0, (minX / nw) * 100 - pad)
      const top = Math.max(0, (minY / nh) * 100 - pad)
      const width = Math.min(100 - left, ((maxX - minX) / nw) * 100 + pad * 2)
      const height = Math.min(100 - top, ((maxY - minY) / nh) * 100 + pad * 2)
      return { left, top, width, height }
    })
})

function loadHistory() {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    if (raw) history.value = JSON.parse(raw)
  } catch {
    history.value = []
  }
}

function saveHistory() {
  try {
    const trimmed = history.value.slice(-50)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(trimmed))
  } catch {
    // 存不进去就算了，不影响使用
  }
}

function pushHistory(items) {
  if (!items.length) return
  const now = Date.now()
  const annotated = items.map((r) => ({ ...r, time: now }))
  for (const item of annotated) {
    const last = history.value[history.value.length - 1]
    if (last && last.text === item.text && last.format === item.format) continue
    history.value.push(item)
  }
  saveHistory()
}

function clearHistory() {
  history.value = []
  localStorage.removeItem(HISTORY_KEY)
}

function clearPreview() {
  preview.value = ''
  results.value = []
}

function triggerFileInput() {
  fileInput.value?.click()
}

function readFileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

async function onDrop(e) {
  dragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  try {
    const dataURL = await readFileAsDataURL(file)
    await decodeFromDataURL(dataURL)
  } catch (e) {
    ElMessage.error(e?.message || e || '读取文件失败')
  }
}

async function onFileSelected(e) {
  const target = e.target
  const file = target.files?.[0]
  if (!file) return
  try {
    const dataURL = await readFileAsDataURL(file)
    await decodeFromDataURL(dataURL)
  } catch (e) {
    ElMessage.error(e?.message || e || '读取文件失败')
  } finally {
    target.value = ''
  }
}

async function decodeFromFile() {
  try {
    const path = await SelectBarcodeImageFile()
    if (!path) return
    const dataURL = await ReadBarcodeImageFile(path)
    await decodeFromDataURL(dataURL)
  } catch (e) {
    ElMessage.error(e?.message || e || '识别失败')
  }
}

async function decodeFromClipboard() {
  loading.value = true
  try {
    // 先取剪贴板图片展示出来，识别失败时用户能看到是什么图
    const dataURL = await ReadBarcodeClipboardImage()
    await decodeFromDataURL(dataURL)
  } catch (e) {
    ElMessage.error({ message: e?.message || e || '剪贴板识别失败', grouping: true })
  } finally {
    loading.value = false
  }
}

async function decodeFromDataURL(dataURL) {
  preview.value = dataURL
  results.value = []
  imgVersion.value = 0
  loading.value = true
  try {
    const res = await DecodeBarcodeImageData(dataURL)
    handleResults(res)
  } catch (e) {
    results.value = []
    ElMessage.error({ message: e?.message || e || '识别失败', grouping: true })
  } finally {
    loading.value = false
  }
}

async function captureScreen() {
  loading.value = true
  try {
    const shot = await CaptureBarcodeScreen('all')
    captureImage.value = shot.dataURL
    captureVisible.value = true
  } catch (e) {
    ElMessage.error(e?.message || e || '截图失败')
  } finally {
    loading.value = false
  }
}

async function onCaptureConfirm(rect) {
  captureVisible.value = false
  try {
    const img = new Image()
    await new Promise((resolve, reject) => {
      img.onload = () => resolve()
      img.onerror = () => reject(new Error('截图加载失败'))
      img.src = captureImage.value
    })

    const pad = { x: rect.w * 0.08, y: rect.h * 0.08 }
    const x = Math.max(0, rect.x - pad.x)
    const y = Math.max(0, rect.y - pad.y)
    const w = Math.min(img.naturalWidth - x, rect.w + pad.x * 2)
    const h = Math.min(img.naturalHeight - y, rect.h + pad.y * 2)
    if (w < 8 || h < 8) {
      ElMessage.error('选区太小，请重新框选')
      return
    }

    const canvas = document.createElement('canvas')
    canvas.width = Math.round(w)
    canvas.height = Math.round(h)
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('画布创建失败')
    ctx.drawImage(img, x, y, w, h, 0, 0, canvas.width, canvas.height)
    await decodeFromDataURL(canvas.toDataURL('image/png'))
  } catch (e) {
    loading.value = false
    ElMessage.error(e?.message || e || '选区识别失败')
  }
}

async function onCaptureFullScreen() {
  captureVisible.value = false
  await decodeFromDataURL(captureImage.value)
}

async function decodeFullScreen() {
  loading.value = true
  try {
    const res = await DecodeBarcodeScreenshot('all')
    handleResults(res)
  } catch (e) {
    ElMessage.error({ message: e?.message || e || '整屏识别失败', grouping: true })
  } finally {
    loading.value = false
  }
}

function handleResults(items) {
  results.value = items || []
  pushHistory(results.value)
  if (results.value.length) {
    ElMessage.success(`识别成功，共 ${results.value.length} 条`)
  } else {
    ElMessage.warning('未识别到二维码或条形码')
  }
}

async function copyAll() {
  if (!results.value.length) return
  const text = results.value.map((r) => r.text).join('\n')
  try {
    await WriteBarcodeClipboardText(text)
    ElMessage.success(`已复制 ${results.value.length} 条结果`)
  } catch (e) {
    ElMessage.error(e?.message || e || '复制失败')
  }
}

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.decode-panel {
  height: 100%;
}

.left-col,
.right-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
}

.toolbar {
  padding: 12px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.drop-zone {
  padding: 32px 20px;
  text-align: center;
}

.drop-text {
  margin-top: 10px;
  color: #4b5563;
  font-size: 14px;
}

.drop-sub {
  margin-top: 4px;
  color: #9ca3af;
  font-size: 12px;
}

.preview-box,
.result-box,
.history-box {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 14px;
  min-height: 0;
  overflow: hidden;
}

.preview-header,
.result-header,
.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  font-weight: 600;
  color: #374151;
  flex-shrink: 0;
}

.result-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.decode-preview {
  flex: 1;
  min-height: 0;
}

.preview-stack {
  position: relative;
  display: inline-block;
  font-size: 0;
  max-width: 100%;
}

.preview-stack img {
  display: block;
  max-width: 100%;
  max-height: 62vh;
  object-fit: contain;
}

.detect-box {
  position: absolute;
  border: 2px solid #409eff;
  border-radius: 4px;
  background: rgba(64, 158, 255, 0.12);
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.6) inset;
  pointer-events: none;
}

.detect-index {
  position: absolute;
  top: -10px;
  left: -2px;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  border-radius: 9px;
  background: #409eff;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  padding: 0 4px;
}

.right-col {
  display: flex;
  flex-direction: column;
}

.result-box {
  flex: 1.6;
}

.history-box {
  flex: 1;
}
</style>
