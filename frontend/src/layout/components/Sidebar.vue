<template>
  <div class="sidebar" :class="{ collapsed: settings.sidebarCollapsed }">
    <div v-if="showLogo" class="logo">
      <span v-if="!settings.sidebarCollapsed" class="logo-text">Go Admin</span>
      <span v-else class="logo-mini">G</span>
      <!-- 折叠/展开按钮（位于 Logo 右侧） -->
      <div
        class="logo-fold"
        :title="settings.sidebarCollapsed ? '展开菜单' : '折叠菜单'"
        @click="toggleCollapse"
      >
        <el-icon :size="16"><Expand v-if="settings.sidebarCollapsed" /><Fold v-else /></el-icon>
      </div>
    </div>
    <el-scrollbar class="menu-scroll">
      <el-menu
        :default-active="activeMenu"
        :default-openeds="defaultOpeneds"
        :collapse="settings.sidebarCollapsed"
        :collapse-transition="false"
        router
        class="side-menu"
      >
        <!-- 首页固定入口（Dashboard 静态路由，不在后端菜单中） -->
        <el-menu-item index="/index">
          <el-icon v-if="settings.menuIcon"><HomeFilled /></el-icon>
          <template #title>{{ menuTitle('首页') }}</template>
        </el-menu-item>
        <sidebar-item v-for="menu in menus" :key="menu.menuId" :menu="menu" :base="base" />
      </el-menu>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useSettingsStore } from '@/store/settings';
import { HomeFilled, Fold, Expand } from '@element-plus/icons-vue';
import SidebarItem from './SidebarItem.vue';
import { useMenuTitle } from '@/i18n/translate';

const menuTitle = useMenuTitle();

const props = defineProps({
  // 菜单数据覆盖：mix 布局传当前一级目录的子菜单；默认全量菜单
  menusOverride: { type: Array, default: null },
  // 路径前缀：mix 布局传当前一级目录段（如 /system），子菜单 path 为相对段需补齐；默认空
  base: { type: String, default: '' },
  // 是否显示顶部 Logo（desktop 布局 Logo 在顶栏，侧边栏不显示）
  showLogo: { type: Boolean, default: true }
});

const route = useRoute();
const userStore = useUserStore();
const settings = useSettingsStore();

const menus = computed(() => props.menusOverride || userStore.menus);
const activeMenu = computed(() => route.path);
const defaultOpeneds = computed(() =>
  menus.value.filter((m) => m.menuType === 'M').map((m) => '/' + m.path)
);

// 折叠/展开侧边栏（与 Navbar 原折叠按钮行为一致：持久化）
function toggleCollapse() {
  settings.sidebarCollapsed = !settings.sidebarCollapsed;
  settings.save();
}
</script>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background-color: var(--go-sidebar-bg);
}

.logo {
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8px 0 18px;
  color: var(--go-sidebar-logo-text, #fff);
  font-size: 18px;
  font-weight: 600;
  background-color: var(--go-sidebar-logo-bg);
  background-image: linear-gradient(180deg, rgba(255, 255, 255, 0.05), rgba(255, 255, 255, 0));
  white-space: nowrap;
  overflow: hidden;
}

.logo-text {
  overflow: hidden;
  text-overflow: ellipsis;
}

.logo-mini {
  font-size: 16px;
}

/* Logo 右侧折叠按钮 */
.logo-fold {
  width: 26px;
  height: 26px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  color: var(--go-sidebar-logo-text, #fff);
  opacity: 0.85;
  transition: background 0.2s, opacity 0.2s;
}

.logo-fold:hover {
  background: rgba(255, 255, 255, 0.14);
  opacity: 1;
}

/* 折叠态：Logo 区仅 G + 按钮（60px 内紧凑排布） */
.sidebar.collapsed .logo {
  padding: 0 2px 0 6px;
  justify-content: flex-start;
  gap: 0;
}

.sidebar.collapsed .logo-fold {
  width: 30px;
  height: 30px;
}

.menu-scroll {
  flex: 1;
  overflow: hidden;
}

.side-menu {
  border-right: none;
  /* 覆盖 Element Plus el-menu 颜色为侧边栏主题变量 */
  --el-menu-bg-color: var(--go-sidebar-bg);
  --el-menu-text-color: var(--go-sidebar-text);
  --el-menu-active-color: var(--go-sidebar-active-text);
  --el-menu-hover-bg-color: var(--go-sidebar-hover);
  --el-menu-hover-text-color: var(--go-sidebar-text);
}

/* 折叠态菜单项文字隐藏（el-menu collapse 会自动处理宽度） */
.side-menu:not(.el-menu--collapse) {
  width: 100%;
}
</style>
