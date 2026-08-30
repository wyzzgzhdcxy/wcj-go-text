<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.input-label {
  display: inline-block;
  font-size: 13px;
  color: #909399;
  margin-right: 6px;
  vertical-align: middle;
}

.preview-img {
  max-width: 200px;
  max-height: 200px;
  display: block;
  margin: 8px auto;
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
            <span class="input-label">图片大小</span>
            <el-input-number v-model="imgSizeRaw" :min="16" :max="1920" :step="16"
                             controls-position="right" @change="font2Image"/>
          </td>
          <td>
            <span class="input-label">圆角半径</span>
            <el-input-number v-model="cornerRadius" :min="0" :max="1000"
                             controls-position="right" @change="font2Image"/>
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
        <td><img v-show="!!pngUrl.length" class="preview-img" :src="pngUrl" alt="png图片"></td>
        <td><img v-show="!!jpgUrl.length" class="preview-img" :src="jpgUrl" alt="jpg图片"></td>
        <td><img v-show="!!icoUrl.length" class="preview-img" :src="icoUrl" alt="ico图片"></td>
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
      imgSizeRaw: 256,
      jpgUrl: '',
      pngUrl: '',
      pngPath: '',
      icoUrl: ''
    }
  },

  mounted() {
    this.font2Image()
  },

  methods: {
    async selectFile() {
      try {
        const result = await SelectFile()
        if (result) {
          this.pngPath = result
        }
      } catch (error) {
        console.error('选择文件失败:', error)
        this.$message.error('选择文件失败: ' + error)
      }
    },
    convertPng() {
      if (!this.pngPath) {
        this.$message.warning('请先导入要转换的 PNG 图片')
        return
      }
      PngToIcon(this.pngPath).then(() => {
        ElNotification({
          title: 'png图片格式转换结果',
          message: '转换成功（ICO/GIF/BMP 已生成在原图片目录）',
          position: 'bottom-right',
          type: 'success',
        })
      }).catch(error => {
        ElNotification({
          title: 'png图片格式转换结果',
          message: '转换失败: ' + error,
          position: 'bottom-right',
          type: 'error',
        })
      })
    },
    openExplorer() {
      OpenImgExplorer().catch(error => {
        this.$message.error('打开图片目录失败: ' + error)
      })
    },
    font2Image() {
      if (!this.imgText.length || this.imgSizeRaw === 0) {
        return;
      }
      const corner = isNaN(Number(this.cornerRadius)) ? 10 : Number(this.cornerRadius);
      Font2Image(this.imgText, this.imgSizeRaw, corner).then(result => {
        this.pngUrl = result.pngUrl || ''
        this.icoUrl = result.icoUrl || ''
        this.jpgUrl = result.jpgUrl || ''
      }).catch(err => {
        this.$message.error("图片生成异常:" + err)
      })
    }
  }
}
</script>