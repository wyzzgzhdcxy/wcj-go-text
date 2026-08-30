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
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入明文或密文（解密时按密文格式粘贴）"></textarea>
        </div>
      </div>

      <div class="toolbar-group">
        <div class="toolbar">
          <span class="toolbar-label">算法</span>
          <button v-for="a in algoList" :key="a.value" class="chip" :class="{ primary: algo === a.value }" @click="algo = a.value">
            {{ a.label }}
          </button>
          <span class="toolbar-spacer"></span>
          <span class="toolbar-label" title="密文格式：加密输出 / 解密输入">密文</span>
          <button class="chip" :class="{ primary: outEncoding === 'base64' }" @click="outEncoding = 'base64'">Base64</button>
          <button class="chip" :class="{ primary: outEncoding === 'hex' }" @click="outEncoding = 'hex'">HEX</button>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">模式</span>
          <button v-for="m in modeList" :key="m.value" class="chip" :class="{ primary: mode === m.value }" @click="mode = m.value" :title="m.tip">
            {{ m.label }}
          </button>
          <span class="toolbar-spacer"></span>
          <span class="toolbar-label" title="密钥/IV 文本的解析方式">密钥格式</span>
          <button v-for="k in keyEncList" :key="k.value" class="chip" :class="{ primary: keyEncoding === k.value }" @click="keyEncoding = k.value">
            {{ k.label }}
          </button>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">密钥</span>
          <div class="inline-input key-input">
            <input v-model="key" :placeholder="keyPlaceholder" spellcheck="false">
          </div>
          <button class="chip icon-only" @click="genKey" title="生成随机密钥和IV">
            <el-icon><MagicStick /></el-icon>
          </button>
          <template v-if="mode !== 'ECB'">
            <span class="toolbar-label">{{ mode === 'GCM' ? 'Nonce' : 'IV' }}</span>
            <div class="inline-input iv-input">
              <input v-model="iv" :placeholder="ivPlaceholder" spellcheck="false">
            </div>
          </template>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">操作</span>
          <button class="chip primary" @click="run('加密')"><span class="icon">🔒</span>加密</button>
          <button class="chip primary" @click="run('解密')"><span class="icon">🔓</span>解密</button>
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
          <textarea v-model="result" class="text-panel" placeholder="加密/解密结果会显示在这里" disabled></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CopyToClipboard, SymCrypto, SymRandomKey } from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";
import { Delete, Sort, DocumentCopy, MagicStick } from "@element-plus/icons-vue";

// 各算法的密钥长度（字节）
const KEY_LENS = { aes128: 16, aes192: 24, aes256: 32, des: 8, '3des': 24, sm4: 16 };

export default {
  components: { Delete, Sort, DocumentCopy, MagicStick },
  data() {
    return {
      algo: 'aes128',
      mode: 'CBC',
      outEncoding: 'base64',
      keyEncoding: 'text',
      key: '1234567890123456',
      iv: '1234567890123456',
      inputData: '',
      result: '',
      algoList: [
        { value: 'aes128', label: 'AES-128' },
        { value: 'aes192', label: 'AES-192' },
        { value: 'aes256', label: 'AES-256' },
        { value: 'des',    label: 'DES' },
        { value: '3des',   label: '3DES' },
        { value: 'sm4',    label: 'SM4' },
      ],
      modeList: [
        { value: 'ECB', label: 'ECB', tip: '电子密码本，不需要IV' },
        { value: 'CBC', label: 'CBC', tip: '密码分组链接，PKCS7填充' },
        { value: 'CTR', label: 'CTR', tip: '计数器模式（流模式）' },
        { value: 'GCM', label: 'GCM', tip: '认证加密（仅AES/SM4），自动附加认证标签' },
        { value: 'CFB', label: 'CFB', tip: '密码反馈模式（流模式）' },
        { value: 'OFB', label: 'OFB', tip: '输出反馈模式（流模式）' },
      ],
      keyEncList: [
        { value: 'text',   label: '文本' },
        { value: 'hex',    label: 'HEX' },
        { value: 'base64', label: 'Base64' },
      ],
    }
  },
  computed: {
    inputStats() { return this.stats(this.inputData) },
    resultStats() { return this.stats(this.result) },
    keyLen() { return KEY_LENS[this.algo] || 16 },
    keyPlaceholder() {
      const base = this.keyLen + ' 字节，不足自动补0';
      if (this.keyEncoding === 'hex') return '密钥HEX（' + this.keyLen * 2 + '位）';
      if (this.keyEncoding === 'base64') return '密钥Base64';
      return '密钥（' + base + '）';
    },
    ivPlaceholder() {
      const len = this.mode === 'GCM' ? 12 : (this.algo === 'des' || this.algo === '3des' ? 8 : 16);
      if (this.keyEncoding === 'hex') return (this.mode === 'GCM' ? 'Nonce' : 'IV') + 'HEX（' + len * 2 + '位）';
      return (this.mode === 'GCM' ? 'Nonce' : 'IV') + '（' + len + ' 字节）';
    },
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
    async genKey() {
      try {
        const r = await SymRandomKey(this.algo, this.mode, this.keyEncoding);
        this.key = r.key;
        this.iv = r.iv;
        ElNotification({ title: '已生成', message: '已生成随机密钥' + (r.iv ? '和IV' : ''), position: 'bottom-right', type: 'success' });
      } catch (e) {
        ElNotification({ title: '生成失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async run(op) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      if (!this.key) {
        ElNotification({ title: '密钥为空', message: '请输入密钥', position: 'bottom-right', type: 'error' });
        return;
      }
      if (op === '解密' && this.mode !== 'ECB' && !this.iv) {
        ElNotification({ title: 'IV为空', message: '请输入' + (this.mode === 'GCM' ? 'Nonce' : 'IV'), position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        this.result = await SymCrypto(this.inputData, this.algo, this.mode, this.key, this.iv, this.keyEncoding, this.outEncoding, op);
        ElNotification({ title: op + '成功', message: '', position: 'bottom-right', type: 'success' });
      } catch (e) {
        this.result = '';
        ElNotification({ title: op + '失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
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
.inline-input.key-input input {
  width: 220px;
  font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
}
.inline-input.iv-input input {
  width: 160px;
  font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
}
</style>
