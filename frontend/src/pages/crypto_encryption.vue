<template>
  <el-tabs v-model="activeName" @tab-click="handleClick" :tab-position="'left'">
    <el-tab-pane name="genKey" style="width:1100px;">
      <template #label>
        <span class="fa-solid fa-square-caret-right icon_color1" style="font-size: 20px;"></span>
        <span style="font-size: 20px">RSA生成密钥</span>
      </template>
        <div class="left" id="mainLeft">
          <div id="appDiv">
            <el-input placeholder="密钥位数" v-model="keyBits" style="width: 400px">
              <template #prepend>密钥位数</template>
            </el-input>
            <BR/>
            <el-button style="margin-top: 10px" type="success" v-on:click="genKey">生成密钥文件</el-button>
          </div>
        </div>
    </el-tab-pane>
    <el-tab-pane name="cryptoStr" style="width: 1100px">
      <template #label>
        <span class="fa-solid fa-square-caret-right icon_color1" style="font-size: 20px;"></span>
        <span style="font-size: 20px">RSA加密字符串</span>
      </template>
        <!-- 密钥导入区域 -->
        <div class="import-row">
          <el-button type="primary" @click="importRsaKeyFile(0)" icon="el-icon-upload" class="import-btn">
            导入 RSA 公钥
          </el-button>
          <div class="file-path">
            <el-tag type="info" v-if="!pubkPath">未选择公钥文件</el-tag>
            <el-tag type="success" v-else>
              <el-icon>
                <document/>
              </el-icon>
              {{ pubkPath }}
            </el-tag>
          </div>
        </div>
        <div class="import-row">
          <el-button type="primary" @click="importRsaKeyFile(1)" icon="el-icon-upload" class="import-btn">
            导入 RSA 私钥
          </el-button>
          <div class="file-path">
            <el-tag type="info" v-if="!prikPath">未选择私钥文件</el-tag>
            <el-tag type="warning" v-else>
              <el-icon>
                <document/>
              </el-icon>
              {{ prikPath }}
            </el-tag>
          </div>
        </div>

        <!-- 输入区域 -->
        <div class="key-import-section">
          <el-input type="textarea" :rows="12" placeholder="请输入要加密/解密的文本..." v-model="inputData"
                    resize="none" class="input-textarea"/>
        </div>
        <!-- 操作按钮区域 -->
        <div class="action-buttons">
          <el-button type="success" @click="rsaCryptoStr" icon="el-icon-lock" :disabled="!pubkPath || !inputData">
            加密
          </el-button>
          <el-button type="warning" @click="rsaDeCryptoStr" icon="el-icon-unlock" :disabled="!prikPath || !inputData">
            解密
          </el-button>
        </div>

        <!-- 结果展示区域 -->
        <div class="result-section">
          <el-input
              type="textarea"
              :rows="19" placeholder="加密/解密结果将显示在这里..."
              v-model="result" readonly
              resize="none" class="result-textarea">
            <template #append>
              <el-button @click="copyResult" icon="el-icon-document-copy" title="复制结果" :disabled="!result"/>
            </template>
          </el-input>
        </div>
    </el-tab-pane>
    <el-tab-pane name="cryptoFile"  style="width: 1100px">
      <template #label>
        <span class="fa-solid fa-square-caret-right icon_color1" style="font-size: 20px;"></span>
        <span style="font-size: 20px">RSA加密文件</span>
      </template>
        <!-- 密钥导入区域 -->
        <div class="import-row">
          <el-button
              type="primary"
              @click="importRsaKeyFile(0)"
              icon="el-icon-upload"
              class="import-btn"
          >
            导入 RSA 公钥
          </el-button>
          <div class="file-path">
            <el-tag type="info" v-if="!pubkPath">未选择公钥文件</el-tag>
            <el-tag type="success" v-else>
              <el-icon>
                <document/>
              </el-icon>
              {{ pubkPath }}
            </el-tag>
          </div>
        </div>

        <div class="import-row">
          <el-button
              type="primary"
              @click="importRsaKeyFile(1)"
              icon="el-icon-upload"
              class="import-btn"
          >
            导入 RSA 私钥
          </el-button>
          <div class="file-path">
            <el-tag type="info" v-if="!prikPath">未选择私钥文件</el-tag>
            <el-tag type="warning" v-else>
              <el-icon>
                <document/>
              </el-icon>
              {{ prikPath }}
            </el-tag>
          </div>
        </div>

        <!-- 待加密文件导入区域 -->
        <div class="import-row">
          <el-button
              type="primary"
              @click="importOriFile"
              icon="el-icon-upload"
              class="import-btn"
          >
            导入待加密文件
          </el-button>
          <div class="file-path">
            <el-tag type="info" v-if="!oriFilePath">未选择待加密文件</el-tag>
            <el-tag v-else>
              <el-icon>
                <document/>
              </el-icon>
              {{ oriFilePath }}
            </el-tag>
          </div>
        </div>

        <!-- 操作按钮区域 -->
        <div class="action-buttons">
          <el-button
              type="success"
              @click="rsaCryptoStr"
              icon="el-icon-lock"
              :disabled="!pubkPath || !oriFilePath"
          >
            加密文件
          </el-button>
          <el-button
              type="warning"
              @click="rsaDeCryptoStr"
              icon="el-icon-unlock"
              :disabled="!prikPath || !oriFilePath"
          >
            解密文件
          </el-button>
        </div>
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import {Document} from '@element-plus/icons-vue'
import {GenerateKey, OpenTmpDir, RsaCryptoStr, RsaDeCryptoStr, SelectFile} from "../wailsjs/go/app/App.js";

export default {
  components: {
    Document
  },
  data() {
    return {
      pubkPath: "C:\\Users\\wangchaojun\\AppData\\Local\\wtools\\keys\\public.pem",
      prikPath: "C:\\Users\\wangchaojun\\AppData\\Local\\wtools\\keys\\private.pem",
      inputData: "",
      result: "",
      oriFilePath: "",
      keyBits: 2048,
      activeName: 'cryptoStr',
    }
  },
  methods: {
    async rsaCryptoStr() {
      try {
        this.result = "加密中..."
        this.result = await RsaCryptoStr(this.pubkPath, this.inputData)
        this.$message.success('加密成功')
      } catch (error) {
        this.$message.error('加密失败: ' + error)
        this.result = "加密失败: " + error
      }
    },
    async rsaDeCryptoStr() {
      try {
        this.result = "解密中..."
        const res = await RsaDeCryptoStr(this.prikPath, this.inputData)
        this.result = res
        this.$message.success('解密成功')
      } catch (error) {
        this.$message.error('解密失败: ' + error)
        this.result = "解密失败: " + error
      }
    },
    async importRsaKeyFile(prik) {
      try {
        const path = await SelectFile()
        if (prik === 0) {
          this.pubkPath = path
          this.$message.success('公钥文件导入成功')
        } else {
          this.prikPath = path
          this.$message.success('私钥文件导入成功')
        }
      } catch (error) {
        this.$message.error('文件选择失败: ' + error)
      }
    },

    async importOriFile() {
      try {
        const path = await SelectFile()
        this.oriFilePath = path
        this.$message.success('公钥文件导入成功')
      } catch (error) {
        this.$message.error('文件选择失败: ' + error)
      }
    },
    copyResult() {
      navigator.clipboard.writeText(this.result)
          .then(() => {
            this.$message.success('结果已复制到剪贴板')
          })
          .catch(err => {
            this.$message.error('复制失败: ' + err)
          })
    },
    genKey() {
      let that = this;
      if (that.keyBits === 0) {
        this.$message.error('keyBits不能为空');
        return;
      }
      GenerateKey(that.keyBits).then(res => {
        console.log(res);
        this.$message.info('密钥生成成功!');
        OpenTmpDir("/keys")
      })
    }
  }
}
</script>

<style scoped>
@import './css/crypto_encryption.css';
</style>
