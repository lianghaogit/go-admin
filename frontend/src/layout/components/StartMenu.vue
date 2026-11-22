<template>
  <!-- 基础面板：一级目录（相对桌面画布定位，左下角弹出） -->
  <div class="sm-flyout sm-base" @click.stop>
    <div class="sm-header">
      <span class="sm-title">{{ t('layout.start') }}</span>
    </div>
    <div class="sm-search">
      <el-icon class="sm-search-ico" :size="14"><Search /></el-icon>
      <input
        v-model="keyword"
        class="sm-search-input"
        :placeholder="t('layout.searchMenuPlaceholder')"
        @keydown.enter="enterSearch"
        @keydown.down.prevent="moveSearch(1)"
        @keydown.up.prevent="moveSearch(-1)"
      />
    </div>
    <div v-if="!keyword" class="sm-list">
      <div
        v-for="item in level1"
        :key="item.menuId"
        class="sm-item"
        :class="{ 'sm-item-sel': selected[0] === item }"
        @click="onItem(item, 0)"
      >
        <el-icon v-if="item.icon && showIcon" class="sm-ico"
          ><component :is="safeIcon(item.icon)"
        /></el-icon>
        <span class="sm-item-name">{{ menuTitle(item.menuName) }}</span>
        <el-icon v-if="hasKids(item)" class="sm-arrow"><ArrowRight /></el-icon>
      </div>
    </div>
    <!-- 搜索模式：扁平结果列表（标题/父路径匹配，点击直达） -->
    <div v-else class="sm-list">
      <div v-if="!searchResults.length" class="sm-empty">{{ t('layout.searchMenuEmpty') }}</div>
      <div
        v-for="(r, i) in searchResults"
        :key="r.item.menuId"
        class="sm-item"
        :class="{ 'sm-item-sel': searchActive === i }"
        @mouseenter="searchActive = i"
        @click="openSearch(r)"
      >
        <el-icon v-if="r.item.icon && showIcon" class="sm-ico"
          ><component :is="safeIcon(r.item.icon)"
        /></el-icon>
        <span class="sm-item-name">{{ menuTitle(r.item.menuName) }}</span>
        <span class="sm-parent">{{ r.parentTitle }}</span>
      </div>
    </div>
  </div>
  <!-- 级联浮层：点击目录逐级弹出（每级右移 176px，不被父面板裁剪；窄屏右对齐视口防溢出） -->
  <div
    v-for="(panel, i) in subPanels"
    v-show="!keyword"
    :key="'sub' + i"
    class="sm-flyout"
    :style="{ left: flyoutLeft(i) + 'px' }"
    @click.stop
  >
    <div class="sm-header">
      <span class="sm-title">{{ menuTitle(parentNames[i]) }}</span>
    </div>
    <div class="sm-list">
      <div
        v-for="item in panel"
        :key="item.menuId"
        class="sm-item"
        :class="{ 'sm-item-sel': selected[i + 1] === item }"
        @click="onItem(item, i + 1)"
      >
        <el-icon v-if="item.icon && showIcon" class="sm-ico"
          ><component :is="safeIcon(item.icon)"
        /></el-icon>
        <span class="sm-item-name">{{ menuTitle(item.menuName) }}</span>
        <el-icon v-if="hasKids(item)" class="sm-arrow"><ArrowRight /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { PropType } from 'vue';
import { ArrowRight, Search } from '@element-plus/icons-vue';
import { useSettingsStore } from '@/store/settings';
import { safeIcon } from '@/utils/iconNames';
import { useMenuTitle } from '@/i18n/translate';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const menuTitle = useMenuTitle();

// 菜单项（后端菜单树节点，宽松类型）
interface MenuItem {
  menuId: number;
  menuName: string;
  menuType?: string;
  icon?: string;
  path?: string;
  children?: MenuItem[];
}

const props = defineProps({
  menus: { type: Array as PropType<MenuItem[]>, default: () => [] }
});
const emit = defineEmits(['open', 'close']);

const settings = useSettingsStore();
const showIcon = computed(() => settings.menuIcon);

// 层级面板：panels[0] = 顶层；点击目录后逐级追加 children 面板
const panels = ref([] as MenuItem[][]);
// 每级选中的菜单项（用于拼接完整路径与面板标题）
const selected = ref([] as MenuItem[]);

const level1 = computed(() => (props.menus || []).filter((m) => m.menuType !== 'F'));
const subPanels = computed(() => panels.value.slice(1));
// 级联面板标题 = 该面板父级选中的目录名（subPanels[i] 的父级即 selected[i]）
const parentNames = computed(() =>
  subPanels.value.map((_, i) => selected.value[i]?.menuName || '')
);

// 级联面板水平定位：正常逐级右移 176px；超出视口右侧时右对齐视口边缘（窄屏防溢出）
function flyoutLeft(i) {
  const base = 8 + (i + 1) * 176;
  const w = typeof window !== 'undefined' ? window.innerWidth : 1200;
  if (base + 176 > w) return Math.max(8, w - 8 - 176);
  return base;
}

// 判断菜单是否含子菜单
function hasKids(item) {
  return (item.children || []).some((c) => c.menuType !== 'F');
}

// 点击开始菜单项（含级联路径跳转）
function onItem(item, level) {
  selected.value = selected.value.slice(0, level).concat([item]);
  const kids = (item.children || []).filter((c) => c.menuType !== 'F');
  if (kids.length) {
    // 目录：截断右侧层级，弹出下一层
    panels.value = panels.value.slice(0, level + 1).concat([kids]);
  } else {
    // 菜单叶子：拼接完整路径并打开
    let path = '';
    for (const it of selected.value) {
      path += '/' + it.path;
    }
    emit('open', path);
  }
}

watch(
  () => props.menus,
  () => {
    panels.value = [level1.value];
    selected.value = [];
  },
  { immediate: true }
);

// ===== 开始菜单搜索 =====
const keyword = ref('');
const searchActive = ref(0);
interface SearchHit {
  item: MenuItem;
  pathStack: MenuItem[];
  parentTitle: string;
}
const searchResults = computed<SearchHit[]>(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return [];
  const flat: SearchHit[] = [];
  const walk = (items: MenuItem[], stack: MenuItem[]) => {
    (items || []).forEach((m) => {
      if (m.menuType === 'F') return;
      const kids = (m.children || []).filter((c) => c.menuType !== 'F');
      const next = stack.concat([m]);
      if (kids.length) walk(kids, next);
      else
        flat.push({
          item: m,
          pathStack: next,
          parentTitle: next
            .slice(0, -1)
            .map((x) => menuTitle(x.menuName))
            .join(' / ')
        });
    });
  };
  walk(level1.value, []);
  return flat.filter(({ item, pathStack }) => {
    const title = menuTitle(item.menuName).toLowerCase();
    const parents = pathStack
      .slice(0, -1)
      .map((x) => menuTitle(x.menuName))
      .join('/')
      .toLowerCase();
    return title.includes(kw) || parents.includes(kw) || (item.path || '').toLowerCase().includes(kw);
  });
});

function openSearch(hit: SearchHit) {
  let path = '';
  hit.pathStack.forEach((it) => (path += '/' + it.path));
  emit('open', path);
}

function enterSearch() {
  if (searchResults.value.length) openSearch(searchResults.value[searchActive.value]);
}

function moveSearch(dir: number) {
  const n = searchResults.value.length;
  if (!n) return;
  searchActive.value = (searchActive.value + dir + n) % n;
}
</script>

<style scoped>
.sm-flyout {
  position: absolute;
  left: 8px;
  bottom: 46px;
  width: 176px;
  height: 420px;
  max-height: calc(100vh - 56px);
  display: flex;
  flex-direction: column;
  background: var(--go-startmenu-bg, rgba(244, 246, 250, 0.97));
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  box-shadow: 0 12px 36px rgba(0, 21, 41, 0.24);
  z-index: 3000;
  overflow: hidden;
  backdrop-filter: blur(10px);
}

.sm-header {
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  padding: 0 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--go-navbar-text, #303133);
  border-bottom: 1px solid var(--go-border-color, #e4e7ed);
}

.sm-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 6px;
  box-sizing: border-box;
}

.sm-search {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 8px 2px;
  padding: 0 8px;
  height: 30px;
  border: 1px solid var(--go-border-color, #d8dce5);
  border-radius: 2px;
  background: var(--go-card-bg, rgba(255, 255, 255, 0.8));
}

.sm-search-ico {
  color: #909399;
  flex-shrink: 0;
}

.sm-search-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--go-navbar-text, #303133);
}

.sm-search-input::placeholder {
  color: #a8abb2;
}

.sm-parent {
  flex-shrink: 0;
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 11px;
  color: #909399;
}

.sm-empty {
  padding: 24px 0;
  text-align: center;
  font-size: 13px;
  color: #909399;
}

.sm-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 10px;
  border-radius: 2px;
  font-size: 13px;
  color: var(--go-navbar-text, #303133);
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.12s;
}

.sm-item:hover {
  background: rgba(64, 158, 255, 0.12);
}

.sm-item.sm-item-sel {
  background: rgba(64, 158, 255, 0.2);
  color: var(--go-topnav-active, #409eff);
}

.sm-ico {
  color: var(--go-topnav-active, #409eff);
  flex-shrink: 0;
}

.sm-item-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sm-arrow {
  flex-shrink: 0;
  color: #909399;
}
</style>
