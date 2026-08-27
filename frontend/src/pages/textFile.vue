<template>
  <div class="page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.chars }} 字符</span></div>
          <div class="card-actions">
            <button class="chip" @click="importFile"><span class="icon">📂</span>选择文件</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="输入文件路径或内容”"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">编码</span>
        <button class="chip" @click="run('base64编码')">Base64 编码</button>
        <button class="chip" @click="run('base64解码')">Base64 解码</button>
        <button class="chip" @click="run('md5')">MD5</button>
        <button class="chip" @click="run('sha1')">SHA1</button>
        <button class="chip" @click="run('sha256')">SHA256</button>
        <button class="chip" @click="run('sha512')">SHA512</button>
        <span class="file-path" v-if="filepath">{{ filepath }}</span>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符</span></div>
          <div class="card-actions">
            <button class="chip primary" :disabled="!result" @click="copyText">复制结果</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="result" class="text-panel" placeholder="处理结果会显示在这里”" disabled></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CopyToClipboard, FileEncode, SelectFile } from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";

export default {
  data() {
    return {
      inputData: '',
      result: '',
      filepath: ''
    }
  },
  computed: {
    inputStats() { return { chars: (this.inputData || '').length } },
    resultStats() { return { chars: (this.result || '').length } }
  },
  methods: {
    clearInput() { this.inputData = ''; this.result = ''; this.filepath = ''; },
    async copyText() {
      if (!this.result) return;
      await CopyToClipboard(this.result);
      ElNotification({ title: '已复制', message: '结果已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
    async importFile() {
      try {
        const result = await SelectFile();
        if (result) {
          this.filepath = result;
          this.inputData = result;
        }
      } catch (e) {
        console.error('选择文件失败:', e);
      }
    },
    async run(val) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先选择文件或输入内容', position: 'bottom-right', type: 'error' });
        return;
      }
      this.result = await FileEncode(this.inputData, val);
    }
  }
}
</script>

<style scoped>
.file-path {
  color: var(--text-tertiary);
  font-size: 12px;
  margin-left: 8px;
  word-break: break-all;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
