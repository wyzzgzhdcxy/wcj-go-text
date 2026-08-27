<style scoped>
.m3u8-task-container {
  display: flex;
  height: 100%;
  gap: 10px;
  padding: 0;
  box-sizing: border-box;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-count {
  display: inline-block;
  background: #f56c6c;
  color: white;
  border-radius: 10px;
  padding: 0 6px;
  font-size: 11px;
  margin-left: 4px;
  line-height: 16px;
  min-width: 16px;
  text-align: center;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.toolbar {
  padding: 10px 15px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-list {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
  background: #fff;
}

.task-item {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 12px;
  border-left: 4px solid #409eff;
  transition: all 0.3s;
}

.task-item.downloading {
  border-left-color: #67c23a;
  background: #f0f9eb;
}

.task-item.merging {
  border-left-color: #e6a23c;
  background: #fdf6ec;
}

.task-item.completed {
  border-left-color: #909399;
  opacity: 0.8;
}

.task-item.failed {
  border-left-color: #f56c6c;
  background: #fef0f0;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.task-actions-inline {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-title {
  font-weight: bold;
  color: #303133;
  font-size: 14px;
  word-break: break-all;
  flex: 1;
}

.task-status {
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  margin-left: 10px;
}

.task-status.pending {
  background: #e4e7ed;
  color: #606266;
}

.task-status.downloading {
  background: #67c23a;
  color: white;
}

.task-status.merging {
  background: #e6a23c;
  color: white;
}

.task-status.completed {
  background: #909399;
  color: white;
}

.task-status.failed {
  background: #f56c6c;
  color: white;
}

.task-url {
  font-size: 12px;
  color: #909399;
  word-break: break-all;
  margin-bottom: 10px;
}

.task-info {
  font-size: 12px;
  color: #606266;
  margin-bottom: 10px;
}

.task-progress {
  margin-top: 8px;
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #606266;
}

.task-actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.add-dialog :deep(.el-dialog__body) {
  padding: 20px 30px;
}

.form-item {
  margin-bottom: 18px;
}

.form-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  display: block;
}

.empty-tip {
  text-align: center;
  color: #909399;
  padding: 50px 20px;
  font-size: 14px;
}
</style>

<template>
  <div class="m3u8-task-container">
    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 工具栏 -->
      <div class="toolbar">
        <span style="font-weight: bold; font-size: 16px;">下载任务列表</span>
        <div class="toolbar-right">
          <el-radio-group v-model="filterType" size="default">
            <el-radio-button label="all">
              全部<span class="filter-count" v-if="taskCounts.all > 0">{{ taskCounts.all }}</span>
            </el-radio-button>
            <el-radio-button label="downloading">
              下载中<span class="filter-count" v-if="taskCounts.downloading > 0">{{ taskCounts.downloading }}</span>
            </el-radio-button>
            <el-radio-button label="completed">
              已完成<span class="filter-count" v-if="taskCounts.completed > 0">{{ taskCounts.completed }}</span>
            </el-radio-button>
          </el-radio-group>
          <el-button
            v-if="filterType === 'completed' && taskCounts.completed > 0"
            type="danger"
            plain
            @click="clearCompleted"
          >
            清空列表
          </el-button>
          <el-tooltip :content="'同时下载任务数：' + settings.maxConcurrent" placement="bottom">
            <el-button circle @click="openSettings">
              <el-icon><Setting /></el-icon>
            </el-button>
          </el-tooltip>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            <span>添加</span>
          </el-button>
        </div>
      </div>

      <!-- 任务列表 -->
      <div class="task-list">
        <div v-if="filteredTasks.length === 0" class="empty-tip">
          暂无{{ filterType === 'all' ? '' : filterType === 'downloading' ? '下载中' : '已完成' }}任务
        </div>

        <div
          v-for="task in filteredTasks"
          :key="task.id"
          class="task-item"
          :class="task.status"
        >
          <div class="task-header">
            <span class="task-title">{{ task.title || '未命名' }}</span>
            <div class="task-actions-inline">
              <el-button
                v-if="task.status === 'downloading' || task.status === 'merging'"
                type="danger"
                size="small"
                @click="cancelDownload(task)"
              >
                取消
              </el-button>
              <el-button
                size="small"
                @click="removeTask(task)"
              >
                删除
              </el-button>
              <span v-if="task.status === 'pending'" class="task-status pending">等待中</span>
              <span v-else-if="task.status === 'downloading'" class="task-status downloading">下载中</span>
              <span v-else-if="task.status === 'merging'" class="task-status merging">合并中</span>
              <span v-else-if="task.status === 'completed'" class="task-status completed">已完成</span>
              <span v-else-if="task.status === 'failed'" class="task-status failed">失败</span>
              <el-button
                v-if="task.status === 'pending' || task.status === 'failed'"
                type="primary"
                size="small"
                @click="startDownload(task)"
              >
                {{ task.status === 'failed' ? '重试' : '开始' }}
              </el-button>
              <el-button
                v-if="task.status === 'completed'"
                type="success"
                size="small"
                @click="openOutputDir(task)"
              >
                打开目录
              </el-button>
            </div>
          </div>

          <div class="task-url">{{ task.url }}</div>

          <div class="task-info">
            <span>保存路径：{{ task.outputDir }}</span>
            <span style="margin-left: 15px;">线程数：{{ task.threadCount }}</span>
            <span v-if="task.status === 'failed' && task.errorMsg" style="margin-left: 15px; color: #f56c6c;">
              错误：{{ task.errorMsg }}
            </span>
          </div>

          <div class="task-progress">
            <el-progress :percentage="task.progress" :stroke-width="10" style="flex: 1; margin-right: 10px;" />
            <span>下载速度：{{ task.speed }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加任务对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="添加下载任务"
      width="500px"
      class="add-dialog"
    >
      <div class="form-item">
        <label class="form-label">M3U8地址（支持批量，每行一个）</label>
        <el-input
          v-model="newTask.url"
          type="textarea"
          :rows="6"
          placeholder="支持两种格式：
① 单个URL：https://example.com/video.m3u8
② 文件名+URL：视频名	https://example.com/video.m3u8"
          @change="onUrlChange"
        />
      </div>

      <div class="form-item">
        <label class="form-label">保存路径 *</label>
        <el-input
          v-model="newTask.outputDir"
          placeholder="点击右侧按钮选择目录"
          readonly
        >
          <template #append>
            <el-button @click="selectOutputDir">选择</el-button>
          </template>
        </el-input>
      </div>

      <div class="form-item">
        <label class="form-label">保存文件名</label>
        <el-input
          v-model="newTask.title"
          placeholder="可选，留空自动命名"
          clearable
        />
      </div>

      <div class="form-item">
        <label class="form-label">下载线程数</label>
        <el-input-number
          v-model="newTask.threadCount"
          :min="1"
          :max="64"
          style="width: 100%"
        />
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addTask" :disabled="!canAddTask">添加</el-button>
      </template>
    </el-dialog>

    <!-- 设置对话框 -->
    <el-dialog v-model="settingsVisible" title="下载设置" width="420px">
      <div class="form-item">
        <label class="form-label">同时下载任务数量（1 - 10）</label>
        <el-input-number
          v-model="settingsDraft.maxConcurrent"
          :min="1"
          :max="10"
          style="width: 100%"
        />
        <div style="font-size: 12px; color: #909399; margin-top: 6px;">
          超出此数量的任务将进入等待队列，逐个开始。
        </div>
      </div>
      <template #footer>
        <el-button @click="settingsVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSettings">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {DownloadM3u8, SelectDirectory, OpenExplorer, OpenUrl} from "../wailsjs/go/app/App.js";
import {EventsOn, EventsOff} from "../wailsjs/runtime/runtime.js";
import {ElMessage, ElMessageBox} from "element-plus";
import {Setting, Plus} from "@element-plus/icons-vue";

// 全局存储下载任务列表，避免切换页面数据丢失
if (!window.m3u8TaskList) {
  window.m3u8TaskList = []
}

// 全局标记是否已注册事件监听器
if (window.m3u8ProgressHandlerRegistered === undefined) {
  window.m3u8ProgressHandlerRegistered = false
}

export default {
  name: 'M3u8TaskDownload',
  components: { Setting, Plus },
  data() {
    return {
      filterType: 'all',
      tasks: window.m3u8TaskList,
      dialogVisible: false,
      settingsVisible: false,
      settings: {
        maxConcurrent: parseInt(localStorage.getItem('m3u8MaxConcurrent')) || 3
      },
      settingsDraft: {
        maxConcurrent: 3
      },
      newTask: {
        url: '',
        title: '',
        outputDir: '',
        threadCount: 16
      },
      statusText: {
        pending: '等待中',
        downloading: '下载中',
        merging: '合并中',
        completed: '已完成',
        failed: '失败'
      },
      eventHandler: null
    }
  },
  watch: {
    tasks: {
      handler(newTasks) {
        window.m3u8TaskList = newTasks
      },
      deep: true
    }
  },
  computed: {
    filteredTasks() {
      if (this.filterType === 'all') {
        return this.tasks
      } else if (this.filterType === 'downloading') {
        return this.tasks.filter(t => t.status === 'downloading' || t.status === 'pending' || t.status === 'merging')
      } else if (this.filterType === 'completed') {
        return this.tasks.filter(t => t.status === 'completed')
      }
      return this.tasks
    },
    taskCounts() {
      return {
        all: this.tasks.length,
        downloading: this.tasks.filter(t => t.status === 'downloading' || t.status === 'pending' || t.status === 'merging').length,
        completed: this.tasks.filter(t => t.status === 'completed').length
      }
    },
    canAddTask() {
      return this.newTask.url.trim() && this.newTask.outputDir
    }
  },
  mounted() {
    this.setupProgressListener()
  },
  methods: {
    setupProgressListener() {
      // 避免重复注册
      if (window.m3u8ProgressHandlerRegistered) {
        return
      }
      window.m3u8ProgressHandlerRegistered = true

      const handler = (data) => {
        const status = data.status
        const percent = data.percent
        const taskId = data.taskId

        // 按taskId查找任务
        let task = null
        if (taskId) {
          task = window.m3u8TaskList.find(t => t.id === taskId)
        }

        if (!task) return

        if (status === 'starting') {
          task.progress = 0
        } else if (status === 'completed') {
          // 后端在 N_m3u8DL-RE 进程退出且最终视频文件已生成后才发送 completed，
          // 此时才算真正完成下载+合并
          task.status = 'completed'
          task.progress = 100
          ElMessage.success((task.title || '任务') + ' 下载完成')
          this.scheduleTasks()
        } else if (status === 'failed') {
          task.status = 'failed'
          const msg = data.message || data.speed || '下载失败'
          task.errorMsg = msg
          ElMessage.error((task.title || '任务') + ' 下载失败: ' + msg)
          this.scheduleTasks()
        } else if (status === 'merging') {
          // 后端在分片全部下载完成、正在合并/封装最终视频时发送 merging，
          // 此时任务仍处于进行中，不应移动到已完成页面
          if (task.status === 'downloading') {
            task.status = 'merging'
          }
          task.progress = Math.round(percent || 99)
        } else if (percent !== undefined) {
          task.progress = Math.round(percent)
          if (data.speed) {
            task.speed = data.speed
          }
        }
      }

      this.eventHandler = handler
      EventsOn('m3u8_progress', handler)
    },

    showAddDialog() {
      this.newTask = {
        url: '',
        title: '',
        outputDir: 'D:/',
        threadCount: 16
      }
      this.dialogVisible = true
    },

    openSettings() {
      this.settingsDraft = { ...this.settings }
      this.settingsVisible = true
    },

    saveSettings() {
      const v = parseInt(this.settingsDraft.maxConcurrent)
      if (!v || v < 1 || v > 10) {
        ElMessage.warning('请输入 1 - 10 之间的数字')
        return
      }
      this.settings.maxConcurrent = v
      localStorage.setItem('m3u8MaxConcurrent', String(v))
      this.settingsVisible = false
      ElMessage.success('设置已保存')
      // 设置变化后尝试补齐空闲位
      this.scheduleTasks()
    },

    activeCount() {
      return this.tasks.filter(t => t.status === 'downloading' || t.status === 'merging').length
    },

    scheduleTasks() {
      const cap = this.settings.maxConcurrent
      const active = this.activeCount()
      const slots = cap - active
      if (slots <= 0) return
      const queued = this.tasks.filter(t => t.status === 'pending')
      for (let i = 0; i < Math.min(slots, queued.length); i++) {
        this.startDownload(queued[i])
      }
    },

    async selectOutputDir() {
      try {
        const dir = await SelectDirectory()
        if (dir) {
          this.newTask.outputDir = dir
        }
      } catch (e) {
        ElMessage.error('选择目录失败')
      }
    },

    onUrlChange() {
      // 批量输入时，自动填充文件名
      if (this.newTask.url && !this.newTask.title) {
        const lines = this.newTask.url.split('\n').filter(l => l.trim())
        if (lines.length === 1) {
          const httpPos = lines[0].toLowerCase().indexOf('http')
          if (httpPos > 0) {
            this.newTask.title = lines[0].substring(0, httpPos).trim()
          }
        }
      }
    },

    addTask() {
      if (!this.newTask.url.trim()) {
        ElMessage.warning('请填写M3U8地址')
        return
      }
      if (!this.newTask.outputDir) {
        ElMessage.warning('请选择保存路径')
        return
      }

      // 解析批量输入
      const lines = this.newTask.url.split('\n').filter(l => l.trim())
      const userTitle = (this.newTask.title || '').trim()
      let addedCount = 0
      for (const line of lines) {
        const httpPos = line.toLowerCase().indexOf('http')
        if (httpPos === -1) continue
        const lineTitle = line.substring(0, httpPos).trim()
        const url = line.substring(httpPos).trim()

        // 命名优先级：行内前缀 > 单行时用户输入的保存文件名 > 自动生成
        let title = lineTitle
        if (!title && lines.length === 1 && userTitle) {
          title = userTitle
        }
        if (!title) {
          title = this.generateTitle(url)
        }
        title = this.sanitizeFilename(title)

        const task = {
          id: Date.now().toString() + '_' + addedCount,
          url: url,
          title: title,
          outputDir: this.newTask.outputDir,
          threadCount: this.newTask.threadCount,
          status: 'pending',
          progress: 0,
          speed: ''
        }
        this.tasks.unshift(task)
        this.startDownload(task)
        addedCount++
      }

      this.dialogVisible = false
      if (addedCount > 0) {
        ElMessage.success('已添加 ' + addedCount + ' 个任务')
      } else {
        ElMessage.warning('没有有效的M3U8链接')
      }
    },

    generateTitle(url) {
      const parts = url.split('/')
      const lastPart = parts[parts.length - 1]
      return lastPart.replace('.m3u8', '') || 'video_' + Date.now()
    },

    sanitizeFilename(name) {
      // 移除 Windows 文件名非法字符：\ / : * ? " < > | 以及控制字符
      let s = name.replace(/[\\/:*?"<>|\x00-\x1f]/g, '_')
      // 首尾的空格和点号在 Windows 下不允许
      s = s.replace(/^[\s.]+|[\s.]+$/g, '')
      // 限制长度，保留扩展名空间
      if (s.length > 100) s = s.substring(0, 100)
      return s || 'video_' + Date.now()
    },

    async startDownload(task) {
      if (task.status === 'downloading' || task.status === 'merging') return

      // 并发限制：达到上限时保持 pending，等待空闲槽位
      if (this.activeCount() >= this.settings.maxConcurrent) {
        task.status = 'pending'
        return
      }

      task.status = 'downloading'
      task.progress = 0
      task.speed = ''

      // 不等待结果，让进度事件更新状态
      DownloadM3u8({
        url: task.url,
        title: task.title,
        outputDir: task.outputDir,
        threadCount: task.threadCount,
        taskId: task.id
      }).then(result => {
        if (!result.success) {
          task.status = 'failed'
          task.errorMsg = result.message
        }
        // 成功时由进度事件最终标记完成；完成后补齐队列
        this.$nextTick(() => this.scheduleTasks())
      }).catch(err => {
        task.status = 'failed'
        task.errorMsg = err.message
        this.$nextTick(() => this.scheduleTasks())
      })
    },

    cancelDownload(task) {
      // 后端暂不支持取消，这里只是标记
      task.status = 'failed'
      task.errorMsg = '用户取消'
      ElMessage.info('已取消下载')
    },

    openOutputDir(task) {
      console.log('openOutputDir:', task.outputDir, task.outputPath)
      if (task.outputDir) {
        OpenExplorer(task.outputDir)
      } else if (task.outputPath) {
        const lastSlash = Math.max(task.outputPath.lastIndexOf('/'), task.outputPath.lastIndexOf('\\'))
        OpenExplorer(lastSlash > 0 ? task.outputPath.substring(0, lastSlash) : task.outputDir)
      }
    },

    removeTask(task) {
      const index = this.tasks.findIndex(t => t.id === task.id)
      if (index !== -1) {
        this.tasks.splice(index, 1)
      }
    },

    clearCompleted() {
      if (this.taskCounts.completed === 0) return
      ElMessageBox.confirm(
        '确定清空所有已完成任务吗？',
        '提示',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      ).then(() => {
        this.tasks = this.tasks.filter(t => t.status !== 'completed')
        ElMessage.success('已清空已完成任务')
      }).catch(() => {})
    }
  }
}
</script>
