<template>
  <div
      style="width: 100%; height: 100%; background: white; border-radius: 8px; padding: 16px; display: flex; flex-direction: column; overflow: hidden;">
    <!-- 刷新按钮和统计信息 -->
    <div style="margin-bottom: 12px; flex-shrink: 0; display: flex; align-items: center; gap: 16px;">
      <el-button type="primary" @click="loadAllEnvVars" :loading="loading" size="small">
        <span v-if="!loading">🔄 刷新</span>
        <span v-else>加载中...</span>
      </el-button>
      <el-button @click="openEnvEditor" size="small">
        🌐 打开环境变量窗口
      </el-button>

      <!-- 统计信息 -->
      <div style="display: flex; gap: 12px; flex-wrap: wrap;">
        <el-tag type="primary" size="small">用户变量: {{ userEnvVars.length }}</el-tag>
        <el-tag type="success" size="small">系统变量: {{ systemEnvVars.length }}</el-tag>
        <el-tag type="info" size="small">进程变量: {{ processEnvVars.length }}</el-tag>
        <el-tag v-if="commonKeys.length > 0" type="warning" size="small" @click="showCommonDialog = true" style="cursor: pointer;">
          重复变量: {{ commonKeys.length }}
        </el-tag>
        <el-tag v-if="emptyValueKeys.length > 0" type="danger" size="small" @click="showEmptyDialog = true" style="cursor: pointer;">
          空值变量: {{ emptyValueKeys.length }}
        </el-tag>
      </div>

      <el-button size="small" @click="backupUserEnvVars" :loading="backing">
        💾 备份用户环境变量
      </el-button>
      <el-button size="small" @click="showRestoreDialog = true">
        ♻️ 恢复用户环境变量
      </el-button>

      <el-input v-model="filterKey" placeholder="输入变量名过滤" clearable size="small"
                style="margin-left: auto; width: 220px;">
        <template #prefix>🔍</template>
      </el-input>
    </div>

    <!-- 标签页 -->
    <el-tabs v-model="activeTab" class="env-tabs" style="flex: 1; min-height: 0;">
      <!-- 用户环境变量 -->
      <el-tab-pane label="用户环境变量" name="user">
        <div v-if="userEnvVars.length === 0 && !loading" style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">👤</div>
          <div>暂无用户环境变量</div>
        </div>
        <div v-else style="flex: 1; min-height: 0;">
          <el-table :data="filteredUserVars" stripe height="100%" style="width: 100%;">
            <el-table-column prop="name" label="变量名" width="250">
              <template #default="scope">
                <span v-if="isCommonKey(scope.row.name)" style="font-weight: 600; color: #e6a23c;">{{ scope.row.name }} ⚠️</span>
                <span v-else style="font-weight: 500; color: #409eff;">{{ scope.row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="变量值" min-width="600">
              <template #default="scope">
                <span v-if="scope.row.value === ''" style="color: #f56c6c;">(空值)</span>
                <el-tooltip v-else :content="scope.row.value" placement="top" :show-after="300">
                  <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 600px;">{{ scope.row.value }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="scope">
                <el-button type="danger" size="small" @click="deleteUserEnvVar(scope.row.name)" :loading="deleting">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 系统环境变量 -->
      <el-tab-pane label="系统环境变量" name="system">
        <el-alert
            title="提示：系统环境变量需要管理员权限才能删除"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 8px; flex-shrink: 0;">
        </el-alert>
        <div v-if="systemEnvVars.length === 0 && !loading" style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">💻</div>
          <div>暂无系统环境变量</div>
        </div>
        <div v-else style="flex: 1; min-height: 0;">
          <el-table :data="filteredSystemVars" stripe height="100%" style="width: 100%;">
            <el-table-column prop="name" label="变量名" width="250">
              <template #default="scope">
                <span v-if="isCommonKey(scope.row.name)" style="font-weight: 600; color: #e6a23c;">{{ scope.row.name }} ⚠️</span>
                <span v-else style="font-weight: 500; color: #67c23a;">{{ scope.row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="变量值" min-width="600">
              <template #default="scope">
                <span v-if="scope.row.value === ''" style="color: #f56c6c;">(空值)</span>
                <el-tooltip v-else :content="scope.row.value" placement="top" :show-after="300">
                  <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 600px;">{{ scope.row.value }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="scope">
                <el-button type="danger" size="small" @click="deleteSystemEnvVar(scope.row.name)" :loading="deleting">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 进程环境变量 -->
      <el-tab-pane label="进程环境变量" name="process">
        <el-alert
            title="提示：进程环境变量是当前进程的只读副本，删除不会影响系统环境变量"
            type="info"
            :closable="false"
            show-icon
            style="margin-bottom: 8px; flex-shrink: 0;">
        </el-alert>
        <div v-if="processEnvVars.length === 0 && !loading" style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">⚙️</div>
          <div>暂无进程环境变量</div>
        </div>
        <div v-else style="flex: 1; min-height: 0;">
          <el-table :data="filteredProcessVars" stripe height="100%" style="width: 100%;">
            <el-table-column prop="name" label="变量名" width="250">
              <template #default="scope">
                <span style="font-weight: 500; color: #909399;">{{ scope.row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="变量值" min-width="600">
              <template #default="scope">
                <span v-if="scope.row.value === ''" style="color: #f56c6c;">(空值)</span>
                <el-tooltip v-else :content="scope.row.value" placement="top" :show-after="300">
                  <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 600px;">{{ scope.row.value }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="scope">
                <el-tag type="info" size="small">只读</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- PATH 环境变量 -->
      <el-tab-pane label="PATH 变量" name="path">
        <div v-if="duplicatePathCount > 0" style="margin-bottom: 8px; flex-shrink: 0;">
          <el-alert
              :title="`提示：检测到 ${duplicatePathCount} 个重复的 PATH 条目，重复条目可能会导致命令执行顺序不确定`"
              type="warning"
              :closable="false"
              show-icon>
          </el-alert>
        </div>
        <div v-else style="margin-bottom: 8px; flex-shrink: 0;">
          <el-alert
              title="未检测到重复的 PATH 条目"
              type="success"
              :closable="false"
              show-icon>
          </el-alert>
        </div>
        <div v-if="pathEntries.length === 0 && !loading" style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">📁</div>
          <div>暂无 PATH 信息</div>
        </div>
        <div v-else style="flex: 1; min-height: 0;">
          <el-table :data="filteredPathEntries" stripe height="100%" style="width: 100%;">
            <el-table-column prop="path" label="路径" min-width="500">
              <template #default="scope">
                <span v-if="scope.row.isDuplicate" style="font-weight: 600; color: #e6a23c;">{{ scope.row.path }} ⚠️ 重复</span>
                <el-tooltip v-else :content="scope.row.path" placement="top" :show-after="300">
                  <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 700px;">{{ scope.row.path }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column prop="source" label="来源" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.source === 'User' ? 'primary' : scope.row.source === 'System' ? 'success' : 'info'" size="small">
                  {{ scope.row.source === 'User' ? '用户' : scope.row.source === 'System' ? '系统' : '进程' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="scope">
                <el-tag v-if="scope.row.isDuplicate" type="warning" size="small">重复</el-tag>
                <el-tag v-else type="success" size="small">正常</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 重复变量弹窗 -->
    <el-dialog v-model="showCommonDialog" title="用户与系统重复的环境变量" width="600px">
      <el-alert
          title="以下环境变量在用户环境和系统环境中都存在，可能导致不可预期的行为"
          type="warning"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;">
      </el-alert>
      <el-table :data="commonKeysDetail" stripe max-height="400px">
        <el-table-column prop="name" label="变量名" width="200">
          <template #default="scope">
            <span style="font-weight: 600; color: #e6a23c;">{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="用户环境变量值" min-width="180">
          <template #default="scope">
            <el-tooltip :content="scope.row.userValue || '(空值)'" placement="top" :show-after="300">
              <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 200px;">{{ scope.row.userValue || '(空值)' }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="系统环境变量值" min-width="180">
          <template #default="scope">
            <el-tooltip :content="scope.row.systemValue || '(空值)'" placement="top" :show-after="300">
              <span style="display: inline-block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 200px;">{{ scope.row.systemValue || '(空值)' }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 空值变量弹窗 -->
    <el-dialog v-model="showEmptyDialog" title="值为空的环境变量" width="700px">
      <el-alert
          title="以下环境变量的值为空，可能会导致程序运行异常"
          type="danger"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;">
      </el-alert>
      <el-table :data="emptyValueDetail" stripe max-height="400px">
        <el-table-column prop="name" label="变量名" width="200">
          <template #default="scope">
            <span :style="{ color: scope.row.source === 'user' ? '#409eff' : scope.row.source === 'system' ? '#67c23a' : '#909399' }">{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="所属环境" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.source === 'user' ? 'primary' : scope.row.source === 'system' ? 'success' : 'info'" size="small">
              {{ scope.row.source === 'user' ? '用户环境' : scope.row.source === 'system' ? '系统环境' : '进程环境' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前值">
          <template #default="scope">
            <span style="color: #f56c6c;">(空值)</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 恢复用户环境变量备份对话框 -->
    <el-dialog title="恢复用户环境变量备份" v-model="showRestoreDialog" width="700px"
               :close-on-click-modal="false" @open="loadBackups">
      <div v-if="loadingBackups" style="text-align: center; padding: 40px; color: #909399;">
        加载备份列表...
      </div>
      <div v-else-if="backups.length === 0" style="text-align: center; padding: 40px;">
        <div style="font-size: 48px; color: #909399;">📦</div>
        <p style="margin-top: 12px; color: #909399;">暂无备份文件</p>
      </div>
      <div v-else>
        <el-alert
            title="注意：恢复环境变量将覆盖当前设置"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 16px;">
        </el-alert>
        <el-table :data="backups" stripe max-height="400px">
          <el-table-column prop="name" label="备份文件名" min-width="220"/>
          <el-table-column prop="modified" label="修改时间" width="170"/>
          <el-table-column label="大小" width="100">
            <template #default="scope">
              {{ formatFileSize(scope.row.size) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button type="primary" size="small" @click="restoreBackup(scope.row)" :loading="restoringBackup">
                恢复
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showRestoreDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, computed, onMounted} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';

// 导入 Wails 后端方法
import {
  GetUserEnvVars,
  GetSystemEnvVars,
  GetProcessEnvVars,
  DeleteUserEnvVar,
  DeleteSystemEnvVar,
  GetPathInfo,
  OpenEnvironmentEditor,
  BackupUserEnvVars,
  GetEnvBackups,
  RestoreEnvBackup
} from '../wailsjs/go/app/App.js';

// 数据
const activeTab = ref('user');
const loading = ref(false);
const deleting = ref(false);
const userEnvVars = ref([]);
const systemEnvVars = ref([]);
const processEnvVars = ref([]);
const pathEntries = ref([]);
const showCommonDialog = ref(false);
const showEmptyDialog = ref(false);
const filterKey = ref('');
const backing = ref(false);
const showRestoreDialog = ref(false);
const restoringBackup = ref(false);
const backups = ref([]);
const loadingBackups = ref(false);

// 按关键字过滤（变量名 / PATH 条目），不区分大小写
const matchFilter = (text) => {
  const q = filterKey.value.trim().toLowerCase();
  return !q || String(text || '').toLowerCase().includes(q);
};
const filteredUserVars = computed(() => userEnvVars.value.filter(v => matchFilter(v.name)));
const filteredSystemVars = computed(() => systemEnvVars.value.filter(v => matchFilter(v.name)));
const filteredProcessVars = computed(() => processEnvVars.value.filter(v => matchFilter(v.name)));
const filteredPathEntries = computed(() => displayPathEntries.value.filter(p => matchFilter(p.path)));

// 计算用户和系统环境变量中都存在的key
const commonKeys = computed(() => {
  const userKeys = new Set(userEnvVars.value.map(v => v.name));
  const systemKeys = new Set(systemEnvVars.value.map(v => v.name));
  return [...userKeys].filter(key => systemKeys.has(key));
});

// 重复变量的详细信息
const commonKeysDetail = computed(() => {
  return commonKeys.value.map(key => {
    const userVar = userEnvVars.value.find(v => v.name === key);
    const systemVar = systemEnvVars.value.find(v => v.name === key);
    return {
      name: key,
      userValue: userVar?.value || '',
      systemValue: systemVar?.value || ''
    };
  });
});

// 值为空的变量
const emptyValueKeys = computed(() => {
  const empty = [];
  userEnvVars.value.forEach(v => {
    if (v.value === '') {
      empty.push({ name: v.name, source: 'user' });
    }
  });
  systemEnvVars.value.forEach(v => {
    if (v.value === '') {
      empty.push({ name: v.name, source: 'system' });
    }
  });
  processEnvVars.value.forEach(v => {
    if (v.value === '') {
      empty.push({ name: v.name, source: 'process' });
    }
  });
  return empty;
});

// 值为空的变量详细信息
const emptyValueDetail = computed(() => emptyValueKeys.value);

// 重复的 PATH 条目数量
const duplicatePathCount = computed(() => {
  return pathEntries.value.filter(p => p.isDuplicate).length;
});

// 显示的 PATH 条目（无重复时只显示进程的）
const displayPathEntries = computed(() => {
  if (duplicatePathCount.value > 0) {
    return pathEntries.value;
  }
  return pathEntries.value.filter(p => p.source === 'Process');
});

// 检查是否为重复key
const isCommonKey = (name) => {
  return commonKeys.value.includes(name);
};

// 打开系统环境变量编辑窗口
async function openEnvEditor() {
  try {
    await OpenEnvironmentEditor();
  } catch (error) {
    console.error('打开环境变量编辑窗口失败:', error);
    ElMessage.error('打开环境变量编辑窗口失败: ' + error);
  }
}

// 备份用户环境变量
async function backupUserEnvVars() {
  backing.value = true;
  try {
    const dir = await BackupUserEnvVars();
    ElMessage.success('用户环境变量备份成功: ' + dir);
  } catch (error) {
    console.error('用户环境变量备份失败:', error);
    ElMessage.error('用户环境变量备份失败: ' + error);
  } finally {
    backing.value = false;
  }
}

// 加载备份列表
async function loadBackups() {
  loadingBackups.value = true;
  try {
    backups.value = (await GetEnvBackups()) || [];
  } catch (error) {
    console.error('加载备份列表失败:', error);
    ElMessage.error('加载备份列表失败: ' + error);
  } finally {
    loadingBackups.value = false;
  }
}

// 恢复备份
async function restoreBackup(backup) {
  try {
    await ElMessageBox.confirm(
        `确定要恢复备份 "${backup.name}" 吗？此操作将覆盖当前用户环境变量。`,
        '确认恢复',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        }
    );

    restoringBackup.value = true;
    try {
      await RestoreEnvBackup(backup.path);
      ElMessage.success('环境变量恢复成功');
      showRestoreDialog.value = false;
      // 刷新列表（本页读取的是注册表，恢复后立即可见）
      await loadAllEnvVars();
    } catch (error) {
      console.error('恢复备份失败:', error);
      ElMessage.error('恢复备份失败: ' + error);
    } finally {
      restoringBackup.value = false;
    }
  } catch {
    // 用户取消操作
  }
}

// 格式化文件大小
function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// 加载所有环境变量
async function loadAllEnvVars() {
  loading.value = true;
  try {
    const [user, system, process, pathInfo] = await Promise.all([
      GetUserEnvVars(),
      GetSystemEnvVars(),
      GetProcessEnvVars(),
      GetPathInfo()
    ]);
    userEnvVars.value = user || [];
    systemEnvVars.value = system || [];
    processEnvVars.value = process || [];
    pathEntries.value = pathInfo || [];
  } catch (error) {
    console.error('加载环境变量失败:', error);
    ElMessage.error('加载环境变量失败: ' + error);
  } finally {
    loading.value = false;
  }
}

// 删除用户环境变量
async function deleteUserEnvVar(name) {
  try {
    await ElMessageBox.confirm(
        `确定要删除用户环境变量 "${name}" 吗？`,
        '确认删除',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        }
    );

    deleting.value = true;
    try {
      await DeleteUserEnvVar(name);
      ElMessage.success(`用户环境变量 "${name}" 已删除`);
      // 刷新列表
      await loadAllEnvVars();
    } catch (error) {
      console.error('删除用户环境变量失败:', error);
      ElMessage.error('删除用户环境变量失败: ' + error);
    } finally {
      deleting.value = false;
    }
  } catch {
    // 用户取消操作
  }
}

// 删除系统环境变量
async function deleteSystemEnvVar(name) {
  try {
    await ElMessageBox.confirm(
        `确定要删除系统环境变量 "${name}" 吗？此操作需要管理员权限。`,
        '确认删除',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        }
    );

    deleting.value = true;
    try {
      await DeleteSystemEnvVar(name);
      ElMessage.success(`系统环境变量 "${name}" 已删除`);
      // 刷新列表
      await loadAllEnvVars();
    } catch (error) {
      console.error('删除系统环境变量失败:', error);
      ElMessage.error('删除系统环境变量失败: ' + error);
    } finally {
      deleting.value = false;
    }
  } catch {
    // 用户取消操作
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadAllEnvVars();
});
</script>

<style scoped>
/* 让标签页内容区撑满页面剩余高度，表格填满内容区 */
.env-tabs {
  display: flex;
  flex-direction: column;
}
.env-tabs :deep(.el-tabs__header) {
  flex-shrink: 0;
}
.env-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.env-tabs :deep(.el-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>