import { defineStore } from 'pinia';
import { applyThemeColors } from '@/utils/theme';
import i18n, { saveLanguage, type Lang } from '@/i18n';

const STORAGE_KEY = 'go-admin-settings';

// 默认布局配置（对标 RuoYi 布局配置）
const defaultSettings = {
  themeColor: '#409EFF', // 主题色
  darkMode: false, // 暗黑模式
  sidebarTheme: 'dark', // 侧边栏主题 dark / light
  layoutMode: 'side', // 布局模式 side（左侧边栏）/ rightSide（右侧边栏）/ top（顶部导航）/ bottom（底部导航）/ mix（混合）/ dial（转盘）/ desktop（桌面）
  showLogo: true, // 显示 Logo
  breadcrumb: true, // 显示面包屑
  showTagsView: true, // 多页签
  showFooter: false, // 显示页脚
  menuIcon: true, // 菜单图标
  size: 'default', // 组件尺寸 large / default / small
  sidebarCollapsed: false, // 侧边栏折叠（localStorage 持久化）
  drawerVisible: false, // 布局配置抽屉开关
  desktopBg: '', // 桌面布局背景：''=默认渐变；'url(...)'=自定义图片；其他=CSS background 值
  language: 'zh' as Lang // 界面语言 zh / en
};

// 读取持久化的布局配置
function loadSaved() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    // 排除抽屉开关：抽屉状态不持久化，避免刷新后自动打开
    const rest = { ...saved };
    delete rest.drawerVisible;
    return { ...defaultSettings, ...rest };
  } catch (e) {
    return { ...defaultSettings };
  }
}

export const useSettingsStore = defineStore('settings', {
  state: () => ({ ...loadSaved() }),
  actions: {
    // 应用主题到 DOM（主题色 CSS 变量 + 暗黑/侧边栏主题 class）
    applyTheme() {
      applyThemeColors(this.themeColor);
      const root = document.documentElement;
      root.classList.toggle('dark', this.darkMode);
      root.classList.toggle('sidebar-light', this.sidebarTheme === 'light');
    },
    // 初始化：读持久化配置并应用
    init() {
      Object.assign(this, loadSaved());
      this.applyTheme();
      // 同步 vue-i18n 语言
      i18n.global.locale.value = this.language;
    },
    save() {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          themeColor: this.themeColor,
          darkMode: this.darkMode,
          sidebarTheme: this.sidebarTheme,
          layoutMode: this.layoutMode,
          showLogo: this.showLogo,
          breadcrumb: this.breadcrumb,
          showTagsView: this.showTagsView,
          showFooter: this.showFooter,
          menuIcon: this.menuIcon,
          size: this.size,
          sidebarCollapsed: this.sidebarCollapsed,
          desktopBg: this.desktopBg,
          language: this.language
        })
      );
    },
    // 切换界面语言（同步 vue-i18n locale + 持久化）
    setLanguage(lang: Lang) {
      this.language = lang;
      i18n.global.locale.value = lang;
      saveLanguage(lang);
      this.save();
    },
    // 恢复默认
    reset() {
      Object.assign(this, defaultSettings);
      this.applyTheme();
      i18n.global.locale.value = this.language;
      this.save();
    }
  }
});
