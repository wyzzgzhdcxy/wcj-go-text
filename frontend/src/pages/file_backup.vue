<template>

  <!-- 骨架屏占位 -->
  <div v-if="loading" style="width: 100%;height:100%;margin-top: 300px">
    <h2 style="width: 90%;margin-left: 5%">{{ actionDesc }}:[{{ progressValue }}%]</h2>
    <div class="progress-track" style="width: 90%;margin-left: 5%">
      <div class="progress-thumb" :style="{ width: progressValue + '%' }">
        <span class="thumb-icon">🚀</span>
      </div>
    </div>
  </div>

  <!-- 输入对话框 -->
  <el-dialog v-model="dialogVisible"
             title="修改备份文件夹" width="30%" :close-on-click-modal="false"
             :close-on-press-escape="false" :show-close="false">

    <!-- 表单内容 -->
    <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
      <el-form-item :label="`文件夹 ${index + 1}`" prop="name" v-for="(dir, index) in analysisResults.oriDir"
                    :key="index">
        <el-input v-model="analysisResults.oriDir[index]" placeholder="请输入名称"></el-input>
      </el-form-item>
    </el-form>

    <!-- 对话框底部按钮 -->
    <template #footer>
      <el-button @click="dialogVisible=false">取消</el-button>
      <el-button type="primary" @click="submit">提交</el-button>
    </template>
  </el-dialog>

  <div v-if="!loading" class="folder-analytics-app">
    <div class="analytics-header">
      <h1><i class="fas fa-folder-open"></i>文件备份还原</h1>
      <div class="summary-stats" v-if="analysisResults.items.length">
        <div class="stat-badge">
          <i class="fas fa-folder"></i>
          <span>{{ analysisResults.items.length }} 个文件夹</span>
        </div>
        <div class="stat-badge accent">
          <i class="fas fa-file-alt"></i>
          <span>{{ analysisResults.totalFiles }} 个文件</span>
        </div>
        <div class="stat-badge warning">
          <i class="fas fa-hdd"></i>
          <span>{{ formatSize(analysisResults.totalSize) }}</span>
        </div>
        <div class="stat-badge alert-success" @click="changeOriDir">
          <i class="fas fa-hdd"></i>
          <span>设置</span>
        </div>
      </div>
    </div>

    <div class="folder-grid">
      <div v-for="(result, index) in analysisResults.items"
           :key="result.folderPath"
           :class="['folder-card', {expanded: result.showDetails}]"
           :style="cardStyle(result)">
        <div class="card-header">
          <div class="folder-meta">
            <div class="folder-icon">
              <i class="fas" :class="result.showDetails ? 'fa-folder-open' : 'fa-folder'"></i>
            </div>
            <div class="folder-info">
              <div class="folder-name">{{ getFolderName(result.folderPath) }}</div>
              <div class="folder-path">{{ result.folderPath }}</div>
            </div>
          </div>
          <div class="folder-actions">
            <button class="action-btn backup-btn" @click.stop="backupFolder(result)" title="备份文件夹">
              <i class="fas fa-cloud-upload-alt"></i>
              <span class="btn-tooltip">备份</span>
            </button>
            <button class="action-btn restore-btn" @click.stop="restoreFolder(result)" title="还原文件夹">
              <i class="fas fa-cloud-download-alt"></i>
              <span class="btn-tooltip">还原</span>
            </button>
          </div>
        </div>

        <div class="card-stats">
          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-weight-hanging"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">总大小</div>
              <div class="stat-value">{{ result.humanTotalSize }}</div>
            </div>
          </div>

          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-file"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">最新文件</div>
              <div class="stat-value truncate" :title="result.latestFileName">
                {{ result.latestFileName }}
              </div>
            </div>
          </div>

          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-clock"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">最后修改</div>
              <div class="stat-value">{{ formatTime(result.latestModifyTime) }}</div>
            </div>
          </div>


          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-file"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">文件个数</div>
              <div class="stat-value truncate" :title="result.fileLength">
                {{ result.fileLength }}
              </div>
            </div>
          </div>

          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-file"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">备份文件</div>
              <div class="stat-value truncate" :title="result.tarFn">
                {{ getFileName(result.tarFn) }}
              </div>
            </div>
          </div>


          <div class="stat-item">
            <div class="stat-icon">
              <i class="fas fa-clock"></i>
            </div>
            <div class="stat-content">
              <div class="stat-label">备份时间</div>
              <div class="stat-value">{{ formatTime(result.tarLastModifyTime) }}</div>
            </div>
          </div>

        </div>
      </div>
    </div>


    <div style="margin-top: 100px">
      <el-tag type="danger">忽略的文件夹列表</el-tag>
      <span v-for="(item, index) in analysisResults.ignoreList" :key="index">
                  <el-tag type="warning">  {{ item }}</el-tag></span>

      <el-tag type="danger">备份仓库</el-tag>
      <el-tag type="warning">{{ analysisResults.backupDir }}</el-tag>

    </div>

  </div>
</template>


<script>
import {AnalyzeFolders, BackUp, ModifyOriDir, RestoreFiles} from "../wailsjs/go/app/App.js";
import {ElNotification} from "element-plus";

export default {
  name: 'FolderAnalyzer',
  data() {
    return {
      formData: {},
      loading: true,
      progressValue: 0,
      actionDesc: "数据加载中",
      newFolderPath: '',
      dialogVisible: false,
      analysisResults: [],
      colorPalette: [
        {bg: '#fff0f3', header: '#ff758f', icon: '#ff758f'},
        {bg: '#f0f7ff', header: '#3a86ff', icon: '#3a86ff'},
        {bg: '#f0fff4', header: '#38b000', icon: '#38b000'},
        {bg: '#fff9f0', header: '#ff9e00', icon: '#ff9e00'},
        {bg: '#f5f0ff', header: '#9d4edd', icon: '#9d4edd'},
      ]

    }
  },
  computed: {},
  mounted() {
    this.analyzeFolders()
  },
  methods: {

    submit() {
      let that = this;
      console.dir(this.analysisResults.oriDir)
      ModifyOriDir(this.analysisResults.oriDir).then(errorInfo => {
        this.dialogVisible = false
        ElNotification({
          title: '修改备份文件夹结果',
          message: "成功",
          position: 'bottom-right',
          type: 'success'
        })
        that.analyzeFolders()
      })
    },

    changeOriDir() {
      this.dialogVisible = true
    },
    cardStyle(result) {
      let color
      if (result.tarFn !== "" && result.tarLastModifyTime > result.latestModifyTime) {
        color = this.colorPalette[1]
      } else {
        color = this.colorPalette[0]
      }
      return {
        '--card-bg': color.bg,
        '--card-header': color.header,
        '--card-icon': color.icon
      }
    },
    getFolderName(path) {
      return path.split(/[\\/]/).pop()
    },
    getFileName(path) {
      return path.split(/[\\/]/).pop()
    },
    formatSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },
    formatTime(timeStr) {
      const date = new Date(timeStr)
      return date.toLocaleString()
    },

    backupFolder(result) {
      let that = this;
      that.progressValue = 0
      that.actionDesc = "正在数据备份中"
      let intervalTask = setInterval(function () {
        that.progressValue++
        if (that.progressValue > 100) {
          that.progressValue = 0
        }
      }, 500)
      this.loading = true
      // 这里添加备份逻辑
      console.log(result.folderPath)
      BackUp(result.folderPath).then(() => {
        that.actionDesc = "数据备份成功,重新加载数据中"
        this.LoadAnalyzeData(intervalTask)
      })
    },

    restoreFolder(result) {
      // 这里添加还原逻辑
      console.log('还原文件夹:', result.folderPath);
      this.$confirm(`确定要使用tar文件[${result.tarFn}]还原文件夹 ${this.getFolderName(result.folderPath)} 吗?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        let that = this;
        that.progressValue = 0
        that.actionDesc = "正在数据恢复中"
        let intervalTask = setInterval(function () {
          that.progressValue++
          if (that.progressValue > 100) {
            that.progressValue = 0
          }
        }, 500)
        this.loading = true

        RestoreFiles(result.tarFn, result.folderPath).then(result => {
          that.actionDesc = "数据恢复成功,重新加载数据中"
          this.LoadAnalyzeData(intervalTask)
        })
      });
    },

    LoadAnalyzeData(intervalTask) {
      let that = this;
      AnalyzeFolders().then(result => {
        try {
          that.analysisResults = result
          console.dir(that.analysisResults)
        } catch (err) {
          console.error("Error analyzing folders:", err)
          alert("Error analyzing folders: " + err)
        } finally {
          console.log(that.loading)
          that.loading = false
          clearInterval(intervalTask)
        }
      })
    },


    async analyzeFolders() {
      let that = this;
      let intervalTask = setInterval(function () {
        that.progressValue++
        if (that.progressValue > 100) {
          that.progressValue = 0
        }
      }, 50)
      that.loading = true
      that.LoadAnalyzeData(intervalTask);
    }
  }
}
</script>

<style scoped>
@import './css/file_backup.css';
</style>