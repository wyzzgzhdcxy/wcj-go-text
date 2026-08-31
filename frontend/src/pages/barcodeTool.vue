<template>
  <div class="barcode-page">
    <el-tabs v-model="activeTab" type="border-card" class="main-tabs">
      <el-tab-pane name="generate">
        <template #label>
          <span class="tab-label"><el-icon><CirclePlus /></el-icon> 生成</span>
        </template>
        <GeneratePanel />
      </el-tab-pane>

      <el-tab-pane name="decode">
        <template #label>
          <span class="tab-label"><el-icon><ZoomIn /></el-icon> 识别</span>
        </template>
        <DecodePanel />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { CirclePlus, ZoomIn } from '@element-plus/icons-vue'
import GeneratePanel from '../components/GeneratePanel.vue'
import DecodePanel from '../components/DecodePanel.vue'
import { barcodeActiveTab } from '../store'

// 直接绑到 store，保证点结果列表「去生成」能切回生成页
const activeTab = computed({
  get: () => barcodeActiveTab.value,
  set: (v) => { barcodeActiveTab.value = v },
})
</script>

<style scoped>
.barcode-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 0;
  box-sizing: border-box;
}

.main-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
  border-radius: 8px;
  overflow: hidden;
  background: #ffffff;
}

.main-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 16px;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
</style>
