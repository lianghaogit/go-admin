<template>
  <div class="app-main">
    <!-- 多页签（布局配置可关闭）：side/rightSide 布局已并入 Navbar（与面包屑一行），此处仅 top/bottom/mix 布局显示 -->
    <TagsView v-if="settings.showTagsView && settings.layoutMode !== 'side' && settings.layoutMode !== 'rightSide'" />
    <!-- 刷新时通过 refreshKey 重新挂载当前视图；keep-alive 按标签缓存页面状态；路由切换 fade-transform 动画 -->
    <div class="app-main-body">
      <router-view v-slot="{ Component }">
        <transition name="fade-transform" mode="out-in">
          <keep-alive :include="tagsStore.cachedViews">
            <component :is="Component" :key="tagsStore.refreshKey + '-' + $route.path" />
          </keep-alive>
        </transition>
      </router-view>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from '@/store/settings';
import { useTagsViewStore } from '@/store/tagsView';
import TagsView from './TagsView.vue';

const settings = useSettingsStore();
const tagsStore = useTagsViewStore();
</script>

<style scoped>
.app-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--go-main-bg, #f0f2f5);
}

.app-main-body {
  flex: 1;
  overflow: auto;
  padding: 16px;
}

/* 路由切换动画（fade-transform，对标 RuoYi） */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.25s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
