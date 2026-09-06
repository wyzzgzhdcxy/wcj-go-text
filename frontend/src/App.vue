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
          <template v-for="g in visibleTextGroups" :key="g.key">
            <div v-if="g.items.length" class="nav-group-label">{{ g.label }}</div>
            <router-link
              v-for="item in g.items"
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
          </template>
        </template>

        <template v-else>
          <template v-for="g in visibleToolsGroups" :key="g.key">
            <div v-if="g.items.length" class="nav-group-label">{{ g.label }}</div>
            <router-link
              v-for="item in g.items"
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
          </template>
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
import { GetBuildTime } from './wailsjs/go/app/App.js';

// 侧边栏分组定义
const TEXT_GROUPS = [
  { key: 'encode',  label: '编码',   icon: 'Link' },
  { key: 'process', label: '处理',   icon: 'Operation' },
  { key: 'format',  label: '格式',   icon: 'MagicStick' },
  { key: 'file',    label: '文件',   icon: 'Folder' },
  { key: 'codeGen', label: '代码生成', icon: 'Memo' },
  { key: 'system',  label: '系统工具', icon: 'SetUp' },
];
const TOOLS_GROUPS = [
  { key: 'video',    label: '视频下载', icon: 'Monitor' },
  { key: 'music',    label: '音乐',   icon: 'Headset' },
  { key: 'generate', label: '生成',   icon: 'SetUp' },
  { key: 'image',    label: '图片',   icon: 'Picture' },
  { key: 'video2',   label: '视频',   icon: 'Film' },
  { key: 'settings', label: '设置',   icon: 'Setting' },
];

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
      activeTab: this.loadActiveTab(),
      collapsed: this.loadCollapsed(),
      buildTime: '',
      hiddenMenus: this.loadHiddenMenus(),
      textNav: {
        encode: [
          { path: '/textCommonEncode', label: '常用编码', icon: 'Link' },
          { path: '/symmetricCrypto', label: '对称加密', icon: 'Lock' },
          { path: '/crypto_encryption',label: '公钥算法',   icon: 'Key' },
        ],
        process: [
          { path: '/textBothEnds',     label: '简单文本', icon: 'Sort' },
          { path: '/TextBasicTools',    label: '文本差集', icon: 'Operation' },
        ],
        format: [
          { path: '/JsonTableView',     label: 'JSON表格', icon: 'DataAnalysis' },
          { path: '/JsonTools',         label: 'JSON工具', icon: 'Tickets' },
          { path: '/SqlTools',          label: 'SQL工具',  icon: 'Histogram' },
        ],
        file: [
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
          { path: '/cmdExecute',        label: '命令执行', icon: 'Promotion' },
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
          { path: '/barcodeTool',   label: '二维码 / 条形码', icon: 'PriceTag' },
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
    // 按分组定义顺序输出可见分组
    visibleTextGroups() {
      return TEXT_GROUPS.map(g => ({ ...g, items: this.visibleTextNav[g.key] || [] }));
    },
    visibleToolsGroups() {
      return TOOLS_GROUPS.map(g => ({ ...g, items: this.visibleToolsNav[g.key] || [] }));
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
    loadActiveTab() {
      try {
        const v = localStorage.getItem('wcj_active_tab');
        return v === 'tools' ? 'tools' : 'text';
      } catch (e) {
        return 'text';
      }
    },
    loadCollapsed() {
      try {
        return localStorage.getItem('wcj_sidebar_collapsed') === 'true';
      } catch (e) {
        return false;
      }
    },
    buildMenuGroups() {
      const groups = [];
      TEXT_GROUPS.forEach(g => {
        const items = this.textNav[g.key] || [];
        if (items.length) groups.push({ tab: 'text', label: g.label, icon: g.icon, items: [...items] });
      });
      TOOLS_GROUPS.forEach(g => {
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
          '/m3u8TaskDownload',
          '/musicSearch', '/musicSource',
          '/imageTool', '/emojiTool', '/textToImage', '/textToSpeech',
          '/barcodeTool',
          '/videoTool',
        ];
        const allTextPaths = [
          '/textCommonEncode',
          '/symmetricCrypto',
          '/textBothEnds',
          '/timeConvert','/tpl','/categorizeFiles','/JsonTableView',
          '/JavaTools','/cmdExecute','/SqlTools',
          '/TextBasicTools','/cronExp','/fileSync','/salary',
          '/idcard','/rename','/envCheck','/shutdown','/crypto_encryption',
          '/expresson','/JsonTools','/FileBackup','/fileSplitMerge',
          '/envVariables','/crypto_gen_key','/menuSettings',
        ];
        if (toolPaths.includes(path)) this.activeTab = 'tools';
      },
    },
    activeTab(v) {
      try { localStorage.setItem('wcj_active_tab', v); } catch (e) { /* ignore */ }
    },
    collapsed(v) {
      try { localStorage.setItem('wcj_sidebar_collapsed', String(!!v)); } catch (e) { /* ignore */ }
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
  padding: 6px 10px;
  border-radius: 4px;
}
.footer-menu:hover {
  background: rgba(64, 158, 255, 0.1);
}
.footer-menu-icon {
  font-size: 26px;
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
  padding: 14px 0 10px 16px;
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
  justify-content: flex-end;
  padding: 14px 8px 10px 0;
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
  padding: 5px 0;
}
.app-sidebar.collapsed .nav-item .nav-icon {
  width: 26px;
  height: 26px;
  font-size: 22px;
}
.app-sidebar.collapsed .tab-btn .el-icon {
  font-size: 22px;
}
.app-sidebar.collapsed .footer-menu {
  padding: 6px 10px;
}
.app-sidebar.collapsed .footer-menu-icon {
  font-size: 28px;
}
.app-sidebar.collapsed .sidebar-footer {
  justify-content: flex-end;
  padding-right: 8px;
}
</style>
