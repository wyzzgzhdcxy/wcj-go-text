<template xmlns="http://www.w3.org/1999/html">


  <div style="display: flex;width: 100%">
    <div style="flex: 0 0 calc(100% - 300px);width: calc(100% - 300px)">
      <my-json-editor v-model="inputData" height="100%" width="100%"/>
    </div>

    <div class="edit_right_div">
      <div style="height: 50px"></div>
      <div>
        <el-button type="success" v-on:click="jsonFormat">格式化</el-button>
        <el-button type="success" v-on:click="jsonCompress">压缩</el-button>
        <hr>
        <el-button style="margin-top: 10px;margin-bottom: 10px" type="success" v-on:click="json2Java">生成Java代码</el-button>
        <el-button style="margin-top: 10px;margin-bottom: 10px" type="success" v-on:click="json2Go">生成Go代码</el-button>
        <hr>
        <el-button type="success" v-on:click="json2Sql">json转sql</el-button>
      </div>
    </div>


    <div class="editValue" style="background-color: white" v-if="editDisplay">
      <div class="edit_header">
        <div class="header-buttons">
          <el-button type="primary" @click="closeEditView">关闭</el-button>
        </div>
      </div>
      <div style="height: calc(100% - 50px)">
        <my-sql-editor v-if="language === 'sql'" v-model="editValue" height="100%" width="100%"/>
        <my-java-editor v-if="language === 'java'" v-model="editValue" height="100%" width="100%"/>
        <my-go-editor v-if="language === 'go'" v-model="editValue" height="100%" width="100%"/>
      </div>

    </div>
  </div>

</template>

<script setup>
import MyJsonEditor from "./component/myJsonEditor.vue";
import MySqlEditor from "./component/mySqlEditor.vue";
import MyJavaEditor from "./component/myJavaEditor.vue";
import MyGoEditor from "./component/myGoEditor.vue";
</script>

<style scoped>


table {
  border-spacing: 0;
  border-collapse: collapse;
}


div {
  width: 100%;
  height: 100%;
}


</style>


<script>
import {GenGoCodeByJsonStr, GenJavaCodeByJsonStr, ReadDemoFile} from "../wailsjs/go/app/App.js";
import {w} from "./js/fun.js";


export default {
  data() {
    return {
      editKey: "",
      editValue: "",
      editDisplay: false,
      language: "java",
      tableList: [],
      inputData: ''
    }
  },
  mounted() {
    this.initData();
  },
  computed: {},
  methods: {
    async initData() {
      this.inputData = await ReadDemoFile('/json_eg1.json');
    },

    jsonCompress() {
      this.inputData = JSON.stringify(JSON.parse(this.inputData)).replace(/\s+/g, '');
    },

    json2Sql() {
      try {
        let that = this;
        let jsonObj = JSON.parse(this.inputData);
        that.editValue = w.arrayToSQLInsert(jsonObj);
        that.editDisplay = true
        that.language = "sql"
      } catch (err) {
        this.$message.error("请求异常:" + err)
      }
    },

    jsonFormat() {
      this.inputData = JSON.stringify(JSON.parse(this.inputData), null, 2);
    },
    closeEditView() {
      this.editDisplay = false
    },
    async json2Go() {
      this.editValue = await GenGoCodeByJsonStr(this.inputData)
      this.editDisplay = true
      this.language = "go"
    },
    async json2Java() {
      this.editValue = await GenJavaCodeByJsonStr(this.inputData)
      this.language = "java"
      this.editDisplay = true
    }
  }
}
</script>