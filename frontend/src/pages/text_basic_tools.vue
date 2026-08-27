<style>
table {
  border-spacing: 0;
  border-collapse: collapse;
}
</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%;height: 100%">
    <hr style="margin-top: 0">
    <div id="appDiv" style="display: flex; height: 100%;">
      <div style="flex: 1;" class="blanchedalmond_div">
        <my-text-editor style="height: 45%;border: #ACBED1" v-model="text1"/>
        <hr style="color: #1a365d">
        <my-text-editor style="height: 55%" v-model="text2"/>
      </div>

      <div style="flex: 1;height: 100%">
        <div style="height: 35px;" class="blanchedalmond_div">
          <el-button type="success" v-on:click="textDifference">差集[text1-text2]</el-button>
          <el-button type="success" v-on:click="textIntersection">交集</el-button>
          <el-button type="success" v-on:click="textUnion">并集</el-button>
        </div>
        <my-text-editor v-model="result" language="java" style="height: 100%"/>
      </div>
    </div>
  </div>
</template>

<script>
import {TextDifference, TextIntersection, TextUnion} from "../wailsjs/go/app/App.js";
import MyTextEditor from "./component/myTextEditor.vue";

export default {
  components: {MyTextEditor: MyTextEditor},
  data() {
    return {
      text1: '{\n' +
          '  "positionId": "pos_img_left_bottom",\n' +
          '  "parentStyle": {\n' +
          '    "style": 2\n' +
          '  },\n' +
          '  "showType": 1,\n' +
          '  "tagList": []\n' +
          '}',
      text2: '{\n' +
          '  "positionId": "pos_img_left_bottom",\n' +
          '  "parentStyle": {\n' +
          '    "style": 2\n' +
          '  },\n' +
          '  "showType": 1,\n' +
          '  "tagList": []\n' +
          '}',
      result: ''
    }
  },
  computed: {},

  mounted() {
  },


  methods: {

    formatValue(value) {
      if (value === null) return 'null'
      if (typeof value === 'object') return JSON.stringify(value)
      return value
    },

    textUnion() {
      let that = this;
      TextUnion(that.text1, that.text2).then(text => {
        that.result = text
      })
    },
    textIntersection() {
      let that = this;
      TextIntersection(that.text1, that.text2).then(text => {
        that.result = text
      })
    },
    textDifference() {
      let that = this;
      console.log(that.text1, that.text2);
      TextDifference(that.text1, that.text2).then(text => {
        console.log(text);
        that.result = text
      })
    },
  }
}
</script>