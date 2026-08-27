<template>
  <div class="page">
    <div class="card">
      <div class="card-head">
        <div class="card-title">菜单显示设置</div>
        <div class="card-actions">
          <button class="chip" @click="showAll">全部显示</button>
          <button class="chip" @click="hideAll">全部隐藏</button>
          <button class="chip primary" @click="saveSettings">保存</button>
        </div>
      </div>
      <div class="card-body">
        <div v-for="group in menuGroups" :key="group.tab + group.label" class="menu-group">
          <div class="group-head">
            <el-icon class="group-icon"><component :is="group.icon" /></el-icon>
            <span class="group-label">{{ group.tab === 'text' ? '文本' : '工具' }} · {{ group.label }}</span>
            <span class="group-count">({{ visibleCount(group) }}/{{ group.items.length }})</span>
          </div>
          <div class="checkbox-grid">
            <label
              v-for="item in group.items"
              :key="item.path"
              class="menu-item"
              :class="{ checked: !hiddenList.includes(item.path) }"
            >
              <input
                type="checkbox"
                :checked="!hiddenList.includes(item.path)"
                @change="toggleItem(item.path)"
                class="native-cb"
              />
              <el-icon class="item-icon" v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span class="item-label" :title="item.label">{{ item.label }}</span>
              <span class="item-path" :title="item.path">{{ item.path }}</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus';
import {
  Link, Lock, Document, Sort, Operation, Scissor, MagicStick, DataLine,
  Folder, Tools, SetUp,
  VideoCamera, VideoPlay, Monitor, Download, Headset, Search,
  Picture, Film, Setting, Microphone,
  Edit, Calendar, Key, Tickets, Files, Coin, Promotion, Timer, Histogram,
  Connection, ChatLineRound, DocumentCopy, Switch, Crop, List, Memo,
  DataAnalysis, ScaleToOriginal, PriceTag, Iphone, Avatar, Postcard,
} from '@element-plus/icons-vue';

export default {
  name: 'MenuSettings',
  components: {
    Link, Lock, Document, Sort, Operation, Scissor, MagicStick, DataLine,
    Folder, Tools, SetUp,
    VideoCamera, VideoPlay, Monitor, Download, Headset, Search,
    Picture, Film, Setting, Microphone,
    Edit, Calendar, Key, Tickets, Files, Coin, Promotion, Timer, Histogram,
    Connection, ChatLineRound, DocumentCopy, Switch, Crop, List, Memo,
    DataAnalysis, ScaleToOriginal, PriceTag, Iphone, Avatar, Postcard,
  },
  data() {
    return {
      hiddenList: [],
      menuGroups: [
        { tab: 'text', label: '编码',   icon: 'Link',         items: [] },
        { tab: 'text', label: '处理',   icon: 'Operation',    items: [] },
        { tab: 'text', label: '格式',   icon: 'MagicStick',   items: [] },
        { tab: 'text', label: '文件',   icon: 'Folder',       items: [] },
        { tab: 'text', label: '代码生成', icon: 'Memo',         items: [] },
        { tab: 'text', label: '系统工具', icon: 'SetUp',        items: [] },
        { tab: 'tools', label: '视频下载', icon: 'Monitor',     items: [] },
        { tab: 'tools', label: '音乐',   icon: 'Headset',      items: [] },
        { tab: 'tools', label: '生成',   icon: 'SetUp',        items: [] },
        { tab: 'tools', label: '图片',   icon: 'Picture',      items: [] },
        { tab: 'tools', label: '视频',   icon: 'Film',         items: [] },
        { tab: 'tools', label: '设置',   icon: 'Setting',      items: [] },
      ],
    };
  },
  async mounted() {
    try {
      const raw = localStorage.getItem('wcj_hidden_menus');
      this.hiddenList = raw ? JSON.parse(raw) : [];
    } catch (e) {
      this.hiddenList = [];
    }
    this.loadGroups();
  },
  methods: {
    loadGroups() {
      if (window.__wcjMenuData__) {
        this.menuGroups = window.__wcjMenuData__;
      }
    },
    toggleItem(path) {
      const idx = this.hiddenList.indexOf(path);
      if (idx >= 0) {
        this.hiddenList.splice(idx, 1);
      } else {
        this.hiddenList.push(path);
      }
    },
    visibleCount(group) {
      return group.items.filter(it => !this.hiddenList.includes(it.path)).length;
    },
    showAll() {
      this.hiddenList = [];
    },
    hideAll() {
      const all = [];
      this.menuGroups.forEach(g => g.items.forEach(it => all.push(it.path)));
      this.hiddenList = all;
    },
    saveSettings() {
      localStorage.setItem('wcj_hidden_menus', JSON.stringify(this.hiddenList));
      ElMessage.success('保存成功，正在刷新...');
      // 延迟刷新，先让成功提示展示出来，再重载页面使菜单隐藏/显示设置立即生效
      setTimeout(() => {
        window.location.reload();
      }, 500);
    },
  },
};
</script>

<style scoped>
.page {
  width: 100%;
  height: 100%;
  padding: 16px;
  box-sizing: border-box;
  overflow-y: auto;
  overflow-x: hidden;
  display: block; /* 覆盖全局 .page 的 flex 布局，改为普通块级，配合 overflow 正常滚动 */
}
.card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  min-width: 0;
  display: block; /* 覆盖全局 .card 的 flex 布局，避免被 flex-shrink 压缩裁切内容 */
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid #ebeef5;
  flex-wrap: wrap;
  gap: 10px;
}
.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.card-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.chip {
  padding: 6px 14px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  background: #fff;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  transition: all 0.2s;
}
.chip:hover {
  border-color: #409eff;
  color: #409eff;
}
.chip.primary {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}
.chip.primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
  color: #fff;
}
.card-body {
  padding: 20px;
  min-width: 0;
  display: block; /* 覆盖全局 .card-body 的 flex 布局，否则菜单分组会被横向挤压 */
}
.menu-group {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px dashed #ebeef5;
  min-width: 0;
}
.menu-group:last-child {
  border-bottom: none;
  margin-bottom: 0;
}
.group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}
.group-icon {
  color: #409eff;
  font-size: 16px;
  flex-shrink: 0;
}
.group-label {
  letter-spacing: 0.3px;
}
.group-count {
  font-size: 12px;
  color: #909399;
  font-weight: normal;
}
.checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 8px;
  width: 100%;
}
.menu-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  min-width: 0;
  overflow: hidden;
  background: #fafafa;
}
.menu-item:hover {
  background: #f0f5ff;
  border-color: #d4e4ff;
}
.menu-item.checked {
  background: #ecf5ff;
  border-color: #b3d8ff;
}
.native-cb {
  flex-shrink: 0;
  cursor: pointer;
  margin: 0;
}
.item-icon {
  color: #409eff;
  font-size: 14px;
  flex-shrink: 0;
}
.item-label {
  font-size: 13px;
  color: #303133;
  flex-shrink: 0;
  white-space: nowrap;
}
.item-path {
  font-size: 10px;
  color: #c0c4cc;
  font-family: monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
  text-align: right;
}
</style>
