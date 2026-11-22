<template>
  <div class="tags-view">
    <div class="tags-scroll">
      <router-link
        v-for="tag in visitedViews"
        :key="tag.path"
        :to="tag.path"
        class="tags-item"
        :class="{ active: isActive(tag) }"
        @click.middle="closeTag(tag)"
        @contextmenu.prevent="openMenu(tag, $event)"
      >
        <el-icon v-if="tag.affix" :size="12" style="margin-right: 2px"><HomeFilled /></el-icon>
        {{ menuTitle(tag.title) }}
        <el-icon v-if="!tag.affix" :size="11" class="tag-close" @click.prevent.stop="closeTag(tag)">
          <Close />
        </el-icon>
      </router-link>
    </div>

    <!-- 右键菜单 -->
    <ul
      v-show="menuVisible"
      class="context-menu"
      :style="{ left: menuLeft + 'px', top: menuTop + 'px' }"
    >
      <li @click="refreshSelectedTag">{{ t('tagsView.refresh') }}</li>
      <li v-if="!selectedTag?.affix" @click="closeSelectedTag">{{ t('tagsView.close') }}</li>
      <li @click="closeOthersTags">{{ t('tagsView.closeOthers') }}</li>
      <li @click="closeAllTags">{{ t('tagsView.closeAll') }}</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { HomeFilled, Close } from '@element-plus/icons-vue';
import { useTagsViewStore } from '@/store/tagsView';
import { useSettingsStore } from '@/store/settings';
import { useMenuTitle } from '@/i18n/translate';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const tagsStore = useTagsViewStore();
const settings = useSettingsStore();
const menuTitle = useMenuTitle();

const visitedViews = computed(() => tagsStore.visitedViews);
const menuVisible = ref(false);
const menuLeft = ref(0);
const menuTop = ref(0);
const selectedTag = ref(undefined);

// 判断页签是否激活
function isActive(tag) {
  return tag.path === route.path;
}

// 关闭页签
function closeTag(tag) {
  if (tag.affix) return;
  tagsStore.delView(tag);
  // 关闭的是当前页则跳转到最后一个标签
  if (isActive(tag)) {
    toLastTag();
  }
}

// 跳转到最后一个页签
function toLastTag() {
  const views = tagsStore.visitedViews;
  const lastView = views[views.length - 1];
  if (lastView) {
    router.push(lastView.path);
  } else {
    router.push('/index');
  }
}

// 打开页签右键菜单
function openMenu(tag, e) {
  selectedTag.value = tag;
  const offsetLeft = e.clientX;
  const offsetTop = e.clientY;
  menuLeft.value = offsetLeft;
  menuTop.value = offsetTop;
  menuVisible.value = true;
}

// 关闭页签右键菜单
function closeMenu() {
  menuVisible.value = false;
}

// 刷新当前页签
function refreshSelectedTag() {
  tagsStore.refreshView();
  closeMenu();
}

// 关闭当前页签
function closeSelectedTag() {
  closeTag(selectedTag.value);
  closeMenu();
}

// 关闭其他页签
function closeOthersTags() {
  tagsStore.delOthersViews(selectedTag.value);
  closeMenu();
}

// 关闭全部页签
function closeAllTags() {
  tagsStore.delAllViews();
  router.push('/index');
  closeMenu();
}

// 点击页面其他区域关闭右键菜单
function handleGlobalClick() {
  closeMenu();
}

onMounted(() => {
  // 首页标签常驻（affix），再添加当前路由
  tagsStore.addView({ path: '/index', meta: { title: '首页' } });
  tagsStore.addView(route);
  document.addEventListener('click', handleGlobalClick);
  // 标签自动滚动到可见区域
  nextTick(() => {
    const el = document.querySelector('.tags-item.active');
    el?.scrollIntoView({ inline: 'nearest' });
  });
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleGlobalClick);
});

// 路由变化时添加标签并高亮滚动
watch(
  () => route.path,
  () => {
    tagsStore.addView(route);
    nextTick(() => {
      const el = document.querySelector('.tags-item.active');
      el?.scrollIntoView({ inline: 'nearest' });
    });
  }
);

// 关闭标签时若 settings.showTagsView 关闭同步清空
watch(
  () => settings.showTagsView,
  (v) => {
    if (!v) {
      tagsStore.delAllViews();
    }
  }
);
</script>

<style scoped>
.tags-view {
  height: 32px;
  background: var(--go-navbar-bg, #fff);
  border-bottom: 1px solid var(--go-border-color, #e4e7ed);
  display: flex;
  align-items: center;
  flex-shrink: 0;
  position: relative;
}

.tags-scroll {
  display: flex;
  align-items: center;
  overflow-x: auto;
  padding: 0 8px;
  gap: 6px;
  white-space: nowrap;
  scrollbar-width: none;
}

.tags-scroll::-webkit-scrollbar {
  display: none;
}

.tags-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  height: 22px;
  padding: 0 10px;
  font-size: 12px;
  color: var(--go-navbar-text, #303133);
  background: transparent;
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.2s;
}

.tags-item:hover {
  color: var(--go-topnav-active, #409eff);
}

.tags-item.active {
  color: #fff;
  background: var(--go-topnav-active, #409eff);
  border-color: var(--go-topnav-active, #409eff);
}

.tag-close {
  border-radius: 50%;
  padding: 1px;
}

.tag-close:hover {
  background: rgba(255, 255, 255, 0.4);
}

.context-menu {
  position: fixed;
  z-index: 3000;
  list-style: none;
  margin: 0;
  padding: 5px 0;
  background: var(--go-bg-page);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  font-size: 13px;
  min-width: 110px;
}

.context-menu li {
  padding: 7px 16px;
  cursor: pointer;
  color: var(--go-text-primary);
}

.context-menu li:hover {
  background: var(--go-fill-light);
  color: #409eff;
}
</style>
