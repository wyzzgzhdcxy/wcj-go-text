<template>
  <div class="page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.chars }} 字符<</span></div>
          <div class="card-actions">
            <button class="chip" @click="clearInput">清空</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入或粘贴文本"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">算法</span>
        <button class="chip" @click="run('sha1')">SHA1</button>
        <button class="chip" @click="run('sha256')">SHA256</button>
        <button class="chip" @click="run('sha512')">SHA512</button>
        <button class="chip" @click="run('hex编码')">HEX 编码</button>
        <button class="chip" @click="run('hex解码')">HEX 解码</button>
        <button class="chip" @click="run('ascii编码')">ASCII 编码</button>
        <button class="chip" @click="run('ascii解码')">ASCII 解码</button>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符<</span></div>
          <div class="card-actions">
            <button class="chip primary" :disabled="!result" @click="copyText">复制结果</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="result" class="text-panel" placeholder="处理结果会显示在这里" disabled></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CopyToClipboard, TextEncode } from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";

export default {
  data() {
    return {
      inputData: '11111111\n22222222\n33333333333',
      result: ''
    }
  },
  computed: {
    inputStats() { return { chars: (this.inputData || '').length } },
    resultStats() { return { chars: (this.result || '').length } }
  },
  methods: {
    clearInput() { this.inputData = ''; this.result = ''; },
    async copyText() {
      if (!this.result) return;
      await CopyToClipboard(this.result);
      ElNotification({ title: '已复制', message: '结果已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
    async run(val) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      this.result = await TextEncode(this.inputData, val);
    }
  }
}
</script>
