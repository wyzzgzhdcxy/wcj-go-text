import { createRouter, createWebHistory } from 'vue-router';
import { SetTitle } from '../../wailsjs/go/main/App.js';

const routes = [
    {path: '/', redirect: '/textCommonEncode'},
    {path: '/textCommonEncode', component: () => import('../textCommonEncode.vue'), meta: { title: '常用编码' }},
    {path: '/textHashEncode', component: () => import('../textHashEncode.vue'), meta: { title: '哈希编码' }},
    {path: '/textNormal', component: () => import('../textNormal.vue'), meta: { title: '普通文本' }},
    {path: '/textBothEnds', component: () => import('../textBothEnds.vue'), meta: { title: '文本两端' }},
    {path: '/textSort', component: () => import('../textSort.vue'), meta: { title: '文本排序' }},
    {path: '/textRemove', component: () => import('../textRemove.vue'), meta: { title: '去除文本' }},
    {path: '/textFormat', component: () => import('../textFormat.vue'), meta: { title: '格式转换' }},
    {path: '/textChar', component: () => import('../textChar.vue'), meta: { title: '字符转换' }},
    {path: '/textFile', component: () => import('../textFile.vue'), meta: { title: '文件处理' }},
];

const router = createRouter({
    history: createWebHistory(),
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

export default router;