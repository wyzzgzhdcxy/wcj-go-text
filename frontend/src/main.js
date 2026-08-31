import {createApp} from 'vue'
import App from './App.vue'
import router from './pages/js/route.js'
import {GetStartupTime, SetTitle} from './wailsjs/go/app/App.js'
import './style.css';

// Element Plus 按需引入：只引入实际用到的组件，避免全量打包
import {
  ElAlert, ElButton, ElCheckbox, ElCheckboxButton, ElCheckboxGroup,
  ElCol, ElCollapse, ElCollapseItem, ElDatePicker, ElDialog, ElDivider,
  ElDropdown, ElDropdownItem, ElDropdownMenu, ElForm, ElFormItem, ElIcon,
  ElInput, ElInputNumber, ElLoading, ElMenu, ElMenuItem, ElOption,
  ElOptionGroup, ElProgress, ElRadio, ElRadioButton, ElRadioGroup, ElRow,
  ElSelect, ElStatistic, ElSubMenu, ElSwitch, ElTable, ElTableColumn,
  ElTabPane, ElTabs, ElTag, ElTooltip,
} from 'element-plus'

// 组件样式（含 ElMessage/ElMessageBox/ElNotification 服务样式）
import 'element-plus/es/components/alert/style/css'
import 'element-plus/es/components/button/style/css'
import 'element-plus/es/components/checkbox/style/css'
import 'element-plus/es/components/checkbox-button/style/css'
import 'element-plus/es/components/checkbox-group/style/css'
import 'element-plus/es/components/col/style/css'
import 'element-plus/es/components/collapse/style/css'
import 'element-plus/es/components/collapse-item/style/css'
import 'element-plus/es/components/date-picker/style/css'
import 'element-plus/es/components/dialog/style/css'
import 'element-plus/es/components/divider/style/css'
import 'element-plus/es/components/dropdown/style/css'
import 'element-plus/es/components/dropdown-item/style/css'
import 'element-plus/es/components/dropdown-menu/style/css'
import 'element-plus/es/components/form/style/css'
import 'element-plus/es/components/form-item/style/css'
import 'element-plus/es/components/icon/style/css'
import 'element-plus/es/components/input/style/css'
import 'element-plus/es/components/input-number/style/css'
import 'element-plus/es/components/loading/style/css'
import 'element-plus/es/components/menu/style/css'
import 'element-plus/es/components/menu-item/style/css'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import 'element-plus/es/components/notification/style/css'
import 'element-plus/es/components/option/style/css'
import 'element-plus/es/components/option-group/style/css'
import 'element-plus/es/components/progress/style/css'
import 'element-plus/es/components/radio/style/css'
import 'element-plus/es/components/radio-button/style/css'
import 'element-plus/es/components/radio-group/style/css'
import 'element-plus/es/components/row/style/css'
import 'element-plus/es/components/select/style/css'
import 'element-plus/es/components/statistic/style/css'
import 'element-plus/es/components/sub-menu/style/css'
import 'element-plus/es/components/switch/style/css'
import 'element-plus/es/components/table/style/css'
import 'element-plus/es/components/table-column/style/css'
import 'element-plus/es/components/tab-pane/style/css'
import 'element-plus/es/components/tabs/style/css'
import 'element-plus/es/components/tag/style/css'
import 'element-plus/es/components/tooltip/style/css'

// 只引入实际用到的图标（全量约 293 个，这里 52 个）
import {
  Avatar, Calendar, ChatLineRound, CirclePlus, Coffee, Coin, Connection, CopyDocument, Crop,
  DataAnalysis, DataLine, Delete, Document, DocumentCopy, Download, Edit,
  Expand, Files, Film, Folder, FolderOpened, Fold, FullScreen, Headset, Histogram, Iphone, Key, Link,
  List, Loading, Lock, MagicStick, Memo, Menu, Microphone, Monitor,
  Operation, Picture, PictureFilled, PictureRounded, Plus, Postcard,
  PriceTag, Promotion, Refresh, ScaleToOriginal, Scissor, Search, Setting, SetUp,
  Sort, Switch, Tickets, Timer, Tools, VideoCamera, VideoPlay, ZoomIn,
} from '@element-plus/icons-vue'

const components = [
  ElAlert, ElButton, ElCheckbox, ElCheckboxButton, ElCheckboxGroup,
  ElCol, ElCollapse, ElCollapseItem, ElDatePicker, ElDialog, ElDivider,
  ElDropdown, ElDropdownItem, ElDropdownMenu, ElForm, ElFormItem, ElIcon,
  ElInput, ElInputNumber, ElMenu, ElMenuItem, ElOption, ElOptionGroup,
  ElProgress, ElRadio, ElRadioButton, ElRadioGroup, ElRow, ElSelect,
  ElStatistic, ElSubMenu, ElSwitch, ElTable, ElTableColumn, ElTabPane,
  ElTabs, ElTag, ElTooltip,
]

const icons = {
  Avatar, Calendar, ChatLineRound, CirclePlus, Coffee, Coin, Connection, CopyDocument, Crop,
  DataAnalysis, DataLine, Delete, Document, DocumentCopy, Download, Edit,
  Expand, Files, Film, Folder, FolderOpened, Fold, FullScreen, Headset, Histogram, Iphone, Key, Link,
  List, Loading, Lock, MagicStick, Memo, Menu, Microphone, Monitor,
  Operation, Picture, PictureFilled, PictureRounded, Plus, Postcard,
  PriceTag, Promotion, Refresh, ScaleToOriginal, Scissor, Search, Setting, SetUp,
  Sort, Switch, Tickets, Timer, Tools, VideoCamera, VideoPlay, ZoomIn,
}

const app = createApp(App)

components.forEach(c => app.use(c))
app.use(ElLoading)  // 注册 v-loading 指令

for (const [name, component] of Object.entries(icons)) {
  app.component(name, component)
}

app.use(router)
app.mount('#app')

// 显示启动耗时
GetStartupTime().then(ms => {
  SetTitle(document.title + ` - 启动耗时: ${ms}ms`)
})
