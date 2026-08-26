<template>
  <div class="app-shell">
    <aside class="app-sidebar">
      <div class="sidebar-brand">
        <div class="brand-logo">文</div>
        <div>
          <div class="brand-title">文本工具箱</div>
          <div class="brand-subtitle">Text Toolkit</div>
        </div>
      </div>

      <div class="sidebar-tabs">
        <div class="tabs-bar">
          <button class="tab-btn" :class="{ active: activeTab === 'text' }" @click="activeTab = 'text'">
            <el-icon><Document /></el-icon>
            <span>文本</span>
          </button>
          <button class="tab-btn" :class="{ active: activeTab === 'tools' }" @click="activeTab = 'tools'">
            <el-icon><Tools /></el-icon>
            <span>工具</span>
          </button>
        </div>
      </div>

      <nav class="sidebar-nav">
        <template v-if="activeTab === 'text'">
          <div class="nav-group-label">编码</div>
          <router-link
            v-for="item in textNav.encode"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div class="nav-group-label">处理</div>
          <router-link
            v-for="item in textNav.process"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div class="nav-group-label">格式</div>
          <router-link
            v-for="item in textNav.format"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div class="nav-group-label">文件</div>
          <router-link
            v-for="item in textNav.file"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>
        </template>

        <template v-else>
          <div class="nav-group-label">视频下载</div>
          <router-link
            v-for="item in toolsNav.video"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon">
              <component :is="item.icon" v-if="item.icon" />
              <span v-else class="nav-emoji">{{ item.emoji }}</span>
            </el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div class="nav-group-label">音乐</div>
          <router-link
            v-for="item in toolsNav.music"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div class="nav-group-label">生成</div>
          <router-link
            v-for="item in toolsNav.generate"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>
        </template>
      </nav>

      <div class="sidebar-footer">
        <span class="dot"></span>
        <span>版本:1.0</span>
      </div>
    </aside>

    <main class="app-main">
      <div class="app-body">
        <router-view></router-view>
      </div>
    </main>
  </div>
</template>

<script>
import {
  Link, Lock, Document, Sort, Operation, Scissor, MagicStick, DataLine,
  Folder, Tools, SetUp,
  VideoCamera, VideoPlay, Monitor, Download, Headset, Search,
} from '@element-plus/icons-vue';

export default {
  name: 'App',
  components: {
    Link, Lock, Document, Sort, Operation, Scissor, MagicStick, DataLine,
    Folder, Tools, SetUp,
    VideoCamera, VideoPlay, Monitor, Download, Headset, Search,
  },
  data() {
    return {
      activeTab: 'text',
      textNav: {
        encode: [
          { path: '/textCommonEncode', label: '常用编码', icon: 'Link' },
          { path: '/textHashEncode',   label: '哈希编码', icon: 'Lock' },
          { path: '/textNormal',       label: '普通文本', icon: 'Document' },
        ],
        process: [
          { path: '/textBothEnds', label: '文本两端', icon: 'Sort' },
          { path: '/textSort',     label: '文本排序', icon: 'Operation' },
          { path: '/textRemove',   label: '去除文本', icon: 'Scissor' },
        ],
        format: [
          { path: '/textFormat', label: '格式转换', icon: 'MagicStick' },
          { path: '/textChar',   label: '字符转换', icon: 'DataLine' },
        ],
        file: [
          { path: '/textFile', label: '文件处理', icon: 'Folder' },
        ],
      },
      toolsNav: {
        video: [
          { path: '/ytdlp',            label: 'yt-dlp 下载', icon: 'Monitor' },
          { path: '/downloadList',     label: 'B站视频', icon: 'Download' },
          { path: '/m3u8TaskDownload', label: 'M3U8 下载', icon: 'VideoPlay' },
        ],
        music: [
          { path: '/musicSearch', label: '音乐搜索', icon: 'Search' },
          { path: '/musicSource', label: '音乐解析', icon: 'Headset' },
        ],
        generate: [
          { path: '/appGenerator', label: '应用生成', icon: 'SetUp' },
        ],
      },
    };
  },
  watch: {
    '$route.path': {
      immediate: true,
      handler(path) {
        const toolPaths = [
          '/appGenerator',
          '/ytdlp', '/downloadList', '/m3u8TaskDownload',
          '/musicSearch', '/musicSource',
        ];
        if (toolPaths.includes(path)) this.activeTab = 'tools';
      },
    },
  },
};
</script>

<style scoped>
.nav-emoji {
  font-size: 18px;
  line-height: 1;
}
</style>
