<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%; height: 100%; display: flex; flex-direction: column;">
    <div style="padding: 10px;">
      <el-button type="primary" v-on:click="openFile">📂 打开文件夹</el-button>
      <el-button type="primary" v-on:click="fileRename">🔐 文件名加密</el-button>
      <el-button type="primary" v-on:click="recovery">↩️ 复原</el-button>
      <el-button type="warning" v-on:click="cleanFilename">🧹 清理文件名</el-button>
    </div>

    <div style="flex: 1; overflow: auto;">
      <single-table v-bind:data="data"/>
    </div>

    <div style="height: 30px; line-height: 30px; padding: 0 10px; background: #f5f7fa; border-top: 1px solid #e4e7ed; font-size: 12px; color: #606266; margin-top: -1px;">
      <span>📁 {{ currentDir || '未选择文件夹' }}</span>
      <span style="margin-left: 20px;">📊 {{ fileCount }} 个文件</span>
    </div>
  </div>
</template>

<script>
import {ListFile, Recovery, RenameFile, CleanFilename, SelectDirectory} from "../wailsjs/go/main/App.js";

export default {
  data() {
    return {
      currentDir: '',
      fileCount: 0,
      data: []
    }
  },

  methods: {
    async openFile() {
      const dir = await SelectDirectory();
      if (dir) {
        this.currentDir = dir;
        this.listFile();
      }
    },
    fileRename() {
      let that = this;
      RenameFile(that.currentDir).then(function (res) {
        that.data = res;
        that.fileCount = res.length;
      })
    },
    listFile() {
      let that = this;
      ListFile(that.currentDir).then(function (res) {
        that.data = res;
        that.fileCount = res.length;
      })
    },
    recovery() {
      let that = this;
      Recovery(that.currentDir).then(function (res) {
        that.data = res;
        that.fileCount = res.length;
      })
    },
    cleanFilename() {
      let that = this;
      CleanFilename(that.currentDir).then(function (res) {
        that.data = res;
        that.fileCount = res.length;
      })
    }
  }
}
</script>