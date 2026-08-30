<style>
</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="padding: 20px;">
    <div style="margin-bottom: 20px;">
      <el-button type="primary" @click="genIdNo">身份证号生成</el-button>
      <el-button type="primary" @click="parseNo">解析身份证号</el-button>
      <el-button type="success" @click="generatorBankNo">银行卡号生成</el-button>
      <el-button type="success" @click="parseBankCardNo">解析银行卡号</el-button>
    </div>

    <div style="margin-bottom: 20px;">
      <el-input v-model="ipValue" placeholder="输入IP地址" style="width: 200px;"></el-input>
      <el-button type="warning" @click="parseIp" style="margin-left: 10px;">IP解析</el-button>
      <span v-if="ipRes" style="margin-left: 10px;">{{ ipRes }}</span>
    </div>

    <div style="height: 1px;background-color: #e8eaed;margin-bottom: 20px"></div>

    <div v-for="(item, index) in data" :key="index">
      <el-tag type="success" :disable-transitions="true" style="margin: 5px;">{{item}}</el-tag>
    </div>
  </div>
</template>

<script>

import {GenIdNo, ParseIdNo, GenerateBankCardNo, ParseBankCard, IpParse} from "../wailsjs/go/app/App.js";
import {ElMessage, ElMessageBox} from 'element-plus';

export default {
  data() {
    return {
      data: [],
      ipValue: '',
      ipRes: ''
    }
  },

  methods: {
    async generatorBankNo() {
      let that = this;
      ElMessageBox.prompt('请输入要生成的银行卡号个数', '银行卡号生成', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: '10',
        inputType: 'number',
        inputPattern: /^[1-9]\d{0,2}$/,
        inputErrorMessage: '请输入1-100之间的数字'
      }).then(({ value }) => {
        let num = parseInt(value);
        if (num <= 0 || num > 100) {
          ElMessage.error('请输入1-100之间的数字');
          return;
        }
        GenerateBankCardNo(num).then(resData => {
          that.data = resData;
        })
      }).catch(() => {});
    },

    async parseBankCardNo() {
      let that = this;
      ElMessageBox.prompt('请输入银行卡号', '解析银行卡号', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'number',
      }).then(({ value }) => {
        ParseBankCard(value).then(resData => {
          that.data = [resData];
        })
      }).catch(() => {});
    },

    async genIdNo() {
      let that = this;
      ElMessageBox.prompt('请输入要生成的身份证号个数', '身份证号生成', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: '10',
        inputType: 'number',
        inputPattern: /^[1-9]\d{0,2}$/,
        inputErrorMessage: '请输入1-100之间的数字'
      }).then(({ value }) => {
        let num = parseInt(value);
        if (num <= 0 || num > 100) {
          ElMessage.error('请输入1-100之间的数字');
          return;
        }
        GenIdNo(num).then(resData => {
          that.data = resData;
        })
      }).catch(() => {});
    },


    async parseNo() {
      let that = this;
      ElMessageBox.prompt('请输入身份证号', '解析身份证号', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'number',
      }).then(({ value }) => {
        ParseIdNo(value).then(resData => {
          that.data = [resData];
        })
      }).catch(() => {});
    },

    async parseIp() {
      let that = this;
      IpParse(that.ipValue).then(resData => {
        that.ipRes = resData;
      })
    }
  }
}
</script>