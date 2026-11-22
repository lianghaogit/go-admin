import * as ElementPlusIconsVue from '@element-plus/icons-vue';

// 全部已注册的 Element Plus 图标名集合（main.js 已全量注册，供菜单动态渲染）
const iconNames = new Set(Object.keys(ElementPlusIconsVue));

// 菜单图标兜底：无效/未注册的图标名降级为 Menu，避免渲染空白图标
export function safeIcon(name) {
  return name && iconNames.has(name) ? name : 'Menu';
}
