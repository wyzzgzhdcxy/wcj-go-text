<template>
  <div class="page cat-page">
    <!-- 目录与参数 -->
    <div class="toolbar">
      <span class="toolbar-label">目标目录</span>
      <el-input
        v-model="dir"
        class="dir-input"
        placeholder="点击右侧按钮选择要归类的目录"
        readonly
        @click="selectDir"
      >
        <template #append>
          <el-button @click="selectDir">
            <el-icon class="btn-icon"><FolderOpened /></el-icon>选择目录
          </el-button>
        </template>
      </el-input>
      <button class="chip" :disabled="!dir" @click="openDir" title="在资源管理器中打开">
        <el-icon class="btn-icon"><Folder /></el-icon>打开目录
      </button>
      <span class="toolbar-spacer"></span>
      <label class="inline-input">获取前缀数量
        <input v-model.number="prefixCount" type="number" min="1" max="100" style="width: 56px;" />
      </label>
    </div>

    <!-- 分类前缀 -->
    <div class="card cat-input-card">
      <div class="card-head">
        <div class="card-title">分类前缀 <span class="badge">{{ categoryList.length }} 个分类</span></div>
        <div class="card-actions">
          <button class="chip" :disabled="!dir || loadingPrefixes" @click="fetchPrefixes">
            <el-icon class="btn-icon"><Refresh /></el-icon>{{ loadingPrefixes ? '获取中...' : '获取前缀' }}
          </button>
          <button class="chip" :disabled="!categories" @click="clearInput" title="清空">
            <el-icon class="btn-icon"><Delete /></el-icon>
          </button>
        </div>
      </div>
      <div class="card-body">
        <textarea
          v-model="categories"
          class="text-panel"
          placeholder="每行输入一个文件名前缀，也可用逗号分隔，例如：&#10;张宇&#10;张碧晨"
        ></textarea>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="toolbar">
      <button class="chip primary" :disabled="!canCategorize || categorizing" @click="categorize">
        <el-icon class="btn-icon"><Finished /></el-icon>{{ categorizing ? '归类中...' : '开始归类' }}
      </button>
      <span class="cat-hint">将把目录下以这些前缀开头的文件移动到同名子文件夹中</span>
      <span class="toolbar-spacer"></span>
      <span v-if="!dir" class="cat-hint warn">请先选择目录</span>
    </div>

    <!-- 归类结果 -->
    <div class="card cat-result-card">
      <div class="card-head">
        <div class="card-title">
          归类结果
          <span v-if="resultSummary" class="badge">{{ resultSummary }}</span>
        </div>
        <div class="card-actions" v-if="hasResult">
          <button class="chip" @click="openDir">
            <el-icon class="btn-icon"><Folder /></el-icon>打开目录
          </button>
        </div>
      </div>
      <div class="card-body cat-result-body">
        <div v-if="!hasResult" class="cat-empty">
          <el-icon class="cat-empty-icon"><Box /></el-icon>
          <div>尚无归类结果，选择目录并点击「开始归类」</div>
        </div>
        <div v-else class="cat-result-list">
          <div v-for="cat in resultRows" :key="cat.name" class="cat-row">
            <div class="cat-row-head">
              <el-icon class="btn-icon"><FolderOpened /></el-icon>
              <span class="cat-name">{{ cat.name }}/</span>
              <span class="badge" :class="{ zero: cat.count === 0 }">{{ cat.count }} 个文件</span>
            </div>
            <div v-if="cat.count" class="cat-files">{{ cat.files.join('、') }}</div>
            <div v-else class="cat-files none">未匹配到文件</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {CategorizeFiles, GetFilePrefixesInDir, SelectDirectory, OpenExplorer} from "../wailsjs/go/app/App.js";
import {ElMessageBox, ElNotification} from "element-plus";
import {Delete, Folder, FolderOpened, Refresh, Finished, Box} from "@element-plus/icons-vue";

export default {
  name: 'CategorizeFiles',
  components: {Delete, Folder, FolderOpened, Refresh, Finished, Box},
  data() {
    return {
      dir: '',
      prefixCount: 20,
      categories: '',
      categorizing: false,
      loadingPrefixes: false,
      result: null,
    }
  },
  computed: {
    categoryList() {
      return this.categories
        .split(/[\n\r,，、]+/)
        .map(s => s.trim())
        .filter(Boolean);
    },
    canCategorize() {
      return !!this.dir && this.categoryList.length > 0;
    },
    hasResult() {
      return !!this.result && Object.keys(this.result).length > 0;
    },
    resultRows() {
      if (!this.result) return [];
      return Object.keys(this.result)
        .sort()
        .map(name => ({name, files: this.result[name], count: (this.result[name] || []).length}));
    },
    resultSummary() {
      if (!this.hasResult) return '';
      const total = this.resultRows.reduce((n, r) => n + r.count, 0);
      return `${this.resultRows.length} 个分类 · 共移动 ${total} 个文件`;
    },
  },
  methods: {
    async selectDir() {
      const dir = await SelectDirectory();
      if (dir) {
        this.dir = dir;
      }
    },
    openDir() {
      if (this.dir) {
        OpenExplorer(this.dir);
      }
    },
    async fetchPrefixes() {
      this.loadingPrefixes = true;
      try {
        const prefixes = await GetFilePrefixesInDir(this.dir, this.prefixCount || 20);
        if (prefixes && prefixes.length) {
          this.categories = prefixes.join('\n');
          ElNotification({title: '获取前缀', message: `已获取 ${prefixes.length} 个前缀`, type: 'success', duration: 3000});
        } else {
          ElNotification({title: '获取前缀', message: '目录下未找到符合规则的前缀（文件名需包含“-”和中文）', type: 'warning', duration: 4000});
        }
      } catch (e) {
        ElNotification({title: '获取前缀失败', message: String(e), type: 'error', duration: 4000});
      } finally {
        this.loadingPrefixes = false;
      }
    },
    clearInput() {
      this.categories = '';
      this.result = null;
    },
    async categorize() {
      const count = this.categoryList.length;
      // 归类会移动文件，先确认
      try {
        await ElMessageBox.confirm(
          `将把「${this.dir}」下匹配 ${count} 个前缀的文件移动到同名子文件夹，是否继续？`,
          '确认归类',
          {confirmButtonText: '开始归类', cancelButtonText: '取消', type: 'warning'}
        );
      } catch (e) {
        return; // 用户取消
      }
      this.categorizing = true;
      try {
        this.result = await CategorizeFiles(this.dir, JSON.stringify(this.categoryList));
        const total = this.resultRows.reduce((n, r) => n + r.count, 0);
        ElNotification({
          title: '归类完成',
          message: `共移动 ${total} 个文件到 ${this.resultRows.length} 个分类文件夹`,
          type: 'success',
          duration: 4000,
        });
      } catch (e) {
        ElNotification({title: '归类失败', message: String(e), type: 'error', duration: 4000});
      } finally {
        this.categorizing = false;
      }
    },
  }
}
</script>

<style scoped>
.cat-page {
  padding: 10px 14px 14px;
  gap: 8px;
}

.dir-input {
  width: 420px;
  max-width: 55%;
}

.btn-icon {
  margin-right: 4px;
}

.cat-input-card {
  flex: 1 1 0;
}

.cat-result-card {
  flex: 0 0 auto;
  max-height: 45%;
}

.cat-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-left: 8px;
}

.cat-hint.warn {
  color: var(--warning);
}

.cat-result-body {
  overflow-y: auto;
}

.cat-empty {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 28px 0;
  color: var(--text-tertiary);
  font-size: 13px;
}

.cat-empty-icon {
  font-size: 34px;
  color: var(--text-tertiary);
  opacity: 0.6;
}

.cat-result-list {
  width: 100%;
  padding: 4px 14px 12px;
  overflow-y: auto;
}

.cat-row {
  padding: 8px 0;
  border-bottom: 1px solid var(--border-light);
}

.cat-row:last-child {
  border-bottom: none;
}

.cat-row-head {
  display: flex;
  align-items: center;
  gap: 6px;
}

.cat-row-head .btn-icon {
  color: var(--accent);
  margin-right: 0;
}

.cat-name {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-primary);
}

.badge.zero {
  background: var(--bg-soft);
  color: var(--text-tertiary);
}

.cat-files {
  margin: 4px 0 0 22px;
  font-size: 12.5px;
  color: var(--text-secondary);
  line-height: 1.7;
  word-break: break-all;
}

.cat-files.none {
  color: var(--text-tertiary);
}
</style>
