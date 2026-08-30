<template>
  <div class="page gen-page">
    <div class="split">
      <div class="card">
        <div class="card-head">
          <div class="card-title">生成参数</div>
        </div>
        <div class="card-body">
          <div class="form-grid">
            <div class="field">
              <label class="label">应用名称 <span class="required">*</span></label>
              <input class="text-input" :class="{ invalid: attempted && !appName }" v-model="appName" placeholder="将写入 wails.json 的 name 和 outputfilename" />
            </div>

            <div class="field">
              <label class="label">窗口标题 <span class="required">*</span></label>
              <input class="text-input" :class="{ invalid: attempted && !title }" v-model="title" placeholder="将写入 wails.json 的 title" />
            </div>

            <div class="field">
              <label class="label">跳转链接 <span class="required">*</span></label>
              <input class="text-input" :class="{ invalid: attempted && !redirectURL }" v-model="redirectURL" placeholder="如 https://chat.deepseek.com/，将替换 index.html 中的跳转地址" />
            </div>

            <div class="field">
              <label class="label">输出目录</label>
              <input class="text-input" v-model="outputDir" placeholder="留空则导出到 D:\\" />
              <button class="chip" @click="onSelectOutputDir">选择目录</button>
            </div>

            <div class="field">
              <label class="label">图标文件</label>
              <input class="text-input" v-model="iconPath" placeholder="支持 ICO/PNG，PNG 会自动转换为 ICO" readonly />
              <button class="chip" @click="onSelectIcon">选择文件</button>
            </div>
          </div>
        </div>
      </div>

      <div class="toolbar">
        <span class="toolbar-label">操作</span>
        <button class="chip primary" :disabled="loading" @click="onBuild">
          <span v-if="loading">⏳ 构建中…</span>
          <span v-else>🚀 开始生成并打包</span>
        </button>
      </div>

      <div v-if="loading" class="card progress-card">
        <div class="progress-bar">
          <div class="progress-inner"></div>
        </div>
        <div class="progress-text">正在编译模板并打包，请稍候…</div>
      </div>

      <div v-if="resultMsg" class="card result-card" :class="{ success: resultSuccess, error: !resultSuccess }">
        <div class="result-icon">
          <span v-if="resultSuccess">✅</span>
          <span v-else>❌</span>
        </div>
        <div class="result-text">{{ resultMsg }}</div>
      </div>
    </div>
  </div>
</template>

<script>
import {
  BuildProject,
  SelectIconFile,
  SelectOutputDir,
} from '../wailsjs/go/app/App.js';
import { ElNotification } from 'element-plus';

export default {
  data() {
    return {
      appName: '',
      title: '',
      iconPath: '',
      outputDir: '',
      redirectURL: 'https://chat.deepseek.com/',
      loading: false,
      resultMsg: '',
      resultSuccess: false,
      attempted: false,
    };
  },
  computed: {
    requiredFields() {
      return [
        { key: 'appName', label: '应用名称' },
        { key: 'title', label: '窗口标题' },
        { key: 'redirectURL', label: '跳转链接' },
      ];
    },
    missingFields() {
      return this.requiredFields.filter(f => !this[f.key] || !String(this[f.key]).trim());
    },
  },
  methods: {
    async onSelectOutputDir() {
      const dir = await SelectOutputDir();
      if (dir) this.outputDir = dir;
    },
    async onSelectIcon() {
      const file = await SelectIconFile();
      if (file) this.iconPath = file;
    },
    async onBuild() {
      this.attempted = true;
      if (this.missingFields.length > 0) {
        const names = this.missingFields.map(f => f.label).join('、');
        ElNotification({
          title: '请填写必填项',
          message: `缺少：${names}`,
          position: 'bottom-right',
          type: 'warning',
        });
        return;
      }

      this.loading = true;
      this.resultMsg = '';
      this.resultSuccess = false;
      try {
        const res = await BuildProject({
          appName: this.appName,
          title: this.title,
          iconPath: this.iconPath,
          outputDir: this.outputDir,
          redirectURL: this.redirectURL,
        });
        this.resultSuccess = res.success;
        this.resultMsg = res.success ? `打包成功！输出：${this.outputDir || 'D:\\'}` : res.message;
        ElNotification({
          title: this.resultSuccess ? '完成' : '失败',
          message: this.resultMsg,
          position: 'bottom-right',
          type: this.resultSuccess ? 'success' : 'error',
        });
      } catch (e) {
        this.resultSuccess = false;
        this.resultMsg = '构建出错: ' + e;
        ElNotification({ title: '出错', message: this.resultMsg, position: 'bottom-right', type: 'error' });
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<style scoped>
.gen-page {
  padding: 0;
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
  width: 100%;
  max-width: 720px;
  padding: 4px 4px 8px;
}

.field {
  display: grid;
  grid-template-columns: 110px 1fr auto;
  align-items: center;
  gap: 12px;
}

.label > span.required {
  color: var(--error);
  margin-left: 2px;
}

.label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
}

.text-input {
  width: 100%;
  height: 36px;
  padding: 0 14px;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.text-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.text-input[readonly] {
  background: var(--bg-soft);
  color: var(--text-tertiary);
  cursor: pointer;
}

.text-input.invalid {
  border-color: var(--error);
  background: #FFF5F5;
}

.text-input.invalid:focus {
  border-color: var(--error);
  box-shadow: 0 0 0 3px rgba(245, 63, 63, 0.15);
}

.progress-card {
  padding: 18px 20px;
  gap: 8px;
}

.progress-bar {
  height: 6px;
  background: var(--bg-soft);
  border-radius: 3px;
  overflow: hidden;
  flex-shrink: 0;
}

.progress-inner {
  height: 100%;
  background: linear-gradient(90deg, var(--accent), #7B5BFF);
  border-radius: 3px;
  animation: progress-indeterminate 1.5s ease-in-out infinite;
}

.progress-text {
  font-size: 13px;
  color: var(--text-secondary);
}

.result-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 20px;
  flex-direction: row;
  flex: 0 0 auto;
}

.result-icon {
  font-size: 22px;
  line-height: 1;
  flex-shrink: 0;
}

.result-text {
  font-size: 13px;
  line-height: 1.7;
  word-break: break-all;
}

.result-card.success {
  background: #F0FFF4;
  border: 1px solid #C6F6D5;
  color: #276749;
}

.result-card.error {
  background: #FFF5F5;
  border: 1px solid #FED7D7;
  color: #C53030;
}

@keyframes progress-indeterminate {
  0% { width: 0%; margin-left: 0%; }
  50% { width: 60%; margin-left: 20%; }
  100% { width: 0%; margin-left: 100%; }
}
</style>