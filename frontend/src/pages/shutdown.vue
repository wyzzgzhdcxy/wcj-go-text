<style scoped>
.shutdown-container {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.shutdown-card {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.shutdown-option {
  margin-bottom: 20px;
}

.option-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  display: block;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.el-input-number {
  width: 200px;
}

.result-text {
  margin-top: 16px;
  padding: 12px;
  border-radius: 4px;
  font-size: 14px;
}

.success {
  background: #f0f9eb;
  color: #67c23a;
  border: 1px solid #e1f3d8;
}

.error {
  background: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fde2e2;
}

.warning-text {
  color: #e6a23c;
  font-size: 13px;
  margin-top: 8px;
}

.divider {
  height: 1px;
  background: #ebeef5;
  margin: 20px 0;
}
</style>

<template>
  <div class="shutdown-container">
    <div class="page-title">
      <span>⏻</span>
      <span>定时关机</span>
    </div>

    <!-- 倒计时关机 -->
    <div class="shutdown-card">
      <div class="card-title">⏱️ 倒计时关机</div>
      <div class="shutdown-option">
        <label class="option-label">设置倒计时（秒）：</label>
        <div class="option-row">
          <el-input-number v-model="shutdownSeconds" :min="1" :max="3600"></el-input-number>
          <el-button type="primary" @click="shutdownAfterSeconds">执行关机</el-button>
        </div>
        <div class="warning-text">提示：设置后将立即执行关机倒计时</div>
      </div>
    </div>

    <!-- 定时关机 -->
    <div class="shutdown-card">
      <div class="card-title">📅 定时关机</div>
      <div class="shutdown-option">
        <label class="option-label">设置具体关机时间：</label>
        <div class="option-row">
          <el-date-picker
            v-model="shutdownTime"
            type="datetime"
            placeholder="选择关机时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            :picker-options="pickerOptions"
            style="width: 300px;"
          ></el-date-picker>
          <el-button type="primary" @click="shutdownAt">设置定时</el-button>
        </div>
        <div class="warning-text">提示：请选择一个未来的时间点进行关机</div>
      </div>
    </div>

    <!-- 屏幕工具 -->
    <div class="shutdown-card">
      <div class="card-title">🖥️ 屏幕工具</div>
      <div class="shutdown-option">
        <label class="option-label">屏幕控制：</label>
        <div class="option-row">
          <el-button type="info" @click="turnOffDisplay">熄灭屏幕</el-button>
        </div>
        <div class="warning-text">提示：点击按钮立即关闭显示器</div>
      </div>
    </div>

    <!-- 取消关机 -->
    <div class="shutdown-card">
      <div class="card-title">❌ 取消关机</div>
      <div class="shutdown-option">
        <el-button type="warning" @click="cancelShutdown">取消关机计划</el-button>
      </div>
    </div>

    <!-- 执行结果 -->
    <div v-if="shutdownResult" class="result-text" :class="resultClass">
      {{ shutdownResult }}
    </div>
  </div>
</template>

<script>
import {ShutdownAfterSeconds, ShutdownAt, CancelShutdown, TurnOffDisplayString} from "../wailsjs/go/app/App.js";
import dayjs from 'dayjs';

export default {
  name: 'Shutdown',
  data() {
    return {
      shutdownSeconds: 60,
      shutdownTime: dayjs().format('YYYY-MM-DD HH:mm:ss'),
      shutdownResult: '',
      pickerOptions: {
        disabledDate(time) {
          // 允许当前时间，只禁用过去的时间（1秒容差）
          return time.getTime() < Date.now() - 1000;
        }
      }
    };
  },
  computed: {
    resultClass() {
      if (this.shutdownResult.includes('错误') || this.shutdownResult.includes('失败')) {
        return 'error';
      }
      return 'success';
    }
  },
  methods: {
    shutdownAfterSeconds() {
      let that = this;
      ShutdownAfterSeconds(this.shutdownSeconds).then(res => {
        that.shutdownResult = res;
      }).catch(err => {
        that.shutdownResult = '错误: ' + err;
      });
    },
    shutdownAt() {
      let that = this;
      ShutdownAt(this.shutdownTime).then(res => {
        that.shutdownResult = res;
      }).catch(err => {
        that.shutdownResult = '错误: ' + err;
      });
    },
    cancelShutdown() {
      let that = this;
      CancelShutdown().then(res => {
        that.shutdownResult = res;
      }).catch(err => {
        that.shutdownResult = '错误: ' + err;
      });
    },
    turnOffDisplay() {
      let that = this;
      TurnOffDisplayString().then(res => {
        that.shutdownResult = res;
      }).catch(err => {
        that.shutdownResult = '错误: ' + err;
      });
    }
  }
};
</script>
