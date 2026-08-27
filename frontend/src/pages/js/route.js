import { createRouter, createWebHashHistory } from 'vue-router';
import { SetTitle } from '../../wailsjs/go/app/App.js';

const LAST_ROUTE_KEY = 'wcj_last_route';

function getLastRoute() {
    try {
        const v = localStorage.getItem(LAST_ROUTE_KEY);
        return v && v.startsWith('/') && v !== '/' ? v : null;
    } catch (e) {
        return null;
    }
}

const routeDefs = [
    {path: '/textCommonEncode', component: () => import('../textCommonEncode.vue'), meta: { title: '常用编码' }},
    {path: '/textHashEncode', component: () => import('../textHashEncode.vue'), meta: { title: '哈希编码' }},
    {path: '/textNormal', component: () => import('../textNormal.vue'), meta: { title: '普通文本' }},
    {path: '/textBothEnds', component: () => import('../textBothEnds.vue'), meta: { title: '文本两端' }},
    {path: '/textSort', component: () => import('../textSort.vue'), meta: { title: '文本排序' }},
    {path: '/textRemove', component: () => import('../textRemove.vue'), meta: { title: '去除文本' }},
    {path: '/textFormat', component: () => import('../textFormat.vue'), meta: { title: '格式转换' }},
    {path: '/textChar', component: () => import('../textChar.vue'), meta: { title: '字符转换' }},
    {path: '/textFile', component: () => import('../textFile.vue'), meta: { title: '文件处理' }},
    {path: '/appGenerator', component: () => import('../appGenerator.vue'), meta: { title: '应用生成' }},

    {path: '/ytdlp', component: () => import('../ytdlp.vue'), meta: { title: 'yt-dlp 下载' }},
    {path: '/downloadList', component: () => import('../downloadList.vue'), meta: { title: 'B站视频' }},
    {path: '/m3u8TaskDownload', component: () => import('../m3u8TaskDownload.vue'), meta: { title: 'M3U8 下载' }},
    {path: '/musicSearch', component: () => import('../musicSearch.vue'), meta: { title: '音乐搜索' }},
    {path: '/musicSource', component: () => import('../musicSource.vue'), meta: { title: '音乐解析' }},

    {path: '/imageTool', component: () => import('../imageTool.vue'), meta: { title: '图片工具' }},
    {path: '/emojiTool', component: () => import('../emojiTool.vue'), meta: { title: 'Emoji工具' }},
    {path: '/textToImage', component: () => import('../textToImage.vue'), meta: { title: '文字生图' }},
    {path: '/textToSpeech', component: () => import('../textToSpeech.vue'), meta: { title: '文字转语音' }},
    {path: '/videoTool', component: () => import('../videoTool.vue'), meta: { title: '视频处理' }},
    {path: '/systemSetting', component: () => import('../systemSetting.vue'), meta: { title: '系统设置' }},

    {path: '/timeConvert', component: () => import('../timeConvert.vue'), meta: { title: '时间转换' }},
    {path: '/tpl', component: () => import('../tpl.vue'), meta: { title: '模板工具' }},
    {path: '/categorizeFiles', component: () => import('../categorizeFiles.vue'), meta: { title: '文件归类' }},
    {path: '/JsonTableView', component: () => import('../jsonTableView.vue'), meta: { title: 'JSON表格' }},
    {path: '/JavaTools', component: () => import('../java_tools.vue'), meta: { title: 'Java代码工具' }},
    {path: '/cmdExecute', component: () => import('../cmdExecute.vue'), meta: { title: '命令行执行' }},
    {path: '/cmdManager', component: () => import('../cmdManager.vue'), meta: { title: '命令行管理' }},
    {path: '/SqlTools', component: () => import('../sqlTools.vue'), meta: { title: 'SQL工具' }},
    {path: '/TextBasicTools', component: () => import('../text_basic_tools.vue'), meta: { title: '文本差集' }},
    {path: '/cronExp', component: () => import('../cronExp.vue'), meta: { title: 'cron表达式' }},
    {path: '/fileSync', component: () => import('../fileSync.vue'), meta: { title: '文件同步' }},
    {path: '/salary', component: () => import('../salary.vue'), meta: { title: '工资计算器' }},
    {path: '/idcard', component: () => import('../idcard.vue'), meta: { title: '卡号工具' }},
    {path: '/rename', component: () => import('../rename.vue'), meta: { title: '文件重命名' }},
    {path: '/envCheck', component: () => import('../envCheck.vue'), meta: { title: '环境检测' }},
    {path: '/shutdown', component: () => import('../shutdown.vue'), meta: { title: '定时关机' }},
    {path: '/crypto_gen_key', component: () => import('../crypto_encryption.vue'), meta: { title: '加解密' }},
    {path: '/crypto_encryption', component: () => import('../crypto_encryption.vue'), meta: { title: '加解密' }},
    {path: '/expresson', component: () => import('../expresson.vue'), meta: { title: '表达式计算' }},
    {path: '/JsonTools', component: () => import('../json_tools.vue'), meta: { title: 'JSON工具' }},
    {path: '/FileBackup', component: () => import('../file_backup.vue'), meta: { title: '文件备份还原' }},
    {path: '/fileSplitMerge', component: () => import('../fileSplitMerge.vue'), meta: { title: '文件分割合并' }},
    {path: '/envVariables', component: () => import('../envVariables.vue'), meta: { title: '环境变量' }},
    {path: '/menuSettings', component: () => import('../menuSettings.vue'), meta: { title: '菜单设置' }},
];

const validPaths = new Set(routeDefs.map(r => r.path));
const savedRoute = getLastRoute();
const homePath = savedRoute && validPaths.has(savedRoute) ? savedRoute : '/textCommonEncode';

const routes = [
    {path: '/', redirect: homePath},
    ...routeDefs,
];

const router = createRouter({
    history: createWebHashHistory(),
    routes,
});

router.beforeEach((to, from, next) => {
    if (to.meta.title) {
        const title = '文本工具箱 - ' + to.meta.title;
        document.title = title;
        SetTitle(title);
    }
    next();
});

router.afterEach((to) => {
    if (to.path && to.path !== '/') {
        try {
            localStorage.setItem(LAST_ROUTE_KEY, to.path);
        } catch (e) { /* ignore */ }
    }
});

export default router;