<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.function-group {
  margin: 20px 0;
  padding: 20px;
  border-radius: 8px;
  background-color: #f8f9fa;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.function-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 15px;
  color: #409eff;
}

.function-description {
  margin-bottom: 15px;
  line-height: 1.6;
  color: #606266;
}

.button-group {
  display: flex;
  gap: 15px;
}
</style>

<template>
  <div style="width: 100%">
    <!-- 屏幕工具 -->
    <div class="function-group" style="width: 600px">
      <div class="function-title">屏幕工具</div>
      <div class="function-description">
        点击按钮立即熄灭显示器，移动鼠标或按任意键即可重新点亮屏幕：
      </div>
      <div class="button-group">
        <el-button type="info" v-on:click="turnOffDisplay">🖥️ 熄灭屏幕</el-button>
      </div>
    </div>

    <!-- 注册表启动项 -->
    <div class="function-group" style="width: 600px">
      <div class="function-title">注册表启动项</div>
      <div class="function-description">
        点击按钮将自动打开注册表编辑器并定位到开机启动项，方便查看和管理开机自启动程序：
      </div>
      <div style="display: flex; align-items: center; gap: 8px; margin: 10px 0;">
        <code style="font-size: 13px; color: #606266;">HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run</code>
        <el-button size="small" :icon="DocumentCopy" @click="copyPath" title="复制路径"></el-button>
      </div>
      <el-button type="warning" v-on:click="openRegistryStartup">⚙️ 打开注册表</el-button>
    </div>

    <!-- 备份还原功能组 -->
    <div class="function-group" style="width: 600px">
      <div class="function-title">环境配置备份与还原</div>
      <div class="function-description">
        此功能用于备份和还原您的重要系统配置文件和开发环境设置。备份内容包括：系统环境变量、Git配置(SSH密钥和全局配置)、常用开发工具的配置文件等。还原功能可以将这些配置一键恢复到新系统或不同设备上，节省您重复配置环境的时间。建议在系统重装前或更换电脑时使用此功能备份您的开发环境配置。
      </div>
      <div class="button-group">
        <el-button type="success" v-on:click="backUpEnvFiles" icon="DocumentAdd">备份常用文件</el-button>
        <el-button type="warning" v-on:click="restoreEnvFiles" icon="Refresh">还原常用文件</el-button>
      </div>
    </div>
  </div>
</template>

<script>
import {BackUpEnvFiles, OpenRegistryStartup, RestoreEnvFiles, TurnOffDisplayString} from "../wailsjs/go/app/App.js";
import {ElNotification} from "element-plus";
import {DocumentCopy} from "@element-plus/icons-vue";

export default {
  data() {
    return {
      req: {}
    }
  },

  computed: {
    DocumentCopy() {
      return DocumentCopy;
    }
  },

  mounted() {
  },

  methods: {
    backUpEnvFiles() {
      BackUpEnvFiles().then(errorInfo => {
        ElNotification({
          title: '备份数据结果',
          message: errorInfo === "" ? "配置备份成功！建议将备份文件保存在安全位置。" : errorInfo,
          position: 'bottom-right',
          type: errorInfo === "" ? 'success' : 'error',
          duration: 3000
        })
      })
    },

    restoreEnvFiles() {
      this.$confirm('此操作将覆盖当前配置，请确认已备份最新配置', '警告', {
        confirmButtonText: '确定还原',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        RestoreEnvFiles().then(errorInfo => {
          ElNotification({
            title: '还原数据结果',
            message: errorInfo === "" ? "配置还原成功！部分设置可能需要重启应用才能生效。" : errorInfo,
            position: 'bottom-right',
            type: errorInfo === "" ? 'success' : 'error',
            duration: 3000
          })
        })
      }).catch(() => {
        this.$message({
          type: 'info',
          message: '已取消还原操作'
        });
      });
    },

    turnOffDisplay() {
      TurnOffDisplayString().then(res => {
        this.$message.success(res || "已熄灭屏幕");
      }).catch(err => {
        this.$message.error("熄灭屏幕失败: " + err);
      });
    },

    async openRegistryStartup() {
      try {
        await OpenRegistryStartup();
        this.$message.success("已打开注册表并定位到开机启动项");
      } catch (err) {
        console.error("打开注册表启动项失败:", err);
        this.$message.error("打开注册表启动项失败: " + err.message);
      }
    },

    copyPath() {
      const path = "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run";
      navigator.clipboard.writeText(path).then(() => {
        this.$message.success("路径已复制到剪贴板");
      }).catch(() => {
        this.$message.error("复制失败");
      });
    }
  }
}
</script>