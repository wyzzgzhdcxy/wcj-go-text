<template>
  <div style="padding: 0;margin: 0;" ref="editorContainer" class="editor-container"></div>
</template>

<script setup>
import {computed, onMounted, ref, watch} from 'vue'
import {EditorView} from '@codemirror/view'
import {EditorState} from '@codemirror/state'
import {basicSetup} from 'codemirror'
import {oneDark} from '@codemirror/theme-one-dark'
import {indentUnit} from '@codemirror/language'
import {java} from '@codemirror/lang-java'

//传递组件参数
const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  language: {
    type: String,
    default: 'javascript'
  },
  height: {
    type: String,
    default: '100%'
  },
  width: {
    type: String,
    default: '100%'
  }
})

const emit = defineEmits(['update:modelValue'])

const editorContainer = ref(null)
const currentLanguage = ref(props.language)

let editorView = null


const languageExtensions = {
  java: java()
}


const extensions = computed(() => [
  basicSetup,
  languageExtensions[currentLanguage.value],
  oneDark,
  indentUnit.of('  '),
  EditorView.updateListener.of(update => {
    if (update.docChanged) {
      const doc = update.state.doc.toString()
      emit('update:modelValue', doc)
    }
  })
])
let isUpdatingFromParent = false // 防止父组件更新导致的循环

onMounted(() => {
  editorView = new EditorView({
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        basicSetup,
        java(),
        // 添加内容变化监听器
        EditorView.updateListener.of(update => {
          //这里只会进入用户直接编辑内容的事件，对页面按钮调用函数的值改变不做处理，在watch事件里处理
          if (update.docChanged && !isUpdatingFromParent) {
            // 获取编辑器最新内容
            let newValue = update.state.doc.toString()
            // 触发更新事件
            emit('update:modelValue', newValue)
          }
        }),
        EditorView.theme({
          "&": {height: props.height, width: props.width},
          ".cm-scroller": {overflow: "auto"}
        })
      ]
    }),
    parent: editorContainer.value
  })
})

watch(extensions, (newExtensions) => {
  if (editorView) {
    editorView.dispatch({
      effects: EditorState.reconfigure(newExtensions)
    })
  }
})


//按女设置值之类的操作会走这里
watch(() => props.modelValue, (newValue) => {
  if (editorView && newValue !== editorView.state.doc.toString()) {
    isUpdatingFromParent = true
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: newValue
      }
    })
    isUpdatingFromParent = false
  }
})

</script>

<style>
.editor-container {
  overflow: hidden;
}
</style>