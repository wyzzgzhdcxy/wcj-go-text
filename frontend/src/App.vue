<template>
  <div id="app-layout">
    <div class="main-wrapper">
      <!-- 左侧工具导航 -->
      <div class="sidebar">
        <div class="nav-items">
          <router-link to="/textCommonEncode" class="nav-item" :class="{ active: $route.path === '/textCommonEncode' }">
            <span class="nav-icon">🔗</span>
            <span class="nav-tooltip">常用编码</span>
          </router-link>
          <router-link to="/textHashEncode" class="nav-item" :class="{ active: $route.path === '/textHashEncode' }">
            <span class="nav-icon">#️⃣</span>
            <span class="nav-tooltip">哈希编码</span>
          </router-link>
          <router-link to="/textNormal" class="nav-item" :class="{ active: $route.path === '/textNormal' }">
            <span class="nav-icon">📝</span>
            <span class="nav-tooltip">普通文本</span>
          </router-link>
          <router-link to="/textBothEnds" class="nav-item" :class="{ active: $route.path === '/textBothEnds' }">
            <span class="nav-icon">〰️</span>
            <span class="nav-tooltip">文本两端</span>
          </router-link>
          <router-link to="/textSort" class="nav-item" :class="{ active: $route.path === '/textSort' }">
            <span class="nav-icon">⬆️</span>
            <span class="nav-tooltip">文本排序</span>
          </router-link>
          <router-link to="/textRemove" class="nav-item" :class="{ active: $route.path === '/textRemove' }">
            <span class="nav-icon">✂️</span>
            <span class="nav-tooltip">去除文本</span>
          </router-link>
          <router-link to="/textFormat" class="nav-item" :class="{ active: $route.path === '/textFormat' }">
            <span class="nav-icon">🔧</span>
            <span class="nav-tooltip">格式转换</span>
          </router-link>
          <router-link to="/textChar" class="nav-item" :class="{ active: $route.path === '/textChar' }">
            <span class="nav-icon">🔀</span>
            <span class="nav-tooltip">字符转换</span>
          </router-link>
          <router-link to="/textFile" class="nav-item" :class="{ active: $route.path === '/textFile' }">
            <span class="nav-icon">📁</span>
            <span class="nav-tooltip">文件处理</span>
          </router-link>
        </div>
      </div>

      <!-- 主内容区 -->
      <div class="main-content">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  overflow: hidden;
}

#app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
}

.main-wrapper {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 左侧导航栏 */
.sidebar {
  width: 48px;
  min-width: 48px;
  background: #ffffff;
  border-right: 1px solid #e8eaed;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
  z-index: 100;
}

.nav-items {
  flex: 1;
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  scrollbar-width: none;
}

.nav-items::-webkit-scrollbar {
  display: none;
}

.nav-item {
  width: 100%;
  padding: 10px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  color: #606266;
  text-decoration: none;
  transition: background-color 0.2s;
}

.nav-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.nav-item.active {
  color: #409eff;
  background-color: #ecf5ff;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background-color: #409eff;
  border-radius: 0 2px 2px 0;
}

.nav-icon {
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-tooltip {
  position: absolute;
  left: calc(100% + 8px);
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 11px;
  white-space: nowrap;
  z-index: 1000;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s;
}

.nav-item:hover .nav-tooltip {
  opacity: 1;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f5f7fa;
}
</style>
