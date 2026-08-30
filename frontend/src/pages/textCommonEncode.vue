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
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入或粘贴文本"></textarea>
        </div>
      </div>

      <div class="toolbar-group">
        <div class="toolbar">
          <span class="toolbar-label">分组</span>
          <button class="chip" :class="{ primary: group === 'common' }" @click="group = 'common'">常用</button>
          <button class="chip" :class="{ primary: group === 'hash' }" @click="group = 'hash'">哈希</button>
          <button class="chip" :class="{ primary: group === 'encode' }" @click="group = 'encode'">编码</button>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">操作</span>

        <template v-if="group === 'common'">
          <button class="chip" @click="run('转json')"><span class="icon">{}</span>转JSON</button>
          <button class="chip" @click="run('全部大写')"><span class="icon">Aa</span>全大写</button>
          <button class="chip" @click="run('全部小写')"><span class="icon">aa</span>全小写</button>
          <button class="chip" @click="run('删除重复行')"><span class="icon">⊟</span>去重</button>
          <button class="chip" @click="run('删除空行')"><span class="icon">∕</span>去空行</button>
          <button class="chip" @click="run('去除不可见字符')"><span class="icon">⎕</span>去不可见</button>
          <button class="chip" @click="run('下划线转驼峰')"><span class="icon">→</span>下划线转驼峰</button>
          <button class="chip" @click="run('驼峰转下划线')"><span class="icon">←</span>驼峰转下划线</button>
          <button class="chip" @click="run('反转字符串')"><span class="icon">↔</span>反转</button>
          <button class="chip" @click="run('去首尾空白')"><span class="icon">⇥</span>去首尾空白</button>
          <button class="chip" @click="run('压缩空格')"><span class="icon">␣</span>压缩空格</button>
        </template>

        <template v-else-if="group === 'hash'">
          <button class="chip" @click="run('md5')">MD5 摘要</button>
          <button class="chip" @click="run('sm3')">SM3</button>
          <button class="chip" @click="run('crc32')">CRC32</button>
          <button class="chip" @click="run('sha1')">SHA1</button>
          <button class="chip" @click="run('sha224')">SHA224</button>
          <button class="chip" @click="run('sha256')">SHA256</button>
          <button class="chip" @click="run('sha384')">SHA384</button>
          <button class="chip" @click="run('sha512')">SHA512</button>
        </template>

        <template v-else>
          <button class="chip" @click="run('url编码')">URL 编码</button>
          <button class="chip" @click="run('url解码')">URL 解码</button>
          <button class="chip" @click="run('base64编码')">Base64 编码</button>
          <button class="chip" @click="run('base64解码')">Base64 解码</button>
          <button class="chip" @click="run('base32编码')">Base32 编码</button>
          <button class="chip" @click="run('base32解码')">Base32 解码</button>
          <button class="chip" @click="run('hex编码')">HEX 编码</button>
          <button class="chip" @click="run('hex解码')">HEX 解码</button>
          <button class="chip" @click="run('unicode编码')">Unicode 编码</button>
          <button class="chip" @click="run('unicode解码')">Unicode 解码</button>
          <button class="chip" @click="run('html编码')">HTML 编码</button>
          <button class="chip" @click="run('html解码')">HTML 解码</button>
          <button class="chip" @click="run('二进制编码')">二进制 编码</button>
          <button class="chip" @click="run('二进制解码')">二进制 解码</button>
          <button class="chip" @click="run('ascii编码')">ASCII 编码</button>
          <button class="chip" @click="run('ascii解码')">ASCII 解码</button>
        </template>
        </div>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符 · {{ resultStats.lines }} 行</span></div>
          <div class="card-actions">
            <button class="chip icon-only" :disabled="!result" @click="swap" title="→输入">
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
import { CopyToClipboard, TextEncode } from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";
import { Delete, Sort, DocumentCopy } from "@element-plus/icons-vue";

export default {
  components: { Delete, Sort, DocumentCopy },
  data() {
    return {
      group: 'common',
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
    // 统一换行符为 \n，避免 Windows CRLF 导致按行处理失效
    normalizeEol() {
      return this.inputData.replace(/\r\n/g, '\n');
    },
    async run(val) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      if (val === '转json') {
        const norm = this.normalizeEol();
        // 末尾换行视为行终止符而不是空行
        const lines = norm.endsWith('\n') ? norm.slice(0, -1).split('\n') : norm.split('\n');
        this.result = JSON.stringify(lines, null, 4);
        return;
      }
      if (val === '全部大写') {
        this.result = this.inputData.toUpperCase();
        return;
      }
      if (val === '全部小写') {
        this.result = this.inputData.toLowerCase();
        return;
      }
      if (val === '删除重复行') {
        this.result = [...new Set(this.normalizeEol().split('\n'))].join('\n');
        return;
      }
      if (val === '删除空行') {
        // 空行及只含空白字符的行都移除
        this.result = this.normalizeEol().split('\n').filter(l => l.trim() !== '').join('\n');
        return;
      }
      if (val === '去除不可见字符') {
        // 零宽字符、软连字符、BOM、词连接符等
        this.result = this.inputData.replace(/[\u200B-\u200F\u2060\u2061-\u2064\uFEFF\u00AD\u180E\u202A-\u202E]/g, '');
        return;
      }
      if (val === '下划线转驼峰') {
        const s = this.inputData.replace(/^_+|_+$/g, '');
        this.result = s.split(/_+/).map((w, i) =>
          i === 0 ? w : (w ? w.charAt(0).toUpperCase() + w.slice(1) : w)
        ).join('');
        return;
      }
      if (val === '去首尾空白') {
        this.result = this.inputData.trim();
        return;
      }
      this.result = await TextEncode(this.inputData, val);
    }
  }
}
</script>

<style scoped>
.toolbar-group {
  display: flex;
  flex-direction: column;
  gap: 0;
  flex-shrink: 0;
}
.toolbar-group .toolbar {
  border-radius: 0;
  box-shadow: none;
}
.toolbar-group .toolbar:first-child {
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}
.toolbar-group .toolbar:last-child {
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}
.chip.icon-only {
  padding: 6px 8px;
}
.chip.icon-only .el-icon {
  font-size: 15px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>
