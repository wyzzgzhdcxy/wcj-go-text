<style scoped>
.music-source-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  box-sizing: border-box;
  overflow: hidden;
}

.input-section {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.id-input {
  flex: 1;
}

.result-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  min-height: 0;
}

.result-card {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px;
}

.result-url {
  font-size: 12px;
  color: #606266;
  margin-bottom: 8px;
  word-break: break-all;
}

.redirect-url {
  font-size: 11px;
  color: #909399;
  margin-bottom: 8px;
  word-break: break-all;
}

.result-content {
  font-size: 13px;
  color: #303133;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px;
}

.loading {
  color: #409eff;
}

.error {
  color: #f56c6c;
}

.download-btn {
  margin-top: 8px;
}
</style>

<template>
  <div class="music-source-wrapper">
    <div class="input-section">
      <el-input
        v-model="songId"
        class="id-input"
        placeholder="输入歌曲ID..."
        @keyup.enter="fetchAllSources"
      >
        <template #prepend>歌曲ID</template>
      </el-input>
      <el-button type="primary" @click="fetchAllSources" :loading="loading">
        {{ loading ? '解析中...' : '解析' }}
      </el-button>
      <span style="color: #909399; font-size: 12px; margin-left: 8px;">音乐ID获取地址：<a href="#" @click.prevent="openMusicSite" style="color: #409eff; cursor: pointer;">https://music.163.com/</a></span>
    </div>

    <div class="output-dir-section" style="margin-bottom: 16px; display: flex; gap: 10px; align-items: center;">
      <span style="color: #606266; font-size: 13px;">输出目录:</span>
      <el-input v-model="outputDir" style="flex: 1; max-width: 400px;" readonly />
      <el-button type="primary" size="small" @click="selectOutputDir">选择目录</el-button>
    </div>

    <div v-if="audioFileName" style="margin-bottom: 16px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 13px;">
      <span style="color: #67c23a;">完整路径:</span> {{ outputDir }}\{{ audioFileName }}
    </div>

    <div class="result-section">
      <div
        v-for="(source, index) in sources"
        :key="index"
        class="result-card"
      >
        <div class="result-url">{{ source.url }}</div>
        <div v-if="source.redirectUrl" class="redirect-url">最终地址: {{ source.redirectUrl }}</div>
        <div v-if="source.loading" class="loading">加载中...</div>
        <div v-else-if="source.error" class="error">{{ source.error }}</div>
        <div v-else-if="source.data" class="result-content">
          <div v-if="source.audioUrl" style="margin-bottom: 8px;">
            <span style="color: #67c23a;">音频:</span> {{ source.audioUrl }}
            <span v-if="source.contentType" style="margin-left: 8px; color: #909399;">类型: {{ source.contentType }}</span>
            <span v-if="source.duration" style="margin-left: 8px; color: #909399;">时长: {{ source.duration }}</span>
            <span v-if="source.fileSize" style="margin-left: 8px; color: #909399;">大小: {{ source.fileSize }}</span>
            <el-button type="primary" size="small" style="margin-left: 12px;" @click="downloadItem(source.audioUrl, 'audio', audioFileName)">下载音频</el-button>
          </div>
          <div v-if="source.picUrl" style="margin-bottom: 8px;">
            <span style="color: #67c23a;">封面:</span> {{ source.picUrl }}
            <el-button type="primary" size="small" style="margin-left: 12px;" @click="downloadItem(source.picUrl, 'pic', audioFileName)">下载封面</el-button>
          </div>
          <div v-if="source.lrcUrl" style="margin-bottom: 8px;">
            <span style="color: #67c23a;">歌词:</span> {{ source.lrcUrl }}
            <el-button type="primary" size="small" style="margin-left: 12px;" @click="downloadItem(source.lrcUrl, 'lrc', audioFileName)">下载歌词</el-button>
          </div>
          <div>{{ source.data }}</div>
        </div>
        <el-button
          v-if="source.redirectUrl && index >= 2 && !source.audioUrl"
          type="success"
          size="small"
          class="download-btn"
          @click="downloadItem(source.redirectUrl, 'audio', audioFileName)"
        >
          下载音频
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { DownloadFile, GetFileInfo, GetRedirectUrl, OpenUrl, SelectDirectory } from "../wailsjs/go/main/App.js";

export default {
  name: "MusicSource",
  data() {
    return {
      songId: "",
      sources: [],
      loading: false,
      audioFileName: "",
      outputDir: "D:\\音乐",
    };
  },
  methods: {
    async fetchAllSources() {
      if (!this.songId) {
        this.$message.warning("请输入歌曲ID");
        return;
      }

      const baseUrls = [
        { url: "https://api.injahow.cn/meting/?type=song&id={id}&raw=1", followRedirect: false },
        { url: "https://music-api.gdstudio.xyz/api.php?types=url&id={id}&source=netease", followRedirect: false },
        { url: "https://api.qijieya.cn/meting/?type=url&id={id}", followRedirect: true },
        { url: "https://meting.mikus.ink/api?server=netease&type=url&id={id}", followRedirect: true },
      ];

      this.loading = true;
      this.sources = [];

      // Initialize sources with replaced URLs
      const filledUrls = baseUrls.map(item => ({
        ...item,
        url: item.url.replace("{id}", this.songId),
        redirectUrl: null,
        loading: true,
        error: null,
        data: null
      }));

      this.sources = filledUrls;

      // Fetch all sources
      for (let i = 0; i < this.sources.length; i++) {
        await this.fetchSource(this.sources[i], i);
      }

      this.loading = false;
    },

    async fetchSource(source, index) {
      try {
        if (source.followRedirect) {
          // Use Go backend to follow redirects and get final URL
          const finalUrl = await GetRedirectUrl(source.url);
          this.sources[index].redirectUrl = finalUrl || "获取失败";
          // Get file info for the redirect URL
          if (finalUrl) {
            const fileInfoStr = await GetFileInfo(finalUrl);
            if (fileInfoStr) {
              const info = JSON.parse(fileInfoStr);
              if (info.size > 0) {
                this.sources[index].fileSize = this.formatSize(info.size);
              }
              if (info.contentType) {
                this.sources[index].contentType = info.contentType;
              }
            }
          }
        } else {
          const response = await fetch(source.url);
          const text = await response.text();
          this.sources[index].data = text;
          // Try to parse JSON and extract audio URL for duration
          try {
            const json = JSON.parse(text);
            // Handle array response
            const item = Array.isArray(json) ? json[0] : json;
            if (item) {
              if (item.url) {
                this.sources[index].audioUrl = item.url;
                // Get duration asynchronously
                this.getAudioDuration(item.url, index);
              }
              if (item.pic) {
                this.sources[index].picUrl = item.pic;
              }
              if (item.lrc) {
                this.sources[index].lrcUrl = item.lrc;
              }
              // Store artist and name for filename
              if (item.artist || item.name) {
                this.sources[index].artist = item.artist || "";
                this.sources[index].name = item.name || "";
                // Set shared audio filename from first source
                if (index === 0) {
                  this.audioFileName = `${item.artist || ""}-${item.name || ""}`.replace(/["/\\]/g, "");
                }
              }
            }
          } catch (e) {
            // Not JSON, ignore
          }
        }
        this.sources[index].loading = false;
      } catch (error) {
        this.sources[index].error = "请求失败: " + error.message;
        this.sources[index].loading = false;
      }
    },

    async getAudioDuration(url, index) {
      try {
        // Get file info using HEAD request
        const fileInfoStr = await GetFileInfo(url);
        if (fileInfoStr) {
          const info = JSON.parse(fileInfoStr);
          if (info.size > 0) {
            this.sources[index].fileSize = this.formatSize(info.size);
          }
          if (info.contentType) {
            this.sources[index].contentType = info.contentType;
          }
          if (info.filename) {
            this.sources[index].filename = info.filename;
          }
        }

        // Get final URL for audio element
        const audio = new Audio();
        audio.preload = "metadata";
        audio.src = url;
        audio.onloadedmetadata = () => {
          const duration = audio.duration;
          if (isFinite(duration)) {
            const mins = Math.floor(duration / 60);
            const secs = Math.floor(duration % 60);
            this.sources[index].duration = `${mins}:${secs.toString().padStart(2, '0')}`;
          }
          audio.src = "";
        };
        audio.onerror = () => {
          this.sources[index].duration = "无法获取";
          audio.src = "";
        };
      } catch (e) {
        this.sources[index].duration = "无法获取";
      }
    },

    formatSize(bytes) {
      if (bytes < 1024) return bytes + " B";
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
      if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + " MB";
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + " GB";
    },

    openMusicSite() {
      OpenUrl("https://music.163.com/");
    },

    async selectOutputDir() {
      const dir = await SelectDirectory();
      if (dir) {
        this.outputDir = dir;
      }
    },

    async downloadItem(url, type_, filename) {
      if (!url) return;
      console.log("Download params:", { url, type_, filename, outputDir: this.outputDir });
      try {
        await DownloadFile(url, type_, filename, this.outputDir);
        this.$message.success("下载成功");
      } catch (e) {
        console.error("Download error:", e);
        this.$message.error("下载失败: " + e.message);
      }
    },
  },
};
</script>
