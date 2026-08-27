<template>
  <div class="custom-titlebar draggable" @dblclick="handleDoubleClick">
    <!-- 左侧：Logo 和 名称 -->
    <div class="titlebar-left">
      <div class="app-logo">🛠️</div>
      <div class="app-name">我的工具箱</div>
    </div>

    <!-- 中间：菜单栏 -->
    <div class="titlebar-menu">
      <el-menu
        mode="horizontal"
        :ellipsis="false"
        :default-active="currentRoute"
        @select="handleMenuSelect"
        class="titlebar-el-menu"
      >
        <el-sub-menu v-for="cat in categories" :key="cat.name" :index="cat.name">
          <template #title>{{ cat.icon }} {{ cat.name }}</template>
          <el-menu-item
            v-for="item in cat.items"
            :key="item.link"
            :index="'/' + item.link"
          >
            {{ item.icon || item.emoji }} {{ item.name || item.text }}
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </div>

    <!-- 右侧：窗口控制按钮 -->
    <div class="titlebar-controls">
      <!-- 最小化 -->
      <div class="control-btn minimize" @mousedown.prevent="handleMinimize" title="最小化">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect x="1" y="5.5" width="10" height="1" fill="currentColor"/>
        </svg>
      </div>
      <!-- 最大化/还原 -->
      <div class="control-btn maximize" @mousedown.prevent="handleMaximize" :title="isMaximized ? '还原' : '最大化'">
        <svg v-if="!isMaximized" width="12" height="12" viewBox="0 0 12 12">
          <rect x="1.5" y="1.5" width="9" height="9" fill="none" stroke="currentColor" stroke-width="1"/>
        </svg>
        <svg v-else width="12" height="12" viewBox="0 0 12 12">
          <rect x="3" y="0.5" width="8" height="8" fill="none" stroke="currentColor" stroke-width="1"/>
          <rect x="0.5" y="3" width="8" height="8" fill="#2c3e50" stroke="currentColor" stroke-width="1"/>
        </svg>
      </div>
      <!-- 关闭 -->
      <div class="control-btn close" @mousedown.prevent="handleClose" title="关闭">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <line x1="1" y1="1" x2="11" y2="11" stroke="currentColor" stroke-width="1.2"/>
          <line x1="11" y1="1" x2="1" y2="11" stroke="currentColor" stroke-width="1.2"/>
        </svg>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export default {
  name: 'CustomTitleBar',
  setup() {
    const route = useRoute();
    const router = useRouter();

    const isMaximized = ref(false);
    const menuData = ref([]);
    const currentRoute = computed(() => route.path);

    // 加载菜单配置
    const loadMenu = async () => {
      try {
        const { ReadAssetFile } = await import('../../wailsjs/go/app/App.js');
        const content = await ReadAssetFile('/config/menu.json');
        const items = JSON.parse(content);

        // 按类别分组
        const categoryMap = {};
        items.forEach(item => {
          if (!categoryMap[item.category]) {
            categoryMap[item.category] = [];
          }
          categoryMap[item.category].push(item);
        });

        // 类别配置（按优先级排序）
        const categoryOrder = ['文件工具', '视频工具', '代码编辑工具', '图像工具', '网络工具', '系统工具', '其他工具'];
        const categoryIcons = {
          '文件工具': '📁',
          '视频工具': '🎬',
          '代码编辑工具': '📄',
          '图像工具': '🖼️',
          '网络工具': '🌐',
          '系统工具': '⚙️',
          '其他工具': '📦',
        };

        const cats = [];
        categoryOrder.forEach(name => {
          if (name === '代码编辑工具') {
            // 文本/代码工具：过滤掉text工具，只保留其他文本相关工具
            const otherItems = categoryMap[name] ? categoryMap[name].filter(item => item.link !== 'text') : [];
            if (otherItems.length > 0) {
              cats.push({
                name,
                icon: categoryIcons[name] || '📂',
                items: otherItems,
              });
            }
          } else if (categoryMap[name] && categoryMap[name].length > 0) {
            cats.push({
              name,
              icon: categoryIcons[name] || '📂',
              items: categoryMap[name],
            });
          }
        });

        menuData.value = cats;
      } catch (e) {
        console.error('加载菜单失败:', e);
      }
    };

    const categories = computed(() => menuData.value);

    // 窗口控制
    const handleMinimize = () => {
      window.runtime.WindowMinimise();
    };

    // 双击标题栏最大化/还原
    const handleDoubleClick = () => {
      handleMaximize();
    };

    const handleMaximize = async () => {
      if (isMaximized.value) {
        window.runtime.WindowUnmaximise();
      } else {
        window.runtime.WindowMaximise();
      }
      isMaximized.value = !isMaximized.value;
    };

    const handleClose = () => {
      window.runtime.Quit();
    };

    // 菜单选择
    const handleMenuSelect = (index, keyPath) => {
      if (index.startsWith('/')) {
        router.push(index);
      }
    };

    // 检查窗口是否最大化
    const checkMaximized = async () => {
      try {
        isMaximized.value = await window.runtime.WindowIsMaximised();
      } catch (e) {
        console.error('检查最大化状态失败:', e);
      }
    };

    onMounted(() => {
      loadMenu();
      checkMaximized();
    });

    return {
      isMaximized,
      categories,
      currentRoute,
      handleMinimize,
      handleDoubleClick,
      handleMaximize,
      handleClose,
      handleMenuSelect,
    };
  },
};
</script>

<style>
.custom-titlebar {
  display: flex;
  align-items: center;
  height: 38px;
  border-bottom: 1px solid rgba(0,0,0,0.2);
  user-select: none;
  -webkit-app-region: drag;
}

.custom-titlebar.draggable {
  --wails-draggable: drag;
}

.custom-titlebar :deep(.titlebar-menu),
.custom-titlebar :deep(.titlebar-controls),
.custom-titlebar :deep(.el-menu) {
  -webkit-app-region: no-drag;
  pointer-events: auto;
}

.titlebar-left {
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 8px;
  height: 38px;
  box-sizing: border-box;
  flex-shrink: 0;
  min-width: 120px;
}

.app-logo {
  font-size: 22px;
  line-height: 1;
}

.app-name {
  font-size: 15px;
  font-weight: 600;
  color: #ecf0f1;
  line-height: 1;
}

.titlebar-menu {
  flex: 1;
  display: flex;
  align-items: center;
  height: 38px;
  box-sizing: border-box;
}

.titlebar-el-menu {
  border-bottom: none !important;
  background: transparent !important;
  --el-menu-text-color: #ecf0f1 !important;
  --el-menu-hover-text-color: #409eff !important;
  --el-menu-hover-bg-color: rgba(255,255,255,0.1) !important;
  height: 38px !important;
  display: flex !important;
  align-items: center !important;
}

.titlebar-el-menu .el-menu-item,
.titlebar-el-menu .el-sub-menu__title {
  height: 48px !important;
  line-height: 48px !important;
  font-size: 13px !important;
  padding: 0 12px !important;
  color: #ecf0f1 !important;
  display: flex !important;
  align-items: center !important;
}

.titlebar-el-menu .el-sub-menu__title {
  gap: 4px;
}

.titlebar-el-menu .el-sub-menu .el-menu {
  max-height: 400px;
  overflow-y: auto;
}

.titlebar-controls {
  display: flex;
  align-items: center;
  height: 100%;
  flex-shrink: 0;
  margin-left: auto;
  padding-right: 0;
  position: relative;
}

.control-btn {
  width: 46px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #ecf0f1;
  transition: background-color 0.15s, color 0.15s;
}

.control-btn:hover {
  background-color: rgba(255,255,255,0.1);
  color: #409eff;
}

.control-btn.close:hover {
  background-color: #e74c3c;
  color: #fff;
}

.control-btn.close {
  margin-left: auto;
  position: relative;
  z-index: 1001;
}

/* 颜色选择器 */
.color-picker-wrapper {
  position: relative;
}

.color-btn {
  width: 46px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #ecf0f1;
}

.color-btn:hover {
  background-color: rgba(255,255,255,0.1);
}

.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 2px solid rgba(255,255,255,0.3);
}

.color-picker-dropdown {
  position: fixed;
  top: 42px;
  right: 10px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  padding: 12px;
  z-index: 10000;
  min-width: 180px;
}

.color-presets {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 12px;
}

.color-preset {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: transform 0.15s, border-color 0.15s;
}

.color-preset:hover {
  transform: scale(1.1);
  border-color: #409eff;
}

.color-input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  border-top: 1px solid #eee;
  padding-top: 8px;
}

.color-input {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  padding: 0;
}

.color-input-label {
  font-size: 12px;
  color: #666;
}
</style>
