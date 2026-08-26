import {createApp} from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './pages/js/route.js'
import {GetStartupTime, SetTitle} from './wailsjs/go/main/App.js'
import './style.css';

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}

app.use(ElementPlus)
app.use(router)
app.mount('#app')

// 显示启动耗时
GetStartupTime().then(ms => {
    SetTitle(document.title + ` - 启动耗时: ${ms}ms`)
})