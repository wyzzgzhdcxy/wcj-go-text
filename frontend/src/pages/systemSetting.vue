<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.td_result {
  border: 1px solid #ACBED1;
}
</style>

<template xmlns="http://www.w3.org/1999/html">
  <!-- 输入对话框 -->
  <el-dialog v-model="dialogVisible"
             title="添加配置参数" width="30%" :close-on-click-modal="false"
             :close-on-press-escape="false" :show-close="false">

    <!-- 表单内容 -->
    <el-form :model="formData" ref="formRef" label-width="100px">
      <el-input placeholder="配置key" v-model="requestKey">
        <template #prepend>配置key</template>
      </el-input>
      <el-input style="margin-top: 10px" placeholder="配置val" v-model="requestValue">
        <template #prepend>配置val</template>
      </el-input>
    </el-form>

    <!-- 对话框底部按钮 -->
    <template #footer>
      <el-button @click="dialogVisible=false">取消</el-button>
      <el-button type="primary" @click="addConfigValue">提交</el-button>
    </template>
  </el-dialog>


  <div class="left">
    <div id="appDiv">
      <table>
        <tbody>
        <tr>
          <td>
            <el-input placeholder="配置KEY" v-model="query.key">
              <template #prepend>配置KEY</template>
            </el-input>
          </td>
          <td>
          </td>
          <td>
            <el-button type="primary" v-on:click="list">🔍 查询</el-button>
          </td>
          <td>
            <div style="color: red">{{ tipText }}</div>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
    <div style="height: 1px;background-color: black;margin-top: 10px;margin-bottom: 10px"></div>
    <div style="width: 100%">
      <table style="width: 100%">
        <tbody>
        <tr style="background-color: cadetblue">
          <td style="width: 40px" class='td_result'>序号</td>
          <td style="width: 250px" class='td_result'>key</td>
          <td class='td_result'>value</td>
          <td class='td_result'>操作</td>
        </tr>
        <tr v-for="(value, key,index) in data" :key="key">
          <td class='td_result'>{{ index }}</td>
          <td class='td_result'>{{ key }}</td>
          <td class='td_result'>{{ truncateText(value, 60) }}</td>
          <td class='td_result'>
            <el-button type="danger" v-on:click="deleteConfig(key)" :icon="Delete"></el-button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import {AddConfigValue, DeleteConfig, ListConfig} from "../wailsjs/go/main/App.js";
import {Delete} from "@element-plus/icons-vue";

export default {
  data() {
    return {
      formData: {},
      dialogVisible: false,
      query: {
        key: "system."
      },
      data: {},
      requestKey: "",
      requestValue: "",
      copy: {},
      setting: {
        "videoType": "all"
      },
      videoTypes: ["all", "like", "doubleFlying"],
      activeName: 'first',
      tipText: ""
    }
  },
  mounted() {
    let that = this;
    that.list();
  },
  computed: {
    Delete() {
      return Delete
    }

  },
  methods: {

    addConfigValue() {
      console.log("addConfigValue" + this.requestKey + this.requestValue);
      AddConfigValue("system.setting." + this.requestKey, this.requestValue).then(response => {
        this.dialogVisible = false;
        this.list();
      })

    },
    deleteConfig(key) {
      let that = this;
      DeleteConfig(key).then(res => {
        that.$message.info("删除成功:" + res)
        that.list()
      });
    },

    truncateText(text, maxLength) {
      if (!text) return "";
      return text.length > maxLength ? text.slice(0, maxLength) + "..." : text;
    },
    list() {
      let that = this;
      ListConfig(that.query.key).then((res) => {
        that.data = res;
        console.log(that.data);
      })
    }

  }
}
</script>