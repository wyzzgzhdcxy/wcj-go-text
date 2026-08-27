<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.td_style {
  border: 1px solid #ACBED1;
  text-align: right;
}
</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <div id="appDiv">
      <table>
        <tbody>
          <tr>
            <td>
              <el-input placeholder="税前工资" v-model="salaryAmount" v-on:change="computerSalary">
                <template #prepend>税前工资</template>
              </el-input>
            </td>
            <td>
              <el-input placeholder="社保基数" v-model="socialSecurityBase">
                <template #prepend>社保基数</template>
              </el-input>
            </td>
            <td>
              <el-input placeholder="免征额" v-model="exemptionAmount">
                <template #prepend>免征额</template>
              </el-input>
            </td>
          </tr>
          <tr>
            <td colspan="4">
              <el-input placeholder="每个月的额外收入,用逗号隔开" v-model="extraSalary">
                <template #prepend>每个月的额外收入,用逗号隔开</template>
              </el-input>
            </td>
          </tr>
          <tr>
            <td>
              <el-button type="primary" v-on:click="computerSalary">计算</el-button>
            </td>
            <td></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div style="height: 1px;background-color: black;margin-top: 10px;margin-bottom: 10px"></div>
    <div id="result">
      <myTable v-bind:data="data.List" v-bind:count="100" :class-stype="tdStyle" v-bind:header="header"></myTable>
    </div>
    <div style="height: 1px;background-color: black;margin-top: 10px;margin-bottom: 10px"></div>
    <div>
      <table style="border: black;border:1px" id="resultTable">
        <tbody>
          <tr>
            <td>
              <el-tag type="success" :disable-transitions="true">税前总收入:</el-tag>
            </td>
            <td>{{ data.TotalSalary }}元</td>
            <td>
              <el-tag type="success" :disable-transitions="true">税后总收入:</el-tag>
            </td>
            <td>{{ data.TotalRealSalary }}元</td>
            <td>
              <el-tag type="success" :disable-transitions="true">社保总缴纳:</el-tag>
            </td>
            <td>{{ data.SocialBxTotal }}元</td>
          </tr>
          <tr>
            <td style="color: rebeccapurple;font-size: 12px">注意:免征额=含5000+专项扣除</td>
          </tr>
        </tbody>
      </table>
    </div>


    <div style="height: 1px;background-color: black;margin-top: 10px;margin-bottom: 10px"></div>
    <H3>个税计算规则：</H3>
    <div id="result">
      <table style="border: black;border:1px" id="resultTable">
        <tbody>
          <tr style="background-color: #c2e7b0">
            <th class='td_result'>级数</th>
            <th style="width: 300px" class='td_result'>全年应纳税所得额</th>
            <th class='td_result'>税率(%)</th>
            <th class='td_result'>速算扣除</th>
          </tr>
          <tr style="text-align: right" v-for="(item, index) in data.TaxRuleList" :key="index">
            <td class='td_result'>{{ index }}</td>
            <td class='td_result'>超过{{ item.Start }}元至{{ item.End }}的部分</td>
            <td class='td_result'>{{ item.TaxRate }}</td>
            <td class='td_result'>{{ item.Deduct }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>


import {w} from "./js/fun.js";
import {SalaryList} from "../wailsjs/go/app/App.js";

export default {
  data() {
    return {
      salaryAmount: "10000",
      extraSalary: "652.73,493.33,700,940,860",
      socialSecurityBase: "10000",
      exemptionAmount: "6500",
      cardType: "idCard",
      tdStyle: "td_style_2",
      inputValue: 10,
      header: ["月份", "税前收入", "养老保险", "医疗保险", "公积金保险", "生育保险", "失业保险", "社保合计", "当月个税", "全年纳税额", "免征额", "已缴税额", "税后收入"],
      taxRateHeader: ["级数", "全年应纳税所得额", "税率(%)", "速算扣除"],
      data: []
    }
  },
  filters: {
    numFilter(value) {
      // 截取当前数据到小数点后两位
      let realVal = Number(value).toFixed(4)
      return realVal
    }
  },

  mounted() {
    this.computerSalary();
  },
  methods: {

    async computerSalary() {
      let that = this;
      SalaryList({
        salaryAmount: that.salaryAmount,
        exemptionAmount: that.exemptionAmount,
        socialSecurityBase: that.socialSecurityBase,
        extraSalary: that.extraSalary
      }).then(res => {
        console.log(res);
        that.data = res;
      })
    }
  }
}
</script>