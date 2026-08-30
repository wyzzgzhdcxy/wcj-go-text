<template>
  <div
      style="width: 100%; height: 100%; background: white; border-radius: 8px; padding: 24px; display: flex; flex-direction: column; overflow: hidden;">
    <!-- 标题区域 -->
    <div style="margin-bottom: 24px; flex-shrink: 0;">
      <h2 style="margin: 0; font-size: 20px; font-weight: 600; color: #303133;">🔍 检测</h2>
      <p style="margin: 8px 0 0 0; font-size: 13px; color: #909399;">检测电脑上的开发环境配置是否正常</p>
    </div>

    <!-- 操作按钮 -->
    <div style="margin-bottom: 20px; display: flex; gap: 12px; flex-shrink: 0;">
      <el-button type="primary" @click="checkEnvironment" :loading="checking">
        <span v-if="!checking">🔄 开始检测</span>
        <span v-else>检测中...</span>
      </el-button>
    </div>

    <!-- 检测结果 -->
    <div v-if="results.length > 0" style="flex: 1; overflow-y: auto; overflow-x: hidden;">
      <div
          v-for="(item, index) in results"
          :key="index"
          :style="{
          padding: '16px',
          marginBottom: '12px',
          border: '1px solid',
          borderRadius: '8px',
          backgroundColor: getStatusBg(item.status),
          borderColor: getStatusBorder(item.status)
        }"
      >
        <!-- 标题行 -->
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 24px;">{{ getIcon(item.name) }}</span>
            <span style="font-size: 16px; font-weight: 600; color: #303133;">{{ item.name }}</span>
          </div>
          <el-tag :type="getStatusType(item.status)" size="default">
            {{ getStatusText(item.status) }}
          </el-tag>
        </div>

        <!-- 版本信息 -->
        <div v-if="item.version" style="margin-bottom: 6px; font-size: 13px; color: #606266;">
          <span style="color: #909399;">版本:</span> {{ item.version }}
        </div>

        <!-- 路径信息 -->
        <div v-if="item.path" style="margin-bottom: 6px; font-size: 13px; color: #606266;">
          <span style="color: #909399;">路径:</span> {{ item.path }}
        </div>

        <!-- 描述信息 -->
        <div v-if="item.message" style="margin-bottom: 6px; font-size: 13px; color: #606266;">
          {{ item.message }}
        </div>

        <!-- 详细信息（可展开） -->
        <el-collapse v-if="item.detail" style="margin-top: 8px;">
          <el-collapse-item title="详细信息" name="detail">
            <div
                style="font-size: 12px; color: #606266; white-space: pre-wrap; word-break: break-all; background: #f5f7fa; padding: 12px; border-radius: 4px;">
              {{ item.detail }}
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>

    <!-- 初始状态 -->
    <div v-else-if="!checking"
         style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399;">
      <div style="font-size: 64px; margin-bottom: 16px;">🔍</div>
      <div style="font-size: 14px;">点击"开始检测"按钮开始检测环境配置</div>
    </div>

    <!-- 统计信息 -->
    <div v-if="results.length > 0"
         style="margin-top: 20px; padding-top: 16px; border-top: 1px solid #e8eaed; flex-shrink: 0;">
      <div style="display: flex; gap: 24px; font-size: 13px;">
        <div>
          <span style="color: #909399;">总计:</span>
          <span style="font-weight: 600; color: #303133; margin-left: 4px;">{{ results.length }}</span>
        </div>
        <div>
          <span style="color: #909399;">已安装:</span>
          <span style="font-weight: 600; color: #67c23a; margin-left: 4px;">{{ getInstalledCount() }}</span>
        </div>
        <div>
          <span style="color: #909399;">未安装:</span>
          <span style="font-weight: 600; color: #f56c6c; margin-left: 4px;">{{ getNotInstalledCount() }}</span>
        </div>
      </div>
    </div>

    <!-- 恢复环境备份对话框（已移至环境变量页面） -->
  </div>
</template>

<script setup>
import {ref} from 'vue';
import {ElMessage} from 'element-plus';

// 导入 Wails 后端方法
import {CheckEnvironment} from '../wailsjs/go/app/App.js';

// 数据
const checking = ref(false);
const results = ref([]);

// 检查环境
async function checkEnvironment() {
  checking.value = true;
  try {
    const result = await CheckEnvironment();
    results.value = result || [];
  } catch (error) {
    console.error('环境检测失败:', error);
    ElMessage.error('环境检测失败: ' + error);
  } finally {
    checking.value = false;
  }
}

// 获取图标
function getIcon(name) {
  const icons = {
    'Java': '☕',
    'Maven': '📦',
    'Gradle': '🐘',
    'Python': '🐍',
    'Node.js': '💚',
    'Golang': '🐹',
  };
  return icons[name] || '🔧';
}

// 获取状态背景色
function getStatusBg(status) {
  const bgColors = {
    'installed': '#f0f9ff',
    'not_installed': '#fef2f2',
    'error': '#fefce8',
  };
  return bgColors[status] || '#f5f7fa';
}

// 获取状态边框色
function getStatusBorder(status) {
  const borderColors = {
    'installed': '#bae6fd',
    'not_installed': '#fecaca',
    'error': '#fef08a',
  };
  return borderColors[status] || '#e5e7eb';
}

// 获取状态类型
function getStatusType(status) {
  const typeMap = {
    'installed': 'success',
    'not_installed': 'danger',
    'error': 'warning',
  };
  return typeMap[status] || 'info';
}

// 获取状态文本
function getStatusText(status) {
  const textMap = {
    'installed': '已安装',
    'not_installed': '未安装',
    'error': '异常',
  };
  return textMap[status] || '未知';
}

// 获取已安装数量
function getInstalledCount() {
  return results.value.filter(item => item.status === 'installed').length;
}

// 获取未安装数量
function getNotInstalledCount() {
  return results.value.filter(item => item.status === 'not_installed').length;
}

// 组件挂载时自动检测
checkEnvironment();
</script>
