<template>
  <div class="app-shell">
    <aside class="app-sidebar" :class="{ collapsed }">
      <div class="sidebar-header">
        <div class="sidebar-title">文本工具箱</div>
        <button class="sidebar-collapse" @click="collapsed = !collapsed" :title="collapsed ? '展开菜单' : '收起菜单'">
          <el-icon><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
        </button>
      </div>

      <div class="sidebar-tabs">
        <div class="tabs-bar">
          <button class="tab-btn" :class="{ active: activeTab === 'text' }" @click="activeTab = 'text'">
            <span>文本</span>
          </button>
          <button class="tab-btn" :class="{ active: activeTab === 'tools' }" @click="activeTab = 'tools'">
            <span>工具</span>
          </button>
        </div>
      </div>

      <nav class="sidebar-nav">
        <template v-if="activeTab === 'text'">
          <div v-if="visibleTextNav.encode.length" class="nav-group-label">编码</div>
          <router-link
            v-for="item in visibleTextNav.encode"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleTextNav.process.length" class="nav-group-label">处理</div>
          <router-link
            v-for="item in visibleTextNav.process"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleTextNav.format.length" class="nav-group-label">格式</div>
          <router-link
            v-for="item in visibleTextNav.format"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleTextNav.file.length" class="nav-group-label">文件</div>
          <router-link
            v-for="item in visibleTextNav.file"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleTextNav.codeGen.length" class="nav-group-label">代码生成</div>
          <router-link
            v-for="item in visibleTextNav.codeGen"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleTextNav.system.length" class="nav-group-label">系统工具</div>
          <router-link
            v-for="item in visibleTextNav.system"
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
          <div v-if="visibleToolsNav.video.length" class="nav-group-label">视频下载</div>
          <router-link
            v-for="item in visibleToolsNav.video"
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

          <div v-if="visibleToolsNav.music.length" class="nav-group-label">音乐</div>
          <router-link
            v-for="item in visibleToolsNav.music"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleToolsNav.generate.length" class="nav-group-label">生成</div>
          <router-link
            v-for="item in visibleToolsNav.generate"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleToolsNav.image.length" class="nav-group-label">图片</div>
          <router-link
            v-for="item in visibleToolsNav.image"
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

          <div v-if="visibleToolsNav.video2.length" class="nav-group-label">视频</div>
          <router-link
            v-for="item in visibleToolsNav.video2"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>

          <div v-if="visibleToolsNav.settings.length" class="nav-group-label">设置</div>
          <router-link
            v-for="item in visibleToolsNav.settings"
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
        <span>版本:1.0 ({{ buildTime || '加载中...' }})</span>
        <el-dropdown trigger="click" @command="handleMenuCmd" class="footer-menu">
          <el-icon class="footer-menu-icon"><Menu /></el-icon>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="all">全部</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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
  Picture, Film, Setting, Microphone,
  Edit, Calendar, Key, Tickets, Files, Coin, Promotion, Timer, Histogram,
  Connection, ChatLineRound, DocumentCopy, Switch, Crop, List, Memo,
  DataAnalysis, ScaleToOriginal, PriceTag, Iphone, Avatar, Postcard,
  Menu, Fold, Expand,
} from '@element-plus/icons-vue';
import { GetBuildTime } from './wailsjs/go/main/App.js';

export default {
  name: 'App',
  components: {
    Link, Lock, Document, Sort, Operation, Scissor, MagicStick, DataLine,
    Folder, Tools, SetUp,
    VideoCamera, VideoPlay, Monitor, Download, Headset, Search,
    Picture, Film, Setting, Microphone,
    Edit, Calendar, Key, Tickets, Files, Coin, Promotion, Timer, Histogram,
    Connection, ChatLineRound, DocumentCopy, Switch, Crop, List, Memo,
    DataAnalysis, ScaleToOriginal, PriceTag, Iphone, Avatar, Postcard,
    Menu, Fold, Expand,
  },
  created() {
    // 在子组件挂载前就生成菜单数据，保证“菜单显示设置”页无论以何种方式进入都有数据
    this.buildMenuGroups();
  },
  async mounted() {
    try {
      const t = await GetBuildTime();
      this.buildTime = (t && t.length) ? t : '未注入';
    } catch (e) {
      console.error('Failed to get build time:', e);
      this.buildTime = '调用失败';
    }
  },
  data() {
    return {
      activeTab: 'text',
      collapsed: false,
      buildTime: '',
      hiddenMenus: this.loadHiddenMenus(),
      textNav: {
        encode: [
          { path: '/textCommonEncode', label: '常用编码', icon: 'Link' },
          { path: '/textHashEncode',   label: '哈希编码', icon: 'Lock' },
          { path: '/textNormal',       label: '普通文本', icon: 'Document' },
          { path: '/crypto_encryption',label: '加解密',   icon: 'Key' },
        ],
        process: [
          { path: '/textBothEnds',     label: '文本两端', icon: 'Sort' },
          { path: '/textSort',         label: '文本排序', icon: 'Operation' },
          { path: '/textRemove',       label: '去除文本', icon: 'Scissor' },
          { path: '/TextBasicTools',    label: '文本差集', icon: 'Operation' },
        ],
        format: [
          { path: '/textFormat',        label: '格式转换', icon: 'MagicStick' },
          { path: '/textChar',          label: '字符转换', icon: 'DataLine' },
          { path: '/JsonTableView',     label: 'JSON表格', icon: 'DataAnalysis' },
          { path: '/JsonTools',         label: 'JSON工具', icon: 'Tickets' },
          { path: '/SqlTools',          label: 'SQL工具',  icon: 'Histogram' },
        ],
        file: [
          { path: '/textFile',          label: '文件处理', icon: 'Folder' },
          { path: '/rename',            label: '重命名',   icon: 'Edit' },
          { path: '/fileSplitMerge',    label: '分割合并', icon: 'Crop' },
          { path: '/fileSync',          label: '文件同步', icon: 'Connection' },
          { path: '/FileBackup',        label: '备份还原', icon: 'Folder' },
          { path: '/categorizeFiles',   label: '文件归类', icon: 'Files' },
        ],
        codeGen: [
          { path: '/tpl',               label: '模板工具', icon: 'Memo' },
          { path: '/JavaTools',         label: 'Java工具', icon: 'Coffee' },
        ],
        system: [
          { path: '/cmdExecute',        label: '命令行',   icon: 'Promotion' },
          { path: '/cmdManager',        label: '命令管理', icon: 'List' },
          { path: '/envCheck',          label: '环境检测', icon: 'ChatLineRound' },
          { path: '/envVariables',      label: '环境变量', icon: 'SetUp' },
          { path: '/shutdown',          label: '定时关机', icon: 'Timer' },
          { path: '/cronExp',           label: 'cron',    icon: 'Calendar' },
          { path: '/expresson',         label: '表达式',   icon: 'ScaleToOriginal' },
          { path: '/salary',            label: '工资计算', icon: 'Coin' },
          { path: '/timeConvert',       label: '时间转换', icon: 'Timer' },
          { path: '/idcard',            label: '卡号工具', icon: 'Postcard' },
          { path: '/systemSetting',     label: '系统设置', icon: 'Setting' },
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
        image: [
          { path: '/imageTool',     label: '图片工具',     icon: 'Picture' },
          { path: '/emojiTool',     label: 'Emoji工具',    icon: 'PictureFilled' },
          { path: '/textToImage',   label: '文字生图',     icon: 'PictureRounded' },
          { path: '/textToSpeech',  label: '文字转语音',   icon: 'Microphone' },
        ],
        video2: [
          { path: '/videoTool',     label: '视频处理',     icon: 'Film' },
        ],
        settings: [
          { path: '/systemSetting', label: '系统设置',     icon: 'Setting' },
        ],
      },
    };
  },
  computed: {
    visibleTextNav() {
      const filterFn = (items) => (items || []).filter(it => !this.hiddenMenus.includes(it.path));
      return {
        encode: filterFn(this.textNav.encode),
        process: filterFn(this.textNav.process),
        format: filterFn(this.textNav.format),
        file: filterFn(this.textNav.file),
        codeGen: filterFn(this.textNav.codeGen),
        system: filterFn(this.textNav.system),
      };
    },
    visibleToolsNav() {
      const filterFn = (items) => (items || []).filter(it => !this.hiddenMenus.includes(it.path));
      return {
        video: filterFn(this.toolsNav.video),
        music: filterFn(this.toolsNav.music),
        generate: filterFn(this.toolsNav.generate),
        image: filterFn(this.toolsNav.image),
        video2: filterFn(this.toolsNav.video2),
        settings: filterFn(this.toolsNav.settings),
      };
    },
  },
  methods: {
    loadHiddenMenus() {
      try {
        const raw = localStorage.getItem('wcj_hidden_menus');
        return raw ? JSON.parse(raw) : [];
      } catch (e) {
        return [];
      }
    },
    buildMenuGroups() {
      const textGroupDefs = [
        { key: 'encode',  label: '编码',   icon: 'Link' },
        { key: 'process', label: '处理',   icon: 'Operation' },
        { key: 'format',  label: '格式',   icon: 'MagicStick' },
        { key: 'file',    label: '文件',   icon: 'Folder' },
        { key: 'codeGen', label: '代码生成', icon: 'Memo' },
        { key: 'system',  label: '系统工具', icon: 'SetUp' },
      ];
      const toolsGroupDefs = [
        { key: 'video',    label: '视频下载', icon: 'Monitor' },
        { key: 'music',    label: '音乐',   icon: 'Headset' },
        { key: 'generate', label: '生成',   icon: 'SetUp' },
        { key: 'image',    label: '图片',   icon: 'Picture' },
        { key: 'video2',   label: '视频',   icon: 'Film' },
        { key: 'settings', label: '设置',   icon: 'Setting' },
      ];
      const groups = [];
      textGroupDefs.forEach(g => {
        const items = this.textNav[g.key] || [];
        if (items.length) groups.push({ tab: 'text', label: g.label, icon: g.icon, items: [...items] });
      });
      toolsGroupDefs.forEach(g => {
        const items = this.toolsNav[g.key] || [];
        if (items.length) groups.push({ tab: 'tools', label: g.label, icon: g.icon, items: [...items] });
      });
      window.__wcjMenuData__ = groups;
    },
    handleMenuCmd(cmd) {
      if (cmd === 'all') {
        this.buildMenuGroups();
        this.$router.push('/menuSettings');
      }
    },
  },
  watch: {
    '$route.path': {
      immediate: true,
      handler(path) {
        const toolPaths = [
          '/appGenerator',
          '/ytdlp', '/downloadList', '/m3u8TaskDownload',
          '/musicSearch', '/musicSource',
          '/imageTool', '/emojiTool', '/textToImage', '/textToSpeech',
          '/videoTool',
        ];
        const allTextPaths = [
          '/textCommonEncode','/textHashEncode','/textNormal',
          '/textBothEnds','/textSort','/textRemove',
          '/textFormat','/textChar','/textFile',
          '/timeConvert','/tpl','/categorizeFiles','/JsonTableView',
          '/JavaTools','/cmdExecute','/cmdManager','/SqlTools',
          '/TextBasicTools','/cronExp','/fileSync','/salary',
          '/idcard','/rename','/envCheck','/shutdown','/crypto_encryption',
          '/expresson','/JsonTools','/FileBackup','/fileSplitMerge',
          '/envVariables','/crypto_gen_key','/menuSettings',
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
.sidebar-footer {
  position: relative;
}
.footer-menu {
  margin-left: auto;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}
.footer-menu:hover {
  background: rgba(64, 158, 255, 0.1);
}
.footer-menu-icon {
  font-size: 16px;
  color: #909399;
}
.footer-menu :deep(.el-dropdown-menu__item) {
  font-size: 13px;
}

/* ---------- Sidebar header ---------- */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 10px;
  flex-shrink: 0;
  gap: 8px;
}
.sidebar-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sidebar-collapse {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border: none;
  background: transparent;
  color: var(--accent);
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.15s, color 0.15s;
}
.sidebar-collapse:hover {
  background: var(--accent-soft);
  color: var(--accent-hover);
}

/* ---------- Tabs (千问胶囊风格) ---------- */
.sidebar-tabs {
  padding: 4px 12px 8px;
  background: transparent;
  border-bottom: none;
  flex-shrink: 0;
}
.tabs-bar {
  display: flex;
  gap: 0;
  background: var(--bg-soft);
  border-radius: var(--radius-md);
  padding: 3px;
}
.tab-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 34px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  border-radius: 10px;
  cursor: pointer;
  transition: background-color 0.18s, color 0.18s, box-shadow 0.18s;
  user-select: none;
}
.tab-btn:hover:not(.active) {
  color: var(--text-secondary);
}
.tab-btn.active {
  background: var(--bg-card);
  color: var(--text-primary);
  font-weight: 600;
  box-shadow: var(--shadow-card);
}
.tab-btn.active::after {
  display: none;
}

/* ---------- Collapsed ---------- */
.app-sidebar {
  transition: width 0.2s ease, min-width 0.2s ease;
}
.app-sidebar.collapsed {
  width: 60px;
  min-width: 60px;
}
.app-sidebar.collapsed .sidebar-header {
  justify-content: center;
  padding: 14px 0 10px;
}
.app-sidebar.collapsed .sidebar-title,
.app-sidebar.collapsed .tab-btn span,
.app-sidebar.collapsed .nav-item span,
.app-sidebar.collapsed .nav-group-label,
.app-sidebar.collapsed .sidebar-footer > span {
  display: none;
}
.app-sidebar.collapsed .tabs-bar {
  flex-direction: column;
  gap: 3px;
}
.app-sidebar.collapsed .tab-btn {
  height: 36px;
}
.app-sidebar.collapsed .nav-item {
  justify-content: center;
  padding: 9px 0;
}
.app-sidebar.collapsed .sidebar-footer {
  justify-content: center;
}
</style>
