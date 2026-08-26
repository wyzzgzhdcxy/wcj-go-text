<style scoped>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.td_result {
  border: 1px solid #ACBED1;
}
</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <div id="appDiv">
      <table>
        <tbody>
        <tr>
          <td></td>
          <td v-if="!hiddenInput"><textarea id="jsonData" cols="120" rows="20" v-model="jsonData"
                                            @input="jsonChange"></textarea></td>
          <td id="td_filter">
            <div v-if="hiddenInput" style="margin-bottom: 10px">
              <el-checkbox-group v-model="checkboxGroup1">
                <el-checkbox-button @change="convertTable" v-for="it in checkBoxFilterList" :label="it"
                                    :key="it">
                  {{ it }}
                </el-checkbox-button>
              </el-checkbox-group>
            </div>
          </td>
        </tr>
        <tr>
          <td></td>
          <td>
            <el-button v-if="hiddenInput" type="success" v-on:click="selectAll">全选</el-button>
            <el-button v-if="hiddenInput" type="success" v-on:click="noSelect">不选</el-button>
            <el-button v-if="hiddenInput" type="success" v-on:click="resetInput">重置</el-button>
            <el-button v-if="!hiddenInput" type="success" v-on:click="getArrEle">获取数组元素</el-button>
            <el-button v-if="!hiddenInput" v-for="item in arrFieldList" type="info" v-on:click="gotoTablePage">
              {{ item }}
            </el-button>
          </td>
          <td></td>
        </tr>
        </tbody>
      </table>
    </div>

    <HR/>
    <div id="result">
      <table style="border: black;border:1px" id="resultTable">
        <tbody>
        <tr style="background-color: #c2e7b0">
          <th v-if="tableList !== undefined && tableList.length > 0" class='td_result'>序号</th>
          <th v-for="(item, index) in checkboxGroup1" class='td_result'>{{ item }}</th>
        </tr>
        <tr v-for="(item, index) in tableList" :key="index">
          <td class='td_result'>{{ index }}</td>
          <td v-for="(it, index) in item" class='td_result'>{{ it }}</td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>

import {ReadDemoFile} from "../wailsjs/go/main/App.js";

export default {
  data() {
    return {
      arrFieldSelect: "",
      arrFieldList: [],
      tableList: [],
      jsonData: '',
      hasFilterCheckBox: '',
      hiddenInput: false,
      //所有过滤字段
      checkBoxFilterList: [],
      //选中的过滤字段
      checkboxGroup1: []
    }
  },
  computed: {
    getCheckBoxSelectField() {
      return this.checkBoxFilterList.filter((item) => {
        return item.checked
      })
    }
  },

  mounted() {
    this.initData();
  },


  methods: {

    async initData() {
      this.jsonData = await ReadDemoFile('/json_eg1.json');
    },

    // initializeEditor() {
    //   const editor = CodeMirror.fromTextArea(this.$refs.editor, {
    //     // 在这里可以配置CodeMirror的选项
    //     mode: 'javascript',
    //     lineNumbers: true,
    //   });
    //   // 可以在这里对CodeMirror进行进一步的配置或使用editor实例
    // },
    getArrEle() {
      let that = this;
      let jsonData = JSON.parse(this.jsonData);
      if (Array.isArray(jsonData)) {
        that.arrFieldList.push("all");
      } else {
        that.arrFieldList = [];
        for (let key in jsonData) {
          that.addFilterSelect('', key, jsonData)
        }
      }
    },

    resetInput() {
      location.reload()
    },

    addFilterSelect(pKey, key, jsonObj) {
      let that = this;
      let obj = jsonObj[key];
      if (Array.isArray(obj)) {
        let filterKey = pKey + "." + key;
        if (pKey === '') {
          filterKey = key;
        }
        that.arrFieldList.push(filterKey);
        return;
      } else {
        if (typeof obj == 'object') {
          for (let k1 in obj) {
            that.addFilterSelect(key, k1, obj)
          }
        }
      }
    },
    genFilterCheckBox() {
      this.hasFilterCheckBox = true;
      let keysList = [];
      for (let key in this.getJsonData()[0]) {
        keysList.push(key);
      }
      this.checkBoxFilterList = keysList;
      this.checkboxGroup1 = keysList.slice(0, 10);
    },
    gotoTablePage(event) {
      let that = this;
      //没有过滤选择页面
      that.arrFieldSelect = event.target.innerText;
      if (that.arrFieldSelect.length == 0) {
        this.$message.error('没有获取到数组元素');
        return;
      }
      this.genFilterCheckBox();
      this.convertTable()
    },
    convertTable() {
      let that = this;
      that.hiddenInput = true;
      let jsonArrValue = this.getJsonData();
      let tempTableList = [];
      jsonArrValue.forEach(function (jsonItem) {
        let tableItem = {};
        that.checkboxGroup1.forEach(function (item) {
          tableItem[item] = jsonItem[item];
        })
        tempTableList.push(tableItem);
      })
      this.tableList = tempTableList;
    },
    selectAll() {
      this.checkboxGroup1 = this.checkBoxFilterList
      this.convertTable();
    },
    jsonChange(event) {
      this.checkBoxFilterList = [];
    },
    noSelect() {
      this.checkboxGroup1 = []
      this.convertTable();
    },
    getJsonData() {
      let that = this;
      let jsonArrValue = "";
      try {
        jsonArrValue = JSON.parse(this.jsonData);
        if (that.arrFieldSelect == '' || that.arrFieldSelect == 'all' || that.arrFieldSelect == null) {
          return jsonArrValue;
        } else {
          let funcStr = "function parseData(v){return v." + that.arrFieldSelect + "}";
          let testFunc = eval("(false || " + funcStr + ")");
          return testFunc(jsonArrValue);
        }
      } catch (err) {
        alert("数据格式错误，请输入json数组" + err)
      }
      return jsonArrValue;
    }
  }
}
</script>