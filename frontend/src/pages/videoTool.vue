<style scoped>
.video-tool {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dir-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.dir-path {
  flex: 1;
  background: #f5f7fa;
  padding: 7px 12px;
  border-radius: 4px;
  font-size: 13px;
  color: #303133;
  border: 1px solid #e4e7ed;
  word-break: break-all;
}

.dir-path.empty {
  color: #c0c4cc;
}

.action-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.bar-divider {
  width: 1px;
  height: 22px;
  background: #dcdfe6;
  margin: 0 6px;
  flex-shrink: 0;
}

.list-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}

.table-toolbar {
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.select-count {
  font-size: 12px;
  color: #909399;
}

.table-scroll {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
}

.file-table th,
.file-table td {
  border-bottom: 1px solid #ebeef5;
  padding: 6px 10px;
  text-align: left;
  font-size: 13px;
}

.file-table th {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
  position: sticky;
  top: 0;
  z-index: 1;
}

.file-table tbody tr:hover td {
  background: #ecf5ff;
}

.file-name {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: #909399;
  font-size: 12px;
  white-space: nowrap;
}

.empty {
  color: #909399;
  text-align: center;
  padding: 40px 0;
  font-size: 13px;
}

.log-area {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 10px 12px;
  background: #f8f9fa;
  height: 150px;
  overflow-y: auto;
  font-size: 12px;
  line-height: 1.8;
  color: #303133;
  font-family: Consolas, Monaco, monospace;
  flex-shrink: 0;
  margin-top: 12px;
}

.result-card {
  margin-top: 12px;
  padding: 12px 14px;
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

.result-row {
  margin-top: 4px;
  color: #606266;
}

.failed-list {
  margin-top: 10px;
}

.failed-item {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
  word-break: break-all;
}

.dialog-row {
  margin-top: 12px;
  display: flex;
  align-items: center;
}

.dialog-label {
  display: inline-block;
  width: 84px;
  flex-shrink: 0;
}

.dialog-unit {
  margin: 0 8px;
}

.dialog-tip {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

.kf-table {
  width: 100%;
  border-collapse: collapse;
}

.kf-table th,
.kf-table td {
  padding: 7px 12px;
  text-align: left;
  border-bottom: 1px solid #e4e7ed;
}

.kf-table tbody tr:hover td {
  background: #ecf5ff;
}
</style>

<template>
  <!-- 声音分离对话框 -->
  <el-dialog v-model="audioDialogVisible" title="声音分离" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>要处理的文件个数：<strong>{{ selectedFiles.length }}</strong> 个</div>
      <div class="dialog-row">
        <span class="dialog-label">输出格式：</span>
        <el-select v-model="audioFormat" style="width: 150px;">
          <el-option value="mp3" label="MP3" />
          <el-option value="aac" label="AAC" />
          <el-option value="wav" label="WAV" />
          <el-option value="flac" label="FLAC" />
        </el-select>
      </div>
      <div class="dialog-row">
        <span class="dialog-label">并行线程：</span>
        <el-select v-model="audioThread" style="width: 150px;">
          <el-option :value="2" label="2" />
          <el-option :value="4" label="4" />
          <el-option :value="8" label="8" />
          <el-option :value="16" label="16" />
        </el-select>
      </div>
    </div>
    <template #footer>
      <el-button @click="audioDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmExtractAudio">开始处理</el-button>
    </template>
  </el-dialog>

  <!-- 抽帧对话框 -->
  <el-dialog v-model="framesDialogVisible" title="视频抽帧" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div class="dialog-row">
        <span class="dialog-label">每个视频抽帧数：</span>
        <el-input-number v-model="frameCount" :min="1" :max="100" />
      </div>
    </div>
    <template #footer>
      <el-button @click="framesDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmExtractFrames">开始抽帧</el-button>
    </template>
  </el-dialog>

  <!-- 旋转对话框 -->
  <el-dialog v-model="rotateDialogVisible" title="视频旋转" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div class="dialog-row">
        <span class="dialog-label">旋转角度：</span>
        <el-select v-model="rotateAngle" style="width: 150px;">
          <el-option :value="90" label="顺时针90°" />
          <el-option :value="180" label="180°" />
          <el-option :value="270" label="逆时针90°" />
        </el-select>
      </div>
      <div class="dialog-row">
        <el-checkbox v-model="rotateFast">快速旋转（仅改方向元数据，推荐 MP4/MOV，无损）</el-checkbox>
      </div>
    </div>
    <template #footer>
      <el-button @click="rotateDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmRotate">开始旋转</el-button>
    </template>
  </el-dialog>

  <!-- 剪切对话框（片头/片尾共用） -->
  <el-dialog v-model="trimDialogVisible" :title="trimPosition === 'start' ? '去除片头' : '去除片尾'" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频</div>
      <div class="dialog-row">
        <span class="dialog-label">{{ trimPosition === 'start' ? '片头时长：' : '片尾时长：' }}</span>
        <el-input-number v-model="trimMinutes" :min="0" :max="60" :step="1" style="width: 100px;" />
        <span class="dialog-unit">分</span>
        <el-input-number v-model="trimSeconds" :min="0" :max="59" :step="1" style="width: 100px;" />
        <span class="dialog-unit">秒</span>
      </div>
      <div class="dialog-tip">共 {{ trimMinutes * 60 + trimSeconds }} 秒，将无损切割视频，输出为新文件</div>
    </div>
    <template #footer>
      <el-button @click="trimDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmTrim">开始处理</el-button>
    </template>
  </el-dialog>

  <!-- 合并选中对话框 -->
  <el-dialog v-model="mergeDialogVisible" title="合并选中视频" width="420px" :close-on-click-modal="false">
    <div style="text-align: left; line-height: 1.8;">
      <div>已选择 <strong>{{ selectedFiles.length }}</strong> 个视频，按列表顺序合并</div>
      <div class="dialog-row">
        <span class="dialog-label">输出文件名：</span>
        <el-input v-model="mergeOutputName" placeholder="不含扩展名" style="width: 200px;" />
        <span style="color: #909399; font-size: 12px;">.mp4</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="mergeDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmMergeSelected">开始合并</el-button>
    </template>
  </el-dialog>

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
        <table class="kf-table">
          <thead>
            <tr style="background: #f5f7fa;">
              <th>#</th>
              <th>时间</th>
              <th>秒</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="kf in (keyframesData.keyframes || [])" :key="kf.index">
              <td>{{ kf.index }}</td>
              <td style="font-family: monospace;">{{ kf.timeStr }}</td>
              <td style="color: #909399; font-size: 12px;">{{ kf.time.toFixed(3) }}s</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div style="margin-top: 12px; color: #909399; font-size: 12px;">
        使用"去除片头"时，建议切在第一个关键帧之后，可确保切割点落在关键帧上
      </div>
    </div>
    <div v-else style="text-align: center; padding: 20px; color: #909399;">
      未获取到关键帧信息
    </div>
    <template #footer>
      <el-button @click="keyframesDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>

  <div
    class="video-tool"
    v-loading="processing"
    :element-loading-text="loadingText"
    element-loading-background="rgba(255, 255, 255, 0.75)"
  >
    <!-- 目录栏 -->
    <div class="dir-bar">
      <el-button type="primary" @click="selectDir">选择目录</el-button>
      <div class="dir-path" :class="{ empty: !videoDir }">{{ videoDir || '未选择目录，请先选择视频所在目录' }}</div>
      <el-button :disabled="!videoDir" @click="openVideoDir">打开目录</el-button>
    </div>

    <!-- 操作栏（处理中由整页 loading 遮罩统一屏蔽） -->
    <div class="action-bar">
      <el-button type="primary" :disabled="selectedFiles.length === 0" @click="audioDialogVisible = true">声音分离</el-button>
      <el-button type="primary" :disabled="selectedFiles.length === 0" @click="startExtractFrames">抽帧</el-button>
      <div class="bar-divider"></div>
      <el-button type="warning" :disabled="selectedFiles.length === 0" @click="startTrim('start')">去除片头</el-button>
      <el-button type="warning" :disabled="selectedFiles.length === 0" @click="startTrim('end')">去除片尾</el-button>
      <el-button type="warning" :disabled="selectedFiles.length === 0" @click="startRotate">旋转</el-button>
      <div class="bar-divider"></div>
      <el-button type="success" :disabled="selectedFiles.length < 2" @click="startMergeSelected">合并选中</el-button>
      <el-button type="success" :disabled="subDirs.length === 0" @click="startBatchMerge">批量合并</el-button>
      <div class="bar-divider"></div>
      <el-button type="info" :disabled="loadingResolutions || !videoDir" :loading="loadingResolutions" @click="showResolutions">详情</el-button>
      <el-button type="info" :disabled="selectedFiles.length !== 1" @click="showKeyframes">关键帧</el-button>
      <el-button type="info" :disabled="!videoDir" @click="startClassifyByResolution">按分辨率分类</el-button>
    </div>

    <!-- 文件列表 -->
    <div class="list-card">
      <div v-if="fileList.length > 0" class="table-toolbar">
        <el-checkbox v-model="isAllSelected" @change="toggleSelectAll">全选</el-checkbox>
        <el-button size="small" @click="selectNone">清空选择</el-button>
        <span class="select-count">已选择：{{ selectedFiles.length }} / {{ filteredFileList.length }}</span>
        <span v-if="totalDuration > 0" class="select-count" style="color: #67c23a;">（选中总时长：{{ formatDuration(totalDuration) }}）</span>
        <span v-if="subDirs.length > 0" class="select-count" style="color: #67c23a;">（发现 {{ subDirs.length }} 个含视频的子目录，可批量合并）</span>
        <span style="flex: 1;"></span>
        <el-input
          v-model="fileNameFilter"
          placeholder="文件名过滤（输入即自动选中匹配项）"
          style="width: 220px;"
          size="small"
          clearable
        />
      </div>

      <div v-if="fileList.length > 0 && filteredFileList.length === 0" class="empty">
        没有匹配"{{ fileNameFilter }}"的文件
      </div>

      <div v-else-if="fileList.length > 0" class="table-scroll">
        <table class="file-table">
          <thead>
            <tr>
              <th style="width: 40px;"></th>
              <th>名称</th>
              <th style="width: 90px;">大小</th>
              <th v-if="showResolutionColumn" style="width: 100px;">分辨率</th>
              <th v-if="showResolutionColumn" style="width: 80px;">帧率</th>
              <th v-if="showResolutionColumn" style="width: 90px;">码率</th>
              <th v-if="showResolutionColumn" style="width: 80px;">编码</th>
              <th v-if="showResolutionColumn" style="width: 110px;">时长</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="file in filteredFileList" :key="file.name">
              <td><el-checkbox v-model="file.selected" @change="updateSelectAllState"/></td>
              <td class="file-name" :title="file.name">{{ file.name }}</td>
              <td class="file-size">{{ formatSize(file.fileSize) }}</td>
              <td v-if="showResolutionColumn" class="file-size" :style="{ color: file.resolution ? '#67c23a' : '#909399' }">
                {{ file.resolution || '-' }}
              </td>
              <td v-if="showResolutionColumn" class="file-size">{{ formatFrameRate(file.frameRate) }}</td>
              <td v-if="showResolutionColumn" class="file-size">{{ file.bitRate || '-' }}</td>
              <td v-if="showResolutionColumn" class="file-size">{{ file.codec || '-' }}</td>
              <td v-if="showResolutionColumn" class="file-size">{{ formatDuration(file.duration) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else-if="loading" class="empty">正在加载目录...</div>
      <div v-else-if="videoDir" class="empty">目录中没有视频文件</div>
      <div v-else class="empty">请先选择视频目录</div>
    </div>

    <!-- 日志区域 -->
    <div class="log-area" ref="logArea">
      <div v-if="logs.length === 0" style="color: #c0c4cc; text-align: center; padding-top: 60px;">
        处理日志将显示在这里...
      </div>
      <div v-for="(log, i) in logs" :key="i">{{ log }}</div>
    </div>

    <!-- 结果 -->
    <div v-if="result" class="result-card" :class="{ error: !result.success }">
      <strong v-if="result.success" style="color: #67c23a;">✅ 处理完成</strong>
      <strong v-else style="color: #f56c6c;">❌ 处理失败</strong>
      <span v-if="result.message && !result.success" style="margin-left: 8px; color: #606266;">{{ result.message }}</span>

      <div v-if="result.totalCount" class="result-row">成功：{{ result.successCount }} / {{ result.totalCount }}</div>
      <div v-if="result.outputDir" class="result-row">输出目录：{{ result.outputDir }}</div>
      <div v-if="result.totalCost" class="result-row">总耗时：{{ result.totalCost }}</div>

      <div v-if="result.failedDirs && result.failedDirs.length > 0" class="failed-list">
        <div style="font-weight: 600; color: #f56c6c;">失败目录：</div>
        <div v-for="(dir, i) in result.failedDirs" :key="'d' + i" class="failed-item">{{ dir }}</div>
      </div>

      <div v-if="result.failedFiles && result.failedFiles.length > 0" class="failed-list">
        <div style="font-weight: 600; color: #f56c6c;">失败文件：</div>
        <div v-for="(file, i) in result.failedFiles" :key="'f' + i" class="failed-item">{{ file }}</div>
      </div>

      <div v-if="result.outputDir" style="margin-top: 10px;">
        <el-button size="small" type="success" @click="openOutputDir">打开输出目录</el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { reactive } from 'vue';
import {
  SelectDirectory, OpenExplorer, GetDirContents,
  BatchExtractAudioByFiles, RotateVideo, ExtractFramesByFiles,
  TrimVideoStartByFiles, TrimVideoEndByFiles,
  MergeVideos, MergeVideosByFiles, ClassifyVideosByResolution, GetKeyframes,
  ScanVideos,
} from '../wailsjs/go/app/App.js';
import { ElMessage, ElMessageBox } from 'element-plus';

// 与后端 myVideo 包支持的视频扩展名保持一致
const VIDEO_EXTS = ['.mp4', '.avi', '.mkv', '.mov', '.wmv', '.flv', '.m4v', '.ts', '.rmvb', '.webm'];

function isVideoFile(name) {
  const lower = (name || '').toLowerCase();
  return VIDEO_EXTS.some(ext => lower.endsWith(ext));
}

// 模块级共享状态：不随页面切换销毁，切走后任务继续、日志继续累积，
// 返回本页时等待遮罩 / 日志 / 结果 / 目录原样恢复
const shared = reactive({
  videoDir: '',
  processing: false,
  loadingText: '正在处理，请稍候...',
  logs: [],
  result: null,
});

let backMsgBound = false;
function bindBackMsgOnce() {
  if (backMsgBound) return;
  backMsgBound = true;
  window.runtime.EventsOn('back_msg', (message) => {
    try {
      const msg = JSON.parse(message);
      shared.logs.push(msg && msg.msg ? (msg.time ? msg.time + ' ' + msg.msg : String(msg.msg)) : message);
    } catch (e) {
      shared.logs.push(message);
    }
  });
}

export default {
  name: 'videoTool',
  data() {
    return {
      fileList: [],
      fileNameFilter: '',
      loading: false,
      selectedFiles: [],
      isAllSelected: false,
      loadingResolutions: false,
      showResolutionColumn: false,
      subDirs: [],
      // 声音分离
      audioDialogVisible: false,
      audioFormat: 'mp3',
      audioThread: 4,
      // 抽帧
      framesDialogVisible: false,
      frameCount: 5,
      // 旋转
      rotateDialogVisible: false,
      rotateAngle: 90,
      rotateFast: true,
      // 剪切（片头/片尾共用）
      trimDialogVisible: false,
      trimPosition: 'start',
      trimMinutes: 0,
      trimSeconds: 10,
      // 合并选中
      mergeDialogVisible: false,
      mergeOutputName: 'merged_video',
      // 关键帧
      keyframesDialogVisible: false,
      keyframesLoading: false,
      keyframesData: null,
    };
  },
  mounted() {
    bindBackMsgOnce();
    // 从其他页面返回时恢复状态：目录、标题、日志、进行中的任务遮罩
    if (shared.videoDir) {
      window.runtime.WindowSetTitle('视频处理 - ' + shared.videoDir);
      this.loadDirContents();
    } else {
      window.runtime.WindowSetTitle('视频处理');
    }
    this.$nextTick(() => {
      if (this.$refs.logArea) {
        this.$refs.logArea.scrollTop = this.$refs.logArea.scrollHeight;
      }
    });
  },
  watch: {
    videoDir(newDir) {
      window.runtime.WindowSetTitle(newDir ? '视频处理 - ' + newDir : '视频处理');
    },
    // 新日志到达时自动滚动到底部（切页期间累积的日志也会触发）
    'logs.length'() {
      this.$nextTick(() => {
        if (this.$refs.logArea) {
          this.$refs.logArea.scrollTop = this.$refs.logArea.scrollHeight;
        }
      });
    },
    fileNameFilter(newFilter) {
      if (newFilter) {
        const filter = newFilter.toLowerCase();
        this.fileList.forEach(f => {
          f.selected = f.name.toLowerCase().includes(filter);
        });
        this.updateSelectAllState();
      }
    },
  },
  computed: {
    // 这些状态存放在模块级 shared 中，切换页面后返回仍能恢复
    videoDir: {
      get() {
        return shared.videoDir;
      },
      set(v) {
        shared.videoDir = v;
      },
    },
    processing() {
      return shared.processing;
    },
    loadingText() {
      return shared.loadingText;
    },
    logs() {
      return shared.logs;
    },
    result() {
      return shared.result;
    },
    filteredFileList() {
      if (!this.fileNameFilter) {
        return this.fileList;
      }
      const filter = this.fileNameFilter.toLowerCase();
      return this.fileList.filter(f => f.name.toLowerCase().includes(filter));
    },
    totalDuration() {
      return this.selectedFiles.reduce((sum, file) => sum + (file.duration > 0 ? file.duration : 0), 0);
    },
  },
  methods: {
    async selectDir() {
      try {
        const dir = await SelectDirectory();
        if (!dir) return;
        this.videoDir = dir;
        shared.result = null;
        shared.logs = [];
        await this.loadDirContents();
      } catch (e) {
        ElMessage.error('选择目录失败: ' + e);
      }
    },

    async loadDirContents() {
      if (!this.videoDir) return;
      this.loading = true;
      this.showResolutionColumn = false;
      try {
        const res = await GetDirContents(this.videoDir);
        const items = res || [];

        // 扫描含视频文件的子目录（供"批量合并"使用）
        this.subDirs = [];
        await Promise.all(items.filter(it => it.isDir).map(async it => {
          try {
            const subRes = await GetDirContents(this.videoDir + '/' + it.name);
            const count = (subRes || []).filter(f => !f.isDir && isVideoFile(f.name)).length;
            if (count > 0) {
              this.subDirs.push({ name: it.name, videoCount: count });
            }
          } catch (e) {
            // 单个子目录读取失败不影响整体
          }
        }));
        this.subDirs.sort((a, b) => a.name.localeCompare(b.name));

        this.fileList = items.filter(it => !it.isDir && isVideoFile(it.name));
        this.fileList.forEach(f => { f.selected = false; });
        this.selectedFiles = [];
        this.isAllSelected = false;
      } catch (e) {
        console.error('加载目录内容失败:', e);
        ElMessage.error('加载目录失败: ' + e);
        this.fileList = [];
        this.subDirs = [];
      } finally {
        this.loading = false;
      }
    },

    openVideoDir() {
      if (this.videoDir) {
        OpenExplorer(this.videoDir);
      }
    },

    openOutputDir() {
      if (shared.result && shared.result.outputDir) {
        OpenExplorer(shared.result.outputDir);
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
        return `${hours}时${mins % 60}分${secs}秒`;
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

    selectNone() {
      this.filteredFileList.forEach(f => f.selected = false);
      this.updateSelectAllState();
    },

    toggleSelectAll(checked) {
      this.filteredFileList.forEach(f => f.selected = !!checked);
      this.updateSelectAllState();
    },

    updateSelectAllState() {
      this.selectedFiles = this.fileList.filter(f => f.selected);
      this.isAllSelected = this.filteredFileList.length > 0 &&
        this.filteredFileList.every(f => f.selected);
    },

    // ===== 声音分离 =====
    confirmExtractAudio() {
      this.audioDialogVisible = false;
      shared.loadingText = '正在声音分离，请稍候...';
      shared.processing = true;
      shared.result = null;
      const fileNames = this.selectedFiles.map(f => f.name);
      BatchExtractAudioByFiles({
        dirPath: this.videoDir,
        fileNames: fileNames,
        format: this.audioFormat,
        threadCount: Number(this.audioThread),
      }).then(res => {
        shared.result = res;
        if (res.success) {
          ElMessage.success('声音分离完成');
        } else {
          ElMessage.warning('处理完成，但有失败文件');
        }
      }).catch(e => {
        ElMessage.error('声音分离失败: ' + e);
        shared.logs.push('❌ 声音分离失败: ' + e);
      }).finally(() => {
        shared.processing = false;
      });
    },

    // ===== 抽帧 =====
    startExtractFrames() {
      this.frameCount = 5;
      this.framesDialogVisible = true;
    },

    confirmExtractFrames() {
      this.framesDialogVisible = false;
      shared.loadingText = '正在抽帧，请稍候...';
      shared.processing = true;
      shared.result = null;
      const fileNames = this.selectedFiles.map(f => f.name);
      ExtractFramesByFiles({
        dirPath: this.videoDir,
        fileNames: fileNames,
        count: this.frameCount,
      }).then(res => {
        shared.result = res;
        if (res.success) {
          ElMessage.success('抽帧完成');
        } else {
          ElMessage.warning('抽帧完成，部分失败');
        }
        return this.loadDirContents();
      }).catch(e => {
        ElMessage.error('抽帧失败: ' + e);
        shared.logs.push('❌ 抽帧失败: ' + e);
      }).finally(() => {
        shared.processing = false;
      });
    },

    // ===== 旋转 =====
    startRotate() {
      this.rotateDialogVisible = true;
    },

    async confirmRotate() {
      this.rotateDialogVisible = false;
      shared.loadingText = '正在旋转视频，请稍候...';
      shared.processing = true;
      shared.result = null;
      const fileNames = this.selectedFiles.map(f => f.name);
      const clockwise = this.rotateAngle !== 270;
      let ok = 0;
      const failedFiles = [];

      try {
        for (let i = 0; i < fileNames.length; i++) {
          try {
            const res = await RotateVideo({
              dirPath: this.videoDir,
              fileName: fileNames[i],
              angle: this.rotateAngle,
              clockwise: clockwise,
              fastRotate: this.rotateFast,
            });
            if (res && res.success) {
              ok++;
              shared.logs.push(`✅ [${i + 1}/${fileNames.length}] ${fileNames[i]} -> ${res.outputPath}`);
            } else {
              const msg = (res && res.message) || '未知错误';
              failedFiles.push(`${fileNames[i]}：${msg}`);
              shared.logs.push(`❌ [${i + 1}/${fileNames.length}] ${fileNames[i]} 失败: ${msg}`);
            }
          } catch (e) {
            failedFiles.push(`${fileNames[i]}：${e}`);
            shared.logs.push(`❌ [${i + 1}/${fileNames.length}] ${fileNames[i]} 异常: ${e}`);
          }
        }

        shared.result = {
          success: ok > 0,
          message: failedFiles.length > 0 ? '部分文件处理失败' : '',
          totalCount: fileNames.length,
          successCount: ok,
          failedFiles: failedFiles,
          outputDir: this.videoDir,
        };
        if (failedFiles.length === 0) {
          ElMessage.success('旋转完成');
        } else if (ok > 0) {
          ElMessage.warning('旋转完成，部分失败');
        } else {
          ElMessage.error('旋转全部失败');
        }
        await this.loadDirContents();
      } finally {
        shared.processing = false;
      }
    },

    // ===== 剪切（片头/片尾） =====
    startTrim(position) {
      this.trimPosition = position;
      this.trimMinutes = 0;
      this.trimSeconds = 10;
      this.trimDialogVisible = true;
    },

    async confirmTrim() {
      const duration = this.trimMinutes * 60 + this.trimSeconds;
      if (duration <= 0) {
        ElMessage.warning('剪切时长必须大于 0');
        return;
      }
      this.trimDialogVisible = false;
      shared.loadingText = this.trimPosition === 'start' ? '正在去除片头，请稍候...' : '正在去除片尾，请稍候...';
      shared.processing = true;
      shared.result = null;
      const fileNames = this.selectedFiles.map(f => f.name);
      const isStart = this.trimPosition === 'start';

      try {
        const req = { dirPath: this.videoDir, fileNames: fileNames, duration: duration };
        const res = isStart ? await TrimVideoStartByFiles(req) : await TrimVideoEndByFiles(req);
        shared.result = res;
        if (res.success) {
          ElMessage.success(isStart ? '去除片头完成' : '去除片尾完成');
        } else {
          ElMessage.warning('处理完成，但有失败文件');
        }
        await this.loadDirContents();
      } catch (e) {
        ElMessage.error('处理失败: ' + e);
        shared.logs.push('❌ 剪切失败: ' + e);
      } finally {
        shared.processing = false;
      }
    },

    // ===== 合并选中 =====
    startMergeSelected() {
      this.mergeOutputName = 'merged_video';
      this.mergeDialogVisible = true;
    },

    confirmMergeSelected() {
      if (!this.mergeOutputName || !this.mergeOutputName.trim()) {
        ElMessage.warning('请输入输出文件名');
        return;
      }
      this.mergeDialogVisible = false;
      shared.loadingText = '正在合并视频，请稍候...';
      shared.processing = true;
      shared.result = null;
      const fileNames = this.selectedFiles.map(f => f.name);
      MergeVideosByFiles({
        dirPath: this.videoDir,
        fileNames: fileNames,
        outputName: this.mergeOutputName.trim(),
      }).then(res => {
        shared.result = {
          success: res.success,
          message: res.message,
          outputDir: res.success ? this.videoDir : '',
        };
        if (res.success) {
          shared.logs.push(`✅ 合并完成: ${res.outputPath || ''}`);
          ElMessage.success('合并完成');
        } else {
          ElMessage.error('合并失败: ' + res.message);
        }
        return this.loadDirContents();
      }).catch(e => {
        ElMessage.error('合并失败: ' + e);
        shared.logs.push('❌ 合并失败: ' + e);
      }).finally(() => {
        shared.processing = false;
      });
    },

    // ===== 批量合并（按子目录） =====
    startBatchMerge() {
      ElMessageBox.confirm(
        `将对 ${this.subDirs.length} 个子目录中的视频分别合并，是否继续？`,
        '批量合并',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      ).then(() => {
        shared.loadingText = '正在批量合并视频，请稍候...';
        shared.processing = true;
        shared.result = null;
        return MergeVideos({ dirPath: this.videoDir, ignoreSort: false });
      }).then(res => {
        shared.result = res;
        if (res.success) {
          ElMessage.success('批量合并完成');
        } else {
          ElMessage.warning('合并完成，但有失败目录');
        }
        return this.loadDirContents();
      }).catch(e => {
        if (e === 'cancel' || e === 'close') return;
        ElMessage.error('合并失败: ' + e);
        shared.logs.push('❌ 批量合并失败: ' + e);
      }).finally(() => {
        shared.processing = false;
      });
    },

    // ===== 详情（扫描分辨率等信息） =====
    async showResolutions() {
      this.loadingResolutions = true;
      shared.result = null;
      this.showResolutionColumn = true;

      try {
        const res = await ScanVideos({ dirPath: this.videoDir });
        if (res.success && res.allVideos && res.allVideos.length > 0) {
          // 将扫描到的详细信息填充到文件列表
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

          shared.logs.push(`=== 视频详情：共 ${res.allVideos.length} 个 ===`);
          const statMap = {};
          res.allVideos.forEach((video, i) => {
            const r = video.resolution || '未知';
            statMap[r] = (statMap[r] || 0) + 1;
            shared.logs.push(`${i + 1}. ${video.fileName}  ${r} | ${this.formatFrameRate(video.frameRate)} | ${video.bitRate || '-'} | ${video.codec || '-'} | ${this.formatDuration(video.duration)}`);
          });
          shared.logs.push('--- 分辨率统计 ---');
          Object.keys(statMap).sort().forEach(r => {
            shared.logs.push(`${r}: ${statMap[r]}个`);
          });
        } else {
          ElMessage.warning(res.message || '未找到视频文件');
        }
      } catch (e) {
        ElMessage.error('获取视频信息失败: ' + e);
        shared.logs.push('❌ 获取视频信息失败: ' + e);
      } finally {
        this.loadingResolutions = false;
      }
    },

    // ===== 关键帧 =====
    showKeyframes() {
      this.keyframesDialogVisible = true;
      this.keyframesLoading = true;
      this.keyframesData = null;
      GetKeyframes({
        dirPath: this.videoDir,
        fileName: this.selectedFiles[0].name,
        maxDuration: 10,
      }).then(res => {
        if (res.success) {
          this.keyframesData = res;
        } else {
          ElMessage.error('获取关键帧失败: ' + res.message);
        }
      }).catch(e => {
        ElMessage.error('获取关键帧失败: ' + e);
      }).finally(() => {
        this.keyframesLoading = false;
      });
    },

    // ===== 按分辨率分类 =====
    startClassifyByResolution() {
      ElMessageBox.confirm(
        '将按分辨率把视频分类到不同文件夹，是否继续？',
        '按分辨率分类',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      ).then(() => {
        shared.loadingText = '正在按分辨率分类，请稍候...';
        shared.processing = true;
        shared.result = null;
        return ClassifyVideosByResolution({ dirPath: this.videoDir });
      }).then(res => {
        shared.result = { ...res, outputDir: this.videoDir };
        if (res.success) {
          ElMessage.success('分类完成');
        } else {
          ElMessage.warning('分类完成，但有失败文件');
        }
        return this.loadDirContents();
      }).catch(e => {
        if (e === 'cancel' || e === 'close') return;
        ElMessage.error('分类失败: ' + e);
        shared.logs.push('❌ 分类失败: ' + e);
      }).finally(() => {
        shared.processing = false;
      });
    },
  },
};
</script>
