// 共享类型：与后端 wcj-go-text/app/app_barcode.go 的 Barcode* 类型保持一致

export interface BarcodeGenerateRequest {
  content: string
  format: string
  size: number
  height: number
  level: string
  fgColor: string
  bgColor: string
  showText: boolean
  margin: number
  logo: string
}

export interface BarcodeImageResult {
  dataURL: string
  width: number
  height: number
}

export interface BarcodeDecodeResult {
  text: string
  format: string
  type: string
  points: { x: number; y: number }[]
}

// 历史记录条目 = 识别结果 + 时间戳
export type BarcodeHistoryItem = BarcodeDecodeResult & { time: number }

export interface BarcodeFormatOption {
  value: string
  label: string
  kind: string
  hint: string
}

export interface BarcodeRect {
  x: number
  y: number
  w: number
  h: number
}
