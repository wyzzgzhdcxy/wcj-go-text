<template>
  <div v-if="visible" class="capture-overlay" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp">
    <canvas ref="canvasRef" class="capture-canvas" />
    <div class="overlay-toolbar" @mousedown.stop @mouseup.stop>
      <el-button size="small" @click="cancelSelection">取消 (Esc)</el-button>
      <el-button size="small" type="primary" :disabled="!validSelection" @click="confirmSelection">识别选区</el-button>
      <el-button size="small" type="success" @click="confirmFullScreen">整屏识别</el-button>
    </div>
    <div class="overlay-tip">拖拽框选要识别的区域，或点击「整屏识别」</div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, computed, onBeforeUnmount } from 'vue'

const props = defineProps({
  visible: Boolean,
  image: String,
})

const emit = defineEmits(['confirm', 'cancel', 'fullScreen'])

const canvasRef = ref(null)
const selecting = ref(false)
const startX = ref(0)
const startY = ref(0)
const endX = ref(0)
const endY = ref(0)
const img = ref(null)

const validSelection = computed(() => {
  const sel = getSelectionRect()
  return sel.w >= 16 && sel.h >= 16
})

function getSelectionRect() {
  const x1 = Math.min(startX.value, endX.value)
  const y1 = Math.min(startY.value, endY.value)
  const x2 = Math.max(startX.value, endX.value)
  const y2 = Math.max(startY.value, endY.value)
  return { x: Math.round(x1), y: Math.round(y1), w: Math.round(x2 - x1), h: Math.round(y2 - y1) }
}

function draw() {
  const canvas = canvasRef.value
  const image = img.value
  if (!canvas || !image) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const w = canvas.width
  const h = canvas.height
  ctx.clearRect(0, 0, w, h)
  ctx.drawImage(image, 0, 0)

  const sel = getSelectionRect()
  if (sel.w > 0 && sel.h > 0) {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.45)'
    ctx.beginPath()
    ctx.rect(0, 0, w, h)
    ctx.rect(sel.x, sel.y, sel.w, sel.h)
    ctx.fill('evenodd')

    ctx.strokeStyle = '#409eff'
    ctx.lineWidth = 2
    ctx.strokeRect(sel.x, sel.y, sel.w, sel.h)

    ctx.fillStyle = '#409eff'
    ctx.font = '12px sans-serif'
    ctx.fillText(`${sel.w} × ${sel.h}`, sel.x + 4, sel.y - 6)
  }
}

watch(
  () => props.visible,
  async (val) => {
    if (!val) return
    await nextTick()
    const canvas = canvasRef.value
    if (!canvas) return
    const image = new Image()
    image.onload = () => {
      img.value = image
      canvas.width = image.naturalWidth
      canvas.height = image.naturalHeight
      draw()
    }
    image.src = props.image
  },
  { immediate: false }
)

function getPos(e) {
  const canvas = canvasRef.value
  const rect = canvas.getBoundingClientRect()
  const ratioX = canvas.width / rect.width
  const ratioY = canvas.height / rect.height
  return {
    x: (e.clientX - rect.left) * ratioX,
    y: (e.clientY - rect.top) * ratioY,
  }
}

function onMouseDown(e) {
  const pos = getPos(e)
  selecting.value = true
  startX.value = pos.x
  startY.value = pos.y
  endX.value = pos.x
  endY.value = pos.y
  draw()
}

function onMouseMove(e) {
  if (!selecting.value) return
  const pos = getPos(e)
  endX.value = pos.x
  endY.value = pos.y
  draw()
}

function onMouseUp() {
  selecting.value = false
  draw()
}

function confirmSelection() {
  const sel = getSelectionRect()
  if (sel.w < 16 || sel.h < 16) return
  emit('confirm', sel)
}

function confirmFullScreen() {
  emit('fullScreen')
}

function cancelSelection() {
  emit('cancel')
}

// Esc 快速退出截图模式
function onKeydown(e) {
  if (e.key === 'Escape') cancelSelection()
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      window.addEventListener('keydown', onKeydown)
    } else {
      window.removeEventListener('keydown', onKeydown)
    }
  }
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.capture-overlay {
  position: fixed;
  inset: 0;
  /* 低于 Element Plus 弹层默认层级（约 2000+），保证识别成功的 toast 能浮在遮罩上 */
  z-index: 1500;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: crosshair;
}

.capture-canvas {
  max-width: 100vw;
  max-height: 100vh;
  object-fit: contain;
}

.overlay-toolbar {
  position: fixed;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 8px;
  backdrop-filter: blur(4px);
}

.overlay-tip {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  color: #fff;
  font-size: 13px;
  padding: 6px 12px;
  background: rgba(0, 0, 0, 0.55);
  border-radius: 4px;
  pointer-events: none;
}
</style>
