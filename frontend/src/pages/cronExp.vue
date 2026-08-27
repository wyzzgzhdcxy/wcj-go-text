<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

</style>

<template>
  <div style="width: 100%">
    <div id="appDiv">
      <table>
        <tbody>
        <tr style="height: 50px">
          <td>
            <el-tag type="success" :disable-transitions="true">cron表达式</el-tag>
          </td>
          <td>
            <el-input placeholder="cron表达式" v-model="cronExpress"/>
          </td>
          <td></td>
          <td>
            <el-tag type="success" :disable-transitions="true">数量</el-tag>
          </td>
          <td>
            <el-input placeholder="数量" v-model.number="count"/>
          </td>
          <td>
            <el-tag type="success" :disable-transitions="true">开始时间</el-tag>
          </td>
          <td style="horiz-align: right">
            <el-date-picker v-model="startDate" type="datetime" placeholder="开始时间"></el-date-picker>
          </td>
          <td>
            <el-button type="success" v-on:click="getNextExecTime">下次执行时间</el-button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>

    <div v-if="data !== undefined && data.length !== undefined &&  data.length !== 0"
         style="border-color:black; margin-top: 10px">
      <table style="border: black;border:1px" id="resultTable">
        <tbody>
        <tr style="background-color: #c2e7b0;border: 1px">
          <td class='td_result'>序号</td>
          <td class='td_result'>时间</td>
          <td class='td_result'>星期</td>
        </tr>
        <tr v-for="(item, index) in data" :key="index">
          <td class='td_result'>{{ index }}</td>
          <td class='td_result'>{{ item.TimeStr }}</td>
          <td class='td_result'>{{ item.Weekday }}</td>
        </tr>
        </tbody>
      </table>
    </div>

    <div style="font-size: 12px;color: rebeccapurple;margin-top: 20px">
      常用cron表达式例子 <br>
      （1）0/2 * * * * ? 表示每2秒 执行任务 <br>
      （1）0 0/2 * * * ? 表示每2分钟 执行任务 <br>
      （1）0 0 2 1 * ? 表示在每月的1日的凌晨2点调整任务 <br>
      （2）0 15 10 ? * MON-FRI 表示周一到周五每天上午10:15执行作业 <br>
      （3）0 15 10 ? 6L 2002-2006 表示2002-2006年的每个月的最后一个星期五上午10:15执行作 <br>
      （4）0 0 10,14,16 * * ? 每天上午10点，下午2点，4点 <br>
      （5）0 0/30 9-17 * * ? 朝九晚五工作时间内每半小时 <br>
      （6）0 0 12 ? * WED 表示每个星期三中午12点 <br>
      （7）0 0 12 * * ? 每天中午12点触发 <br>
      （8）0 15 10 ? * * 每天上午10:15触发 <br>
      （9）0 15 10 * * ? 每天上午10:15触发 <br>
      （10）0 15 10 * * ? 每天上午10:15触发 <br>
      （11）0 15 10 * * ? 2005 2005年的每天上午10:15触发 <br>
      （12）0 * 14 * * ? 在每天下午2点到下午2:59期间的每1分钟触发 <br>
      （13）0 0/5 14 * * ? 在每天下午2点到下午2:55期间的每5分钟触发 <br>
      （14）0 0/5 14,18 * * ? 在每天下午2点到2:55期间和下午6点到6:55期间的每5分钟触发 <br>
      （15）0 0-5 14 * * ? 在每天下午2点到下午2:05期间的每1分钟触发 <br>
      （16）0 10,44 14 ? 3 WED 每年三月的星期三的下午2:10和2:44触发 <br>
      （17）0 15 10 ? * MON-FRI 周一至周五的上午10:15触发 <br>
      （18）0 15 10 15 * ? 每月15日上午10:15触发 <br>
      （19）0 15 10 L * ? 每月最后一日的上午10:15触发 <br>
      （20）0 15 10 ? * 6L 每月的最后一个星期五上午10:15触发 <br>
      （21）0 15 10 ? * 6L 2002-2005 2002年至2005年的每月的最后一个星期五上午10:15触发 <br>
      （22）0 15 10 ? * 6#3 每月的第三个星期五上午10:15触发 <br>
    </div>
  </div>
</template>

<script>
import SingleTable from "./component/singleTable.vue";
import dayjs from "dayjs";
import {NextTimeList} from "../wailsjs/go/app/App.js";


export default {
  components: {SingleTable},
  data() {
    return {
      cronExpress: "0 0 0 * * ?",
      startDate: "",
      count: 10,
      desc: "",
      data: []
    };
  },

  mounted() {
  },
  computed: {},
  methods: {


    getNextExecTime() {
      let that = this;
      let dateStr = ""
      if (that.startDate === '') {
      } else {
        let dayjsDate = dayjs(that.startDate);
        dateStr = dayjsDate.format('YYYY-MM-DD HH:mm:ss')
      }

      NextTimeList(that.cronExpress,dateStr,that.count).then(result => {
        that.data = result;
      })
    }
  },
};
</script>