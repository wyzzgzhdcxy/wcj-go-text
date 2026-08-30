<template xmlns="http://www.w3.org/1999/html">
  <div style="display: flex;width: 100%">
    <div class="left-div">
      <my-sql-editor v-model="inputData"/>
    </div>
    <div class="edit_right_div">
      <div style="height: 80px"></div>
      <el-input style="height: 30px;width: 290px" placeholder="输入查询条件按回车查询" v-model="url">
        <template #prepend>URL:</template>
      </el-input>
      <hr>
      <el-button style="margin-top: 10px" type="success" v-on:click="getDDL">加载ddl</el-button>
      <el-button style="margin-top: 10px" type="success" v-on:click="genDocument">生成数据库文档</el-button>
      <hr>
      <el-button style="margin-top: 10px" type="success" v-on:click="viewValue">生成代码</el-button>
      <el-button style="margin-top: 10px" type="success" v-on:click="genCodeByConnUrl">根据url生成代码</el-button>
    </div>
  </div>

  <div class="editValue" style="background-color: white" v-if="editDisplay">
    <div class="edit_header">
      <el-button type="primary" @click="closeEditView">关闭</el-button>
    </div>
    <my-java-editor v-model="editValue"/>
  </div>
</template>

<style scoped>

div {
  width: 100%;
  height: 100%;
}

</style>


<script>
import {w} from './js/fun.js';
import {
  DdlSql,
  GenCodeByConnUrl,
  GenMySqlDoc,
  GetLastConnUrl,
  OpenExplorer,
  ReadDemoFile
} from "../wailsjs/go/app/App.js";
import MyJavaEditor from "./component/myJavaEditor.vue";
import MySqlEditor from "./component/mySqlEditor.vue";


export default {
  components: {MySqlEditor: MySqlEditor, MyJavaEditor: MyJavaEditor},

  data() {
    return {
      editDisplay: false,
      tableList: [],
      inputData: '',
      editValue: "",
      url: 'root:wangchaojun@tcp(192.168.31.236:3306)/xxl_job'
    }
  },
  mounted() {
    this.initData();
  },
  computed: {},
  methods: {

    genDocument() {
      let that = this;
      if (that.url.length === 0) {
        this.$message.error('数据库连接不能为空');
        return;
      }
      GenMySqlDoc(that.url).then(res => {
        console.log(res);
        this.$message.info('数据库文档生成成功!');
        OpenExplorer(".")
      })
    },

    async initData() {
      this.inputData = await ReadDemoFile('/code_gen.ddl');
      let lastConnUrl = await GetLastConnUrl();
      if (!(lastConnUrl === "")) {
        this.url = lastConnUrl
      }
    },

    getDDL() {
      let that = this;

      DdlSql(that.url).then(function (error,res) {
        if (error.length !==0 ){
          this.$message.error("异常:" + error)
        }
        that.inputData = res;
        if (res.length === 0) {
          this.$message.error("连接数据库异常:" + that.url)
        }
      }).catch((err) => {
        this.$message.error("连接数据库异常:" + err)
      })
    },
    closeEditView() {
      this.editDisplay = false
    },

    editValueJsonFormat() {
      this.editValue = JSON.stringify(JSON.parse(this.editValue), null, 2);
    },
    viewValue() {
      let that = this;
      if (that.inputData.length === 0) {
        this.$message.error('ddl语句为空');
        return;
      }
      w.post(that, "/tools/genCodeByDDL", decodeURIComponent(that.inputData), function (resData) {
        that.editValue = resData;
      })
      this.editDisplay = true
    },
    genCodeByConnUrl() {
      let that = this;
      if (that.url === 0) {
        this.$message.error('url为空');
        return;
      }
      GenCodeByConnUrl(that.url).then(function (resData) {
        that.editValue = resData;
      });
      this.editDisplay = true
    },
  }
}
</script>