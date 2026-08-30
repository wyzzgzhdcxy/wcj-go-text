<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

</style>

<template>
  <div style="width: 100%">
    <table>
      <tbody>
      <tr style="height: 50px">
        <td>
          <el-tag type="success" :disable-transitions="true">当前时间</el-tag>
        </td>
        <td>
          <el-input placeholder="当前时间" v-model="currentDateTime"/>
        </td>
        <td></td>
        <td>
          <el-tag type="success" :disable-transitions="true">当前时间戳</el-tag>
        </td>
        <td>
          <el-input placeholder="当前时间戳" v-model="currentTimeStamp"/>
        </td>
      </tr>
      <tr style="height: 50px">
        <td>
          <el-tag type="success" :disable-transitions="true">时间</el-tag>
        </td>
        <td style="horiz-align: right">
          <el-date-picker v-model="dateTime" type="datetime" placeholder="选择日期时间"></el-date-picker>
        </td>
        <td>
          <el-button type="success" v-on:click="toTimeStamp">==></el-button>
        </td>
        <td>
          <el-tag type="success" :disable-transitions="true">时间戳</el-tag>
        </td>
        <td>
          <el-input placeholder="时间戳" v-model="timeStamp"/>
        </td>
      </tr>
      <tr style="height: 50px">
        <td>
          <el-tag type="success" :disable-transitions="true">时间戳</el-tag>
        </td>
        <td>
          <el-input placeholder="时间戳" v-model="timeStamp"/>
        </td>
        <td>
          <el-button type="success" v-on:click="toDateTime">==></el-button>
        </td>
        <td>
          <el-tag type="success" :disable-transitions="true">时间</el-tag>
        </td>
        <td>
          <el-date-picker v-model="dateTime" type="datetime" placeholder="选择日期时间"></el-date-picker>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import dayjs from 'dayjs';


export default {
  data() {
    return {
      result: "",
      inputData: '11111111\n22222222\n33333333333',
      uppercase: true,
      small: false,
      currentDateTime: "",
      currentTimeStamp: "",
      dateTime: "",
      timeStamp: "",
      activeName: 'first',
      imgUrl: ''
    };
  },

  mounted() {
    this.setCurrentDateTime();
  },
  computed: {},
  methods: {
    toTimeStamp() {
      let that = this;
      let dayjsDate = dayjs(that.dateTime);
      that.timeStamp = dayjsDate.unix()
    },
    toDateTime() {
      let that = this;
      let dayjsDate = dayjs.unix(that.timeStamp);
      that.dateTime = dayjsDate.format('YYYY-MM-DD HH:mm:ss')
    },
    setCurrentDateTime() {
      let that = this;
      let dayjsDate = dayjs(new Date());
      that.currentDateTime = dayjsDate.format('YYYY-MM-DD HH:mm:ss');
      that.currentTimeStamp = dayjsDate.unix()
      setTimeout(this.setCurrentDateTime, 1000);
    },
  },
};
</script>