<template>
  <div class="page-container">
    <div class="content-wrapper">
      <div class="textarea-wrapper">
        <textarea v-model="inputData" class="textarea" placeholder="输入文本..."></textarea>
      </div>
      <div class="button-group">
        <el-button type="success" @click="buttonsClick('转json')">{ } 转JSON</el-button>
        <el-button type="success" @click="buttonsClick('全部大写')">🔠 全部大写</el-button>
        <el-button type="success" @click="buttonsClick('全部小写')">🔡 全部小写</el-button>
        <el-button type="success" @click="buttonsClick('删除重复行')">📋 去重</el-button>
        <el-button type="success" @click="buttonsClick('删除空行')">🗑️ 去空行</el-button>
        <el-button type="success" @click="buttonsClick('去除不可见字符')">👁️ 去不可见</el-button>
        <el-button type="success" @click="buttonsClick('下划线转驼峰')">🐫 转驼峰</el-button>
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
    buttonsClick(val) {
      if (this.inputData === "") {
        ElNotification({title: '数据为空', message: "输入数据为空", position: 'bottom-right', type: 'error'});
        return;
      }
      if (val === "转json") {
        let arr = this.inputData.split("\n");
        this.result = JSON.stringify(arr, null, 4)
      } else if (val === "全部大写") {
        this.result = this.inputData.toUpperCase();
      } else if (val === "全部小写") {
        this.result = this.inputData.toLowerCase();
      } else if (val === "删除重复行") {
        let repeatLines = this.inputData.split("\n");
        let uniqueLines = [...new Set(repeatLines)];
        this.result = uniqueLines.join("\n")
      } else if (val === "删除空行") {
        this.result = this.inputData.replace(/\n\n+/g, "\n");
      } else if (val === "去除不可见字符") {
        this.result = this.inputData.replace(/[\u200B-\u200D\uFEFF]/g, '');
      } else if (val === "下划线转驼峰") {
        this.result = this.inputData.split('_').map((word, index) => {
          if (index === 0) return word;
          return word.charAt(0).toUpperCase() + word.slice(1);
        }).join('');
      }
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