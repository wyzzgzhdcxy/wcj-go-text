<template>
  <el-row :gutter="20" class="generate-panel">
    <el-col :span="10" class="form-col">
      <el-form label-position="top" class="gen-form">
        <el-form-item label="内容">
          <el-input
            v-model="req.content"
            type="textarea"
            :rows="3"
            placeholder="输入网址、文本、商品码、WIFI 信息等"
          />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="码制">
              <el-select v-model="req.format" style="width: 100%">
                <el-option
                  v-for="f in formats"
                  :key="f.value"
                  :label="f.label"
                  :value="f.value"
                />
              </el-select>
              <div class="hint">{{ currentFormat?.hint }}</div>
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="isQR">
            <el-form-item label="纠错等级">
              <el-radio-group v-model="req.level">
                <el-radio-button value="L" title="容错 7%，容量最大" label="L" />
                <el-radio-button value="M" title="容错 15%，推荐" label="M" />
                <el-radio-button value="Q" title="容错 25%" label="Q" />
                <el-radio-button value="H" title="容错 30%，适合加图标" label="H" />
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="isQR ? '边长（像素）' : '宽度（像素）'">
              <el-slider v-model="req.size" :min="64" :max="1024" show-input />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="!isQR">
            <el-form-item label="高度（像素）">
              <el-slider v-model="req.height" :min="20" :max="400" show-input />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="isQR ? '静区（模块数，规范推荐 4）' : '边距（像素）'">
          <el-slider
            v-if="isQR"
            v-model="req.margin"
            :min="0"
            :max="8"
            show-input
          />
          <el-slider v-else v-model="req.margin" :min="0" :max="64" show-input />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="前景色">
              <el-color-picker v-model="req.fgColor" show-alpha />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="背景色">
              <el-color-picker v-model="req.bgColor" show-alpha />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="!isQR">
          <el-checkbox v-model="req.showText">在条形码下方显示内容文字</el-checkbox>
        </el-form-item>

        <el-form-item v-if="isQR" label="中心图标（可选）">
          <div class="logo-row">
            <el-button size="small" @click="selectLogo">选择图标</el-button>
            <el-button v-if="req.logo" size="small" type="danger" text @click="req.logo = ''">
              清除
            </el-button>
            <span v-if="req.logo" class="logo-tip">已选择图标</span>
          </div>
        </el-form-item>

        <el-form-item>
          <div class="toolbar">
            <el-button type="primary" :icon="Refresh" @click="generate" :loading="loading">立即生成</el-button>
            <el-button :icon="CopyDocument" @click="copyImage" :disabled="!result">复制图片</el-button>
            <el-button :icon="Download" @click="saveImage" :disabled="!result">保存</el-button>
            <el-button :icon="Monitor" @click="quickSave" :disabled="!result">保存到桌面</el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-col>

    <el-col :span="14" class="preview-col">
      <div class="preview-box card-like">
        <div class="preview-header">
          <span>预览</span>
          <el-button
            v-if="savedPath"
            size="small"
            text
            type="primary"
            :icon="FolderOpened"
            @click="openSavedFolder"
          >
            打开所在文件夹
          </el-button>
        </div>
        <div class="preview-wrap preview-inner" v-loading="loading">
          <img
            v-if="result?.dataURL"
            :src="result.dataURL"
            alt="生成的码"
            title="点击复制图片"
            @click="copyImage"
          />
          <div v-else class="empty-tip">等待生成</div>
        </div>
        <div v-if="result" class="preview-meta">
          {{ result.width }} × {{ result.height }} 像素
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script setup>
import { reactive, ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Download, FolderOpened, Monitor, Refresh } from '@element-plus/icons-vue'
import {
  ListBarcodeFormats,
  GenerateBarcode,
  SaveBarcodeImage,
  QuickSaveBarcodeToDesktop,
  ShowBarcodeInFolder,
  WriteBarcodeClipboardImage,
  SelectBarcodeLogoFile,
} from '../wailsjs/go/app/App.js'
import { barcodeGenerateSeed } from '../store'

// 先用本地兜底列表渲染，等后端返回后替换，避免首帧码制判断错误
const formats = ref([
  { value: 'qr', label: '二维码 QR Code', kind: 'qr', hint: '文本、网址、中文都支持' },
  { value: 'code128', label: 'Code 128', kind: 'barcode', hint: '最通用的一维码' },
])
const result = ref(null)
const loading = ref(false)
const savedPath = ref('')

const req = reactive({
  content: 'https://example.com',
  format: 'qr',
  size: 320,
  height: 120,
  level: 'M',
  fgColor: '#000000ff',
  bgColor: '#ffffffff',
  showText: true,
  margin: 4,
  logo: '',
})

const isQR = computed(() => !!currentFormat.value && currentFormat.value.kind === 'qr')
const currentFormat = computed(() => formats.value.find((f) => f.value === req.format))

function debounce(fn, wait) {
  let timer
  return (...args) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), wait)
  }
}

const generate = async () => {
  const content = (req.content || '').trim()
  if (!content) {
    result.value = null
    return
  }
  loading.value = true
  try {
    result.value = await GenerateBarcode({ ...req })
    savedPath.value = ''
  } catch (e) {
    ElMessage.error({ message: e?.message || e || '生成失败', grouping: true })
    result.value = null
  } finally {
    loading.value = false
  }
}

const debouncedGenerate = debounce(generate, 350)

watch(
  req,
  () => debouncedGenerate(),
  { deep: true, immediate: false }
)

// 切换码制时把静区/边距重置成该码制的推荐值
watch(
  () => req.format,
  () => {
    req.margin = isQR.value ? 4 : 20
  }
)

// 识别页「去生成」送来的内容：只填充内容，由深层 watcher 统一触发生成
watch(barcodeGenerateSeed, (seed) => {
  if (!seed) return
  req.content = seed.text
})

const selectLogo = async () => {
  try {
    const logo = await SelectBarcodeLogoFile()
    if (logo) req.logo = logo
  } catch (e) {
    ElMessage.error(e?.message || e || '选择图标失败')
  }
}

const copyImage = async () => {
  if (!result.value) return
  try {
    await WriteBarcodeClipboardImage(result.value.dataURL)
    ElMessage.success('图片已复制到剪贴板')
  } catch (e) {
    ElMessage.error(e?.message || e || '复制图片失败')
  }
}

const saveImage = async () => {
  if (!result.value) return
  try {
    const fileName = `code-${req.format}.png`
    const path = await SaveBarcodeImage(result.value.dataURL, fileName)
    if (!path) return
    savedPath.value = path
    ElMessage.success(`已保存到 ${path}`)
  } catch (e) {
    ElMessage.error(e?.message || e || '保存失败')
  }
}

const quickSave = async () => {
  if (!result.value) return
  try {
    const fileName = `code-${req.format}-${Date.now()}.png`
    const path = await QuickSaveBarcodeToDesktop(result.value.dataURL, fileName)
    savedPath.value = path
    ElMessage.success(`已保存到桌面：${path}`)
  } catch (e) {
    ElMessage.error(e?.message || e || '保存到桌面失败')
  }
}

const openSavedFolder = async () => {
  if (!savedPath.value) return
  try {
    await ShowBarcodeInFolder(savedPath.value)
  } catch (e) {
    ElMessage.error(e?.message || e || '打开文件夹失败')
  }
}

onMounted(async () => {
  try {
    const list = await ListBarcodeFormats()
    if (Array.isArray(list) && list.length) formats.value = list
  } catch {
    // 保留兜底列表即可
  }
  generate()
})
</script>

<style scoped>
.generate-panel {
  height: 100%;
}

.form-col,
.preview-col {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.gen-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

.hint {
  color: #9ca3af;
  font-size: 12px;
  line-height: 1.4;
  margin-top: 4px;
}

.logo-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-tip {
  color: #10b981;
  font-size: 12px;
}

.preview-box {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  overflow: hidden;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  margin-bottom: 12px;
  color: #374151;
}

.preview-inner {
  flex: 1;
  min-height: 0;
}

.preview-inner img {
  cursor: pointer;
}

.preview-meta {
  margin-top: 10px;
  text-align: center;
  color: #6b7280;
  font-size: 12px;
}
</style>
