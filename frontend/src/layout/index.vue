<template>
  <div
    class="app-wrapper"
    :class="{
      'layout-top': settings.layoutMode === 'top',
      'layout-bottom': settings.layoutMode === 'bottom',
      'layout-rightSide': settings.layoutMode === 'rightSide',
      'layout-mix': settings.layoutMode === 'mix',
      'layout-desktop': settings.layoutMode === 'desktop'
    }"
  >
    <!-- 侧边栏布局：侧边栏完整菜单 + 顶栏 -->
    <template v-if="settings.layoutMode === 'side'">
      <Sidebar
        class="sidebar-container"
        :style="{
          flex: '0 0 ' + (settings.sidebarCollapsed ? '60px' : '200px'),
          width: settings.sidebarCollapsed ? '60px' : '200px'
        }"
      />
      <div class="main-container">
        <Navbar />
        <AppMain />
        <div v-if="settings.showFooter" class="app-footer">
          Copyright © 2026 Go Admin · Gin + Vue3 + Element Plus + SQLite
        </div>
      </div>
    </template>

    <!-- 右侧边栏布局：内容区在左、侧边栏完整菜单在右（与 side 镜像） -->
    <template v-else-if="settings.layoutMode === 'rightSide'">
      <div class="main-container">
        <Navbar />
        <AppMain />
        <div v-if="settings.showFooter" class="app-footer">
          Copyright © 2026 Go Admin · Gin + Vue3 + Element Plus + SQLite
        </div>
      </div>
      <Sidebar
        class="sidebar-container"
        :style="{
          flex: '0 0 ' + (settings.sidebarCollapsed ? '60px' : '200px'),
          width: settings.sidebarCollapsed ? '60px' : '200px'
        }"
      />
    </template>

    <!-- 顶部导航布局：Logo + 横向一级菜单 + 用户区 单行，无侧边栏 -->
    <template v-else-if="settings.layoutMode === 'top'">
      <div class="top-header">
        <div v-if="settings.showLogo" class="top-logo">Go Admin</div>
        <TopNav class="top-nav-slot" />
        <div class="top-right">
          <div class="nav-icon" title="菜单搜索（Ctrl+K）" @click="openSearch">
            <el-icon :size="18"><Search /></el-icon>
          </div>
          <div
            class="nav-icon"
            :title="isFullscreen ? '退出全屏' : '全屏'"
            @click="toggleFullscreen"
          >
            <el-icon :size="18"><component :is="isFullscreen ? 'Aim' : 'FullScreen'" /></el-icon>
          </div>
          <div class="nav-icon" title="布局配置" @click="settings.drawerVisible = true">
            <el-icon :size="18"><Setting /></el-icon>
          </div>
          <UserDropdown />
        </div>
      </div>
      <div class="top-main">
        <div v-if="settings.breadcrumb" class="top-breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">
              <el-icon :size="13" style="vertical-align: -2px; margin-right: 2px"
                ><HomeFilled
              /></el-icon>
              首页
            </el-breadcrumb-item>
            <template v-for="(item, idx) in breadcrumbs" :key="item.path">
              <el-breadcrumb-item
                v-if="idx < breadcrumbs.length - 1 && item.redirect"
                :to="{ path: item.path }"
              >
                <el-icon
                  v-if="item.meta.icon"
                  :size="13"
                  style="vertical-align: -2px; margin-right: 2px"
                >
                  <component :is="item.meta.icon" />
                </el-icon>
                {{ menuTitle(item.meta.title) }}
              </el-breadcrumb-item>
              <el-breadcrumb-item v-else>{{ menuTitle(item.meta.title) }}</el-breadcrumb-item>
            </template>
          </el-breadcrumb>
        </div>
        <AppMain />
        <div v-if="settings.showFooter" class="app-footer">Copyright © 2026 Go Admin</div>
      </div>
    </template>

    <!-- 底部导航布局：内容区在上、Logo + 横向一级菜单 + 用户区单行在底部（与 top 镜像） -->
    <template v-else-if="settings.layoutMode === 'bottom'">
      <div class="bottom-main">
        <div v-if="settings.breadcrumb" class="top-breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">
              <el-icon :size="13" style="vertical-align: -2px; margin-right: 2px"
                ><HomeFilled
              /></el-icon>
              首页
            </el-breadcrumb-item>
            <template v-for="(item, idx) in breadcrumbs" :key="item.path">
              <el-breadcrumb-item
                v-if="idx < breadcrumbs.length - 1 && item.redirect"
                :to="{ path: item.path }"
              >
                <el-icon
                  v-if="item.meta.icon"
                  :size="13"
                  style="vertical-align: -2px; margin-right: 2px"
                >
                  <component :is="item.meta.icon" />
                </el-icon>
                {{ menuTitle(item.meta.title) }}
              </el-breadcrumb-item>
              <el-breadcrumb-item v-else>{{ menuTitle(item.meta.title) }}</el-breadcrumb-item>
            </template>
          </el-breadcrumb>
        </div>
        <AppMain />
        <div v-if="settings.showFooter" class="app-footer">Copyright © 2026 Go Admin</div>
      </div>
      <div class="bottom-header">
        <div v-if="settings.showLogo" class="top-logo">Go Admin</div>
        <TopNav class="top-nav-slot" />
        <div class="top-right">
          <div class="nav-icon" title="菜单搜索（Ctrl+K）" @click="openSearch">
            <el-icon :size="18"><Search /></el-icon>
          </div>
          <div
            class="nav-icon"
            :title="isFullscreen ? '退出全屏' : '全屏'"
            @click="toggleFullscreen"
          >
            <el-icon :size="18"><component :is="isFullscreen ? 'Aim' : 'FullScreen'" /></el-icon>
          </div>
          <div class="nav-icon" title="布局配置" @click="settings.drawerVisible = true">
            <el-icon :size="18"><Setting /></el-icon>
          </div>
          <UserDropdown />
        </div>
      </div>
    </template>

    <!-- 混合布局：顶栏 Logo + 一级目录菜单；侧边栏显示当前一级目录的子菜单 -->
    <template v-else-if="settings.layoutMode === 'mix'">
      <div class="top-header">
        <div v-if="settings.showLogo" class="top-logo">Go Admin</div>
        <TopNav class="top-nav-slot" />
        <div class="top-right">
          <div
            class="nav-icon"
            :title="settings.sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'"
            @click="toggleSidebar"
          >
            <el-icon :size="18"
              ><component :is="settings.sidebarCollapsed ? 'Expand' : 'Fold'"
            /></el-icon>
          </div>
          <div class="nav-icon" title="菜单搜索（Ctrl+K）" @click="openSearch">
            <el-icon :size="18"><Search /></el-icon>
          </div>
          <div
            class="nav-icon"
            :title="isFullscreen ? '退出全屏' : '全屏'"
            @click="toggleFullscreen"
          >
            <el-icon :size="18"><component :is="isFullscreen ? 'Aim' : 'FullScreen'" /></el-icon>
          </div>
          <div class="nav-icon" title="布局配置" @click="settings.drawerVisible = true">
            <el-icon :size="18"><Setting /></el-icon>
          </div>
          <UserDropdown />
        </div>
      </div>
      <div class="mix-body">
        <!-- 宽窗：flow 侧边栏；窄窗隐藏（由 el-drawer 抽屉承担导航） -->
        <Sidebar
          v-if="!isNarrow"
          class="sidebar-container"
          :menus-override="mixMenus"
          :base="activeTopPath"
          :style="{
            flex: '0 0 ' + (settings.sidebarCollapsed ? '64px' : '210px'),
            width: settings.sidebarCollapsed ? '60px' : '200px'
          }"
        />
        <!-- 窄窗抽屉侧边栏 -->
        <el-drawer
          v-if="isNarrow"
          v-model="mobileSidebarOpen"
          direction="ltr"
          size="210px"
          :with-header="false"
          append-to-body
          class="go-sidebar-drawer"
        >
          <Sidebar :show-logo="false" :menus-override="mixMenus" :base="activeTopPath" />
        </el-drawer>
        <div class="main-container">
          <div v-if="settings.breadcrumb" class="top-breadcrumb">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/' }">
                <el-icon :size="13" style="vertical-align: -2px; margin-right: 2px"
                  ><HomeFilled
                /></el-icon>
                {{ menuTitle('首页') }}
              </el-breadcrumb-item>
              <template v-for="(item, idx) in breadcrumbs" :key="item.path">
                <el-breadcrumb-item
                  v-if="idx < breadcrumbs.length - 1 && item.redirect"
                  :to="{ path: item.path }"
                >
                  <el-icon
                    v-if="item.meta.icon"
                    :size="13"
                    style="vertical-align: -2px; margin-right: 2px"
                  >
                    <component :is="item.meta.icon" />
                  </el-icon>
                  {{ menuTitle(item.meta.title) }}
                </el-breadcrumb-item>
                <el-breadcrumb-item v-else>{{ menuTitle(item.meta.title) }}</el-breadcrumb-item>
              </template>
            </el-breadcrumb>
          </div>
          <AppMain />
          <div v-if="settings.showFooter" class="app-footer">Copyright © 2026 Go Admin</div>
        </div>
      </div>
    </template>

    <!-- 转盘布局：全屏内容 + 右上角 1/4 圆转盘菜单（最小态小圆钮，hover 放大） -->
    <template v-else-if="settings.layoutMode === 'dial'">
      <AppMain />
      <DialMenu />
    </template>

    <!-- Dock 坞布局：全屏内容 + 底部中央悬浮应用坞（macOS 风格，hover 放大 + 提示） -->
    <template v-else-if="settings.layoutMode === 'dock'">
      <AppMain />
      <DockBar />
    </template>

    <!-- 垂直页签布局：左侧垂直页签列（浏览器侧栏标签风格）+ 内容区 -->
    <template v-else-if="settings.layoutMode === 'vtab'">
      <div class="vtab-body">
        <VTabBar />
        <div class="main-container vtab-main">
          <AppMain />
          <div v-if="settings.showFooter" class="app-footer">Copyright © 2026 Go Admin</div>
        </div>
      </div>
    </template>

    <!-- 桌面布局：全屏 Windows 桌面（开始菜单 + 多窗口 + 右下角系统托盘），无传统顶栏/侧边栏 -->
    <template v-else>
      <DesktopArea />
    </template>

    <!-- 布局配置抽屉（挂 layout 顶层，不受布局模式切换影响） -->
    <SettingsDrawer />
    <!-- 全局菜单搜索（Ctrl+K） -->
    <SearchMenu />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Setting, Search, HomeFilled } from '@element-plus/icons-vue';
import Sidebar from './components/Sidebar.vue';
import Navbar from './components/Navbar.vue';
import TopNav from './components/TopNav.vue';
import AppMain from './components/AppMain.vue';
import DesktopArea from './components/DesktopArea.vue';
import DockBar from './components/DockBar.vue';
import VTabBar from './components/VTabBar.vue';
import DialMenu from './components/DialMenu.vue';
import UserDropdown from './components/UserDropdown.vue';
import SettingsDrawer from './components/SettingsDrawer.vue';
import SearchMenu from './components/SearchMenu.vue';
import { useSettingsStore } from '@/store/settings';
import { useUserStore } from '@/store/user';
import { useMenuTitle } from '@/i18n/translate';

const settings = useSettingsStore();
const userStore = useUserStore();
const route = useRoute();
const isFullscreen = ref(false);
// 窄窗口（<768px）：side 布局折叠侧边栏；mix/desktop 布局侧边栏转抽屉浮层
const isNarrow = ref(false);
const mobileSidebarOpen = ref(false);
let narrowApplied = false;
let narrowPrevCollapsed = false;

// 顶部布局面包屑（与侧边栏 Navbar 一致：过滤首页避免重复）
const menuTitle = useMenuTitle();
const breadcrumbs = computed(() =>
  route.matched.filter((item) => item.meta && item.meta.title && item.meta.title !== '首页')
);

// 混合布局：当前一级目录（按路由第一段匹配，如 /system/user → 系统管理）
const activeTopPath = computed(() => {
  const seg = route.path.split('/')[1];
  return seg ? '/' + seg : '';
});

// 混合布局：当前一级目录下的二级菜单（无匹配时兜底第一个目录，保证侧边栏有内容）
const mixMenus = computed(() => {
  const top = userStore.menus.find(
    (m) => m.menuType === 'M' && '/' + m.path === activeTopPath.value
  );
  const source = top || userStore.menus.find((m) => m.menuType === 'M');
  if (!source) return userStore.menus;
  return (source.children || []).filter((c) => c.menuType !== 'F');
});

// 打开全局菜单搜索（Ctrl+K）
function openSearch() {
  // 通知全局 SearchMenu（挂载在 layout 顶层）打开
  window.dispatchEvent(new Event('go-open-search'));
}

// 切换全屏
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {});
  } else {
    document.exitFullscreen().catch(() => {});
  }
}

// mix/desktop 布局：窄窗切换抽屉侧边栏，宽窗切换折叠
function toggleSidebar() {
  if (isNarrow.value) {
    mobileSidebarOpen.value = !mobileSidebarOpen.value;
  } else {
    settings.sidebarCollapsed = !settings.sidebarCollapsed;
    settings.save();
  }
}

// 全屏状态变化回调
function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement;
}

// 窗口尺寸变化回调（侧边栏折叠/桌面窗口自适应）
function onWindowResize() {
  isNarrow.value = window.innerWidth < 768;
  const isSide = settings.layoutMode === 'side' || settings.layoutMode === 'rightSide';
  // side/rightSide 布局：窄窗临时折叠；mix/desktop 布局：CSS 隐藏侧边栏为抽屉，无需折叠
  if (isNarrow.value && !narrowApplied && isSide) {
    narrowPrevCollapsed = settings.sidebarCollapsed;
    settings.sidebarCollapsed = true;
    narrowApplied = true;
  } else if ((!isNarrow.value || !isSide) && narrowApplied) {
    settings.sidebarCollapsed = narrowPrevCollapsed;
    narrowApplied = false;
  }
  if (!isNarrow.value) {
    mobileSidebarOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener('fullscreenchange', onFullscreenChange);
  onWindowResize();
  window.addEventListener('resize', onWindowResize);
  // 布局切换后立即重算窄窗状态
  watch(
    () => settings.layoutMode,
    () => onWindowResize()
  );
});

onBeforeUnmount(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange);
  window.removeEventListener('resize', onWindowResize);
});
</script>

<style scoped>
.app-wrapper {
  display: flex;
  height: 100%;
  width: 100%;
}

/* 顶部导航布局：纵向排列（顶栏在上、内容在下），避免顶栏被水平 flex 挤成半宽 */
.app-wrapper.layout-top,
.app-wrapper.layout-bottom {
  flex-direction: column;
}

/* 混合布局 / 桌面布局：整体纵向（顶栏 + 主体行） */
.app-wrapper.layout-mix,
.app-wrapper.layout-desktop {
  flex-direction: column;
}

/* 垂直页签布局：左侧页签栏 + 内容区 */
.vtab-body {
  flex: 1;
  display: flex;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
}

.vtab-main {
  flex: 1;
  min-width: 0;
}

/* 混合/桌面布局的主体行：侧边栏 + 内容区 */
.mix-body {
  flex: 1;
  display: flex;
  min-height: 0;
  min-width: 0;
}

/* 桌面布局顶栏：通栏系统栏（Logo + 面包屑 + 功能区） */
.desktop-header {
  height: 46px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  width: 100%;
  background: var(--go-navbar-bg, #fff);
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 16px;
}

.desktop-breadcrumb {
  flex: 1;
  min-width: 0;
  padding: 0 18px;
  overflow: hidden;
  white-space: nowrap;
}

.desktop-breadcrumb :deep(.el-breadcrumb) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-container {
  width: 200px;
  height: 100%;
  flex-shrink: 0;
  transition: width 0.28s;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.app-footer {
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #909399;
  background: var(--go-footer-bg, #fff);
  border-top: 1px solid var(--go-border-color, #e4e7ed);
}

/* ===== 顶部导航布局 ===== */
.top-header {
  height: 46px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  width: 100%;
  background: var(--go-navbar-bg, #fff);
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 16px;
}

.top-logo {
  font-size: 18px;
  font-weight: 600;
  color: var(--go-topnav-logo-text, #303133);
  margin-right: 24px;
  white-space: nowrap;
  flex-shrink: 0;
}

.top-nav-slot {
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.top-nav-slot::-webkit-scrollbar {
  display: none;
}

.top-right {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: 16px;
  flex-shrink: 0;
}

.nav-icon {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  color: var(--go-navbar-text, #303133);
  transition: background 0.2s;
}

.nav-icon:hover {
  background: rgba(0, 0, 0, 0.06);
}

.top-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

/* 底部导航布局：内容区在上（纵向 flex） */
.bottom-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

/* 底部导航栏：与顶栏镜像（边框/阴影朝上） */
.bottom-header {
  height: 46px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  width: 100%;
  background: var(--go-navbar-bg, #fff);
  box-shadow: 0 -1px 4px rgba(0, 21, 41, 0.08);
  border-top: 1px solid var(--go-border-color, #e4e7ed);
  padding: 0 16px;
}

/* 右侧边栏布局：侧边栏左缘分隔线（镜像），Logo 区内边距镜像 */
.layout-rightSide .sidebar-container {
  border-left: 1px solid var(--go-border-color, rgba(0, 0, 0, 0.08));
}

.layout-rightSide .sidebar :deep(.logo) {
  padding: 0 18px 0 8px;
}

.top-breadcrumb {
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  padding: 0 16px;
  background: var(--go-main-bg, #f0f2f5);
  font-size: 13px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
}

.top-breadcrumb :deep(.el-breadcrumb) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 窄屏（<480px）：顶栏/底部栏两行——第一行 Logo+右侧操作，第二行横向菜单滚动，避免横向溢出 */
@media (max-width: 480px) {
  .top-header,
  .bottom-header {
    flex-wrap: wrap;
    height: auto;
    padding: 0 12px;
  }

  .desktop-header {
    flex-wrap: wrap;
    height: auto;
    padding: 0 12px;
  }

  .top-logo {
    display: none;
  }

  .desktop-breadcrumb {
    display: none;
  }

  .top-header .nav-icon[title*='搜索'],
  .top-header .nav-icon[title*='全屏'],
  .bottom-header .nav-icon[title*='搜索'],
  .bottom-header .nav-icon[title*='全屏'],
  .desktop-header .nav-icon[title*='搜索'],
  .desktop-header .nav-icon[title*='全屏'] {
    display: none;
  }

  .top-right {
    height: 50px;
    margin-left: auto;
  }

  .top-nav-slot {
    order: 3;
    flex-basis: 100%;
    width: 100%;
    height: 42px;
    border-top: 1px solid var(--go-border-color, #e4e7ed);
  }

  .top-menu :deep(.el-menu-item),
  .top-menu :deep(.el-sub-menu__title) {
    height: 42px;
    line-height: 42px;
  }
}

/* 窄窗（<768px）：mix/desktop 布局侧边栏由 flow 转为 el-drawer 抽屉，避免挤压主内容区 */
@media (max-width: 768px) {
  .layout-mix .sidebar-container,
  .layout-desktop .sidebar-container {
    display: none;
  }
}
</style>

<!-- 全局：窄窗抽屉侧边栏（el-drawer append-to-body，不受 scoped 限制） -->
<style>
.go-sidebar-drawer .el-drawer__body {
  padding: 0;
  display: flex;
}

.go-sidebar-drawer .sidebar {
  width: 100%;
  height: 100%;
}
</style>
