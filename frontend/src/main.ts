import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';

import App from './App.vue';
import router from './router';
import hasPermi from './directive/permission';
import tableHeight from './directive/tableHeight';
import i18n from './i18n';
import './permission';
import './styles/index.css';

const app = createApp(App);

// 全局注册 Element Plus 图标（供菜单动态渲染）
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

app.use(createPinia());
app.use(router);
app.use(i18n);
app.use(ElementPlus, { locale: zhCn });
app.directive('hasPermi', hasPermi);
app.directive('tableHeight', tableHeight);
app.mount('#app');
