<template xmlns="http://www.w3.org/1999/html">


  <div style="display: flex" class="container">
    <div style="flex: 0 0 calc(100% - 300px)">
      <my-java-editor v-model="inputData" language="java"/>
    </div>
    <div style="flex: 0 0 293px;margin-left: 5px; border: 1px solid blanchedalmond;border-radius: 10px;">
      <el-button style="margin-top: 10px;margin-bottom: 10px" type="success" v-on:click="class2Json">转json</el-button>
      <el-button style="margin-top: 10px;margin-bottom: 10px" type="success" v-on:click="class2go">转golang</el-button>
    </div>
  </div>

  <div class="editValue" style="background-color: white" v-if="editDisplay">
    <div class="edit_header">
      <el-button style="height: 28px;width: 50px;margin-left: 10px" type="primary" @click="closeEditView">关闭</el-button>
    </div>
    <div class="edit_right_div">
      <my-sql-editor v-if="language === 'sql'" v-model="editValue"/>
      <my-json-editor v-if="language === 'json'" v-model="editValue"/>
      <my-go-editor v-if="language === 'go'" v-model="editValue"/>
    </div>

  </div>
</template>

<script setup>
import MyJavaEditor from "./component/myJavaEditor.vue";
import MySqlEditor from "./component/mySqlEditor.vue";
import MyJsonEditor from "./component/myJsonEditor.vue";
import MyGoEditor from "./component/myGoEditor.vue";
</script>

<style scoped>




div {
  width: 100%;
  height: 100%;
}

table {
  border-spacing: 0;
  border-collapse: collapse;
}


</style>


<script>
import {GenGoCodeByJsonStr, JavaCodeToJson, ReadDemoFile} from "../wailsjs/go/main/App.js";
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
      this.inputData = await ReadDemoFile('/class2Json_eg1.java');
    },


    async class2Json() {
      this.editValue = await JavaCodeToJson(this.inputData)
      this.editDisplay = true
      this.language = "json"
    },

    async class2go() {
      const jsonStr = await JavaCodeToJson(this.inputData)
      this.editValue = await GenGoCodeByJsonStr(jsonStr)
      this.editDisplay = true
      this.language = "go"
    },
    closeEditView() {
      this.editDisplay = false
    }
  }
}
</script>







