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
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入或粘贴文本”"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">参数</span>
        <label class="inline-input">替换字符
          <input v-model="replaceChar" maxlength="3" />
        </label>
        <span class="toolbar-spacer"></span>
        <button class="chip primary" @click="trimRepeated"><span class="icon">✕</span>去除重复</button>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符<</span></div>
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
import trimRepeated from 'trim-repeated';
import { CopyToClipboard } from "../wailsjs/go/main/App.js";
import { ElNotification } from "element-plus";

export default {
  data() {
    return {
      inputData: '11111111\n22222222\n33333333333',
      result: '',
      replaceChar: '#'
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
    trimRepeated() {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      this.result = trimRepeated(this.inputData, this.replaceChar);
    }
  }
}
</script>
