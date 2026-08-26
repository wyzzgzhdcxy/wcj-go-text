<style scoped>
.music-search-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
  box-sizing: border-box;
  overflow: hidden;
}

.search-section {
  padding: 16px;
  display: flex;
  gap: 10px;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
  align-items: center;
}

.search-input {
  flex: 1;
}

.output-dir-section {
  display: flex;
  gap: 10px;
  align-items: center;
}

.search-results {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 16px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  row-gap: 0px;
  column-gap: 10px;
  align-content: start;
}

.result-card {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.result-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.result-card.selected {
  border-color: #67c23a;
  background: #f0f9eb;
}

.song-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.song-name {
  font-size: 14px;
  font-weight: bold;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.song-artist {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80px;
}

.song-duration {
  font-size: 12px;
  color: #606266;
  margin-left: auto;
  flex-shrink: 0;
}

.source-tag {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  flex-shrink: 0;
}

.source-tag.netease {
  background: #ec414d;
  color: white;
}

.source-tag.qqmusic {
  background: #31c27c;
  color: white;
}

.loading {
  text-align: center;
  color: #409eff;
  padding: 20px;
}

.empty-tip {
  text-align: center;
  color: #909399;
  padding: 40px 20px;
  font-size: 14px;
}

.source-section {
  padding: 16px;
  border-top: 1px solid #e4e7ed;
  flex-shrink: 0;
  max-height: 40%;
  overflow-y: auto;
}

.source-title {
  font-size: 14px;
  font-weight: bold;
  margin-bottom: 12px;
  color: #303133;
}

.source-item {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px 12px;
  margin-bottom: 8px;
  overflow-x: hidden;
}

.source-url {
  font-size: 12px;
  color: #606266;
  word-break: break-all;
  overflow-wrap: break-word;
  margin-bottom: 4px;
}

.source-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}
</style>

<template>
  <div class="music-search-wrapper">
    <div class="search-section">
      <el-input
        v-model="searchKeyword"
        class="search-input"
        placeholder="输入歌曲名称搜索..."
        @keyup.enter="searchMusic"
      >
        <template #prepend>歌名</template>
        <template #append>
          <el-button @click="searchMusic" :loading="searchLoading">搜索</el-button>
        </template>
      </el-input>
      <div class="output-dir-section" style="margin-left: 16px;">
        <span style="color: #606266; font-size: 13px; flex-shrink: 0;">输出目录:</span>
        <el-input v-model="outputDir" style="flex: 1; max-width: 300px;" readonly />
        <el-button type="primary" size="small" @click="selectOutputDir">选择目录</el-button>
      </div>
    </div>

    <div class="search-results">
      <div v-if="searchLoading" class="loading">搜索中...</div>
      <div v-else-if="searchResults.length === 0 && searched" class="empty-tip">
        未找到相关歌曲
      </div>
      <div
        v-for="song in searchResults"
        :key="song.id + song.source"
        class="result-card"
        :class="{ selected: selectedSong && selectedSong.id === song.id && selectedSong.source === song.source }"
        @click="selectSong(song)"
      >
        <div class="song-info">
          <span class="source-tag" :class="song.source">{{ song.source === 'netease' ? '网易云' : 'QQ音乐' }}</span>
          <span class="song-name">{{ song.name }}</span>
          <span class="song-artist">{{ song.artist.length > 10 ? '' : song.artist }}</span>
          <span class="song-duration">{{ formatDuration(song.duration) }}</span>
        </div>
      </div>
    </div>

    <div v-if="selectedSong" class="source-section">
      <div class="source-title">已选择: {{ selectedSong.name }} - {{ selectedSong.artist }}</div>
      <div
        v-for="source in validSources"
        :key="source.url"
        class="source-item"
      >
        <div class="source-url">{{ source.url }}</div>
        <div v-if="source.loading" class="loading" style="padding: 4px;">加载中...</div>
        <div v-else-if="source.error" style="color: #f56c6c; font-size: 12px;">{{ source.error }}</div>
        <div v-else-if="source.data && source.fileSizeBytes >= 1048576" style="font-size: 12px; color: #67c23a;">
          {{ source.audioUrl || '解析失败' }}
        </div>
        <div v-if="source.audioUrl && source.fileSizeBytes >= 1048576" class="source-actions">
          <el-button type="primary" size="small"
            @click="downloadItem(source, 'audio', getFileName())"
            :disabled="source.downloading">
            {{ source.downloading ? '下载中' : (source.downloaded ? '下载完成' : '下载音频') }}{{ source.fileSize && !source.downloaded ? ' (' + source.fileSize + ')' : '' }}
          </el-button>
          <el-button size="small" @click="downloadItem(source, 'pic', getFileName())" :disabled="!source.picUrl || source.downloading">下载封面</el-button>
          <el-button size="small" @click="downloadItem(source, 'lrc', getFileName())" :disabled="!source.lrcUrl || source.downloading">下载歌词</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { DownloadFile, SelectDirectory, SaveMusicAudioSource } from "../wailsjs/go/main/App.js";

export default {
  name: "MusicSearch",
  data() {
    return {
      searchKeyword: "",
      searchLoading: false,
      searchResults: [],
      searched: false,
      selectedSong: null,
      sources: [],
      outputDir: "D:\\音乐",
    };
  },
  computed: {
    validSources() {
      return this.sources.filter(s => !s.audioUrl || s.fileSizeBytes >= 1048576);
    }
  },
  methods: {
    async searchMusic() {
      if (!this.searchKeyword.trim()) {
        this.$message.warning("请输入歌曲名称");
        return;
      }

      this.searchLoading = true;
      this.searchResults = [];
      this.searched = false;

      try {
        const results = await window.go.main.App.SearchMusic(this.searchKeyword);
        this.searchResults = results || [];
        this.searched = true;
      } catch (e) {
        this.$message.error("搜索失败: " + e.message);
      } finally {
        this.searchLoading = false;
      }
    },

    async selectSong(song) {
      this.selectedSong = song;
      this.sources = [];
      this.fetchAllSources();
    },

    async fetchAllSources() {
      if (!this.selectedSong) return;

      let baseUrls = [];
      if (this.selectedSong.source === 'netease') {
        baseUrls = [
          { url: `https://api.injahow.cn/meting/?type=song&id=${this.selectedSong.id}&raw=1`, followRedirect: false },
          { url: `https://music-api.gdstudio.xyz/api.php?types=url&id=${this.selectedSong.id}&source=netease`, followRedirect: false },
        ];
      } else if (this.selectedSong.source === 'qqmusic') {
        // QQ音乐使用API，通过songid获取音频URL
        const songId = this.selectedSong.id;
        baseUrls = [
          { url: `https://api.uomg.com/api/song.url?id=${songId}&format=json`, followRedirect: false },
          { url: `https://music-api.gdstudio.xyz/api.php?types=url&id=${songId}&source=qqmusic`, followRedirect: false },
        ];
      }

      this.sources = baseUrls.map(item => ({
        ...item,
        loading: true,
        error: null,
        data: null,
        audioUrl: null,
        picUrl: null,
        lrcUrl: null,
        fileSize: null,
        fileSizeBytes: null,
        downloading: false,
        downloaded: false,
      }));

      for (let i = 0; i < this.sources.length; i++) {
        await this.fetchSource(this.sources[i], i);
      }
    },

    async fetchSource(source, index) {
      try {
        const response = await fetch(source.url);
        const text = await response.text();
        this.sources[index].data = text;

        try {
          const json = JSON.parse(text);
          const item = Array.isArray(json) ? json[0] : json;
          if (item) {
            if (item.url) {
              this.sources[index].audioUrl = item.url;
              // 获取文件大小
              this.getFileSize(item.url, index);
            }
            if (item.pic) {
              this.sources[index].picUrl = item.pic;
            }
            if (item.lrc) {
              this.sources[index].lrcUrl = item.lrc;
            }
          }
        } catch (e) {
          // Not JSON
        }
        this.sources[index].loading = false;
      } catch (error) {
        this.sources[index].error = "请求失败: " + error.message;
        this.sources[index].loading = false;
      }
    },

    async getFileSize(url, index) {
      try {
        const fileInfoStr = await window.go.main.App.GetFileInfo(url);
        if (fileInfoStr) {
          const info = JSON.parse(fileInfoStr);
          if (info.size > 0) {
            this.sources[index].fileSizeBytes = info.size;
            this.sources[index].fileSize = this.formatSize(info.size);
          }
        }
        // 保存音频源到数据库
        const source = this.sources[index];
        const song = this.selectedSong;
        if (song && source.audioUrl) {
          try {
            await SaveMusicAudioSource(
              song.id,
              song.source,
              source.audioUrl || "",
              source.picUrl || "",
              source.lrcUrl || "",
              source.fileSizeBytes || 0
            );
          } catch (e) {
            console.log("保存音频源失败:", e);
          }
        }
      } catch (e) {
        // ignore
      }
    },

    formatDuration(ms) {
      if (!ms) return "";
      const secs = Math.floor(ms / 1000);
      const mins = Math.floor(secs / 60);
      const s = secs % 60;
      return `${mins}:${s.toString().padStart(2, "0")}`;
    },

    formatSize(bytes) {
      if (!bytes) return "";
      if (bytes < 1024) return bytes + " B";
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
      if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + " MB";
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + " GB";
    },

    getFileName() {
      const song = this.selectedSong;
      if (!song) return "";
      return `${song.artist || ""}-${song.name || ""}`.replace(/["/\\:|*?<>]/g, "");
    },

    async selectOutputDir() {
      const dir = await SelectDirectory();
      if (dir) {
        this.outputDir = dir;
      }
    },

    async downloadItem(source, type_, filename) {
      const url = source[type_ + 'Url'];
      if (!url) return;
      source.downloading = true;
      try {
        await DownloadFile(url, type_, filename, this.outputDir);
        source.downloaded = true;
        this.$message.success("下载成功");
      } catch (e) {
        this.$message.error("下载失败: " + e.message);
      } finally {
        source.downloading = false;
      }
    },
  },
};
</script>
