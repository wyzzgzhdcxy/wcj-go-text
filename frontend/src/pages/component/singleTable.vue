<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}
</style>
<template xmlns="http://www.w3.org/1999/html">
  <div v-if="data !== undefined && data.length !== undefined &&  data.length !== 0"
       style="border-color:black; margin-top: 10px">
    <table style="border: black;border:1px" id="resultTable">
      <tbody>
      <tr v-for="(item, index) in data" :key="index">
        <td class='td_result'>{{ index }}</td>
        <td class='td_result'>{{ item }}</td>
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
    count: Number,
    data: Array,
    showIndex: Boolean
  }
}
</script>