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
  <div>
    <div style="border-color:black; margin-top: 10px;width: 100%;">
      <table>
        <tbody>
        <tr>
          <td>
            <div class="multi-input">
              <div v-for="(task, index) in inputData.task" :key="index" class="input-row">
                <input v-model="task.src" placeholder="源目录" />
                <input v-model="task.dst" placeholder="目标目录" />
                <button @click="swapValues(index)" class="btn-swap">↔️</button>
                <button @click="addRow(index)" class="btn-add">+</button>
                <button @click="removeRow(index)" class="btn-remove" :disabled="task.length <= 1">-</button>
              </div>
            </div>
          </td>
          <td style="width: 30px;"></td>
          <td>
            <el-button type="success" v-on:click="check">校验</el-button>
            <el-button type="success" v-on:click="fileSync">同步</el-button>
          </td>
          <td></td>
        </tr>
        </tbody>
      </table>
    </div>

    <div style="width: 1200px;margin-top: 10px" v-if="rStatus === 1">
      <el-alert title="同步成功" type="success" description="所有数据已同步" show-icon/>
    </div>

    <div v-for="(item, index) in respMssage">{{ index }}-{{ item }}</div>

    <div v-if="tableResult.length" style="border-color:black; margin-top: 10px;width: 1200px">
      <table style="border: black;border:1px;width: 1200px" id="dataTable">
        <tbody>
        <tr>
          <td class='td_result'>序号</td>
          <td class='td_result' style="width: 600px">文件路径</td>
          <td class='td_result'>大小</td>
          <td class='td_result'>修改时间</td>
          <td class='td_result'>目标大小</td>
          <td class='td_result'>目标修改时间</td>
          <td class='td_result'>同步状态</td>
        </tr>
        <tr v-for="(item, index) in tableResult" :key="index">
          <td class='td_result'>{{ index }}</td>
          <td class='td_result'>{{ item.Path }}</td>
          <td class='td_result' style="text-align: right">{{ item.WFile.Size }}</td>
          <td class='td_result'>{{ item.WFile.ModTime }}</td>
          <td class='td_result' style="text-align: right">{{ item.WFile.TargetSize }}</td>
          <td class='td_result'>{{ item.WFile.TargetModeTime }}</td>
          <td class='td_result'>
            <el-tag v-if="item.CyType === 0" key="Right" type="success" effect="dark">已同步</el-tag>
            <el-tag v-if="item.CyType === 2" key="Right" type="danger" effect="dark">未同步</el-tag>
            <el-tag v-if="item.CyType === 1" key="Right" type="warning" effect="dark">差异</el-tag>
            <el-tag v-if="item.CyType === 3" key="Right" type="warning" effect="dark">同步中</el-tag>
            <el-tag v-if="item.CyType === 4" key="Right" type="warning" effect="dark">同步异常</el-tag>
          </td>
        </tr>
        </tbody>
      </table>
    </div>

    <el-divider/>
    <div>
      <el-row>
        <el-col :span="3">
          <el-statistic :value="compareRes.SrcSizeStr">
            <template #title>
              <div style="display: inline-flex; align-items: center">
                源大小/数量
                <el-icon style="margin-left: 4px" :size="12">
                  <Male/>
                </el-icon>
              </div>
            </template>
            <template #suffix>/{{ compareRes.SrcCount }}</template>
          </el-statistic>
        </el-col>

        <el-col :span="3">
          <el-statistic :value="compareRes.DstSizeStr">
            <template #title>
              <div style="display: inline-flex; align-items: center">
                目标大小/数量
                <el-icon style="margin-left: 4px" :size="12">
                  <Male/>
                </el-icon>
              </div>
            </template>
            <template #suffix>/{{ compareRes.DstCount }}</template>
          </el-statistic>
        </el-col>

        <el-col :span="3">
          <el-statistic :value="compareRes.MigrateSizeStr">
            <template #title>
              <div style="display: inline-flex; align-items: center">
                迁移大小/数量
                <el-icon style="margin-left: 4px" :size="12">
                  <Male/>
                </el-icon>
              </div>
            </template>
            <template #suffix>/{{ compareRes.MigrateCount }}</template>
          </el-statistic>
        </el-col>


        <el-col :span="3">
          <el-statistic title="比较耗时" :value="compareRes.CompareCost"/>
        </el-col>
        <el-col :span="3">
          <el-statistic title="迁移耗时" :value="syncRes.MigrateCost">
            <template #suffix>
              <el-icon style="vertical-align: -0.125em">
                <ChatLineRound/>
              </el-icon>
            </template>
          </el-statistic>
        </el-col>
      </el-row>
    </div>
  </div>
</template>


<script setup>
import { ref } from "vue";

const rows = ref([
  { input1: "", input2: "" } // 初始 1 行
]);


</script>


<style scoped>
.multi-input {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.input-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

input {
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

button {
  padding: 8px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-swap {
  background: #2196f3;
  color: white;
}

.btn-add {
  background: #4caf50;
  color: white;
}

.btn-remove {
  background: #f44336;
  color: white;
}

button:disabled {
  background: #cccccc;
  cursor: not-allowed;
}
</style>

<script>
import {Compare, ReadAssetFile, ReadDemoFile} from "../wailsjs/go/main/App.js";

export default {
  data() {
    return {
      inputData: '',
      rStatus: 0,
      syncRes: {},
      websocketUrl: 'ws://127.0.0.1/w/file/fileSync', // 替换为你的WebSocket服务器URL
      tableResult: [],
      compareRes: {}
    }
  },

  computed: {},

  mounted() {
    this.getFileSyncConfig();
  },

  methods: {

    // 交换当前行的两个 input 的值
    swapValues(index){
      const temp = this.inputData.task[index].src;
      this.inputData.task[index].src = this.inputData.task[index].dst;
      this.inputData.task[index].dst=temp;
    },
    // 添加一行（在当前行后面插入）
    addRow(index){
      this.inputData.task.push({ src: "", dst: "" });
    },

// 删除当前行（至少保留 1 行）
    removeRow(index){
      if (this.inputData.task.length > 1) {
        this.inputData.task.splice(index, 1);
      }
    },
    getFileSyncConfig() {
      let that = this;
      ReadDemoFile("/fileSync.json").then((resData) => {
        console.log(resData);
        that.inputData = JSON.parse(resData)
      })
    },

    check() {
      let that = this;
      that.rStatus = 0;
      that.tableResult = [];
      Compare(that.inputData.task).then(resData => {
        that.compareRes = resData
        if (resData.Result != null) {
          that.tableResult = resData.Result
        } else {
          that.rStatus = 1;//1--检查返回，确认文件都已同步
        }
      })
    },

    fileSync() {
      let that = this;
      if (that.tableResult.length === 0) {
        that.$message.error('没有需要同步的数据');
        return
      }
      window.runtime.EventsEmit("file_sync", JSON.stringify(that.inputData.task));
      window.runtime.EventsOn("back_msg", (message) => {
        console.log("Backend says:", message);
        let wsRes = JSON.parse(JSON.parse(message).msg)
        for (let item of that.tableResult) {
          if (item.Path === wsRes.Path) {
            item.CyType = wsRes.Status
          }
        }
      });
    }
  }
}
</script>