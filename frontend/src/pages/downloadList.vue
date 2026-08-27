<style scoped>
.download-list-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  box-sizing: border-box;
  overflow: hidden;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: 600;
}

.task-list {
  flex: 1;
  overflow: auto;
}

.task-item {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.task-info {
  flex: 1;
}

.task-url {
  font-size: 14px;
  color: #333;
  word-break: break-all;
  margin-bottom: 8px;
}

.task-meta {
  font-size: 12px;
  color: #999;
}

.task-status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  margin-right: 8px;
}

.status-pending {
  background: #e6e6e6;
  color: #666;
}

.status-downloading {
  background: #409eff;
  color: #fff;
}

.status-completed {
  background: #67c23a;
  color: #fff;
}

.status-failed {
  background: #f56c6c;
  color: #fff;
}

.status-cancelled {
  background: #909399;
  color: #fff;
}

.platform-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: #f0f0f0;
  color: #666;
  margin-right: 8px;
}

.platform-youtube {
  background: #ff0000;
  color: #fff;
}

.platform-bilibili {
  background: #23ade5;
  color: #fff;
}

.progress-section {
  margin-bottom: 12px;
}

.progress-bar-wrapper {
  height: 8px;
  background: #f0f2f5;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 4px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #409eff, #66b1ff);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
}

.task-actions {
  display: flex;
  gap: 8px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

/* 添加任务对话框样式 */
.add-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

.add-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.url-section {
  display: flex;
  gap: 10px;
}

.url-input {
  flex: 1;
}

.format-output {
  min-height: 220px;
  max-height: 360px;
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.format-output.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.download-section {
  display: flex;
  gap: 10px;
  align-items: center;
}

.format-input {
  width: 200px;
}

.path-section {
  display: flex;
  gap: 8px;
  align-items: center;
  flex: 1;
}

.path-input {
  flex: 1;
}
</style>

<template>
  <div class="download-list-wrapper">
    <div class="header">
      <span class="title">下载列表</span>
      <div class="header-actions">
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon>
          <span>添加任务</span>
        </el-button>
        <el-button size="small" @click="refreshTasks">刷新</el-button>
      </div>
    </div>

    <div class="task-list" v-if="tasks.length > 0">
      <div class="task-item" v-for="task in tasks" :key="task.id">
        <div class="task-header">
          <div class="task-info">
            <div class="task-url">{{ task.url }}</div>
            <div class="task-meta">
              <span class="platform-badge" :class="'platform-' + task.platform">
                {{ task.platform === 'youtube' ? 'YouTube' : 'B站' }}
              </span>
              <span class="task-status" :class="'status-' + task.status">
                {{ statusText(task.status) }}
              </span>
              <span>{{ task.format_id }}</span>
            </div>
          </div>
          <div class="task-actions">
            <el-button
              v-if="task.status === 'downloading'"
              size="small"
              type="danger"
              @click="cancelTask(task.id)"
            >
              取消
            </el-button>
            <el-button
              v-if="task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled'"
              size="small"
              type="text"
              @click="removeTask(task.id)"
            >
              删除
            </el-button>
          </div>
        </div>

        <div class="progress-section" v-if="task.status === 'downloading'">
          <div class="progress-bar-wrapper">
            <div class="progress-bar" :style="{ width: task.progress + '%' }"></div>
          </div>
          <div class="progress-info">
            <span>{{ task.speed || '下载中...' }}</span>
            <span>{{ task.progress.toFixed(1) }}%</span>
            <span v-if="task.eta">ETA: {{ task.eta }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div style="text-align: center;">
        <div style="font-size: 48px; margin-bottom: 16px;">📥</div>
        <div>暂无下载任务</div>
      </div>
    </div>

    <!-- 添加任务对话框 -->
    <el-dialog
      v-model="addDialogVisible"
      title="添加下载任务"
      width="780px"
      class="add-dialog"
      :close-on-click-modal="false"
    >
      <div class="add-content">
        <!-- URL Section -->
        <div class="url-section">
          <el-input
            v-model="addForm.url"
            class="url-input"
            placeholder="输入B站或YouTube视频URL"
            @keyup.enter="listFormats"
          >
            <template #prepend>视频URL</template>
          </el-input>
          <el-button type="primary" @click="listFormats" :loading="isLoading">
            {{ isLoading ? '获取中...' : '获取格式' }}
          </el-button>
        </div>

        <!-- Format Output -->
        <div
          v-if="formatOutput || isLoading"
          class="format-output"
        >
          {{ formatOutput }}
        </div>
        <div v-else class="format-output empty">
          <div style="text-align: center;">
            <div style="font-size: 36px; margin-bottom: 8px;">🎬</div>
            <div>输入视频URL获取可用格式</div>
          </div>
        </div>

        <!-- Download Section -->
        <div class="download-section">
          <el-input
            v-model="addForm.formatId"
            class="format-input"
            placeholder="格式ID，如: 30280"
          />
          <div class="path-section">
            <el-input v-model="addForm.downloadPath" class="path-input" readonly />
            <el-button @click="selectDownloadPath">浏览</el-button>
          </div>
          <el-button
            type="primary"
            @click="submitDownload"
            :disabled="!canAdd"
          >
            添加到下载列表
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import {
  GetDownloadTasks,
  CancelDownloadTask,
  RemoveDownloadTask,
  BilibiliListFormats,
  BilibiliDownload,
  GetYtDlpDownloadPath,
  SelectDirectory,
} from "../wailsjs/go/app/App.js";
import { ElMessage } from "element-plus";
import { Plus } from "@element-plus/icons-vue";

export default {
  name: "DownloadList",
  components: { Plus },
  data() {
    return {
      tasks: [],
      refreshTimer: null,

      addDialogVisible: false,
      addForm: {
        url: "",
        formatId: "30080+30280",
        downloadPath: "",
      },
      isLoading: false,
      formatOutput: "",
    };
  },
  computed: {
    canAdd() {
      return this.addForm.url && this.addForm.formatId && this.addForm.downloadPath;
    },
  },
  mounted() {
    this.refreshTasks();
    this.refreshTimer = setInterval(() => {
      this.refreshTasks();
    }, 1000);
  },
  beforeUnmount() {
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
    }
  },
  methods: {
    async refreshTasks() {
      try {
        const result = await GetDownloadTasks();
        if (result) {
          this.tasks = JSON.parse(result);
        }
      } catch (e) {
        console.error("Failed to get tasks:", e);
      }
    },

    async cancelTask(taskID) {
      try {
        await CancelDownloadTask(taskID);
        this.refreshTasks();
      } catch (e) {
        console.error("Cancel task error:", e);
      }
    },

    async removeTask(taskID) {
      try {
        await RemoveDownloadTask(taskID);
        this.refreshTasks();
      } catch (e) {
        console.error("Remove task error:", e);
      }
    },

    statusText(status) {
      const statusMap = {
        pending: "等待中",
        downloading: "下载中",
        completed: "已完成",
        failed: "失败",
        cancelled: "已取消",
      };
      return statusMap[status] || status;
    },

    async openAddDialog() {
      this.addDialogVisible = true;
      this.addForm = {
        url: "",
        formatId: "30080+30280",
        downloadPath: "",
      };
      this.formatOutput = "";
      try {
        this.addForm.downloadPath = await GetYtDlpDownloadPath();
      } catch (e) {
        console.error("Failed to get download path:", e);
      }
    },

    async listFormats() {
      if (!this.addForm.url) {
        ElMessage.warning("请输入B站或YouTube视频URL");
        return;
      }

      this.isLoading = true;
      this.formatOutput = "";

      try {
        const result = await BilibiliListFormats(this.addForm.url);
        if (result) {
          this.formatOutput = result;
          const isBilibili = /bilibili\.com|b23\.tv/.test(this.addForm.url);
          if (isBilibili) {
            const lines = result.trim().split("\n");
            const lastLines = lines.slice(-5);
            let maxSize = 0;
            let maxId = 0;
            for (const line of lastLines) {
              const parts = line.split("|");
              if (parts.length >= 2) {
                const sizeStr = parts[1].trim().split(/\s+/)[0];
                const idMatch = line.match(/^(\d+)/);
                if (idMatch && sizeStr) {
                  let sizeInMiB = 0;
                  if (sizeStr.includes("GiB")) {
                    sizeInMiB = parseFloat(sizeStr) * 1024;
                  } else if (sizeStr.includes("MiB")) {
                    sizeInMiB = parseFloat(sizeStr);
                  }
                  if (sizeInMiB > maxSize) {
                    maxSize = sizeInMiB;
                    maxId = parseInt(idMatch[1], 10);
                  }
                }
              }
            }
            if (maxId > 0) {
              this.addForm.formatId = maxId + "+30280";
            }
          }
        } else {
          ElMessage.error("获取格式列表失败，请检查URL或cookies文件");
        }
      } catch (e) {
        console.error("List formats error:", e);
        ElMessage.error("获取格式失败: " + e.message);
      } finally {
        this.isLoading = false;
      }
    },

    async selectDownloadPath() {
      const dir = await SelectDirectory();
      if (dir) {
        this.addForm.downloadPath = dir;
      }
    },

    async submitDownload() {
      if (!this.canAdd) return;
      const outputPath = this.addForm.downloadPath + "/%(title)s.%(ext)s";
      try {
        await BilibiliDownload(this.addForm.url, this.addForm.formatId, outputPath);
        ElMessage.success("已添加到下载列表");
        this.addDialogVisible = false;
        this.refreshTasks();
      } catch (e) {
        console.error("Add task error:", e);
        ElMessage.error("添加下载任务失败: " + e.message);
      }
    },
  },
};
</script>