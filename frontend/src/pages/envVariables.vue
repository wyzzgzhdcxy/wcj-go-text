<template>
  <div
      style="width: 95%; height: 94%; background: white; border-radius: 8px; padding: 16px; display: flex; flex-direction: column; overflow: hidden;">
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
    </div>

    <!-- 标签页 -->
    <el-tabs v-model="activeTab" style="flex: 1; min-height: 0;">
      <!-- 用户环境变量 -->
      <el-tab-pane label="用户环境变量" name="user">
        <div v-if="userEnvVars.length === 0 && !loading" style="text-align: center; padding: 60px 20px; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">👤</div>
          <div>暂无用户环境变量</div>
        </div>
        <el-table v-else :data="userEnvVars" stripe style="width: 100%;" max-height="calc(93vh - 200px)">
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
      </el-tab-pane>

      <!-- 系统环境变量 -->
      <el-tab-pane label="系统环境变量" name="system">
        <el-alert
            title="提示：系统环境变量需要管理员权限才能删除"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 8px;">
        </el-alert>
        <div v-if="systemEnvVars.length === 0 && !loading" style="text-align: center; padding: 60px 20px; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">💻</div>
          <div>暂无系统环境变量</div>
        </div>
        <el-table v-else :data="systemEnvVars" stripe style="width: 100%;" max-height="calc(93vh - 230px)">
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
      </el-tab-pane>

      <!-- 进程环境变量 -->
      <el-tab-pane label="进程环境变量" name="process">
        <el-alert
            title="提示：进程环境变量是当前进程的只读副本，删除不会影响系统环境变量"
            type="info"
            :closable="false"
            show-icon
            style="margin-bottom: 8px;">
        </el-alert>
        <div v-if="processEnvVars.length === 0 && !loading" style="text-align: center; padding: 60px 20px; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">⚙️</div>
          <div>暂无进程环境变量</div>
        </div>
        <el-table v-else :data="processEnvVars" stripe style="width: 100%;" max-height="calc(93vh - 230px)">
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
      </el-tab-pane>

      <!-- PATH 环境变量 -->
      <el-tab-pane label="PATH 变量" name="path">
        <div v-if="duplicatePathCount > 0" style="margin-bottom: 8px;">
          <el-alert
              :title="`提示：检测到 ${duplicatePathCount} 个重复的 PATH 条目，重复条目可能会导致命令执行顺序不确定`"
              type="warning"
              :closable="false"
              show-icon>
          </el-alert>
        </div>
        <div v-else style="margin-bottom: 8px;">
          <el-alert
              title="未检测到重复的 PATH 条目"
              type="success"
              :closable="false"
              show-icon>
          </el-alert>
        </div>
        <div v-if="pathEntries.length === 0 && !loading" style="text-align: center; padding: 60px 20px; color: #909399;">
          <div style="font-size: 48px; margin-bottom: 12px;">📁</div>
          <div>暂无 PATH 信息</div>
        </div>
        <el-table v-else :data="displayPathEntries" stripe style="width: 100%;" max-height="calc(93vh - 230px)">
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
  OpenEnvironmentEditor
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