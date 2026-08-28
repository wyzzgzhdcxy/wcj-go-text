<style scoped>
table {
  border-spacing: 0;
  border-collapse: collapse;
}



</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <el-tabs v-model="activeName" @tab-click="handleClick" v-if="!hiddenInput">
      <el-tab-pane label="文本数据" name="first">
        <div id="appDiv" style="width:80%">
          文本框内输入数据集
          <el-input
              type="textarea"
              :rows="20"
              :cols="100"
              placeholder="请输入模板数据"
              v-model="inputData">
          </el-input>
          <el-input placeholder="文本分割符" v-model="splitChar">
            <template #prepend>文本分割符</template>
          </el-input>
          <el-button type="success" style="margin-top: 20px;" v-on:click="parseData">下一步</el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="mysql数据库" name="second">
        <div id="appDiv" style="width: 800px">
          <table>
            <tbody>
            <tr>
              <td colspan="2" style="width: 800px">
                <el-input placeholder="输入查询条件按回车查询" v-model="url">
                  <template #prepend>数据库连接</template>
                </el-input>
              </td>
            </tr>
            <tr>
              <td>选择表名:</td>
              <td>
                <el-select v-model="tableName" placeholder="请选择">
                  <el-option
                      v-for="item in tableNames" :key="item" :label="item" :value="item">
                  </el-option>
                </el-select>
              </td>
            </tr>
            <tr>
              <td>
                <el-button type="success" style="margin-top: 20px;" v-on:click="listTables">表名列表</el-button>
              </td>
              <td>
                <el-button type="success" style="margin-top: 20px;" v-on:click="parseDBData">下一步</el-button>
              </td>
            </tr>
            </tbody>
          </table>
        </div>
      </el-tab-pane>
    </el-tabs>
    <div id="mainLeft" v-if="hiddenInput">
      <my-table v-bind:data="data" v-bind:count="10" :show-index="showIndex"/>
      <div v-if="data.length" style="color: red">注意:一共有{{ data.length }}条数据记录,表格只展示10条</div>
      <div id="tplData" style="margin-top: 10px">
        <textarea id="inputData" cols="160" rows="10" v-model="tplData"></textarea>
      </div>
      <div style="margin-bottom: 5px">
        <el-button type="primary" v-on:click="confirm">生成数据</el-button>
        一共处理{{ line }}条数据
      </div>
      <div id="result">
        <textarea id="resultData" cols="160" rows="20" v-model="resultData"></textarea>
      </div>
    </div>
  </div>
</template>

<script>
import MyTable from "./component/myTable.vue";
import {GetTableData, GetTableNames, ReadDemoFile, TplReplace} from "../wailsjs/go/app/App.js";

export default {
  components: {MyTable},
  data() {
    return {
      activeName: 'first',
      result: "",
      inputData: '',
      data: [],
      data1: [],
      username: '',
      url: 'root:wangchaojun@tcp(192.168.31.236:3306)/xxl_job',
      tableName: "xxl_job_info",
      tableNames: [],
      resultData: '',
      line: 0,
      hiddenInput: false,
      showIndex: false,
      splitChar: ',',
      tplData: ''
    }
  },

  computed: {
    displayedItems() {
      // 返回前10条数据
      return this.data.slice(0, 10);
    },
    filteredResults() {
      const searchLowerCase = this.tableName.toLowerCase();
      return this.tableNames.filter(item => {
        return item.name.toLowerCase().includes(searchLowerCase);
      });
    }
  },

  mounted() {
    this.initData();
  },

  methods: {

    async initData() {
      this.inputData = await ReadDemoFile('/tpl_init.txt');
      this.tplData = await ReadDemoFile('/tpl_eg.txt');
    },

    listTables() {
      let that = this;
      GetTableNames(that.url).then(function (ret) {
        that.tableNames = ret;
      }).catch((err) => {
        that.$message.error("连接数据库异常:" + err)
      });
    },
    handleClick(tab, event) {
      console.log(tab, event);
    },
    next() {
      if (this.active++ > 2) this.active = 0
    },
    confirm() {
      let that = this;
      TplReplace(that.tplData, that.data).then(result => {
        that.resultData = result.Result
        that.line = result.Line
      }).catch((err) => {
        console.log(err);
      })
    },

    parseData() {
      this.hiddenInput = true
      let allArr = [];
      console.time('计时器');
      let arr = this.inputData.split("\n")
      for (let i = 0; i < arr.length; i++) {
        let secondArr = arr[i].split(this.splitChar)
        allArr[i] = secondArr
      }
      console.timeEnd('计时器');
      this.data = allArr
      console.dir(this.data[0][0])
    },
    parseDBData() {
      let that = this;
      GetTableData(that.url, that.tableName).then(function (ret) {
        that.data = ret;
        that.hiddenInput = true
      }).catch((err) => {
        that.$message.error("连接数据库异常:" + err)
      })
    }
  }
}
</script>