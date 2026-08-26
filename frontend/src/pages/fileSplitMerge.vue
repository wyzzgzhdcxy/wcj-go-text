<style>
.split-merge-container {
  padding: 20px;
  max-width: 900px;
  margin: 0 auto;
}

.file-info {
  margin: 15px 0;
  padding: 10px;
  background: #e6f7ff;
  border-radius: 4px;
  border-left: 4px solid #1890ff;
}

.result-info {
  margin-top: 20px;
  padding: 15px;
  background: #f6ffed;
  border-radius: 4px;
  border-left: 4px solid #52c41a;
}

.split-files-list {
  margin-top: 15px;
  max-height: 200px;
  overflow-y: auto;
}

.file-item {
  padding: 5px 10px;
  margin: 3px 0;
  background: #fff;
  border-radius: 3px;
  font-size: 13px;
}

.el-radio-group {
  margin-bottom: 15px;
}
</style>

<template>
  <div class="split-merge-container">
    <el-tabs v-model="activeTab" type="border-card">
      <!-- 文件分割 -->
      <el-tab-pane label="文件分割" name="split">
        <el-form label-width="100px">
          <el-form-item label="选择文件">
            <el-input v-model="splitReq.filePath" placeholder="点击右侧按钮选择文件" readonly>
              <template #append>
                <el-button @click="selectSplitFile">选择文件</el-button>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="分割方式">
            <el-radio-group v-model="splitReq.splitType">
              <el-radio label="splitBySize">按大小分割</el-radio>
              <el-radio label="splitByCount">按个数分割</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item :label="splitReq.splitType === 'splitBySize' ? '每个文件大小' : '分割个数'">
            <el-input v-model="splitReq.splitValueStr" :placeholder="splitReq.splitType === 'splitBySize' ? '如: 10MB, 100MB, 1GB' : '如: 3'">
              <template #append v-if="splitReq.splitType === 'splitBySize'">
                <el-select v-model="splitReq.unit" style="width: 100px">
                  <el-option label="B" value="B"></el-option>
                  <el-option label="KB" value="KB"></el-option>
                  <el-option label="MB" value="MB"></el-option>
                  <el-option label="GB" value="GB"></el-option>
                </el-select>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="输出目录">
            <el-input v-model="splitReq.outputDir" placeholder="点击右侧按钮选择目录" readonly>
              <template #append>
                <el-button @click="selectOutputDir">选择目录</el-button>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="splitFile" :disabled="!canSplit">开始分割</el-button>
            <el-button @click="openSplitDir" :disabled="!splitResult.outputDir">打开输出目录</el-button>
          </el-form-item>
        </el-form>

        <!-- 文件信息 -->
        <div v-if="splitReq.filePath" class="file-info">
          <div><strong>源文件：</strong>{{ splitReq.filePath }}</div>
        </div>

        <!-- 分割结果 -->
        <div v-if="splitResult.success" class="result-info">
          <div><strong>{{ splitResult.message }}</strong></div>
          <div><strong>输出目录：</strong>{{ splitResult.outputDir }}</div>
          <div><strong>总大小：</strong>{{ splitResult.totalSizeStr }}</div>
          <div class="split-files-list">
            <div v-for="(file, index) in splitResult.splitFiles" :key="index" class="file-item">
              {{ index + 1 }}. {{ file }}
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 文件合并 -->
      <el-tab-pane label="文件合并" name="merge">
        <el-form label-width="100px">
          <el-form-item label="选择文件">
            <el-button type="primary" @click="selectMergeFiles">添加文件</el-button>
            <el-button @click="clearMergeFiles" v-if="mergeReq.files.length > 0">清空</el-button>
            <span style="margin-left: 10px">已选择 {{ mergeReq.files.length }} 个文件</span>
          </el-form-item>

          <el-form-item label="文件列表" v-if="mergeReq.files.length > 0">
            <div style="max-height: 200px; overflow-y: auto; border: 1px solid #ddd; padding: 10px;">
              <div v-for="(file, index) in mergeReq.files" :key="index" class="file-item">
                {{ index + 1 }}. {{ file }}
              </div>
            </div>
          </el-form-item>

          <el-form-item label="输出文件">
            <el-input v-model="mergeReq.output" placeholder="点击右侧按钮选择输出位置" readonly>
              <template #append>
                <el-button @click="selectOutputFile">选择位置</el-button>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="mergeFiles" :disabled="!canMerge">开始合并</el-button>
          </el-form-item>
        </el-form>

        <!-- 合并结果 -->
        <div v-if="mergeResult.success" class="result-info">
          <div><strong>{{ mergeResult.message }}</strong></div>
          <div><strong>输出文件：</strong>{{ mergeResult.output }}</div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import {SplitFile, MergeFiles, SelectFile, SelectDirectory, OpenExplorer} from "../wailsjs/go/main/App.js";
import {ElNotification} from "element-plus";

export default {
  name: 'FileSplitMerge',
  data() {
    return {
      activeTab: 'split',
      splitReq: {
        filePath: '',
        outputDir: '',
        splitType: 'splitBySize',
        splitValueStr: '100',
        unit: 'MB'
      },
      mergeReq: {
        files: [],
        output: ''
      },
      splitResult: {},
      mergeResult: {},
      sourceFileSize: 0
    }
  },
  computed: {
    canSplit() {
      return this.splitReq.filePath && this.splitReq.outputDir && this.splitReq.splitValueStr
    },
    canMerge() {
      return this.mergeReq.files.length > 0 && this.mergeReq.output
    }
  },
  mounted() {
  },
  methods: {
    formatSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },

    async selectSplitFile() {
      const filePath = await SelectFile()
      if (filePath) {
        this.splitReq.filePath = filePath
      }
    },

    async selectOutputDir() {
      const dir = await SelectDirectory()
      if (dir) {
        this.splitReq.outputDir = dir
      }
    },

    async splitFile() {
      let splitSizeStr = ''
      if (this.splitReq.splitType === 'splitBySize') {
        splitSizeStr = this.splitReq.splitValueStr + this.splitReq.unit
      }

      const req = {
        filePath: this.splitReq.filePath,
        outputDir: this.splitReq.outputDir,
        splitType: this.splitReq.splitType,
        splitValue: parseInt(this.splitReq.splitValueStr),
        splitSizeStr: splitSizeStr
      }

      this.splitResult = await SplitFile(req)
      ElNotification({
        title: '分割结果',
        message: this.splitResult.message,
        type: this.splitResult.success ? 'success' : 'error',
        duration: 5000
      })
    },

    openSplitDir() {
      if (this.splitResult.outputDir) {
        OpenExplorer(this.splitResult.outputDir)
      }
    },

    async selectMergeFiles() {
      const filePath = await SelectFile()
      if (filePath) {
        this.mergeReq.files.push(filePath)
      }
    },

    async selectOutputFile() {
      const dir = await SelectDirectory()
      if (dir) {
        const fileName = prompt('请输入合并后的文件名:', 'merged_file.bin')
        if (fileName) {
          this.mergeReq.output = dir + '\\' + fileName
        }
      }
    },

    async mergeFiles() {
      const req = {
        files: this.mergeReq.files,
        output: this.mergeReq.output,
        deleteSrc: false
      }

      this.mergeResult = await MergeFiles(req)
      ElNotification({
        title: '合并结果',
        message: this.mergeResult.message,
        type: this.mergeResult.success ? 'success' : 'error',
        duration: 5000
      })
    },

    clearMergeFiles() {
      this.mergeReq.files = []
      this.mergeReq.output = ''
      this.mergeResult = {}
    }
  }
}
</script>
