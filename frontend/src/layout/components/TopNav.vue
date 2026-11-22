<template>
  <div class="top-nav">
    <el-menu
      :default-active="activeMenu"
      mode="horizontal"
      :router="!isMix"
      :ellipsis="false"
      class="top-menu"
      @select="onSelect"
    >
      <!-- 首页固定入口（Dashboard 静态路由，不在后端菜单中） -->
      <el-menu-item index="/index">
        <el-icon><HomeFilled /></el-icon>
        <template #title>{{ menuTitle('首页') }}</template>
      </el-menu-item>
      <top-nav-item v-for="menu in menus" :key="menu.menuId" :menu="menu" />
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useSettingsStore } from '@/store/settings';
import { HomeFilled } from '@element-plus/icons-vue';
import TopNavItem from './TopNavItem.vue';
import { useMenuTitle } from '@/i18n/translate';

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const settings = useSettingsStore();
const menuTitle = useMenuTitle();

const menus = computed(() => userStore.menus);
// 混合布局：顶栏为平铺一级目录（index=目录段），需手动跳转其第一个菜单叶子
const isMix = computed(() => settings.layoutMode === 'mix');
// 一级菜单高亮：按路由第一段匹配（/system/user → /system）；首页固定匹配 /index
const activeMenu = computed(() => {
  if (route.path === '/index' || route.path === '/') return '/index';
  const seg = route.path.split('/')[1];
  return seg ? '/' + seg : '/';
});

// 一级目录 -> 第一个菜单叶子路径（递归跳过目录/按钮）
function firstLeafPath(menu) {
  let node = menu;
  let path = '/' + node.path;
  while (node && node.children && node.children.length) {
    const leaf = node.children.find((ch) => ch.menuType !== 'F');
    if (!leaf) break;
    path += '/' + leaf.path;
    if (leaf.menuType === 'C') break;
    node = leaf;
  }
  return path;
}

// 菜单选择：mix 布局平铺项 index 为目录段（如 /system），映射为第一个菜单叶子跳转
function onSelect(index) {
  if (!isMix.value) return;
  if (index === '/index') {
    router.push('/index');
    return;
  }
  const menu = userStore.menus.find((m) => m.menuType === 'M' && '/' + m.path === index);
  router.push(menu ? firstLeafPath(menu) : index);
}
</script>

<style scoped>
.top-nav {
  background: var(--go-navbar-bg, #fff);
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.top-menu {
  border-bottom: none;
  height: 50px;
  background: transparent;
  /* 顶部导航菜单：文字/选中态随主题 */
  --el-menu-hover-bg-color: var(--go-topnav-hover, rgba(0, 0, 0, 0.06));
  --el-menu-hover-text-color: var(--go-navbar-text, #303133);
  --el-menu-active-color: var(--go-topnav-active, #409eff);
}

.top-menu :deep(.el-menu-item),
.top-menu :deep(.el-sub-menu__title) {
  height: 50px;
  line-height: 50px;
  color: var(--go-navbar-text, #303133);
}

.top-menu :deep(.el-menu-item.is-active) {
  color: var(--go-topnav-active, #409eff);
}
</style>
