<template>
  <div class="page">
    <div class="toolbar">
      <span class="toolbar-label">算法</span>
      <button class="chip" :class="{ primary: algo === 'rsa' }" @click="switchAlgo('rsa')">RSA</button>
      <button class="chip" :class="{ primary: algo === 'sm2' }" @click="switchAlgo('sm2')">SM2</button>
      <span class="toolbar-spacer"></span>
      <span class="toolbar-label">功能</span>
      <button class="chip" :class="{ primary: mode === 'gen' }" @click="switchMode('gen')">生成密钥</button>
      <button class="chip" :class="{ primary: mode === 'text' }" @click="switchMode('text')">文本加解密</button>
      <button class="chip" :class="{ primary: mode === 'file' }" @click="switchMode('file')">文件加解密</button>
      <button class="chip" :class="{ primary: mode === 'sign' }" @click="switchMode('sign')">签名验签</button>
    </div>

    <!-- 生成密钥 -->
    <div v-if="mode === 'gen'" class="split">
      <div class="toolbar">
        <span class="toolbar-label">密钥位数</span>
        <template v-if="algo === 'rsa'">
          <button v-for="b in rsaBits" :key="b" class="chip" :class="{ primary: keyBits === b }" @click="keyBits = b">
            {{ b }}
          </button>
          <span class="gen-hint">位数越高越安全，生成越慢</span>
        </template>
        <template v-else>
          <span class="chip disabled-chip">256</span>
          <span class="gen-hint">SM2 固定 256 位（国密标准）</span>
        </template>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">生成{{ algoName }}密钥对 <span class="badge">{{ algoName }}</span></div>
        </div>
        <div class="gen-body">
          <div class="gen-desc">
            {{ algo === 'rsa' ? '生成 RSA 密钥对（PKCS1 公钥 / PKCS1 私钥，PEM 格式）' : '生成 SM2 国密密钥对（PKCS8 私钥 / 公钥，PEM 格式）' }}，
            公钥用于加密，私钥用于解密。文件保存在密钥目录：
          </div>
          <div class="gen-path">{{ keysDir }}</div>
          <div class="gen-actions">
            <button class="chip primary" :disabled="generating" @click="genKey">
              <span class="icon">{{ generating ? '⏳' : '🔑' }}</span>{{ generating ? '生成中...' : '生成密钥' }}
            </button>
            <button class="chip" @click="openKeysDir"><span class="icon">📂</span>打开目录</button>
            <span class="gen-hint">{{ algo === 'rsa' ? 'rsa_public.pem / rsa_private.pem' : 'sm2_public.pem / sm2_private.pem' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 文本加解密 -->
    <div v-else-if="mode === 'text'" class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">输入 <span class="badge">{{ inputStats.chars }} 字符</span></div>
          <div class="card-actions">
            <button class="chip icon-only" @click="clearText" title="清空">
              <el-icon><Delete /></el-icon>
            </button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入明文或密文（解密时粘贴 Base64 密文）"></textarea>
        </div>
      </div>

      <div class="toolbar-group">
        <div class="toolbar">
          <span class="toolbar-label">公钥</span>
          <div class="inline-input key-path-input">
            <input v-model="pubPath" :placeholder="algoName + '公钥文件路径'" spellcheck="false">
          </div>
          <button class="chip" @click="pickKey('pub')">选择</button>
          <span class="toolbar-label">私钥</span>
          <div class="inline-input key-path-input">
            <input v-model="priPath" :placeholder="algoName + '私钥文件路径'" spellcheck="false">
          </div>
          <button class="chip" @click="pickKey('pri')">选择</button>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">操作</span>
          <button class="chip primary" @click="runText('加密')"><span class="icon">🔒</span>加密</button>
          <button class="chip primary" @click="runText('解密')"><span class="icon">🔓</span>解密</button>
          <span class="gen-hint">加密用公钥，解密用私钥</span>
        </div>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">结果 <span class="badge">{{ resultStats.chars }} 字符</span></div>
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

    <!-- 文件加解密 -->
    <div v-else-if="mode === 'file'" class="split">
      <div class="toolbar">
        <span class="toolbar-label">待处理文件</span>
        <div class="inline-input key-path-input wide">
          <input v-model="filePath" placeholder="选择要加密/解密的文件" spellcheck="false">
        </div>
        <button class="chip" @click="pickFile">选择文件</button>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">操作</span>
        <button class="chip primary" @click="encryptFile"><span class="icon">🔒</span>加密文件</button>
        <button class="chip primary" @click="decryptFile"><span class="icon">🔓</span>解密文件</button>
        <span class="gen-hint">采用数字信封：AES-256-GCM 加密内容，{{ algoName }}公钥加密会话密钥</span>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">处理结果 <span class="badge">{{ fileResult ? '完成' : '待处理' }}</span></div>
          <div class="card-actions">
            <button class="chip icon-only" :disabled="!fileResult" @click="openFileDir" title="打开所在目录">
              <el-icon><FolderOpened /></el-icon>
            </button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="fileResult" class="text-panel" placeholder="加密输出为 原文件名.enc；解密自动识别算法并还原原文件" disabled></textarea>
        </div>
      </div>
    </div>

    <!-- 签名验签 -->
    <div v-else class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">原文 <span class="badge">{{ inputStats.chars }} 字符</span></div>
          <div class="card-actions">
            <button class="chip icon-only" @click="clearSign" title="清空">
              <el-icon><Delete /></el-icon>
            </button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="inputData" class="text-panel" placeholder="在此输入待签名的原文（验签时需与签名时的原文完全一致）"></textarea>
        </div>
      </div>

      <div class="toolbar-group">
        <div class="toolbar">
          <span class="toolbar-label">私钥</span>
          <div class="inline-input key-path-input">
            <input v-model="priPath" :placeholder="algoName + '私钥文件路径'" spellcheck="false">
          </div>
          <button class="chip" @click="pickKey('pri')">选择</button>
          <span class="toolbar-label">公钥</span>
          <div class="inline-input key-path-input">
            <input v-model="pubPath" :placeholder="algoName + '公钥文件路径'" spellcheck="false">
          </div>
          <button class="chip" @click="pickKey('pub')">选择</button>
        </div>

        <div class="toolbar">
          <span class="toolbar-label">操作</span>
          <button class="chip primary" @click="signData"><span class="icon">✍️</span>签名</button>
          <button class="chip primary" @click="verifyData"><span class="icon">✅</span>验签</button>
          <span class="gen-hint">{{ algo === 'rsa' ? 'SHA256 + PKCS1v15' : 'SM3 + 默认UID（标准SM2签名）' }}，签名用私钥，验签用公钥</span>
        </div>
      </div>

      <div class="card">
        <div class="card-head">
          <div class="card-title">签名值 <span class="badge" :class="{ 'verify-pass': verifyStatus === 'pass', 'verify-fail': verifyStatus === 'fail' }">{{ verifyBadge }}</span></div>
          <div class="card-actions">
            <button class="chip primary icon-only" :disabled="!signature" @click="copySignature" title="复制签名">
              <el-icon><DocumentCopy /></el-icon>
            </button>
          </div>
        </div>
        <div class="card-body">
          <textarea v-model="signature" class="text-panel" placeholder="签名后此处显示 Base64 签名值；验签时粘贴待验证的签名" ></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {
  CopyToClipboard, GetKeysDir, SelectFile, OpenExplorer,
  GenerateKey, Sm2GenerateKey,
  RsaCryptoStr, RsaDeCryptoStr, Sm2CryptoStr, Sm2DeCryptoStr,
  AsymEncryptFile, AsymDecryptFile,
  RsaSignStr, RsaVerifyStr, Sm2SignStr, Sm2VerifyStr,
} from "../wailsjs/go/app/App.js";
import { ElNotification } from "element-plus";
import { Delete, Sort, DocumentCopy, FolderOpened } from "@element-plus/icons-vue";

export default {
  components: { Delete, Sort, DocumentCopy, FolderOpened },
  data() {
    return {
      algo: 'rsa',
      mode: 'text',
      keyBits: 2048,
      rsaBits: [1024, 2048, 3072, 4096],
      keysDir: '',
      pubPath: '',
      priPath: '',
      filePath: '',
      inputData: '',
      result: '',
      fileResult: '',
      signature: '',
      verifyStatus: '',
      generating: false,
    }
  },
  computed: {
    inputStats() { return { chars: (this.inputData || '').length } },
    resultStats() { return { chars: (this.result || '').length } },
    algoName() { return this.algo === 'rsa' ? 'RSA' : 'SM2' },
    verifyBadge() {
      if (this.verifyStatus === 'pass') return '验签通过';
      if (this.verifyStatus === 'fail') return '验签失败';
      return this.signature ? '待验证' : '未签名';
    },
  },
  async mounted() {
    try {
      this.keysDir = await GetKeysDir();
    } catch (e) {
      this.keysDir = '';
    }
    this.applyDefaultKeyPaths();
  },
  methods: {
    switchAlgo(algo) {
      this.algo = algo;
      this.applyDefaultKeyPaths();
    },
    switchMode(mode) {
      this.mode = mode;
      this.verifyStatus = '';
    },
    clearSign() {
      this.inputData = '';
      this.signature = '';
      this.verifyStatus = '';
    },
    applyDefaultKeyPaths() {
      const base = (this.keysDir || '').replace(/[\\/]+$/, '');
      const sep = base.includes('\\') || /^[A-Za-z]:/.test(base) ? '\\' : '/';
      this.pubPath = base + sep + this.algo + '_public.pem';
      this.priPath = base + sep + this.algo + '_private.pem';
    },
    async genKey() {
      this.generating = true;
      try {
        let dir;
        if (this.algo === 'rsa') {
          dir = await GenerateKey(this.keyBits);
        } else {
          dir = await Sm2GenerateKey();
        }
        this.keysDir = dir;
        this.applyDefaultKeyPaths();
        ElNotification({ title: '生成成功', message: this.algoName + ' 密钥已保存到 ' + dir, position: 'bottom-right', type: 'success' });
      } catch (e) {
        ElNotification({ title: '生成失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      } finally {
        this.generating = false;
      }
    },
    openKeysDir() {
      if (this.keysDir) OpenExplorer(this.keysDir);
    },
    openFileDir() {
      if (this.fileResult) {
        const idx = Math.max(this.fileResult.lastIndexOf('\\'), this.fileResult.lastIndexOf('/'));
        OpenExplorer(idx > 0 ? this.fileResult.slice(0, idx) : '.');
      }
    },
    async pickKey(which) {
      try {
        const path = await SelectFile();
        if (!path) return;
        if (which === 'pub') this.pubPath = path;
        else this.priPath = path;
      } catch (e) {
        ElNotification({ title: '文件选择失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async pickFile() {
      try {
        const path = await SelectFile();
        if (path) this.filePath = path;
      } catch (e) {
        ElNotification({ title: '文件选择失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    clearText() {
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
    async runText(op) {
      if (!this.inputData) {
        ElNotification({ title: '数据为空', message: '请先输入文本', position: 'bottom-right', type: 'error' });
        return;
      }
      if (op === '加密' && !this.pubPath) {
        ElNotification({ title: '缺少公钥', message: '请先选择或生成公钥', position: 'bottom-right', type: 'error' });
        return;
      }
      if (op === '解密' && !this.priPath) {
        ElNotification({ title: '缺少私钥', message: '请先选择或生成私钥', position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        if (op === '加密') {
          this.result = this.algo === 'rsa'
            ? await RsaCryptoStr(this.pubPath, this.inputData)
            : await Sm2CryptoStr(this.pubPath, this.inputData);
        } else {
          this.result = this.algo === 'rsa'
            ? await RsaDeCryptoStr(this.priPath, this.inputData)
            : await Sm2DeCryptoStr(this.priPath, this.inputData);
        }
        ElNotification({ title: op + '成功', message: this.algoName, position: 'bottom-right', type: 'success' });
      } catch (e) {
        this.result = '';
        ElNotification({ title: op + '失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async encryptFile() {
      if (!this.filePath) {
        ElNotification({ title: '未选择文件', message: '请先选择待加密文件', position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        this.fileResult = await AsymEncryptFile(this.algo, this.pubPath, this.filePath);
        ElNotification({ title: '加密成功', message: this.fileResult, position: 'bottom-right', type: 'success' });
      } catch (e) {
        this.fileResult = '';
        ElNotification({ title: '加密失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async decryptFile() {
      if (!this.filePath) {
        ElNotification({ title: '未选择文件', message: '请先选择待解密文件', position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        this.fileResult = await AsymDecryptFile(this.priPath, this.filePath);
        ElNotification({ title: '解密成功', message: this.fileResult, position: 'bottom-right', type: 'success' });
      } catch (e) {
        this.fileResult = '';
        ElNotification({ title: '解密失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async signData() {
      if (!this.inputData) {
        ElNotification({ title: '原文为空', message: '请先输入待签名内容', position: 'bottom-right', type: 'error' });
        return;
      }
      if (!this.priPath) {
        ElNotification({ title: '缺少私钥', message: '请先选择或生成私钥', position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        this.signature = this.algo === 'rsa'
          ? await RsaSignStr(this.priPath, this.inputData)
          : await Sm2SignStr(this.priPath, this.inputData);
        this.verifyStatus = '';
        ElNotification({ title: '签名成功', message: this.algoName + ' 签名已生成', position: 'bottom-right', type: 'success' });
      } catch (e) {
        this.signature = '';
        this.verifyStatus = '';
        ElNotification({ title: '签名失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async verifyData() {
      if (!this.inputData) {
        ElNotification({ title: '原文为空', message: '请先输入待验证的原文', position: 'bottom-right', type: 'error' });
        return;
      }
      if (!this.signature) {
        ElNotification({ title: '签名为空', message: '请先签名或粘贴待验证的签名值', position: 'bottom-right', type: 'error' });
        return;
      }
      if (!this.pubPath) {
        ElNotification({ title: '缺少公钥', message: '请先选择或生成公钥', position: 'bottom-right', type: 'error' });
        return;
      }
      try {
        const ok = this.algo === 'rsa'
          ? await RsaVerifyStr(this.pubPath, this.inputData, this.signature)
          : await Sm2VerifyStr(this.pubPath, this.inputData, this.signature);
        this.verifyStatus = ok ? 'pass' : 'fail';
        if (ok) {
          ElNotification({ title: '验签通过', message: '签名与原文匹配（' + this.algoName + '）', position: 'bottom-right', type: 'success' });
        } else {
          ElNotification({ title: '验签失败', message: '签名与原文或公钥不匹配', position: 'bottom-right', type: 'error' });
        }
      } catch (e) {
        this.verifyStatus = 'fail';
        ElNotification({ title: '验签失败', message: String(e).replace(/^Error:\s*/, ''), position: 'bottom-right', type: 'error' });
      }
    },
    async copySignature() {
      if (!this.signature) return;
      await CopyToClipboard(this.signature);
      ElNotification({ title: '已复制', message: '签名已复制到剪贴板', position: 'bottom-right', type: 'success' });
    },
  }
}
</script>

<style scoped>
.key-path-input input {
  width: 280px;
  font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
}
.key-path-input.wide input {
  width: 460px;
}
.gen-hint {
  font-size: 12px;
  color: var(--text-tertiary);
}
.gen-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0 14px 14px;
}
.gen-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}
.gen-path {
  font-size: 12.5px;
  color: var(--accent);
  font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
  background: var(--bg-soft);
  border-radius: var(--radius-md);
  padding: 8px 12px;
  word-break: break-all;
}
.gen-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.disabled-chip {
  opacity: 0.7;
  cursor: default;
}
.disabled-chip:hover {
  background: var(--bg-soft);
  color: var(--text-secondary);
}
.badge.verify-pass {
  background: rgba(103, 194, 58, 0.15);
  color: #67c23a;
}
.badge.verify-fail {
  background: rgba(245, 108, 108, 0.15);
  color: #f56c6c;
}
</style>
