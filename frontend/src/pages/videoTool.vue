<style scoped>
.batch-container {
  width: 100%;
  height: 100%;
  padding: 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.header-section {
  margin-bottom: 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.label {
  font-size: 14px;
  color: #606266;
  min-width: 80px;
  flex-shrink: 0;
}

.dir-path-inline {
  background: #f5f7fa;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 13px;
  color: #303133;
  border: 1px solid #e4e7ed;
  word-break: break-all;
  flex: 1;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.table-toolbar {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.select-count {
  font-size: 12px;
  color: #909399;
}

.file-table th,
.file-table td {
  border: 1px solid #e4e7ed;
  padding: 5px 10px;
  text-align: left;
}

.file-table th {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
}

.file-table tr:hover td {
  background: #ecf5ff;
}

.file-name {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-dir {
  color: #409eff;
}

.file-size {
  color: #909399;
  font-size: 12px;
}

.stats-card {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid #e4e7ed;
}

.stats-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.stats-total {
  font-size: 28px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.stats-exts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ext-tag {
  padding: 4px 10px;
  background: #ecf5ff;
  border-radius: 4px;
  font-size: 12px;
  color: #409eff;
  border: 1px solid #b3d8ff;
}

.log-area {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #f8f9fa;
  height: 150px;
  overflow-y: auto;
  font-size: 12px;
  line-height: 1.8;
  color: #303133;
  font-family: Consolas, Monaco, monospace;
  flex-shrink: 0;
}

.result-card {
  margin-top: 12px;
  padding: 14px;
  border-radius: 6px;
  border: 1px solid #67c23a;
  background: #f0f9eb;
  font-size: 13px;
  flex-shrink: 0;
}

.result-card.error {
  border-color: #f56c6c;
  background: #fef0f0;
}

.action-btn {
  min-width: 140px;
}
</style>

<template>
  <!-- 旋转对话框 -->
  <el-dialog v-model="rotateDialogVisible" title="视频旋转" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 80px;">旋转角度：</label>
        <el-select v-model="rotateAngle" style="width: 150px;">
          <el-option value="90" label="顺时针90°" />
          <el-option value="180" label="180°" />
          <el-option value="270" label="逆时针90°" />
        </el-select>
      </div>
      <div style="margin-top: 12px;">
        <el-checkbox v-model="rotateFast">快速旋转（仅改方向元数据，推荐MP4/MOV，无损）</el-checkbox>
      </div>
    </div>
    <template #footer>
      <el-button @click="rotateDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmRotate">开始旋转</el-button>
    </template>
  </el-dialog>

  <!-- 抽帧对话框 -->
  <el-dialog v-model="extractFramesDialogVisible" title="视频抽帧" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 80px;">抽帧数量：</label>
        <el-input-number v-model="extractFrameCount" :min="1" :max="100" />
      </div>
    </div>
    <template #footer>
      <el-button @click="extractFramesDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmExtractFrames">开始抽帧</el-button>
    </template>
  </el-dialog>

  <!-- 去除片尾对话框 -->
  <el-dialog v-model="trimEndDialogVisible" title="去除片尾" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 100px;">片尾时长：</label>
        <el-input-number v-model="trimEndMinutes" :min="0" :max="60" :step="1" style="width: 100px;" />
        <span style="margin: 0 8px;">分</span>
        <el-input-number v-model="trimEndSeconds" :min="0" :max="59" :step="1" style="width: 100px;" />
        <span style="margin-left: 8px;">秒</span>
      </div>
      <div style="margin-top: 8px; color: #909399; font-size: 12px;">
        共 {{ trimEndMinutes }} 分 {{ trimEndSeconds }} 秒，将无损切割视频
      </div>
    </div>
    <template #footer>
      <el-button @click="trimEndDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmTrimEnd">开始处理</el-button>
    </template>
  </el-dialog>

  <!-- 去除片头对话框 -->
  <el-dialog v-model="trimStartDialogVisible" title="去除片头" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 100px;">片头时长：</label>
        <el-input-number v-model="trimStartMinutes" :min="0" :max="60" :step="1" style="width: 100px;" />
        <span style="margin: 0 8px;">分</span>
        <el-input-number v-model="trimStartSeconds" :min="0" :max="59" :step="1" style="width: 100px;" />
        <span style="margin-left: 8px;">秒</span>
      </div>
      <div style="margin-top: 8px; color: #909399; font-size: 12px;">
        共 {{ trimStartMinutes }} 分 {{ trimStartSeconds }} 秒，将无损切割视频
      </div>
    </div>
    <template #footer>
      <el-button @click="trimStartDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmTrimStart">开始处理</el-button>
    </template>
  </el-dialog>

  <!-- 合并选中对话框 -->
  <el-dialog v-model="mergeDialogVisible" title="合并选中视频" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 80px;">输出文件名：</label>
        <el-input v-model="mergeOutputName" placeholder="输入输出文件名（不含扩展名）" style="width: 200px;" />
        <span style="color: #909399; font-size: 12px;">.mp4</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="mergeDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmMergeSelected">开始合并</el-button>
    </template>
  </el-dialog>

  <!-- 声音分离对话框 -->
  <el-dialog v-model="extractDialogVisible" title="确认开始声音分离" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>要处理的文件个数：<strong>{{ selectedFiles.length }}</strong> 个</div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 80px;">输出格式：</label>
        <el-select v-model="extractFormat" style="width: 150px;">
          <el-option value="mp3" label="MP3" />
          <el-option value="aac" label="AAC" />
          <el-option value="wav" label="WAV" />
          <el-option value="flac" label="FLAC" />
        </el-select>
      </div>
      <div style="margin-top: 12px;">
        <label style="display: inline-block; width: 80px;">并行线程：</label>
        <el-select v-model="extractThread" style="width: 150px;">
          <el-option value="2" label="2" />
          <el-option value="4" label="4" />
          <el-option value="8" label="8" />
          <el-option value="16" label="16" />
        </el-select>
      </div>
    </div>
    <template #footer>
      <el-button @click="extractDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmExtract">开始处理</el-button>
    </template>
  </el-dialog>

  <div class="batch-container">
    <div class="header-section">
      <div class="form-row">
        <el-button type="primary" @click="selectDir">打开目录</el-button>
        <el-input
          v-model="fileNameFilter"
          placeholder="输入文件名过滤..."
          style="width: 200px;"
          clearable
          @clear="handleFilterClear"
        >
          <template #prefix>
            <span>🔍</span>
          </template>
        </el-input>
        <span v-if="fileNameFilter" class="select-count" style="color: #409eff;">
          过滤后: {{ filteredFileList.length }} / {{ fileList.length }}
        </span>
        <el-button
          type="warning"
          class="action-btn"
          :loading="processing"
          :disabled="selectedFiles.length === 0"
          @click="startExtract"
        >
          {{ processing ? '处理中...' : '🎵 声音分离' }}
        </el-button>
        <el-button
          type="info"
          :loading="processing"
          :disabled="selectedFiles.length === 0"
          @click="startRotate"
        >
          🔄 旋转
        </el-button>
        <el-button
          type="success"
          :loading="processing"
          :disabled="selectedFiles.length === 0"
          @click="startExtractFrames"
        >
          🖼️ 抽帧
        </el-button>
        <el-button
          type="danger"
          :loading="processing"
          :disabled="selectedFiles.length === 0"
          @click="startTrimStart"
        >
          ✂️ 去除片头
        </el-button>
        <el-button
          type="danger"
          :loading="processing"
          :disabled="selectedFiles.length === 0"
          @click="startTrimEnd"
        >
          ✂️ 去除片尾
        </el-button>
        <el-button
          type="warning"
          :loading="processing"
          :disabled="subDirs.length === 0"
          @click="startBatchMerge"
        >
          🔗 批量合并
        </el-button>
        <el-button
          type="warning"
          :loading="processing"
          :disabled="selectedFiles.length < 2"
          @click="startMergeSelected"
        >
          🔗 合并选中
        </el-button>
        <el-button
          type="info"
          :loading="loadingResolutions"
          @click="showResolutions"
        >
          📐 详情
        </el-button>
        <el-button
          type="info"
          :disabled="selectedFiles.length !== 1"
          @click="showKeyframes"
        >
          🎬 关键帧
        </el-button>
        <el-button
          type="success"
          :loading="processing"
          @click="startClassifyByResolution"
        >
          📁 按分辨率分类
        </el-button>
        <el-button
          type="info"
          @click="openVideoDir"
        >
          📂 打开目录
        </el-button>
      </div>

      <!-- 关键帧对话框 -->
      <el-dialog v-model="keyframesDialogVisible" title="关键帧位置" width="500px">
        <div v-if="keyframesLoading" style="text-align: center; padding: 20px;">
          正在获取关键帧信息...
        </div>
        <div v-else-if="keyframesData" style="text-align: left;">
          <div style="margin-bottom: 12px;">
            <strong>{{ keyframesData.fileName }}</strong>
          </div>
          <div style="margin-bottom: 12px; color: #909399; font-size: 12px;">
            前10秒共找到 <strong style="color: #409eff;">{{ (keyframesData.keyframes || []).length }}</strong> 个关键帧
          </div>
          <div style="max-height: 300px; overflow-y: auto; border: 1px solid #e4e7ed; border-radius: 4px;">
            <table style="width: 100%; border-collapse: collapse;">
              <thead>
                <tr style="background: #f5f7fa;">
                  <th style="padding: 8px 12px; text-align: left; border-bottom: 1px solid #e4e7ed;">#</th>
                  <th style="padding: 8px 12px; text-align: left; border-bottom: 1px solid #e4e7ed;">时间</th>
                  <th style="padding: 8px 12px; text-align: left; border-bottom: 1px solid #e4e7ed;">秒</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="kf in (keyframesData.keyframes || [])" :key="kf.index" style="hover: { background: #ecf5ff };">
                  <td style="padding: 8px 12px; border-bottom: 1px solid #e4e7ed;">{{ kf.index }}</td>
                  <td style="padding: 8px 12px; border-bottom: 1px solid #e4e7ed; font-family: monospace;">{{ kf.timeStr }}</td>
                  <td style="padding: 8px 12px; border-bottom: 1px solid #e4e7ed; color: #909399; font-size: 12px;">{{ kf.time.toFixed(3) }}s</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div style="margin-top: 12px; color: #909399; font-size: 12px;">
            选择"去除片头"时，建议选择第一个关键帧所在时间点，可确保视频开头是关键帧
          </div>
        </div>
        <div v-else style="text-align: center; padding: 20px; color: #909399;">
          未获取到关键帧信息
        </div>
        <template #footer>
          <el-button @click="keyframesDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>

      <!-- 文件列表 -->
      <div v-if="filteredFileList.length > 0" class="file-table">
        <div class="table-toolbar">
          <el-button size="small" @click="selectAll">全选</el-button>
          <el-button size="small" @click="selectNone">都不选</el-button>
          <span class="select-count">已选择：{{ selectedFiles.length }} / {{ filteredFileList.length }}</span>
          <span v-if="totalDuration > 0" class="select-count" style="color: #67c23a;">（总时长：{{ formatDuration(totalDuration) }}）</span>
          <span v-if="subDirs.length > 0" class="select-count" style="color: #67c23a;">（批量合并：{{ subDirs.length }} 个子目录）</span>
        </div>
        <div style="flex: 1; overflow-y: auto; min-height: 0;">
          <table>
            <thead>
              <tr>
                <th style="width: 40px;"><el-checkbox v-model="isAllSelected" @change="toggleSelectAll"/></th>
                <th>名称</th>
                <th>大小</th>
                <th v-if="showResolutionColumn">分辨率</th>
                <th v-if="showResolutionColumn">帧率</th>
                <th v-if="showResolutionColumn">码率</th>
                <th v-if="showResolutionColumn">编码</th>
                <th v-if="showResolutionColumn">时长</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(file, index) in filteredFileList" :key="index">
                <td><el-checkbox v-model="file.selected" @change="updateSelectAllState"/></td>
                <td class="file-name">{{ file.name }}</td>
                <td class="file-size">{{ formatSize(file.fileSize) }}</td>
                <td v-if="showResolutionColumn" class="file-size" :style="{ color: file.resolution ? '#67c23a' : '#909399' }">
                  {{ file.resolution || '-' }}
                </td>
                <td v-if="showResolutionColumn" class="file-size">
                  {{ formatFrameRate(file.frameRate) }}
                </td>
                <td v-if="showResolutionColumn" class="file-size">
                  {{ file.bitRate || '-' }}
                </td>
                <td v-if="showResolutionColumn" class="file-size">
                  {{ file.codec || '-' }}
                </td>
                <td v-if="showResolutionColumn" class="file-size">
                  {{ formatDuration(file.duration) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-else-if="videoDir && !loading" style="color: #909399; text-align: center; padding: 20px;">
        目录为空
      </div>
    </div>

    <!-- 日志区域 -->
    <div class="log-area" ref="logArea">
      <div v-if="logs.length === 0" style="color: #c0c4cc; text-align: center; padding-top: 80px;">
        选择视频目录后点击"扫描目录"开始...
      </div>
      <div v-for="(log, i) in logs" :key="i">{{ log }}</div>
    </div>

    <!-- 结果 -->
    <div v-if="result" class="result-card" :class="{ error: !result.success }">
      <div v-if="result.success">
        <strong style="color: #67c23a;">✅ 批量提取完成！</strong>
        <div style="margin-top: 8px;">
          <div>成功：{{ result.successCount }} / {{ result.totalCount }}</div>
          <div>输出目录：{{ result.outputDir }}</div>
          <div>总耗时：{{ result.totalCost }}</div>
        </div>
        <div style="margin-top: 12px;">
          <el-button size="small" type="success" @click="openOutputDir">打开目录</el-button>
        </div>
      </div>
      <div v-else>
        <strong style="color: #f56c6c;">❌ 提取失败</strong>
        <div style="margin-top: 6px; color: #606266;">{{ result.message }}</div>
      </div>

      <!-- 失败文件列表 -->
      <div v-if="result.failedFiles && result.failedFiles.length > 0" style="margin-top: 12px;">
        <div style="font-weight: 600; color: #f56c6c;">失败文件：</div>
        <div v-for="(file, i) in result.failedFiles" :key="i" style="font-size: 11px; color: #909399; margin-top: 2px;">
          {{ file }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {SelectDirectory, OpenExplorer, GetDirContents} from "../wailsjs/go/app/App.js";
import {BatchExtractAudioByFiles, MergeVideos, MergeVideosByFiles, ClassifyVideosByResolution} from "../wailsjs/go/app/App.js";

export default {
  name: 'batchExtractAudio',
  data() {
    return {
      videoDir: '',
      audioFormat: 'mp3',
      threadCount: 4,
      scanning: false,
      extracting: false,
      videoStats: null,
      result: null,
      logs: [],
      fileList: [],
      fileNameFilter: '',
      loading: false,
      selectedFiles: [],
      isAllSelected: false,
      processing: false,
      loadingResolutions: false,
      showResolutionColumn: false,
      // 对话框状态
      rotateDialogVisible: false,
      rotateAngle: '90',
      rotateFast: true,
      extractFramesDialogVisible: false,
      extractFrameCount: 5,
      extractDialogVisible: false,
      extractFormat: 'mp3',
      extractThread: 4,
      keyframesDialogVisible: false,
      keyframesLoading: false,
      keyframesData: null,
      trimEndDialogVisible: false,
      trimEndMinutes: 0,
      trimEndSeconds: 10,
      trimStartDialogVisible: false,
      trimStartMinutes: 0,
      trimStartSeconds: 10,
      mergeDialogVisible: false,
      mergeOutputName: 'merged_video',
      subDirs: [],
      confirmDialogVisible: false,
      confirmDialog: {
        title: '',
        message: '',
        action: null,
      },
    };
  },
  mounted() {
    window.runtime.EventsOn('back_msg', (message) => {
      try {
        const msg = JSON.parse(message);
        this.logs.push(msg.time + ' ' + msg.msg);
        this.$nextTick(() => {
          if (this.$refs.logArea) {
            this.$refs.logArea.scrollTop = this.$refs.logArea.scrollHeight;
          }
        });
      } catch (e) {
        // 如果不是 JSON 格式，直接显示原始消息
        this.logs.push(message);
      }
    });
  },
  watch: {
    videoDir: {
      handler(newDir) {
        if (newDir) {
          window.runtime.WindowSetTitle('视频处理 - ' + newDir);
        } else {
          window.runtime.WindowSetTitle('视频处理');
        }
      },
    },
    fileNameFilter(newFilter) {
      if (newFilter) {
        // 有过滤条件时，自动选中过滤后的文件
        const filter = newFilter.toLowerCase();
        this.fileList.forEach(f => {
          f.selected = f.name.toLowerCase().includes(filter);
        });
        this.updateSelectAllState();
      }
    },
  },
  beforeUnmount() {
    window.runtime.EventsOff('back_msg');
  },
  computed: {
    filteredFileList() {
      if (!this.fileNameFilter) {
        return this.fileList;
      }
      const filter = this.fileNameFilter.toLowerCase();
      return this.fileList.filter(f => f.name.toLowerCase().includes(filter));
    },
    totalDuration() {
      if (!this.showResolutionColumn) return 0;
      return this.selectedFiles.reduce((sum, file) => {
        const found = this.fileList.find(f => f.name === file.name);
        return sum + (found && found.duration > 0 ? found.duration : 0);
      }, 0);
    },
  },
  methods: {
    handleFilterClear() {
      this.fileNameFilter = '';
    },

    async selectDir() {
      try {
        const dir = await SelectDirectory();
        if (dir) {
          this.videoDir = dir;
          this.videoStats = null;
          this.result = null;
          this.logs = [];
          this.fileList = [];
          await this.loadDirContents();
        }
      } catch (e) {
        console.error('选择目录失败:', e);
      }
    },

    async loadDirContents() {
      if (!this.videoDir) return;
      this.loading = true;
      try {
        const res = await GetDirContents(this.videoDir);
        this.fileList = res || [];

        // 扫描子目录
        this.subDirs = [];
        for (const item of this.fileList) {
          if (item.isDir) {
            const subDirPath = this.videoDir + '/' + item.name;
            const subRes = await GetDirContents(subDirPath);
            const mp4Count = (subRes || []).filter(f => f.name.toLowerCase().endsWith('.mp4')).length;
            if (mp4Count > 0) {
              this.subDirs.push({ name: item.name, mp4Count });
            }
          }
        }

        // 只保留 MP4 文件
        this.fileList = this.fileList.filter(f => !f.isDir && f.name.toLowerCase().endsWith('.mp4'));
        this.fileList.forEach(f => f.selected = false);
      } catch (e) {
        console.error('加载目录内容失败:', e);
        this.fileList = [];
      } finally {
        this.loading = false;
      }
    },

    formatSize(size) {
      if (size === 0) return '-';
      if (size < 1024) return size + ' B';
      if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB';
      if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(1) + ' MB';
      return (size / 1024 / 1024 / 1024).toFixed(2) + ' GB';
    },

    formatDuration(seconds) {
      if (!seconds || seconds <= 0) return '-';
      const mins = Math.floor(seconds / 60);
      const secs = Math.floor(seconds % 60);
      if (mins >= 60) {
        const hours = Math.floor(mins / 60);
        const remainingMins = mins % 60;
        return `${hours}时${remainingMins}分${secs}秒`;
      }
      return `${mins}分${secs}秒`;
    },

    formatFrameRate(frameRate) {
      if (!frameRate) return '-';
      // ffprobe 返回格式如 "25/1"，转换为 "25 fps"
      if (frameRate.includes('/')) {
        const parts = frameRate.split('/');
        const num = parseInt(parts[0]);
        const den = parseInt(parts[1]);
        if (den === 1) return num + ' fps';
        return (num / den).toFixed(2) + ' fps';
      }
      return frameRate + ' fps';
    },

    selectAll() {
      this.filteredFileList.forEach(f => f.selected = true);
      this.updateSelectAllState();
    },

    selectNone() {
      this.filteredFileList.forEach(f => f.selected = false);
      this.updateSelectAllState();
    },

    toggleSelectAll(checked) {
      this.filteredFileList.forEach(f => f.selected = checked);
      this.updateSelectAllState();
    },

    updateSelectAllState() {
      const selected = this.fileList.filter(f => f.selected);
      this.selectedFiles = selected;
      this.isAllSelected = this.filteredFileList.length > 0 && selected.length === this.filteredFileList.length;
    },

    async scanDir() {
      if (!this.videoDir) {
        this.$message.warning('请先选择视频目录');
        return;
      }

      this.scanning = true;
      this.logs = [];
      this.videoStats = null;
      this.result = null;

      try {
        const res = await window.go.app.App.ScanVideoDir({dirPath: this.videoDir});
        this.videoStats = res;

        if (res.success) {
          this.logs.push('扫描完成！');
          if (res.totalCount === 0) {
            this.logs.push('⚠️ 未找到视频文件');
          } else {
            this.logs.push(`共找到 ${res.totalCount} 个视频文件`);
            res.extInfos.forEach(ext => {
              this.logs.push(`  ${ext.ext}: ${ext.count}个`);
            });
          }
        } else {
          this.logs.push('❌ 扫描失败: ' + res.message);
        }
      } catch (e) {
        this.logs.push('❌ 扫描失败: ' + e);
      } finally {
        this.scanning = false;
      }
    },

    async startExtract() {
      if (this.selectedFiles.length === 0) {
        this.$message.warning('请先选择要处理的文件');
        return;
      }

      this.extractFormat = this.audioFormat;
      this.extractThread = this.threadCount;
      this.extractDialogVisible = true;
    },

    async confirmExtract() {
      this.extractDialogVisible = false;
      this.audioFormat = this.extractFormat;
      this.threadCount = this.extractThread;
      await this.doExtract(this.extractFormat, this.extractThread);
    },

    async doExtract(format, threadCount) {
      this.extracting = true;
      this.result = null;
      this.logs = [];

      const fileNames = this.selectedFiles.map(f => f.name);

      try {
        const res = await BatchExtractAudioByFiles({
          dirPath: this.videoDir,
          fileNames: fileNames,
          format: format,
          threadCount: threadCount,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('声音分离完成！');
        } else {
          this.$message.error('处理完成，但有失败文件');
        }
      } catch (e) {
        this.$message.error('处理失败: ' + e);
        this.logs.push('❌ 处理失败: ' + e);
      } finally {
        this.extracting = false;
      }
    },

    startRotate() {
      if (this.selectedFiles.length === 0) {
        this.$message.warning('请先选择要处理的视频');
        return;
      }

      this.rotateAngle = '90';
      this.rotateFast = false;
      this.rotateDialogVisible = true;
    },

    async confirmRotate() {
      this.rotateDialogVisible = false;
      const angle = parseInt(this.rotateAngle);
      const isClockwise = angle !== 270;
      await this.doRotate(angle, isClockwise, this.rotateFast);
    },

    async doRotate(angle, clockwise, fastRotate) {
      this.processing = true;
      this.logs = [];

      const fileNames = this.selectedFiles.map(f => f.name);

      try {
        for (let i = 0; i < fileNames.length; i++) {
          const res = await window.go.app.App.RotateVideo({
            dirPath: this.videoDir,
            fileName: fileNames[i],
            angle: angle,
            clockwise: clockwise,
            fastRotate: fastRotate,
          });
          if (res.success) {
            this.logs.push(`✅ ${fileNames[i]} -> ${res.outputPath}`);
          } else {
            this.logs.push(`❌ ${fileNames[i]} 失败: ${res.message}`);
          }
        }
        this.$message.success('旋转完成');
      } catch (e) {
        this.$message.error('旋转失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    startExtractFrames() {
      if (this.selectedFiles.length === 0) {
        this.$message.warning('请先选择要处理的视频');
        return;
      }

      this.extractFrameCount = 5;
      this.extractFramesDialogVisible = true;
    },

    async confirmExtractFrames() {
      this.extractFramesDialogVisible = false;
      await this.doExtractFrames(this.extractFrameCount);
    },

    async doExtractFrames(count) {
      this.processing = true;
      this.result = null;
      this.logs = [];

      const fileNames = this.selectedFiles.map(f => f.name);

      try {
        const res = await window.go.app.App.ExtractFramesByFiles({
          dirPath: this.videoDir,
          fileNames: fileNames,
          count: count,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('抽帧完成！');
        } else {
          this.$message.warning('抽帧完成，部分失败');
        }
      } catch (e) {
        this.$message.error('抽帧失败: ' + e);
        this.logs.push('❌ 抽帧失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    startTrimEnd() {
      if (this.selectedFiles.length === 0) {
        this.$message.warning('请先选择要处理的视频');
        return;
      }

      this.trimEndMinutes = 0;
      this.trimEndSeconds = 10;
      this.trimEndDialogVisible = true;
    },

    async confirmTrimEnd() {
      this.trimEndDialogVisible = false;
      const duration = this.trimEndMinutes * 60 + this.trimEndSeconds;
      if (duration <= 0) {
        this.$message.warning('片尾时长必须大于0');
        return;
      }
      await this.doTrimEnd(duration);
    },

    startTrimStart() {
      if (this.selectedFiles.length === 0) {
        this.$message.warning('请先选择要处理的视频');
        return;
      }

      this.trimStartMinutes = 0;
      this.trimStartSeconds = 10;
      this.trimStartDialogVisible = true;
    },

    async confirmTrimStart() {
      this.trimStartDialogVisible = false;
      const duration = this.trimStartMinutes * 60 + this.trimStartSeconds;
      if (duration <= 0) {
        this.$message.warning('片头时长必须大于0');
        return;
      }
      await this.doTrimStart(duration);
    },

    async doTrimStart(duration) {
      this.processing = true;
      this.result = null;
      this.logs = [];

      const fileNames = this.selectedFiles.map(f => f.name);

      try {
        const res = await window.go.app.App.TrimVideoStartByFiles({
          dirPath: this.videoDir,
          fileNames: fileNames,
          duration: duration,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('去除片头完成！');
        } else {
          this.$message.warning('处理完成，部分失败');
        }
      } catch (e) {
        this.$message.error('去除片头失败: ' + e);
        this.logs.push('❌ 去除片头失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    async doTrimEnd(duration) {
      this.processing = true;
      this.result = null;
      this.logs = [];

      const fileNames = this.selectedFiles.map(f => f.name);

      try {
        const res = await window.go.app.App.TrimVideoEndByFiles({
          dirPath: this.videoDir,
          fileNames: fileNames,
          duration: duration,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('去除片尾完成！');
        } else {
          this.$message.warning('处理完成，部分失败');
        }
      } catch (e) {
        this.$message.error('去除片尾失败: ' + e);
        this.logs.push('❌ 去除片尾失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    startBatchMerge() {
      if (this.subDirs.length === 0) {
        this.$message.warning('没有子目录可以批量合并');
        return;
      }

      this.$confirm(`将对 ${this.subDirs.length} 个子目录中的视频分别合并，是否继续？`, '批量合并', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(async () => {
        await this.doBatchMerge();
      }).catch(() => {});
    },

    async doBatchMerge() {
      this.processing = true;
      this.result = null;
      this.logs = [];

      try {
        const res = await MergeVideos({
          dirPath: this.videoDir,
          ignoreSort: false,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('批量合并完成！');
        } else {
          this.$message.warning('合并完成，但有部分失败');
        }
      } catch (e) {
        this.$message.error('合并失败: ' + e);
        this.logs.push('❌ 合并失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    startMergeSelected() {
      if (this.selectedFiles.length < 2) {
        this.$message.warning('请至少选择 2 个视频文件');
        return;
      }

      this.mergeOutputName = 'merged_video';
      this.mergeDialogVisible = true;
    },

    confirmMergeSelected() {
      if (!this.mergeOutputName || this.mergeOutputName.trim() === '') {
        this.$message.warning('请输入输出文件名');
        return;
      }
      this.mergeDialogVisible = false;
      this.doMergeSelected(this.mergeOutputName.trim());
    },

    async doMergeSelected(outputName) {
      this.processing = true;
      this.result = null;
      this.logs = [];

      try {
        const res = await MergeVideosByFiles({
          dirPath: this.videoDir,
          fileNames: this.selectedFiles.map(f => f.name),
          outputName: outputName,
        });
        this.result = res;

        if (res.success) {
          this.$message.success('合并完成！');
        } else {
          this.$message.error('合并失败: ' + res.message);
        }
      } catch (e) {
        this.$message.error('合并失败: ' + e);
        this.logs.push('❌ 合并失败: ' + e);
      } finally {
        this.processing = false;
      }
    },

    openOutputDir() {
      if (this.result && this.result.outputDir) {
        OpenExplorer(this.result.outputDir);
      }
    },

    async showResolutions() {
      if (!this.videoDir) {
        this.$message.warning('请先选择视频目录');
        return;
      }

      this.loadingResolutions = true;
      this.logs = [];
      this.result = null;
      this.showResolutionColumn = true;

      try {
        const res = await window.go.app.App.ScanVideos({dirPath: this.videoDir});
        if (res.success && res.allVideos && res.allVideos.length > 0) {
          // 更新文件列表的详细信息
          const resMap = {};
          res.allVideos.forEach(v => {
            resMap[v.fileName] = v;
          });
          this.fileList.forEach(file => {
            const info = resMap[file.name];
            if (info) {
              file.resolution = info.resolution;
              file.frameRate = info.frameRate;
              file.bitRate = info.bitRate;
              file.codec = info.codec;
              file.duration = info.duration;
            }
          });

          this.logs.push('=== 视频信息列表 ===');
          res.allVideos.forEach((video, i) => {
            const duration = this.formatDuration(video.duration);
            const fps = this.formatFrameRate(video.frameRate);
            this.logs.push(`${i + 1}. ${video.fileName}`);
            this.logs.push(`   分辨率: ${video.resolution || '-'} | 帧率: ${fps} | 码率: ${video.bitRate || '-'}`);
            this.logs.push(`   编码: ${video.codec || '-'} | 时长: ${duration}`);
          });
          this.logs.push('');
          this.logs.push(`共 ${res.allVideos.length} 个视频`);

          // 按分辨率分组统计
          const statMap = {};
          res.allVideos.forEach(v => {
            const r = v.resolution || '未知';
            statMap[r] = (statMap[r] || 0) + 1;
          });
          this.logs.push('');
          this.logs.push('=== 分辨率统计 ===');
          Object.keys(statMap).sort().forEach(res => {
            this.logs.push(`${res}: ${statMap[res]}个`);
          });
        } else {
          this.logs.push('未找到视频文件');
        }
      } catch (e) {
        this.$message.error('获取视频信息失败: ' + e);
        this.logs.push('❌ 获取视频信息失败: ' + e);
      } finally {
        this.loadingResolutions = false;
      }
    },

    async showKeyframes() {
      if (this.selectedFiles.length !== 1) {
        this.$message.warning('请先选择一个视频文件');
        return;
      }

      this.keyframesDialogVisible = true;
      this.keyframesLoading = true;
      this.keyframesData = null;

      try {
        const res = await window.go.app.App.GetKeyframes({
          dirPath: this.videoDir,
          fileName: this.selectedFiles[0].name,
          maxDuration: 10,
        });
        this.keyframesData = res;
        if (!res.success) {
          this.$message.error('获取关键帧失败: ' + res.message);
          this.keyframesData = null;
        }
      } catch (e) {
        this.$message.error('获取关键帧失败: ' + e);
        this.keyframesData = null;
      } finally {
        this.keyframesLoading = false;
      }
    },

    async startClassifyByResolution() {
      if (!this.videoDir) {
        this.$message.warning('请先选择视频目录');
        return;
      }

      this.$confirm('将对目录下的视频按分辨率分类到不同文件夹，是否继续？', '按分辨率分类', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(async () => {
        this.processing = true;
        this.result = null;
        this.logs = [];

        try {
          const res = await ClassifyVideosByResolution({
            dirPath: this.videoDir,
          });
          this.result = res;

          if (res.success) {
            this.$message.success('分类完成！');
          } else {
            this.$message.warning('分类完成，但有失败文件');
          }
        } catch (e) {
          this.$message.error('分类失败: ' + e);
          this.logs.push('❌ 分类失败: ' + e);
        } finally {
          this.processing = false;
        }
      }).catch(() => {});
    },

    openVideoDir() {
      if (this.videoDir) {
        OpenExplorer(this.videoDir);
      }
    },
  },
};
</script>