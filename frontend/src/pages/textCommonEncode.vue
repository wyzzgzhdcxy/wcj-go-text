<template>
  <div class="page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.chars }} 字符 · {{ inputStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip" @click="clearInput">清空</button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入或粘贴文本”"></textarea>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">操作</span>
        <button class="chip" @click="run('转json')"><span class="icon">{}</span>转JSON</button>
        <button class="chip" @click="run('全部大写')"><span class="icon">Aa</span>全大写</button>
        <button class="chip" @click="run('全部小写')"><span class="icon">aa</span>全小写</button>
        <button class="chip" @click="run('删除重复行')"><span class="icon">⊟</span>去重</button>
        <button class="chip" @click="run('删除空行')"><span class="icon">∕</span>去空行</button>
        <button class="chip" @click="run('去除不可见字符')"><span class="icon">⎕</span>去不可见</button>
        <button class="chip" @click="run('下划线转驼峰')"><span class="icon">→</span>下划线转驼峰</button>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符 · {{ resultStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip" :disabled="!result" @click="swap">→输入</button>
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
    inputStats() { return this.stats(this.inputData) },
    resultStats() { return this.stats(this.result) }
  },
  methods: {
    stats(s) {
      const t = s || '';
      return { chars: t.length, lines: t ? t.split('\n').length : 0 }
    },
    clearInput() {
      this.inputData = '';
      this.result = '';
    },
    swap() {
      const tmp = this.inputData;
      this.inputData = this.result;
      this.result = tmp;
    },
    async copyText() {
      if (!this.result) return;
      await CopyToClipboard(this.result);
      ElNotification({ title: '已复制', message: '结果已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
    run(val) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      if (val === '转json') {
        const arr = this.inputData.split('\n');
        this.result = JSON.stringify(arr, null, 4);
      } else if (val === '全部大写') {
        this.result = this.inputData.toUpperCase();
      } else if (val === '全部小写') {
        this.result = this.inputData.toLowerCase();
      } else if (val === '删除重复行') {
        this.result = [...new Set(this.inputData.split('\n'))].join('\n');
      } else if (val === '删除空行') {
        this.result = this.inputData.replace(/\n\n+/g, '\n');
      } else if (val === '去除不可见字符') {
        this.result = this.inputData.replace(/[\u200B-\u200D\uFEFF]/g, '');
      } else if (val === '下划线转驼峰') {
        this.result = this.inputData.split('_').map((w, i) =>
          i === 0 ? w : w.charAt(0).toUpperCase() + w.slice(1)
        ).join('');
      }
    }
  }
}
</script>
