<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <div id="appDiv">
      <table>
        <tbody>
        <tr>
          <td>
            <el-input placeholder="文字" v-model="imgText" @input="font2Image">
              <template #prepend>文字</template>
            </el-input>
          </td>
          <td>
            <el-input placeholder="图片大小" v-model="imgSize" @input="font2Image">
              <template #prepend>图片大小</template>
            </el-input>
          </td>
          <td>
            <el-input placeholder="圆角半径" type="number" v-model="cornerRadius" @input="font2Image">
              <template #prepend>圆角半径</template>
            </el-input>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
    <HR/>
    <table>
      <tbody>
      <tr>
        <td>
          <el-tag type="success">png</el-tag>
        </td>
        <td>
          <el-tag type="success">jpg</el-tag>
        </td>
        <td>
          <el-tag type="success">ico</el-tag>
        </td>
      </tr>
      <tr>
        <td><img v-show="!!pngUrl.length" :src="pngUrl" alt="png图片"></td>
        <td><img v-show="!!jpgUrl.length" :src="jpgUrl" alt="jpg图片"></td>
        <td><img v-show="!!icoUrl.length" :src="icoUrl" alt="ico图片"></td>
      </tr>
      <tr>
        <td colspan="3">
          <el-button type="success" v-on:click="openExplorer">打开本地图片路径</el-button>
        </td>
      </tr>
      </tbody>
    </table>
  </div>

  <hr/>
  <div>
    <el-button type="success" v-on:click="selectFile">导入png图片</el-button>
    {{ pngPath }}
    <el-button type="success" v-on:click="convertPng">格式转换</el-button>
  </div>
  <hr/>
  <div>
    图片相关在线工具:<BR/>
    <ul>
      <li><a href="https://emoji.aranja.com/" target="_blank"> Emoji to image</a></li>
    </ul>
  </div>
</template>

<script>
import {Font2Image, OpenImgExplorer, PngToIcon, SelectFile} from "../wailsjs/go/app/App.js";
import {ElNotification} from "element-plus";

export default {
  data() {
    return {
      cornerRadius: 0,
      imgText: '测',
      imgSize: 256,
      jpgUrl: '',
      pngUrl: '',
      pngPath: '',
      icoUrl: ''
    }
  },

  computed: {
    imgSize: {
      get() {
        return parseInt(this.imgSize, 10)
      },
      set(newValue) {
        if (newValue > 1920) {
          this.imgSize = 1920;
        } else {
          if (newValue.length === 0) {
            this.imgSize = 0;
          } else {
            this.imgSize = parseInt(newValue, 10);
          }
        }
      }
    }
  },


  mounted() {
    this.font2Image()
  },

  methods: {
    async selectFile() {
      let that = this;
      try {
        const result = await SelectFile()
        if (result) {
          that.pngPath = result
        }
      } catch (error) {
        console.error('选择目录失败:', error)
      }
    },
    convertPng() {
      let that = this;
      PngToIcon(this.pngPath).then(result => {
        ElNotification({
          title: 'png图片格式转换结果',
          message: result,
          position: 'bottom-right',
          type: 'success',
        })
      })
    },
    openExplorer() {
      OpenImgExplorer()
    },
    font2Image() {
      let that = this;
      if (that.imgText.length === 0 || that.imgSize === 0) {
        return;
      }
      that.result = ''
      let corner = isNaN(that.cornerRadius) ? 10 : Number(that.cornerRadius);
      Font2Image(that.imgText, that.imgSize, corner).then(result => {
        that.pngUrl = result.pngUrl
        that.icoUrl = result.icoUrl
        that.jpgUrl = result.jpgUrl
      }).catch(err => {
        that.$message.error("图片生成异常:" + err)
      })
    }
  }
}
</script>