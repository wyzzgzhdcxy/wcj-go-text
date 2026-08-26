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
          <textarea v-model="inputData" class="text-panel" placeholder="首行为表头，其它行为数据”"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">参数</span>
        <label class="inline-input">分割第          <input v-model="splitChar" maxlength="3" />
        </label>
        <span class="toolbar-spacer"></span>
        <button class="chip primary" @click="toJSON"><span class="icon">{}</span>转JSON</button>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符<</span></div>
          <div class="card-actions">
            <button class="chip primary" :disabled="!result" @click="copyText">复制结果</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="result" class="text-panel" placeholder="JSON 结果会显示在这里”" disabled></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CopyToClipboard, Text2Json } from "../wailsjs/go/main/App.js";
import { ElNotification } from "element-plus";

export default {
  data() {
    return {
      inputData: '11111111\n22222222\n33333333333',
      result: '',
      splitChar: ','
    }
  },
  computed: {
    inputStats() { return { lines: this.inputData ? this.inputData.split('\n').length : 0 } },
    resultStats() { return { chars: (this.result || '').length } }
  },
  methods: {
    clearInput() { this.inputData = ''; this.result = ''; },
    async copyText() {
      if (!this.result) return;
      await CopyToClipboard(this.result);
      ElNotification({ title: '已复制', message: '结果已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
    async toJSON() {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      const r = await Text2Json(this.inputData, this.splitChar);
      this.result = JSON.stringify(r);
    }
  }
}
</script>
