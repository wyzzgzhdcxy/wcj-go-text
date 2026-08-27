<template>
  <div class="git-sync-container">
    <div class="op-zone">
      <div class="op-zone-header">
        <div class="card-header">
          <div class="header-left">
            <el-icon class="header-icon"><FolderOpened /></el-icon>
            <span class="header-title">仓库列表</span>
            <el-tag size="small" type="info">{{ repoList.length }} 个</el-tag>
            <el-tag v-if="filteredRepoList.length !== repoList.length" size="small" type="warning">{{ filteredRepoList.length }} 条结果</el-tag>
          </div>
          <div class="header-right">
            <el-input v-model="searchKeyword" placeholder="搜索仓库名 / 路径" clearable :prefix-icon="Search" class="search-input" />
            <el-select v-model="statusFilter" class="filter-select" placeholder="状态">
              <el-option label="全部状态" value="all" />
              <el-option label="同步成功" value="success" />
              <el-option label="同步失败" value="fail" />
              <el-option label="从未同步" value="none" />
              <el-option label="已启用" value="enabled" />
              <el-option label="已禁用" value="disabled" />
            </el-select>
            <el-button type="primary" @click="selectFolder" :icon="FolderAdd">添加</el-button>
            <el-button type="success" @click="syncAll" :loading="syncing" :icon="Refresh">同步</el-button>
          </div>
        </div>
      </div>
      <div class="op-zone-body">
        <el-table :data="filteredRepoList" border style="width: 100%" class="repo-table">
          <el-table-column label="仓库名称" min-width="180">
            <template #default="scope">
              <el-tooltip :content="statusTip(scope.row)" placement="top">
                <span class="status-dot" :class="getRepoStatus(scope.row)"></span>
              </el-tooltip>
              <el-tooltip :content="scope.row.path" placement="top">
                <span class="repo-name">{{ scope.row.name }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="branch" label="分支" width="80" align="center" />
          <el-table-column label="仅提交" width="80" align="center">
            <template #default="scope">
              <el-tooltip content="仅提交，不推送" placement="top">
                <el-switch v-model="scope.row.commitOnly" :disabled="!scope.row.enabled" @change="(val) => toggleCommitOnly(scope.row, val)" />
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="自动同步" width="100" align="center">
            <template #default="scope">
              <el-switch v-model="scope.row.autoSync" :disabled="!scope.row.enabled" @change="(val) => toggleAutoSync(scope.row, val)" />
            </template>
          </el-table-column>
          <el-table-column label="启用" width="70" align="center">
            <template #default="scope">
              <el-switch v-model="scope.row.enabled" @change="saveRepos" />
            </template>
          </el-table-column>
          <el-table-column label="上次同步" width="120" align="center">
            <template #default="scope">
              <el-tooltip :content="scope.row.lastSyncTime || '从未同步'" placement="top">
                <span v-if="repo.lastSyncTime" class="sync-time">{{ formatRelativeTime(scope.row.lastSyncTime) }}</span>
                <span v-else class="text-muted">-</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="250" align="center">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="packageProject(scope.row)" :loading="scope.row.path === currentPackagingPath" :disabled="!!currentPackagingPath && scope.row.path !== currentPackagingPath">打包</el-button>
              <el-button link type="warning" size="small" @click="softReset(scope.row)" :loading="scope.row.path === currentSoftResettingPath" :disabled="(!!currentPackagingPath || !!currentResettingPath || !!currentSoftResettingPath) && scope.row.path !== currentSoftResettingPath">合并提交</el-button>
              <el-button link type="info" size="small" @click="resetProject(scope.row)" :loading="scope.row.path === currentResettingPath" :disabled="(!!currentPackagingPath || !!currentResettingPath) && scope.row.path !== currentResettingPath">重置</el-button>
              <el-button link type="danger" size="small" @click="removeRepo(scope.row)" :icon="Delete" :disabled="!!currentPackagingPath">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div v-if="syncResults.length > 0" class="result-zone">
      <div class="result-zone-header">
        <div class="header-left">
          <el-icon class="header-icon"><Document /></el-icon>
          <span class="header-title">同步结果</span>
          <el-tag size="small" type="success">{{ syncResults.filter(r => r.success).length }} 成功</el-tag>
          <el-tag size="small" type="danger" v-if="syncResults.some(r => !r.success)">{{ syncResults.filter(r => !r.success).length }} 失败</el-tag>
        </div>
      </div>
      <el-collapse>
        <el-collapse-item v-for="(result, index) in syncResults" :key="index">
          <template #title>
            <div class="result-title">
              <el-tag :type="result.success ? 'success' : 'danger'" size="small">
                {{ result.success ? '成功' : '失败' }}
              </el-tag>
              <span class="result-name">{{ result.name }}</span>
            </div>
          </template>
          <div class="log-section">
            <div class="log-block">
              <div class="log-title">Commit</div>
              <pre class="log-output">{{ result.commitLog || '无' }}</pre>
            </div>
            <div class="log-block">
              <div class="log-title">Pull</div>
              <pre class="log-output">{{ result.pullLog || '无' }}</pre>
            </div>
            <div class="log-block">
              <div class="log-title">Push</div>
              <pre class="log-output">{{ result.pushLog || '无' }}</pre>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>

    <div class="status-bar">
      <span class="status-label">同步:</span>
      <span v-if="latestLogs.length === 0" class="status-empty">暂无</span>
      <template v-else>
        <span v-for="(log, index) in latestLogs" :key="log.id" class="status-item">
          <span :class="log.success ? 'status-success' : 'status-fail'">{{ log.success ? '✓' : '✗' }}</span>
          <span class="status-repo">{{ log.repoName }}</span>
          <span class="status-time">{{ formatTime(log.time) }}</span>
          <span class="status-msg">{{ log.message }}</span>
          <span v-if="index < latestLogs.length - 1" class="status-sep">|</span>
        </span>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FolderAdd, Refresh, Delete, FolderOpened, Document, Search } from '@element-plus/icons-vue'
import {
  LoadGitRepoList,
  GetSyncLogs,
  SelectDirectory,
  GetGitRepoInfo,
  SaveGitRepoList,
  GitSync,
  ResetProject,
  SoftReset,
  PackageProject,
} from '../wailsjs/go/app/App.js'

const repoList = ref([])
const syncing = ref(false)
const currentPackagingPath = ref(null)
const currentResettingPath = ref(null)
const currentSoftResettingPath = ref(null)
const syncResults = ref([])
const syncLogs = ref([])
const searchKeyword = ref('')
const statusFilter = ref('all')
let refreshTimer = null

const latestLogs = computed(() => syncLogs.value.slice(0, 3))

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  if (timeStr.includes('T')) {
    return timeStr.split('T')[1].split('Z')[0]
  }
  return timeStr.split(' ')[1] || timeStr
}

const getRepoStatus = (repo) => {
  if (!repo.enabled) return 'disabled'
  if (repo.lastSyncSuccess === 1) return 'success'
  if (repo.lastSyncSuccess === 0) return 'fail'
  return 'none'
}

const statusTip = (repo) => {
  const map = {
    success: '同步成功',
    fail: '同步失败',
    none: '从未同步',
    disabled: '已禁用',
  }
  return map[getRepoStatus(repo)] || '未知'
}

const formatRelativeTime = (timeStr) => {
  if (!timeStr) return '-'
  const s = String(timeStr).replace(' ', 'T').replace(/Z$/, '').replace(/\.\d+$/, '')
  const date = new Date(s)
  if (isNaN(date.getTime())) return timeStr
  const diff = Date.now() - date.getTime()
  if (diff < 60 * 1000) return '刚刚'
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const dd = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${mm}-${dd}`
}

const filteredRepoList = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  return repoList.value.filter((repo) => {
    if (statusFilter.value === 'enabled' && !repo.enabled) return false
    if (statusFilter.value === 'disabled' && repo.enabled) return false
    if (statusFilter.value === 'success' && getRepoStatus(repo) !== 'success') return false
    if (statusFilter.value === 'fail' && getRepoStatus(repo) !== 'fail') return false
    if (statusFilter.value === 'none' && getRepoStatus(repo) !== 'none') return false

    if (kw) {
      const name = (repo.name || '').toLowerCase()
      const path = (repo.path || '').toLowerCase()
      const branch = (repo.branch || '').toLowerCase()
      if (!name.includes(kw) && !path.includes(kw) && !branch.includes(kw)) return false
    }
    return true
  })
})

const loadRepos = async () => {
  try {
    const result = await LoadGitRepoList()
    if (result.success && result.repos) {
      repoList.value = result.repos
    }
  } catch (error) {
    console.error('加载仓库列表失败:', error)
  }
}

const loadSyncLogs = async () => {
  try {
    const result = await GetSyncLogs({ limit: 50 })
    if (result.success) {
      syncLogs.value = result.logs || []
    }
  } catch (error) {
    console.error('加载同步日志失败:', error)
  }
}

const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(() => {
    loadSyncLogs()
    loadRepos()
  }, 3000)
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const selectFolder = async () => {
  try {
    const dirPath = await SelectDirectory()
    if (dirPath) {
      const result = await GetGitRepoInfo({ path: dirPath })
      if (result.success) {
        const exists = repoList.value.some(r => r.path === result.repo.path)
        if (!exists) {
          result.repo.autoSync = false
          repoList.value.push(result.repo)
          saveRepos()
          ElMessage.success('添加成功')
        } else {
          ElMessage.warning('仓库已存在')
        }
      } else {
        ElMessage.error(result.message || '不是Git仓库')
      }
    }
  } catch (error) {
    ElMessage.error('添加失败: ' + error.message)
  }
}

const removeRepo = async (repo) => {
  try {
    await ElMessageBox.confirm(
      `确定从列表中删除仓库「${repo.name}」？仅移除配置，不会删除本地文件。`,
      '删除确认',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }
  const idx = repoList.value.findIndex(r => r.path === repo.path)
  if (idx === -1) return
  repoList.value.splice(idx, 1)
  saveRepos()
  ElMessage.success('已删除')
}

const toggleAutoSync = (repo, newVal) => {
  if (newVal && repo.commitOnly) {
    repo.commitOnly = false
  }
  saveRepos()
}

const toggleCommitOnly = (repo, newVal) => {
  if (newVal && repo.autoSync) {
    repo.autoSync = false
  }
  saveRepos()
}

const saveRepos = async () => {
  try {
    const result = await SaveGitRepoList({ repos: repoList.value })
    if (!result.success) {
      ElMessage.error(result.message)
    }
  } catch (error) {
    ElMessage.error('保存失败: ' + error.message)
  }
}

const syncAll = async () => {
  const enabledRepos = repoList.value.filter(r => r.enabled)
  if (enabledRepos.length === 0) {
    ElMessage.warning('没有已启用的仓库')
    return
  }

  syncing.value = true
  syncResults.value = []

  try {
    const result = await GitSync({ repos: enabledRepos })
    if (result.success) {
      syncResults.value = result.results
      ElMessage.success(result.message)
      await loadSyncLogs()
      await loadRepos()
    } else {
      ElMessage.error(result.message)
    }
  } catch (error) {
    ElMessage.error('同步失败: ' + error.message)
  } finally {
    syncing.value = false
  }
}

const resetProject = async (repo) => {
  try {
    await ElMessageBox.confirm(
      `将清空本地 Git 历史，重新初始化仓库并强制推送到远端，覆盖远程历史，不可恢复。\n\n项目: ${repo.name}\n路径: ${repo.path}`,
      '确认重置',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    currentResettingPath.value = repo.path
    const result = await ResetProject({ path: repo.path })
    currentResettingPath.value = null

    if (result.success) {
      ElMessage.success({ message: '重置成功', duration: 0 })
      console.log('重置输出:', result.output)
    } else {
      ElMessage.error('重置失败: ' + result.message)
    }
  } catch (error) {
    currentResettingPath.value = null
    if (error !== 'cancel') {
      ElMessage.error('重置失败: ' + error.message)
    }
  }
}

const softReset = async (repo) => {
  try {
    await ElMessageBox.confirm(
      `将本地未推送的提交合并为一次提交（基于 origin/${repo.branch || '当前分支'}，不会推送）。\n\n项目: ${repo.name}\n路径: ${repo.path}`,
      '确认合并提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'info',
      }
    )

    currentSoftResettingPath.value = repo.path
    const result = await SoftReset({ path: repo.path, message: '合并本地未推送的提交' })
    currentSoftResettingPath.value = null

    if (result.success) {
      ElMessage.success({ message: '合并成功', duration: 0 })
      console.log('软重置输出:', result.output)
    } else {
      ElMessage.error('合并失败: ' + result.message)
    }
  } catch (error) {
    currentSoftResettingPath.value = null
    if (error !== 'cancel') {
      ElMessage.error('合并失败: ' + error.message)
    }
  }
}

const packageProject = async (repo) => {
  currentPackagingPath.value = repo.path
  try {
    const result = await PackageProject({ path: repo.path })
    if (result.success) {
      ElMessage.success(result.message + (result.outputDir ? ' 目录: ' + result.outputDir : ''))
    } else {
      ElMessage.error(result.message)
      console.error(result.output)
    }
  } catch (error) {
    ElMessage.error('打包失败: ' + error.message)
  } finally {
    currentPackagingPath.value = null
  }
}

onMounted(() => {
  loadRepos()
  loadSyncLogs()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.git-sync-container {
  padding: 4px;
  background: #f5f7fa;
}

.op-zone {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 8px;
}

.op-zone-header {
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
}

.op-zone-body {
  padding: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.header-title {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 16px;
  color: #409eff;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.search-input {
  width: 190px;
}

.filter-select {
  width: 120px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  vertical-align: middle;
  flex-shrink: 0;
}

.status-dot.success {
  background: #67c23a;
  box-shadow: 0 0 0 3px rgba(103, 194, 58, 0.15);
}

.status-dot.fail {
  background: #f56c6c;
  box-shadow: 0 0 0 3px rgba(245, 108, 108, 0.15);
}

.status-dot.none {
  background: #c0c4cc;
}

.status-dot.disabled {
  background: #e4e7ed;
  border: 1px solid #dcdfe6;
}

.repo-name {
  font-weight: 500;
  color: #303133;
  vertical-align: middle;
}

:deep(.el-table) {
  margin: 0;
  font-size: 13px;
}

:deep(.el-table__row) {
  height: 40px;
}

:deep(.el-table td) {
  padding: 4px 0;
}

:deep(.repo-table .el-table__row:hover) {
  background: #f0f7ff;
}

.result-zone {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 8px;
}

.result-zone-header {
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
}

.result-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-name {
  font-weight: 500;
}

.path-text {
  margin-left: auto;
  color: #909399;
  font-size: 12px;
}

.log-output {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 10px;
  border-radius: 4px;
  max-height: 150px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}

.log-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
}

.log-block {
  border-left: 3px solid #409eff;
  padding-left: 10px;
}

.log-title {
  font-weight: bold;
  margin-bottom: 5px;
  font-size: 12px;
  color: #606266;
}

.text-muted {
  color: #c0c4cc;
}

.sync-time {
  font-size: 12px;
  color: #606266;
}

.status-bar {
  height: 32px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  padding: 0 12px;
  font-size: 12px;
}

.status-label {
  font-weight: bold;
  margin-right: 8px;
  color: #606266;
}

.status-empty {
  color: #909399;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

.status-success {
  color: #67c23a;
  font-weight: bold;
}

.status-fail {
  color: #f56c6c;
  font-weight: bold;
}

.status-repo {
  color: #303133;
  font-weight: 500;
}

.status-time {
  color: #909399;
  font-size: 11px;
}

.status-msg {
  color: #606266;
  margin-left: 6px;
}

.status-sep {
  color: #dcdfe6;
  margin: 0 10px;
}
</style>
