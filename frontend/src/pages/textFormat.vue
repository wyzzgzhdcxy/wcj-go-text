<template>
  <div class="page-container">
    <div class="content-wrapper">
      <div class="textarea-wrapper">
        <textarea v-model="inputData" class="textarea" placeholder="输入文本，每行一个值..."></textarea>
      </div>
      <div class="button-group">
        <el-button type="success" @click="toSqlQuery">🗄️ 转SQL条件</el-button>
      </div>
      <div class="textarea-wrapper">
        <textarea v-model="result" class="textarea" placeholder="结果..." disabled></textarea>
      </div>
      <div class="copy-btn">
        <el-button type="primary" @click="copyText">复制结果</el-button>
      </div>
    </div>
  </div>
</template>

<script>
import {CopyToClipboard} from "../wailsjs/go/main/App.js";
import {ElNotification} from "element-plus";

export default {
  data() {
    return {
      inputData: '11111111\n22222222\n33333333333',
      result: ''
    }
  },
  methods: {
    async copyText() {
      await CopyToClipboard(this.result);
      ElNotification({title: '成功', message: "已复制到剪贴板", position: 'bottom-right', type: 'success'});
    },
    toSqlQuery() {
      if (this.inputData === "") {
        ElNotification({title: '数据为空', message: "输入数据为空", position: 'bottom-right', type: 'error'});
        return;
      }
      let arr = this.inputData.split("\n");
      let sql = "("
      for (let i = 0; i < arr.length; i++) {
        if (i !== arr.length - 1) {
          sql += "'" + arr[i] + "', ";
        } else {
          sql += "'" + arr[i] + "'";
        }
      }
      this.result = sql + ")";
    }
  }
}
</script>

<style scoped>
.page-container {
  height: 100%;
  padding: 0px;
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: relative;
}

.textarea-wrapper {
  flex: 1;
  min-height: 0;
}

.textarea {
  width: 100%;
  height: 100%;
  padding: 12px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  resize: none;
  font-family: monospace;
  font-size: 14px;
  box-sizing: border-box;
}

.textarea:focus {
  outline: none;
  border-color: #409EFF;
}

.button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-left: 10px;
}

.copy-btn {
  position: absolute;
  bottom: 0;
  right: 0;
}
</style>