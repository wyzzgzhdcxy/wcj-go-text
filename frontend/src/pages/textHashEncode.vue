<template>
  <div class="page-container">
    <div class="content-wrapper">
      <div class="textarea-wrapper">
        <textarea v-model="inputData" class="textarea" placeholder="输入文本..."></textarea>
      </div>
      <div class="button-group">
        <el-button type="success" @click="textEncode('sha1')">#️⃣ SHA1</el-button>
        <el-button type="success" @click="textEncode('sha256')">#️⃣ SHA256</el-button>
        <el-button type="success" @click="textEncode('sha512')">#️⃣ SHA512</el-button>
        <el-button type="success" @click="textEncode('hex编码')">🔢 HEX编码</el-button>
        <el-button type="success" @click="textEncode('hex解码')">🔢 HEX解码</el-button>
        <el-button type="success" @click="textEncode('ascii编码')">🔢 ASCII编码</el-button>
        <el-button type="success" @click="textEncode('ascii解码')">🔢 ASCII解码</el-button>
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
import {CopyToClipboard, TextEncode} from "../wailsjs/go/main/App.js";
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
    async textEncode(val) {
      if (this.inputData === "") {
        ElNotification({title: '数据为空', message: "输入数据为空", position: 'bottom-right', type: 'error'});
        return;
      }
      this.result = await TextEncode(this.inputData, val);
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