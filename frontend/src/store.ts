import { ref } from 'vue'

// 二维码工具内部共享的标签页状态
export const barcodeActiveTab = ref<'generate' | 'decode'>('generate')

// 生成页监听这个种子：识别结果点"去生成"时填入内容并触发重生成
export const barcodeGenerateSeed = ref<{ text: string; ts: number } | null>(null)

export function sendToGenerate(text: string) {
  barcodeGenerateSeed.value = { text, ts: Date.now() }
  barcodeActiveTab.value = 'generate'
}
