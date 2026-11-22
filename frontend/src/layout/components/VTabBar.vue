<template>
  <!-- 垂直页签布局：左侧垂直页签栏（浏览器侧栏标签风格），点击切换、hover 关闭 -->
  <div class="vtab-bar">
    <div class="vtab-head">
      <div v-if="settings.showLogo" class="vtab-logo">Go Admin</div>
    </div>
    <div class="vtab-list">
      <div
        v-for="w in tagsStore.visitedViews"
        :key="w.path"
        class="vtab-item"
        :class="{ active: w.path === route.path, dragging: w.path === dragPath }"
        :draggable="!w.affix"
        @click="goTab(w)"
        @auxclick="auxClose(w, $event)"
        @contextmenu.prevent="openCtx(w, $event)"
        @dragstart="dragStart(w, $event)"
        @dragover.prevent="dragOver(w, $event)"
        @drop="dropOn(w)"
        @dragend="dragEnd"
      >
        <el-icon :size="15"><component :is="w.affix ? 'HomeFilled' : w.meta?.icon || 'Document'" /></el-icon>
        <span class="vtab-title">{{ menuTitle(w.title) }}</span>
        <span
          v-if="!w.affix"
          class="vtab-close"
          title="关闭"
          @click.stop="closeTab(w)"
        >
          <el-icon :size="11"><Close /></el-icon>
        </span>
      </div>
    </div>
    <div class="vtab-foot">
      <div class="vtab-act" :title="t('layout.launchpad')" @click="openMenu">
        <el-icon :size="16"><Grid /></el-icon>
      </div>
      <div class="vtab-act" title="全屏" @click="toggleFullscreen">
        <el-icon :size="16"><FullScreen /></el-icon>
      </div>
      <div class="vtab-act" title="布局配置" @click="settings.drawerVisible = true">
        <el-icon :size="16"><Setting /></el-icon>
      </div>
      <div class="vtab-act" title="回到首页" @click="router.push('/index')">
        <el-icon :size="16"><HomeFilled /></el-icon>
      </div>
      <div class="vtab-user">
        <UserDropdown />
      </div>
    </div>
  </div>

  <!-- 页签右键菜单：刷新 / 关闭当前 / 关闭其他 / 关闭所有 -->
  <Teleport to="body">
    <Transition name="vtab-ctx-fade">
      <div v-if="ctx" class="vtab-ctx" :style="{ left: ctx.x + 'px', top: ctx.y + 'px' }" @click.stop>
        <div class="vtab-ctx-item" @click="ctxRefresh">
          <el-icon :size="13"><Refresh /></el-icon>
          <span>{{ t('tagsView.refresh') }}</span>
        </div>
        <div v-if="!ctx.affix" class="vtab-ctx-item" @click="ctxClose">
          <el-icon :size="13"><Close /></el-icon>
          <span>{{ t('tagsView.close') }}</span>
        </div>
        <div class="vtab-ctx-sep"></div>
        <div class="vtab-ctx-item" @click="ctxCloseOthers">
          <el-icon :size="13"><CircleClose /></el-icon>
          <span>{{ t('tagsView.closeOthers') }}</span>
        </div>
        <div class="vtab-ctx-item" @click="ctxCloseAll">
          <el-icon :size="13"><FolderDelete /></el-icon>
          <span>{{ t('tagsView.closeAll') }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>

  <!-- 全部菜单入口：垂直页签只显示已打开的页签，这里补齐完整菜单 -->
  <Teleport to="body">
    <Transition name="vtab-menu">
      <div v-if="menuOpen" class="vtab-menu-mask" @click.self="closeMenu">
        <div class="vtab-menu-panel" @click.stop>
          <div class="vtab-menu-head">
            <el-icon :size="15"><Search /></el-icon>
            <input
              ref="menuSearchInput"
              v-model="menuKeyword"
              class="vtab-menu-search"
              :placeholder="t('layout.searchMenu')"
              @keydown.esc="closeMenu"
              @keydown.enter.prevent="openFirstMenu"
            />
            <button class="vtab-menu-close" @click="closeMenu">
              <el-icon :size="14"><Close /></el-icon>
            </button>
          </div>
          <div class="vtab-menu-body">
            <div v-for="group in menuGroups" :key="group.label" class="vtab-menu-group">
              <div class="vtab-menu-group-title">{{ group.label }}</div>
              <div
                v-for="item in group.apps"
                :key="item.key"
                class="vtab-menu-item"
                @click="openMenuItem(item)"
              >
                <el-icon v-if="item.icon" :size="14"
                  ><component :is="safeIcon(item.icon)"
                /></el-icon>
                <span class="vtab-menu-name">{{ item.label }}</span>
                <span v-if="!menuKeyword" class="vtab-menu-parent">{{ item.parent }}</span>
              </div>
            </div>
            <div v-if="!menuGroups.length" class="vtab-menu-empty">
              {{ t('layout.searchMenuEmpty') }}
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { CircleClose, Close, FolderDelete, FullScreen, Grid, Refresh, Search, Setting, HomeFilled } from '@element-plus/icons-vue';
import { useTagsViewStore } from '@/store/tagsView';
import { useSettingsStore } from '@/store/settings';
import { useUserStore } from '@/store/user';
import { useMenuTitle } from '@/i18n/translate';
import { safeIcon } from '@/utils/iconNames';
import { useI18n } from 'vue-i18n';
import UserDropdown from './UserDropdown.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const tagsStore = useTagsViewStore();
const settings = useSettingsStore();
const userStore = useUserStore();
const menuTitle = useMenuTitle();

const ORDER_KEY = 'go-admin-vtab-order';
const dragPath = ref<string | null>(null);

// ===== 页签右键菜单（刷新 / 关闭当前 / 关闭其他 / 关闭所有） =====
const ctx = ref<{ x: number; y: number; path: string; affix: boolean } | null>(null);

function openCtx(w, e: MouseEvent) {
  const pad = 8;
  const mw = 150;
  const mh = 156;
  ctx.value = {
    x: Math.max(pad, Math.min(e.clientX, window.innerWidth - mw - pad)),
    y: Math.max(pad, Math.min(e.clientY, window.innerHeight - mh - pad)),
    path: w.path,
    affix: !!w.affix
  };
}

function closeCtx() {
  ctx.value = null;
}

// 切换全屏（vtab 无顶栏的常用入口）
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {});
  } else {
    document.exitFullscreen().catch(() => {});
  }
}

function ctxView() {
  return ctx.value ? tagsStore.visitedViews.find((v) => v.path === ctx.value!.path) : undefined;
}

function ctxRefresh() {
  tagsStore.refreshView();
  closeCtx();
}

function ctxClose() {
  const w = ctxView();
  if (w) closeTab(w);
  closeCtx();
}

function ctxCloseOthers() {
  const w = ctxView();
  if (w) tagsStore.delOthersViews(w);
  closeCtx();
}

function ctxCloseAll() {
  tagsStore.delAllViews();
  closeCtx();
}

// 点击菜单外部关闭
function onDocClick(e: MouseEvent) {
  const t = e.target as HTMLElement;
  if (!t.closest('.vtab-ctx')) closeCtx();
}

function goTab(w) {
  if (route.path !== w.path) router.push(w.path);
}

function closeTab(w) {
  tagsStore.delView(w);
  if (route.path === w.path) {
    const views = tagsStore.visitedViews;
    const last = views[views.length - 1];
    router.push(last ? last.path : '/index');
  }
  persistOrder();
}

// 中键关闭
function auxClose(w, e) {
  if (e.button === 1) {
    e.preventDefault();
    closeTab(w);
  }
}

// HTML5 拖拽排序
function dragStart(w, e) {
  if (w.affix) return;
  dragPath.value = w.path;
  e.dataTransfer.effectAllowed = 'move';
  try {
    e.dataTransfer.setData('text/plain', w.path);
  } catch {}
}

function dragOver(w, e) {
  if (!dragPath.value || w.path === dragPath.value) return;
  e.preventDefault();
}

function dropOn(target) {
  if (!dragPath.value) return;
  const arr = [...tagsStore.visitedViews];
  const from = arr.findIndex((v) => v.path === dragPath.value);
  const to = arr.findIndex((v) => v.path === target.path);
  if (from < 0 || to < 0 || from === to) return;
  const [mv] = arr.splice(from, 1);
  arr.splice(to, 0, mv);
  tagsStore.visitedViews = arr;
  dragPath.value = null;
  persistOrder();
}

function dragEnd() {
  dragPath.value = null;
}

function persistOrder() {
  localStorage.setItem(
    ORDER_KEY,
    JSON.stringify(tagsStore.visitedViews.map((v) => v.path))
  );
}

// 启动时按持久化顺序重排（忽略已不存在的路径）
function restoreOrder() {
  try {
    const order = JSON.parse(localStorage.getItem(ORDER_KEY) || '[]');
    if (!Array.isArray(order) || !order.length) return;
    const arr = [...tagsStore.visitedViews];
    arr.sort((a, b) => {
      const ia = order.indexOf(a.path);
      const ib = order.indexOf(b.path);
      if (ia === -1 && ib === -1) return 0;
      if (ia === -1) return 1;
      if (ib === -1) return -1;
      return ia - ib;
    });
    tagsStore.visitedViews = arr;
  } catch {}
}

// ===== 全部菜单（垂直页签只显示已打开页签，这里提供完整菜单入口） =====
const menuOpen = ref(false);
const menuKeyword = ref('');
const menuSearchInput = ref<HTMLInputElement | null>(null);

interface MenuLeaf {
  key: string;
  label: string;
  icon?: string;
  path: string;
  parent: string;
  group: string;
}

function joinMenuPath(chain: any[]) {
  return chain.map((m) => '/' + String(m.path || '').replace(/^\/+/, '')).join('');
}

// 拍平后端菜单树：目录作为分类名，叶子作为可打开项
const flatMenus = computed<MenuLeaf[]>(() => {
  const out: MenuLeaf[] = [];
  const walk = (items: any[], chain: any[], group: string) => {
    (items || []).forEach((m) => {
      if (m.menuType === 'F') return;
      const next = chain.concat([m]);
      const kids = (m.children || []).filter((c: any) => c.menuType !== 'F');
      if (kids.length) {
        walk(kids, next, chain.length === 0 ? menuTitle(m.menuName) : group);
        return;
      }
      out.push({
        key: 'm' + m.menuId,
        label: menuTitle(m.menuName),
        icon: m.icon,
        path: joinMenuPath(next),
        parent: next
          .slice(0, -1)
          .map((x: any) => menuTitle(x.menuName))
          .join(' / '),
        group: chain.length === 0 ? t('layout.launchpadOther') : group
      });
    });
  };
  walk((userStore.menus || []) as any[], [], '');
  return out;
});

// 按分类分组；搜索时只保留匹配项
const menuGroups = computed<{ label: string; apps: MenuLeaf[] }[]>(() => {
  const kw = menuKeyword.value.trim().toLowerCase();
  const items = kw
    ? flatMenus.value.filter(
        (m) =>
          m.label.toLowerCase().includes(kw) ||
          m.parent.toLowerCase().includes(kw) ||
          m.path.toLowerCase().includes(kw)
      )
    : flatMenus.value;
  const order: string[] = [];
  const map = new Map<string, MenuLeaf[]>();
  items.forEach((m) => {
    if (!map.has(m.group)) {
      map.set(m.group, []);
      order.push(m.group);
    }
    map.get(m.group)!.push(m);
  });
  return order.map((g) => ({ label: g, apps: map.get(g)! }));
});

function openMenu() {
  menuKeyword.value = '';
  menuOpen.value = true;
  nextTick(() => menuSearchInput.value?.focus());
}

function closeMenu() {
  menuOpen.value = false;
  menuKeyword.value = '';
}

function openMenuItem(item: MenuLeaf) {
  closeMenu();
  if (item.path) router.push(item.path);
}

// 回车打开搜索到的第一个菜单
function openFirstMenu() {
  const first = menuGroups.value[0]?.apps[0];
  if (first) openMenuItem(first);
}

onMounted(() => {
  restoreOrder();
  // 垂直页签自身负责登记标签（此布局不使用 TagsView，页签不会自动新增）
  tagsStore.addView({ path: '/index', meta: { title: '首页' } });
  tagsStore.addView(route);
  document.addEventListener('click', onDocClick);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick);
});

// 路由变化时登记新页签
watch(
  () => route.path,
  () => {
    tagsStore.addView(route);
    nextTick(persistOrder);
  }
);
</script>

<style scoped>
.vtab-bar {
  width: 128px;
  flex-shrink: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--go-sidebar-bg, #23262b);
  border-right: 1px solid var(--go-border-color, #e4e7ed);
}

html.sidebar-light .vtab-bar {
  background: var(--go-sidebar-bg-light, #f5f7fa);
}

.vtab-head {
  flex-shrink: 0;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(127, 127, 127, 0.18);
}

.vtab-logo {
  font-size: 15px;
  font-weight: 700;
  color: var(--go-navbar-text, #e5e9f0);
}

html.sidebar-light .vtab-logo {
  color: #303133;
}

.vtab-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 6px 0;
}

.vtab-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  margin: 2px 6px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 12px;
  color: var(--go-sidebar-text, #cfd3dc);
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.12s;
}

html.sidebar-light .vtab-item {
  color: #606266;
}

.vtab-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

html.sidebar-light .vtab-item:hover {
  background: rgba(64, 158, 255, 0.12);
}

.vtab-item.dragging {
  opacity: 0.4;
}

.vtab-item.active {
  background: var(--el-color-primary);
  color: #fff;
}

.vtab-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.vtab-close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  opacity: 0;
  transition: opacity 0.12s;
}

.vtab-item:hover .vtab-close {
  opacity: 0.85;
}

.vtab-close:hover {
  background: rgba(255, 255, 255, 0.25);
  opacity: 1;
}

html.sidebar-light .vtab-close:hover {
  background: rgba(0, 0, 0, 0.1);
}

.vtab-foot {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 8px 4px;
  border-top: 1px solid rgba(127, 127, 127, 0.18);
}

.vtab-act {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  color: var(--go-sidebar-text, #cfd3dc);
  cursor: pointer;
}

/* 用户入口（仅头像，点击展开个人中心/修改密码/退出） */
.vtab-user {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.vtab-user :deep(.user-info) {
  padding: 0;
  gap: 0;
  height: 26px;
  cursor: pointer;
}

.vtab-user :deep(.user-name),
.vtab-user :deep(.user-info > .el-icon) {
  display: none;
}

.vtab-user :deep(.user-avatar) {
  width: 26px;
  height: 26px;
}

.vtab-act:hover {
  background: rgba(255, 255, 255, 0.12);
  color: var(--el-color-primary);
}

html.sidebar-light .vtab-act {
  color: #606266;
}

html.sidebar-light .vtab-act:hover {
  background: rgba(64, 158, 255, 0.12);
}

/* ===== 全部菜单浮层 ===== */
.vtab-menu-mask {
  position: fixed;
  inset: 0;
  z-index: 2600;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(14, 16, 20, 0.45);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.vtab-menu-panel {
  width: min(560px, 92vw);
  max-height: 78vh;
  display: flex;
  flex-direction: column;
  background: var(--go-card-bg, #ffffff);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 12px;
  box-shadow: 0 24px 64px rgba(0, 21, 41, 0.32);
  overflow: hidden;
}

.vtab-menu-head {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--go-border-color, #e4e7ed);
  color: #909399;
}

.vtab-menu-search {
  flex: 1;
  min-width: 0;
  height: 28px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--go-navbar-text, #303133);
}

.vtab-menu-search::placeholder {
  color: #a8abb2;
}

.vtab-menu-close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: #909399;
  cursor: pointer;
}

.vtab-menu-close:hover {
  background: rgba(0, 0, 0, 0.06);
  color: var(--el-color-primary);
}

.vtab-menu-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 10px 14px;
}

.vtab-menu-group + .vtab-menu-group {
  margin-top: 6px;
}

.vtab-menu-group-title {
  padding: 8px 6px 6px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: #909399;
}

.vtab-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 8px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--go-navbar-text, #303133);
  cursor: pointer;
}

.vtab-menu-item:hover {
  background: rgba(64, 158, 255, 0.1);
  color: var(--el-color-primary);
}

.vtab-menu-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vtab-menu-parent {
  flex-shrink: 0;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 11px;
  color: #a8abb2;
}

.vtab-menu-empty {
  padding: 40px 0;
  text-align: center;
  font-size: 13px;
  color: #909399;
}

.vtab-menu-enter-active,
.vtab-menu-leave-active {
  transition: opacity 0.2s ease;
}

.vtab-menu-enter-active .vtab-menu-panel,
.vtab-menu-leave-active .vtab-menu-panel {
  transition: transform 0.24s cubic-bezier(0.22, 1, 0.36, 1);
}

.vtab-menu-enter-from,
.vtab-menu-leave-to {
  opacity: 0;
}

.vtab-menu-enter-from .vtab-menu-panel,
.vtab-menu-leave-to .vtab-menu-panel {
  transform: scale(0.94);
}

/* ===== 页签右键菜单 ===== */
.vtab-ctx {
  position: fixed;
  z-index: 4000;
  min-width: 150px;
  padding: 5px;
  border-radius: 8px;
  background: var(--el-bg-color-overlay, #fff);
  border: 1px solid var(--el-border-color-light, rgba(0, 0, 0, 0.1));
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
  user-select: none;
}

.vtab-ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--el-text-color-primary, #303133);
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}

.vtab-ctx-item:hover {
  background: var(--go-primary-soft, rgba(64, 158, 255, 0.12));
  color: var(--go-primary, #409eff);
}

.vtab-ctx-sep {
  height: 1px;
  margin: 4px 6px;
  background: var(--el-border-color-lighter, rgba(0, 0, 0, 0.06));
}

.vtab-ctx-fade-enter-active,
.vtab-ctx-fade-leave-active {
  transition: opacity 0.14s ease, transform 0.14s ease;
}

.vtab-ctx-fade-enter-from,
.vtab-ctx-fade-leave-to {
  opacity: 0;
  transform: scale(0.96);
}
</style>
