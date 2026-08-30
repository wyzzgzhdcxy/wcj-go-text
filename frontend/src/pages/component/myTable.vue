<style scoped>
table {
  border-spacing: 0;
  border-collapse: collapse;
}

.td_style_1 {
  border: 1px solid #ACBED1;
  width: 150px;
  text-align: left;
}

.td_style_2 {
  border: 1px solid #ACBED1;
  width: 90px;
  text-align: right;
}
</style>
<template xmlns="http://www.w3.org/1999/html">
  <div v-if="data !== undefined && data.length !== undefined &&  data.length !== 0"
       style="border-color:black; margin-top: 10px">
    <table style="border: black;border:1px;">
      <tbody>
      <tr style="background-color: #c2e7b0" v-if="header != null && header.length !== undefined && header.length !== 0">
        <th v-for="(item, index) in header" class='td_result'>{{ item }}</th>
      </tr>
      <tr style="background-color: #c2e7b0" v-else>
        <th :class='classStype' v-if="showIndex">序号</th>
        <th :class='classStype' v-for="(it, index) in data[0]">{{ index }}</th>
      </tr>
      <tr v-for="(item, index) in displayedItems" :key="index">
        <td :class='classStype' v-if="showIndex">{{ index }}</td>
        <td v-for="(it, index) in item" :class='classStype'>{{ it }}</td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
// 定义组件的名称为myTable
  name: 'myTable',

  computed: {
    displayedItems() {
      let that = this;
      // 返回前10条数据
      if (that.count == null || that.count === 0) {
        that.count = 10;
      }
      return this.data.slice(0, that.count);
    }
  },
  // 定义组件内部使用的属性
  props: {
    // 自定义一个变量，用于接受父组件（首页或者其他页面）传入的参数值
    header: Array, //给组件传递的参数类型
    count: {
      type: Number,
      default: 10
    },
    classStype: {
      type: String,
      default: "td_style_1"
    },
    data: Array,
    showIndex: Boolean
  }
}
</script>