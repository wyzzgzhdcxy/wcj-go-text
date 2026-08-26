<template>
  <div class="page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip" @click="clearInput">清空</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="每行一项内容”"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">排序</span>
        <button class="chip" @click="forwardSort"><span class="icon">→</span>正向</button>
        <button class="chip" @click="reverseSort"><span class="icon">→</span>逆向</button>
        <button class="chip" @click="reverse"><span class="icon">⇓</span>倒序</button>
        <button class="chip" @click="exchange"><span class="icon">⇓</span>交换输入/结果</button>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip primary" :disabled="!result" @click="copyText">复制结果</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="result" class="text-panel" placeholder="排序结果会显示在这里”" disabled></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CopyToClipboard } from "../wailsjs/go/main/App.js";
import { ElNotification } from "element-plus";

export default {
  data() {
    return {
      inputData: '11111111\n22222222\n33333333333',
      result: ''
    }
  },
  computed: {
    inputStats() { return { lines: this.inputData ? this.inputData.split('\n').length : 0 } },
    resultStats() { return { lines: this.result ? this.result.split('\n').length : 0 } }
  },
  methods: {
    clearInput() { this.inputData = ''; this.result = ''; },
    async copyText() {
      if (!this.result) return;
      await CopyToClipboard(this.result);
      ElNotification({ title: '已复制', message: '结果已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
    needInput() {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return false;
      }
      return true;
    },
    forwardSort() {
      if (!this.needInput()) return;
      const arr = this.inputData.split('\n');
      arr.sort((a, b) => a.localeCompare(b, 'zh-Hans-CN', { sensitivity: 'base' }));
      this.result = arr.join('\n');
    },
    reverseSort() {
      if (!this.needInput()) return;
      const arr = this.inputData.split('\n');
      arr.sort((a, b) => b.localeCompare(a, 'zh-Hans-CN', { sensitivity: 'base' }));
      this.result = arr.join('\n');
    },
    reverse() {
      if (!this.needInput()) return;
      this.result = this.inputData.split('\n').reverse().join('\n');
    },
    exchange() {
      const tmp = this.inputData;
      this.inputData = this.result;
      this.result = tmp;
    }
  }
}
</script>
