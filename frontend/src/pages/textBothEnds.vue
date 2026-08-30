<template>
  <div class="page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.chars }} 字符 · {{ inputStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip icon-only" @click="clearInput" title="清空">
              <el-icon><Delete /></el-icon>
            </button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入或粘贴文本（每行一项）"></textarea>
        </div>
      </div>

      <div class="toolbar-group">
        <div class="toolbar">
          <span class="toolbar-label">分组</span>
          <button class="chip" :class="{ primary: group === 'chars' }" @click="group = 'chars'">添加字符</button>
          <button class="chip" :class="{ primary: group === 'sort' }" @click="group = 'sort'">排序</button>
          <button class="chip" :class="{ primary: group === 'remove' }" @click="group = 'remove'">去除重复</button>
          <button class="chip" :class="{ primary: group === 'format' }" @click="group = 'format'">格式转换</button>
          <button class="chip" :class="{ primary: group === 'char' }" @click="group = 'char'">字符转换</button>
        </div>

        <div class="toolbar">
          <template v-if="group === 'chars'">
            <span class="toolbar-label">添加</span>
            <button class="chip" v-for="c in charList" :key="c" @click="addChar(c)">
              <span class="icon">'</span>{{ c }}<span class="icon">'</span>
            </button>
            <span class="toolbar-spacer"></span>
            <div class="inline-input">
              <input v-model="customChar" maxlength="3" placeholder="自定义" style="width: 60px;">
            </div>
            <button class="chip" @click="addChar(customChar)" :disabled="!customChar">
              <span class="icon">'</span>添加<span class="icon">'</span>
            </button>
          </template>

          <template v-else-if="group === 'sort'">
            <span class="toolbar-label">排序</span>
            <button class="chip" @click="forwardSort"><span class="icon">→</span>正向</button>
            <button class="chip" @click="reverseSort"><span class="icon">→</span>逆向</button>
            <button class="chip" @click="reverseLines"><span class="icon">⇓</span>倒序</button>
          </template>

          <template v-else-if="group === 'remove'">
            <span class="toolbar-label">参数</span>
            <label class="inline-input">替换字符
              <input v-model="replaceChar" maxlength="3" />
            </label>
            <button class="chip primary" @click="trimRepeatedChar"><span class="icon">✕</span>去除重复</button>
          </template>

          <template v-else-if="group === 'format'">
            <span class="toolbar-label">格式</span>
            <button class="chip primary" @click="toSqlQuery"><span class="icon">⊟</span>转SQL 条件</button>
            <span class="group-hint">每行一个值，合并为 SQL IN 条件</span>
          </template>

          <template v-else>
            <span class="toolbar-label">参数</span>
            <label class="inline-input">分割符
              <input v-model="splitChar" maxlength="3" />
            </label>
            <button class="chip primary" @click="toJSON"><span class="icon">{}</span>转JSON</button>
            <span class="group-hint">首行为表头，其它行为数据</span>
          </template>
        </div>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符 · {{ resultStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip icon-only" :disabled="!result" @click="exchange" title="→输入">
              <el-icon><Sort /></el-icon>
            </button>
            <button class="chip primary icon-only" :disabled="!result" @click="copyText" title="复制结果">
              <el-icon><DocumentCopy /></el-icon>
            </button>
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
import trimRepeated from 'trim-repeated';
import { CopyToClipboard, Text2Json } from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";
import { Delete, Sort, DocumentCopy } from "@element-plus/icons-vue";

export default {
  components: { Delete, Sort, DocumentCopy },
  data() {
    return {
      group: 'chars',
      inputData: '11111111\n22222222\n33333333333',
      result: '',
      charList: ["'", '"', '#', 'a'],
      customChar: '',
      replaceChar: '#',
      splitChar: ','
    }
  },
  computed: {
    inputStats() {
      const t = this.inputData || '';
      return { chars: t.length, lines: t ? t.split('\n').length : 0 };
    },
    resultStats() {
      const t = this.result || '';
      return { chars: t.length, lines: t ? t.split('\n').length : 0 };
    }
  },
  methods: {
    clearInput() { this.inputData = ''; this.result = ''; },
    exchange() {
      const tmp = this.inputData;
      this.inputData = this.result;
      this.result = tmp;
    },
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
    // 统一换行符为 \n，避免 Windows CRLF 影响按行处理
    normalizeEol() {
      return this.inputData.replace(/\r\n/g, '\n');
    },
    // 添加字符：每行首尾添加指定字符
    addChar(c) {
      if (!c) return;
      if (!this.needInput()) return;
      this.result = this.inputData.split('\n').map(item => c + item + c).join('\n');
    },
    // 正向排序（按中文拼音）
    forwardSort() {
      if (!this.needInput()) return;
      const arr = this.inputData.split('\n');
      arr.sort((a, b) => a.localeCompare(b, 'zh-Hans-CN', { sensitivity: 'base' }));
      this.result = arr.join('\n');
    },
    // 逆向排序
    reverseSort() {
      if (!this.needInput()) return;
      const arr = this.inputData.split('\n');
      arr.sort((a, b) => b.localeCompare(a, 'zh-Hans-CN', { sensitivity: 'base' }));
      this.result = arr.join('\n');
    },
    // 行倒序
    reverseLines() {
      if (!this.needInput()) return;
      this.result = this.inputData.split('\n').reverse().join('\n');
    },
    // 去除重复：连续重复内容替换为指定字符
    trimRepeatedChar() {
      if (!this.needInput()) return;
      if (!this.replaceChar) {
        ElNotification({ title: '参数为空', message: '请填写替换字符', position: 'bottom-right', type: 'error' });
        return;
      }
      this.result = trimRepeated(this.inputData, this.replaceChar);
    },
    // 格式转换：每行值合并为 SQL IN 条件，如 ('a', 'b', 'c')
    toSqlQuery() {
      if (!this.needInput()) return;
      const arr = this.normalizeEol().split('\n').filter(v => v.trim() !== '');
      if (!arr.length) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      this.result = "(" + arr.map(v => `'${v}'`).join(', ') + ")";
    },
    // 字符转换：首行表头 + 分隔符数据 → JSON 数组
    async toJSON() {
      if (!this.needInput()) return;
      if (!this.splitChar) {
        ElNotification({ title: '参数为空', message: '请填写分割符', position: 'bottom-right', type: 'error' });
        return;
      }
      const r = await Text2Json(this.normalizeEol(), this.splitChar);
      this.result = JSON.stringify(r);
    }
  }
}
</script>

<style scoped>
.group-hint {
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>
