<template>
  <div class="navbar">
    <div class="left">
      <div v-if="settings.breadcrumb" class="breadcrumb">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">
            <el-icon :size="13" style="vertical-align: -2px; margin-right: 2px"
              ><HomeFilled
            /></el-icon>
            {{ menuTitle(t('menu.home')) }}
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
      <!-- 多页签与面包屑同一行（侧边栏布局） -->
      <div class="navbar-tags">
        <TagsView v-if="settings.showTagsView" />
      </div>
    </div>
    <div class="right-menu">
      <div class="nav-icon" :title="t('layout.searchMenu')" @click="openSearch">
        <el-icon :size="18"><Search /></el-icon>
      </div>
      <div
        class="nav-icon"
        :title="isFullscreen ? t('navbar.exitFullscreen') : t('navbar.fullscreen')"
        @click="toggleFullscreen"
      >
        <el-icon :size="18"><component :is="isFullscreen ? 'Aim' : 'FullScreen'" /></el-icon>
      </div>
      <div class="nav-icon" :title="t('layout.settings')" @click="settings.drawerVisible = true">
        <el-icon :size="18"><Setting /></el-icon>
      </div>
      <UserDropdown />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Setting, Search, HomeFilled } from '@element-plus/icons-vue';
import { useSettingsStore } from '@/store/settings';
import { useMenuTitle } from '@/i18n/translate';
import UserDropdown from './UserDropdown.vue';
import TagsView from './TagsView.vue';

const { t } = useI18n();
const route = useRoute();
const settings = useSettingsStore();
const menuTitle = useMenuTitle();

const isFullscreen = ref(false);

// 完整面包屑：route.matched 中含 meta.title 的层级（不含 Layout 根与固定首页，避免 /index 重复）
const breadcrumbs = computed(() =>
  route.matched.filter((item) => item.meta && item.meta.title && item.meta.title !== '首页')
);

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

// 全屏状态变化回调
function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement;
}

onMounted(() => {
  document.addEventListener('fullscreenchange', onFullscreenChange);
});

onBeforeUnmount(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange);
});
</script>

<style scoped>
.navbar {
  height: 48px;
  background: var(--go-navbar-bg, #fff);
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  flex-shrink: 0;
}

.left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.breadcrumb {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  flex-shrink: 0;
  max-width: 45%;
}

.breadcrumb :deep(.el-breadcrumb) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.breadcrumb :deep(.el-breadcrumb__inner),
.breadcrumb :deep(.el-breadcrumb__separator) {
  white-space: nowrap;
}

/* 多页签：与面包屑同一行（剩余空间横向滚动） */
.navbar-tags {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
}

.navbar-tags :deep(.tags-view) {
  height: 34px;
  border-bottom: none;
  background: transparent;
}

.right-menu {
  display: flex;
  align-items: center;
  gap: 12px;
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

/* 窄窗（<768px）：空间有限，面包屑隐藏、隐藏非核心图标，避免导航栏横向溢出 */
@media (max-width: 768px) {
  .breadcrumb {
    display: none;
  }

  .navbar .nav-icon[title*='搜索'],
  .navbar .nav-icon[title*='全屏'] {
    display: none;
  }

  .right-menu {
    gap: 6px;
    min-width: 0;
  }
}
</style>
