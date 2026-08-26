<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <table>
      <tbody>
        <tr>
          <td>目录:</td>
          <td>
            <el-input v-model="req.dir" placeholder="目录路径"></el-input>
          </td>
          <td>
          </td>
        </tr>
        <tr>
          <td colspan="2"><textarea cols="120" rows="20" v-model="req.categories" @input=""
                                    placeholder="请输入多行文本..."></textarea></td>
          <td>
            <el-button type="primary" v-on:click="categorizeFile">归类</el-button>
            <el-button type="primary" v-on:click="getFilePrefixesInDir">获取前缀</el-button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div style="height: 1px;background-color: black;margin-top: 10px;margin-bottom: 10px"></div>
  <div>{{ data }}</div>
</template>

<script>
import {CategorizeFiles, GetFilePrefixesInDir} from "../wailsjs/go/main/App.js";


export default {
  data() {
    return {
      req: {"dir": "F:\\MP4_待分类_待上传", "categories": "张宇,张碧晨"},
      data: []
    }
  },

  mounted() {
  },

  methods: {

    categorizeFile() {
      let that = this;
      CategorizeFiles(that.req.dir, that.req.categories).then(resData => {
        console.dir(resData)
        that.data = ["处理完成"];
      })
    },
    getFilePrefixesInDir() {
      let that = this;
      GetFilePrefixesInDir(that.req.dir, that.req.count).then(resData => {
        console.dir(resData)
        that.req.categories = resData.join(',');
      })
    },
  }
}
</script>