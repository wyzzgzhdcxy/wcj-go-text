<style scoped>
.shutdown-container {
  padding: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
}

.shutdown-card {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 10px;
  padding: 24px 28px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.25s ease, transform 0.25s ease;
}

.shutdown-card:hover {
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
  transform: translateY(-1px);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.shutdown-option {
  margin-bottom: 4px;
}

.shutdown-option:last-child {
  margin-bottom: 0;
}

.option-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 10px;
  display: block;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.el-input-number {
  width: 240px;
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

.warning {
  background: #fdf6ec;
  color: #e6a23c;
  border: 1px solid #faecd8;
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
    <!-- 倒计时关机 -->
    <div class="shutdown-card">
      <div class="card-title">⏱️ 倒计时关机</div>
      <div class="shutdown-option">
        <label class="option-label">设置倒计时（秒）：</label>
        <div class="option-row">
          <el-input-number
            v-model="shutdownSeconds"
            :min="1"
            :max="3600"
            :step="60"
            controls-position="right"
          ></el-input-number>
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
        <label class="option-label">操作：</label>
        <div class="option-row">
          <el-button type="warning" @click="cancelShutdown">取消关机计划</el-button>
        </div>
        <div class="warning-text">提示：如已设置关机计划，点击此按钮可取消</div>
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
import {ElMessage} from 'element-plus';
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
      if (this.shutdownResult.includes('没有定时关机任务')) {
        return 'warning';
      }
      return 'success';
    }
  },
  methods: {
    showResult(msg) {
      this.shutdownResult = msg;
      if (!msg) return;
      if (msg.includes('错误') || msg.includes('失败')) {
        ElMessage.error(msg);
      } else if (msg.includes('没有定时关机任务')) {
        ElMessage.warning(msg);
      } else {
        ElMessage.success(msg);
      }
    },
    shutdownAfterSeconds() {
      let that = this;
      ShutdownAfterSeconds(this.shutdownSeconds).then(res => {
        that.showResult(res);
      }).catch(err => {
        that.showResult('错误: ' + err);
      });
    },
    shutdownAt() {
      let that = this;
      ShutdownAt(this.shutdownTime).then(res => {
        that.showResult(res);
      }).catch(err => {
        that.showResult('错误: ' + err);
      });
    },
    cancelShutdown() {
      let that = this;
      CancelShutdown().then(res => {
        that.showResult(res);
      }).catch(err => {
        that.showResult('错误: ' + err);
      });
    },
    turnOffDisplay() {
      let that = this;
      TurnOffDisplayString().then(res => {
        that.showResult(res);
      }).catch(err => {
        that.showResult('错误: ' + err);
      });
    }
  }
};
</script>
