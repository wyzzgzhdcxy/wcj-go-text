<template>
  <div style="width: 100%">
    <div v-for="(box, index) in formulaBoxes" :key="index" class="formula-box">
      <div class="input-container">
        <input  v-model="box.formula" @input="calculate(index)"
               placeholder="输入公式，例如: 5+3 * 2" class="formula-input">
      </div>

      <div class="result-display">
        <span v-if="box.result !== null" style="font-size: 50px">={{ box.result }}</span>
        <span v-else class="error-message" v-show="box.error">{{ box.error }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import {EvaluableExpression} from "../wailsjs/go/main/App.js";

export default {
  data() {
    return {
      formulaBoxes: []
    }
  },
  methods: {
    addFormulaBox() {
      this.formulaBoxes.push({
        formula: '',
        result: null,
        error: ''
      })
    },
    async calculate(index) {
      if ((this.formulaBoxes.length === index + 1) && (this.formulaBoxes.length < 3)) {
        this.addFormulaBox()
      }
      const box = this.formulaBoxes[index]
      try {
        if (box.formula.trim() === '') {
          box.result = null
          box.error = ''
          return
        }

        // 安全计算EvaluableExpression
        box.result = await EvaluableExpression(box.formula)
        box.error = ''
      } catch (e) {
        box.result = null
        box.error = '公式无效'
      }
    }
  },
  mounted() {
    this.addFormulaBox()
  }
}
</script>

<style>
/* 重置页面背景为透明 */
body, html, #app {
  background-color: transparent !important;
  margin: 0;
  padding: 0;
}

.app-container {
  padding: 20px;
  margin-top: 40px;
  font-family: Arial, sans-serif;
  background-color: transparent; /* 容器透明 */
}


/* 只有这个div有浅蓝色背景 */
.formula-box {
  width: 900px;
  height: 200px;
  background-color: lightgoldenrodyellow; /* 浅蓝色背景 */
  border-radius: 15px;
  margin-bottom: 20px;
  position: relative;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  box-sizing: border-box;
}

.input-container {
  margin-bottom: 15px;
  width: 800px;
}

.formula-input {
  width: 800px;
  height: 60px;
  padding: 10px;
  font-size: 50px;
  border: 1px solid #d9d9d9;
  border-radius: 5px;
  /* 基础样式（保持与父容器一致） */
  background-color: inherit; /* 继承父元素背景色 */
  /* 可选：确保与其他行内元素对齐 */
  vertical-align: middle;
  margin: 0;
}

.formula-input:focus {
  outline: none;
  border-color: #1890ff;
}

.result-display {
  position: absolute;
  left: 600px;
  bottom: 20px;
  font-size: 18px;
  color: #333;
}

.error-message {
  color: #f5222d;
}

</style>