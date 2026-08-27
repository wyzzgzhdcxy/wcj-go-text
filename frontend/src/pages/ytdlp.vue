<style scoped>
.ytdlp-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  box-sizing: border-box;
  overflow: hidden;
}

.url-section {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.url-input {
  flex: 1;
}

.monitor-btn {
  width: 40px;
}

.content-area {
  display: flex;
  gap: 16px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.left-panel {
  flex: 2;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.thumbnail-section {
  display: flex;
  gap: 12px;
}

.thumbnail-box {
  width: 200px;
  height: 120px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
}

.thumbnail-box img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.thumbnail-placeholder {
  color: #909399;
  font-size: 12px;
}

.video-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.title-text {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desc-text {
  font-size: 12px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  line-clamp: 3;
}

.format-section {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px;
}

.format-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.format-select {
  width: 100%;
}

.format-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.download-section {
  display: flex;
  gap: 10px;
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.path-section {
  display: flex;
  gap: 8px;
  align-items: center;
}

.path-input {
  flex: 1;
}

.progress-section {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px;
}

.progress-bar-wrapper {
  height: 20px;
  background: #f0f2f5;
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #409eff, #66b1ff);
  border-radius: 10px;
  transition: width 0.3s ease;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
}

.options-section {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.option-label {
  font-size: 12px;
  color: #606266;
  min-width: 70px;
}

.status-section {
  font-size: 12px;
  color: #606266;
  margin-top: 8px;
}
</style>

<template>
  <div class="ytdlp-wrapper">
    <!-- URL Section -->
    <div class="url-section">
      <el-input
        v-model="url"
        class="url-input"
        placeholder="输入视频URL..."
        @keyup.enter="analyzeVideo"
      >
        <template #prepend>URL</template>
      </el-input>
      <el-button
        class="monitor-btn"
        :type="isMonitor ? 'primary' : 'default'"
        @click="toggleMonitor"
        title="监控剪贴板"
      >
        {{ isMonitor ? '监控中' : '监控' }}
      </el-button>
      <el-button type="primary" @click="analyzeVideo" :loading="isAnalyzing">
        {{ isAnalyzing ? '分析中...' : '分析' }}
      </el-button>
    </div>

    <!-- Content Area -->
    <div class="content-area" v-if="videoInfo">
      <!-- Left Panel -->
      <div class="left-panel">
        <!-- Thumbnail and Info -->
        <div class="thumbnail-section">
          <div class="thumbnail-box">
            <img v-if="videoInfo.thumbnail" :src="videoInfo.thumbnail" alt="thumbnail" />
            <span v-else class="thumbnail-placeholder">无缩略图</span>
          </div>
          <div class="video-info">
            <div class="title-text">{{ videoInfo.title }}</div>
            <div class="desc-text">{{ videoInfo.description }}</div>
            <div class="status-section">
              <span v-if="videoInfo.is_live" style="color: #f56c6c;">直播</span>
              <span v-else>时长: {{ formatDuration(videoInfo.duration) }}</span>
            </div>
          </div>
        </div>

        <!-- Video Format -->
        <div class="format-section" v-if="videoFormats.length > 0">
          <div class="format-label">视频格式</div>
          <el-select
            v-model="selectedVideoFormat"
            class="format-select"
            placeholder="选择视频格式"
          >
            <el-option
              v-for="fmt in videoFormats"
              :key="fmt.format_id"
              :label="getFormatLabel(fmt)"
              :value="fmt.format_id"
            />
          </el-select>
        </div>

        <!-- Audio Format -->
        <div class="format-section" v-if="audioFormats.length > 0">
          <div class="format-label">音频格式</div>
          <el-select
            v-model="selectedAudioFormat"
            class="format-select"
            placeholder="选择音频格式"
          >
            <el-option
              v-for="fmt in audioFormats"
              :key="fmt.format_id"
              :label="getFormatLabel(fmt)"
              :value="fmt.format_id"
            />
          </el-select>
        </div>

        <!-- Chapters -->
        <div class="format-section" v-if="videoInfo.chapters && videoInfo.chapters.length > 0">
          <div class="format-label">章节</div>
          <el-select v-model="selectedChapter" class="format-select" placeholder="选择章节">
            <el-option
              v-for="(ch, idx) in videoInfo.chapters"
              :key="idx"
              :label="`${ch.title} (${formatTime(ch.start_time)} - ${formatTime(ch.end_time)})`"
              :value="idx"
            />
          </el-select>
        </div>

        <!-- Subtitles -->
        <div class="format-section" v-if="hasSubtitles">
          <div class="format-label">字幕</div>
          <el-select v-model="selectedSubtitle" class="format-select" placeholder="选择字幕">
            <el-option label="不下载字幕" value="" />
            <el-option
              v-for="(exts, lang) in videoInfo.subtitles"
              :key="lang"
              :label="`${lang} (${exts.join(', ')})`"
              :value="lang"
            />
          </el-select>
        </div>
      </div>

      <!-- Right Panel -->
      <div class="right-panel">
        <!-- Progress -->
        <div class="progress-section" v-if="isDownloading || downloadProgress">
          <div class="progress-bar-wrapper">
            <div class="progress-bar" :style="{ width: downloadProgress + '%' }"></div>
          </div>
          <div class="progress-info">
            <span>{{ downloadSpeed }}</span>
            <span>{{ downloadProgress.toFixed(1) }}%</span>
            <span>ETA: {{ downloadETA }}</span>
          </div>
        </div>

        <!-- Download Path -->
        <div class="format-section">
          <div class="format-label">下载路径</div>
          <div class="path-section">
            <el-input v-model="downloadPath" class="path-input" readonly />
            <el-button @click="selectDownloadPath">浏览</el-button>
          </div>
        </div>

        <!-- Options -->
        <div class="options-section">
          <div class="format-label">选项</div>

          <div class="option-item">
            <span class="option-label">代理</span>
            <el-switch v-model="options.proxyEnabled" />
            <el-input
              v-if="options.proxyEnabled"
              v-model="options.proxy"
              placeholder="http://proxy:port"
              style="flex: 1;"
            />
          </div>

          <div class="option-item">
            <span class="option-label">限速</span>
            <el-input v-model="options.limitRate" placeholder="如: 1M" style="flex: 1;" />
          </div>

          <div class="option-item">
            <span class="option-label">Cookie</span>
            <el-select v-model="options.cookieType" style="flex: 1;">
              <el-option label="不用Cookie" value="" />
              <el-option label="Chrome" value="chrome" />
              <el-option label="Edge" value="edge" />
              <el-option label="Firefox" value="firefox" />
            </el-select>
          </div>

          <div class="option-item">
            <span class="option-label">嵌入</span>
            <el-checkbox v-model="options.embedThumbnail">缩略图</el-checkbox>
            <el-checkbox v-model="options.embedChapters">章节</el-checkbox>
            <el-checkbox v-model="options.embedSubtitles">字幕</el-checkbox>
          </div>
        </div>

        <!-- Download Buttons -->
        <div class="download-section">
          <el-button
            type="primary"
            @click="startDownload"
            :disabled="!canDownload || isDownloading"
            style="flex: 1;"
          >
            {{ isDownloading ? '下载中...' : '下载' }}
          </el-button>
          <el-button v-if="isDownloading" type="danger" @click="cancelDownload">
            取消
          </el-button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!isAnalyzing" class="content-area" style="align-items: center; justify-content: center;">
      <div style="text-align: center; color: #909399;">
        <div style="font-size: 48px; margin-bottom: 16px;">📺</div>
        <div>输入视频URL开始分析</div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else class="content-area" style="align-items: center; justify-content: center;">
      <el-loading-spinner />
    </div>
  </div>
</template>

<script>
import {
  GetYtDlpVideoInfo,
  DownloadWithYtDlp,
  CancelYtDlpDownload,
  GetYtDlpDownloadPath,
  CheckYtDlpDeps,
  SelectDirectory,
} from "../wailsjs/go/app/App.js";

export default {
  name: "Ytdlp",
  data() {
    return {
      url: "",
      isMonitor: false,
      isAnalyzing: false,
      isDownloading: false,
      videoInfo: null,
      selectedVideoFormat: "",
      selectedAudioFormat: "",
      selectedChapter: null,
      selectedSubtitle: "",
      downloadPath: "",
      downloadProgress: 0,
      downloadSpeed: "",
      downloadETA: "",
      options: {
        proxyEnabled: false,
        proxy: "",
        limitRate: "",
        cookieType: "",
        embedThumbnail: false,
        embedChapters: false,
        embedSubtitles: false,
      },
      depsChecked: false,
      depsAvailable: false,
    };
  },
  computed: {
    videoFormats() {
      if (!this.videoInfo || !this.videoInfo.formats) return [];
      return this.videoInfo.formats.filter(
        (f) => f.type === "video" || f.type === "package"
      );
    },
    audioFormats() {
      if (!this.videoInfo || !this.videoInfo.formats) return [];
      return this.videoInfo.formats.filter(
        (f) => f.type === "audio" || f.type === "package"
      );
    },
    hasSubtitles() {
      return (
        this.videoInfo &&
        this.videoInfo.subtitles &&
        Object.keys(this.videoInfo.subtitles).length > 0
      );
    },
    canDownload() {
      return (
        this.url &&
        this.videoInfo &&
        (this.selectedVideoFormat || this.selectedAudioFormat) &&
        this.downloadPath
      );
    },
  },
  async mounted() {
    // Get default download path
    try {
      this.downloadPath = await GetYtDlpDownloadPath();
    } catch (e) {
      console.error("Failed to get download path:", e);
    }

    // Check dependencies
    try {
      const deps = await CheckYtDlpDeps();
      this.depsAvailable = deps["yt-dlp"] && deps["ffmpeg"];
      if (!this.depsAvailable) {
        this.$message.warning("未检测到 yt-dlp 或 ffmpeg，请确保它们已安装并加入PATH");
      }
    } catch (e) {
      console.error("Failed to check deps:", e);
    }

    // Listen for progress events
    window.runtime.EventsOn("ytdlp_progress", (data) => {
      try {
        if (typeof data === 'string') {
          data = JSON.parse(data);
        }
        this.downloadProgress = data.percent || 0;
        this.downloadSpeed = data.speed || "";
        this.downloadETA = data.eta || "";
        if (data.status === "finished") {
          this.isDownloading = false;
          this.$message.success("下载完成！");
        }
      } catch (e) {
        console.error("Failed to parse progress:", e);
      }
    });

    window.runtime.EventsOn("ytdlp_error", (data) => {
      try {
        if (typeof data === 'string') {
          data = JSON.parse(data);
        }
        this.isDownloading = false;
        this.$message.error("下载失败: " + (data.error || "未知错误"));
      } catch (e) {
        console.error("Failed to parse error:", e);
      }
    });

    // Monitor clipboard
    this.startClipboardMonitor();
  },
  methods: {
    async analyzeVideo() {
      if (!this.url) {
        this.$message.warning("请输入URL");
        return;
      }

      this.isAnalyzing = true;
      this.videoInfo = null;
      this.selectedVideoFormat = "";
      this.selectedAudioFormat = "";

      try {
        const result = await GetYtDlpVideoInfo(this.url);
        if (result) {
          this.videoInfo = JSON.parse(result);
          // Auto-select best formats
          if (this.videoFormats.length > 0) {
            this.selectedVideoFormat = this.videoFormats[0].format_id;
          }
          if (this.audioFormats.length > 0) {
            this.selectedAudioFormat = this.audioFormats[0].format_id;
          }
        } else {
          this.$message.error("获取视频信息失败");
        }
      } catch (e) {
        console.error("Analyze error:", e);
        this.$message.error("分析失败: " + e.message);
      } finally {
        this.isAnalyzing = false;
      }
    },

    toggleMonitor() {
      this.isMonitor = !this.isMonitor;
      this.$message.info(this.isMonitor ? "已开启剪贴板监控" : "已关闭剪贴板监控");
    },

    startClipboardMonitor() {
      // Poll clipboard for URL changes
      this.clipboardTimer = setInterval(async () => {
        if (!this.isMonitor || this.isAnalyzing || this.isDownloading) return;
      }, 1000);
    },

    async selectDownloadPath() {
      const dir = await SelectDirectory();
      if (dir) {
        this.downloadPath = dir;
      }
    },

    async startDownload() {
      if (!this.canDownload) return;

      this.isDownloading = true;
      this.downloadProgress = 0;

      const formatId = this.selectedVideoFormat || this.selectedAudioFormat;
      const outputPath = this.downloadPath + "/%(title)s.%(ext)s";

      const opts = {
        format_id: formatId,
        output_path: outputPath,
        proxy: this.options.proxy,
        proxy_enabled: this.options.proxyEnabled,
        cookie_type: this.options.cookieType,
        use_cookie: !!this.options.cookieType,
        embed_thumbnail: this.options.embedThumbnail,
        embed_chapters: this.options.embedChapters,
        embed_subtitles: this.options.embedSubtitles,
        limit_rate: this.options.limitRate,
        time_range: this.selectedChapter !== null ? `${this.videoInfo.chapters[this.selectedChapter]?.start_time}-${this.videoInfo.chapters[this.selectedChapter]?.end_time}` : "",
        extract_audio: !!this.selectedAudioFormat && !this.selectedVideoFormat,
        audio_format: "mp3",
      };

      try {
        await DownloadWithYtDlp(this.url, opts);
      } catch (e) {
        this.isDownloading = false;
        this.$message.error("启动下载失败: " + e.message);
      }
    },

    cancelDownload() {
      try {
        CancelYtDlpDownload();
        this.isDownloading = false;
        this.$message.info("已取消下载");
      } catch (e) {
        console.error("Cancel error:", e);
      }
    },

    getFormatLabel(fmt) {
      const parts = [];
      if (fmt.resolution) parts.push(fmt.resolution);
      if (fmt.ext) parts.push(fmt.ext);
      if (fmt.vcodec && fmt.vcodec !== "none") parts.push(fmt.vcodec);
      if (fmt.acodec && fmt.acodec !== "none") parts.push(fmt.acodec);
      if (fmt.filesize) parts.push(`(${fmt.filesize})`);
      return parts.join(" ");
    },

    formatDuration(seconds) {
      if (!seconds) return "0:00";
      const h = Math.floor(seconds / 3600);
      const m = Math.floor((seconds % 3600) / 60);
      const s = Math.floor(seconds % 60);
      if (h > 0) {
        return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
      }
      return `${m}:${s.toString().padStart(2, "0")}`;
    },

    formatTime(seconds) {
      return this.formatDuration(seconds);
    },
  },
  beforeUnmount() {
    if (this.clipboardTimer) {
      clearInterval(this.clipboardTimer);
    }
  },
};
</script>
