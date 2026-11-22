<template>
  <!-- Dock 坞布局：底部中央悬浮应用坞（macOS 风格），hover 放大 + 名称提示，点击打开；移开自动沉底 -->
  <div
    class="dock-shell"
    :class="{ 'dock-shell-hidden': hidden, 'dock-shell-lp': launchpadOpen }"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <div class="dock-peek" @mouseenter="onEnter"></div>
    <div class="dock-bar" @click.stop>
      <div
        v-for="it in dockItems"
        :key="it.key"
        class="dock-item"
        :class="{ 'dock-item-open': isDockActive(it) }"
        @mouseenter="hoverKey = it.key"
        @mouseleave="hoverKey = null"
        @click="openItem(it)"
      >
        <div class="dock-icon" :style="iconStyle(it)">
          <el-icon :size="hoverKey === it.key ? 30 : 24"><component :is="it.icon" /></el-icon>
        </div>
        <!-- macOS 运行指示点：该顶层菜单组有页签打开（首页常驻常亮） -->
        <span v-if="itHasOpen(it)" class="dock-dot"></span>
        <span class="dock-tip" :class="{ show: hoverKey === it.key }">{{ it.label }}</span>
      </div>
      <div
        class="dock-item"
        :class="{ 'dock-item-open': launchpadOpen }"
        @mouseenter="hoverKey = 'launchpad'"
        @mouseleave="hoverKey = null"
        @click="toggleLaunchpad"
      >
        <div class="dock-icon dock-icon-launchpad">
          <el-icon :size="hoverKey === 'launchpad' ? 30 : 24"><Grid /></el-icon>
        </div>
        <span class="dock-tip" :class="{ show: hoverKey === 'launchpad' }">{{ t('layout.launchpad') }}</span>
      </div>
      <div class="dock-sep"></div>
      <div class="dock-item" @mouseenter="hoverKey = 'settings'" @mouseleave="hoverKey = null" @click="openSettings">
        <div class="dock-icon dock-icon-plain">
          <el-icon :size="hoverKey === 'settings' ? 30 : 24"><Setting /></el-icon>
        </div>
        <span class="dock-tip" :class="{ show: hoverKey === 'settings' }">{{ t('layout.settings') }}</span>
      </div>
      <!-- 账号入口：个人中心 / 修改密码 / 退出登录（dock 布局没有顶栏，需在此补齐） -->
      <div
        class="dock-item dock-item-account"
        @mouseenter="hoverKey = 'account'"
        @mouseleave="hoverKey = null"
      >
        <div class="dock-icon dock-icon-account">
          <UserDropdown />
        </div>
        <span class="dock-tip" :class="{ show: hoverKey === 'account' }">账号</span>
      </div>
    </div>
  </div>

  <!-- 全菜单（Launchpad）：半透明全屏遮罩 + 应用网格，点击图标打开、点击空白/ESC 关闭 -->
  <Teleport to="body">
    <Transition name="lp-fade">
      <div v-if="launchpadOpen" class="lp-overlay" @click.self="closeLaunchpad">
        <div class="lp-panel" @click.stop>
          <div class="lp-head">
            <div class="lp-search">
              <el-icon :size="15"><Search /></el-icon>
              <input
                ref="launchpadInput"
                v-model="launchpadKeyword"
                class="lp-search-input"
                :placeholder="t('layout.launchpadSearch')"
                @keydown="onSearchKeydown"
              />
              <button
                v-if="launchpadKeyword"
                class="lp-search-clear"
                :title="t('layout.launchpadClear')"
                @click="launchpadKeyword = ''"
              >
                <el-icon :size="14"><CircleClose /></el-icon>
              </button>
            </div>
            <span class="lp-count">{{ t('layout.launchpadCount', { n: launchpadOrderedApps.length }) }}</span>
            <button class="lp-close" :title="t('layout.launchpadClose')" @click="closeLaunchpad">
              <el-icon :size="16"><Close /></el-icon>
            </button>
          </div>

          <div
            v-if="launchpadOpen"
            class="lp-body"
            @wheel.prevent="onWheel"
            @touchstart="onTouchStart"
            @touchend="onTouchEnd"
          >
            <div
              class="lp-track"
              :style="{ transform: `translateX(-${pageIndex * 100}%)` }"
            >
              <div v-for="(page, pi) in launchpadPages" :key="pi" class="lp-page">
                <!-- 分类纵向排列并整体上下居中 -->
                <div class="lp-sections">
                  <div v-for="(sec, si) in page.sections" :key="si" class="lp-section">
                    <div class="lp-group-title">{{ sec.label }}</div>
                    <div
                      class="lp-grid"
                      :style="{ gridTemplateColumns: `repeat(${launchpadColumns}, minmax(0, 1fr))` }"
                    >
                      <div
                        v-for="app in sec.apps"
                        :key="app.key"
                        class="lp-app"
                        :class="{ 'lp-app-active': activeAppKey === app.key }"
                        :data-app-key="app.key"
                        :style="{ animationDelay: appDelay(app) }"
                        role="button"
                        tabindex="0"
                        @mouseenter="activeAppKey = app.key"
                        @focus="activeAppKey = app.key"
                        @click="openApp(app)"
                        @keydown.enter.prevent="openApp(app)"
                        @keydown.space.prevent="openApp(app)"
                      >
                        <div class="lp-app-icon" :style="launchpadIconStyle(app)">
                          <el-icon :size="34"><component :is="safeIcon(app.icon)" /></el-icon>
                        </div>
                        <span class="lp-app-name">{{ app.label }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="!launchpadOrderedApps.length" class="lp-empty">
              {{ t('layout.launchpadEmpty') }}
            </div>
          </div>

          <!-- 分页小圆点：菜单超出单页容量时出现，可点击或左右滑动/滚轮翻页 -->
          <div v-if="launchpadPages.length > 1" class="lp-dots">
            <button
              v-for="(_, i) in launchpadPages"
              :key="i"
              class="lp-dot"
              :class="{ 'lp-dot-active': i === pageIndex }"
              :title="t('layout.launchpadPage', { n: i + 1 })"
              @click="goPage(i)"
            />
          </div>

          <div class="lp-foot">{{ t('layout.launchpadHint') }}</div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { CircleClose, Close, Grid, Search, Setting } from '@element-plus/icons-vue';
import { useSettingsStore } from '@/store/settings';
import { useUserStore } from '@/store/user';
import { useTagsViewStore } from '@/store/tagsView';
import { useMenuTitle } from '@/i18n/translate';
import { safeIcon } from '@/utils/iconNames';
import { useI18n } from 'vue-i18n';
import UserDropdown from './UserDropdown.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const settings = useSettingsStore();
const userStore = useUserStore();
const tagsStore = useTagsViewStore();
const menuTitle = useMenuTitle();

const hoverKey = ref<string | null>(null);
const activePath = computed(() => route.path);

// 自动隐藏：鼠标离开底部 Dock 区域 1.2s 沉底，仅留底部提示条；移到屏幕底部边缘弹出
const hidden = ref(false);
let hideTimer: any = null;

function onEnter() {
  if (hideTimer) clearTimeout(hideTimer);
  hidden.value = false;
}

function onLeave() {
  if (hideTimer) clearTimeout(hideTimer);
  // 全菜单打开期间 Dock 保持可见
  if (launchpadOpen.value) return;
  hideTimer = setTimeout(() => {
    hidden.value = true;
    hoverKey.value = null;
  }, 1200);
}

// 全局 mousemove：底部 120px 条带视为 Dock 激活区
function onDocMove(e: MouseEvent) {
  if (launchpadOpen.value) return;
  const inZone = e.clientY > window.innerHeight - 120;
  if (inZone) {
    if (hideTimer) clearTimeout(hideTimer);
    hideTimer = null;
    if (hidden.value) hidden.value = false;
  } else if (!hideTimer) {
    hideTimer = setTimeout(() => {
      hidden.value = true;
      hoverKey.value = null;
      hideTimer = null;
    }, 1200);
  }
}

onMounted(() => {
  window.addEventListener('mousemove', onDocMove);
  window.addEventListener('keydown', onKeydown);
  window.addEventListener('resize', onResize);
});

onBeforeUnmount(() => {
  if (hideTimer) clearTimeout(hideTimer);
  window.removeEventListener('mousemove', onDocMove);
  window.removeEventListener('keydown', onKeydown);
  window.removeEventListener('resize', onResize);
});

// 图标色板（按索引取，形成多彩 Dock）
const PALETTE = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#9c27b0', '#00bcd4', '#ff9800', '#8bc34a'];

interface DockItem {
  key: string;
  label: string;
  icon: string;
  path?: string;
  menu?: any;
}

// Dock 项 = 回到首页 + 顶层菜单（目录打开其首个子页）
const dockItems = computed<DockItem[]>(() => {
  const list: DockItem[] = [
    { key: 'index', label: menuTitle(t('layout.goHome')), icon: 'HomeFilled', path: '/index' }
  ];
  (userStore.menus || []).forEach((m) => {
    if (m.menuType === 'F') return;
    list.push({ key: String(m.menuId), label: menuTitle(m.menuName), icon: m.icon || 'Folder', menu: m });
  });
  return list;
});

function iconStyle(it: DockItem) {
  const idx = dockItems.value.indexOf(it);
  const color = PALETTE[idx % PALETTE.length];
  return {
    background: `linear-gradient(150deg, ${color}cc, ${color}99)`,
    boxShadow: `0 4px 14px ${color}55`
  };
}

// 目录项按子树前缀高亮（子页面打开时父菜单也保持高亮），首页/叶子精确匹配
function isDockActive(it: DockItem) {
  if (!it.menu) return activePath.value === it.path;
  const base = '/' + String(it.menu.path || '').replace(/^\/+/, '');
  return activePath.value === base || activePath.value.startsWith(base + '/');
}

// 已打开页签所属顶层菜单组（按路径首段归类），用于 macOS 运行指示点
const openGroups = computed(() => {
  const set = new Set<string>();
  tagsStore.visitedViews.forEach((v) => {
    const seg = String(v.path || '').split('/').filter(Boolean)[0];
    if (seg) set.add(seg);
  });
  return set;
});

// 该 Dock 项是否有窗口打开：首页页签常驻常亮；目录项看其首段是否有页签
function itHasOpen(it: DockItem) {
  if (!it.menu) {
    return tagsStore.visitedViews.some((v) => v.path === '/index');
  }
  const seg = String(it.menu.path || '').split('/').filter(Boolean)[0];
  return !!seg && openGroups.value.has(seg);
}

// 顶层菜单第一个可访问叶子路径
function firstLeaf(menu): string {
  const kids = (menu.children || []).filter((c) => c.menuType !== 'F');
  if (!kids.length) return '/' + menu.path;
  for (const k of kids) {
    const sub = firstLeaf(k);
    if (sub) return '/' + menu.path + sub;
  }
  return '';
}

function openItem(it: DockItem) {
  if (launchpadOpen.value) closeLaunchpad();
  if (it.path) {
    router.push(it.path);
    return;
  }
  if (it.menu) {
    const p = firstLeaf(it.menu);
    if (p) router.push(p);
  }
}

// Dock 中的布局配置入口（全菜单打开时先关闭遮罩，避免抽屉被遮罩盖住）
function openSettings() {
  if (launchpadOpen.value) closeLaunchpad();
  settings.drawerVisible = true;
}

// ===== 全菜单（Launchpad） =====
const launchpadOpen = ref(false);
const launchpadKeyword = ref('');
const launchpadInput = ref<HTMLInputElement | null>(null);
// 键盘/悬停选中的应用（回车直接打开）
const activeAppKey = ref<string | null>(null);

interface LaunchpadApp {
  key: string;
  label: string;
  icon: string;
  path: string;
  group: string;
}

// 菜单链拼接完整路径（与路由一致：/一级/二级/叶子）
function chainPath(stack: any[]): string {
  return stack.map((m) => '/' + String(m.path || '').replace(/^\/+/, '')).join('');
}

// 拍平菜单树：目录作为分组，叶子作为可点击应用
const launchpadApps = computed<LaunchpadApp[]>(() => {
  const apps: LaunchpadApp[] = [
    { key: 'home', label: menuTitle(t('layout.goHome')), icon: 'HomeFilled', path: '/index', group: t('layout.launchpadCommon') }
  ];
  const walk = (items: any[], stack: any[], group: string) => {
    (items || []).forEach((m) => {
      if (m.menuType === 'F') return;
      const chain = stack.concat([m]);
      const kids = (m.children || []).filter((c: any) => c.menuType !== 'F');
      if (!kids.length) {
        apps.push({
          key: 'm' + m.menuId,
          label: menuTitle(m.menuName),
          icon: m.icon || 'Grid',
          path: chainPath(chain),
          group: stack.length === 0 ? t('layout.launchpadOther') : group
        });
        return;
      }
      walk(kids, chain, stack.length === 0 ? menuTitle(m.menuName) : group);
    });
  };
  walk(userStore.menus || [], [], '');
  return apps;
});

interface LaunchpadSection {
  label: string;
  apps: LaunchpadApp[];
}

interface LaunchpadPage {
  sections: LaunchpadSection[];
}

// 分类分组（搜索时归入「搜索结果」一组）
const launchpadSections = computed<LaunchpadSection[]>(() => {
  const kw = launchpadKeyword.value.trim().toLowerCase();
  if (kw) {
    const hits = launchpadApps.value.filter(
      (a) =>
        a.label.toLowerCase().includes(kw) ||
        a.group.toLowerCase().includes(kw) ||
        a.path.toLowerCase().includes(kw)
    );
    return hits.length ? [{ label: t('layout.launchpadSearchResult'), apps: hits }] : [];
  }
  const order: string[] = [];
  const map = new Map<string, LaunchpadApp[]>();
  launchpadApps.value.forEach((a) => {
    if (!map.has(a.group)) {
      map.set(a.group, []);
      order.push(a.group);
    }
    map.get(a.group)!.push(a);
  });
  return order.map((g) => ({ label: g, apps: map.get(g)! }));
});

// 按分类顺序拍平（供键盘选中与回车打开）
const launchpadOrderedApps = computed<LaunchpadApp[]>(() =>
  launchpadSections.value.flatMap((s) => s.apps)
);

// ===== 分页（macOS Launchpad 风格：单页网格 + 圆点翻页，不使用滚动条） =====
// 视口尺寸响应式：窗口缩放时重新计算单页容量
const viewport = ref({ w: window.innerWidth, h: window.innerHeight });

function onResize() {
  viewport.value = { w: window.innerWidth, h: window.innerHeight };
}

// 列数：85vw 面板按单元格宽度均分
const launchpadColumns = computed(() => {
  const cell = viewport.value.w <= 640 ? 88 : 112;
  const gap = viewport.value.w <= 640 ? 10 : 16;
  const bodyW = Math.max(160, viewport.value.w * 0.85 - 8);
  return Math.max(1, Math.floor((bodyW + gap) / (cell + gap)));
});

// 单页可用高度（视口 - 上下留白 - 头部 - 圆点 - 底部提示）
const launchpadBodyHeight = computed(() =>
  Math.max(180, viewport.value.h * 0.95 - 124 - 56 - 34 - 30)
);

// 网格行距与分类标题高度（窄屏对应缩小）
const lpRowPitch = computed(() => (viewport.value.w <= 640 ? 106 : 130));
const lpHeaderHeight = computed(() => (viewport.value.w <= 640 ? 30 : 36));

// 按分类顺序装页：分类标题 + 该组图标整体放入，放不下则续页（保证不出现滚动条）
const launchpadPages = computed<LaunchpadPage[]>(() => {
  const cols = launchpadColumns.value;
  const bodyH = launchpadBodyHeight.value;
  const pitch = lpRowPitch.value;
  const headerH = lpHeaderHeight.value;
  const pages: LaunchpadPage[] = [];
  let current: LaunchpadPage = { sections: [] };
  let used = 0;

  const flush = () => {
    if (current.sections.length) {
      pages.push(current);
      current = { sections: [] };
      used = 0;
    }
  };

  launchpadSections.value.forEach((sec) => {
    let offset = 0;
    while (offset < sec.apps.length) {
      // 剩余空间连一行都放不下时先换页
      if (current.sections.length && bodyH - used - headerH < pitch) {
        flush();
        continue;
      }
      const maxRows = Math.max(1, Math.floor((bodyH - used - headerH) / pitch));
      const take = Math.min(sec.apps.length - offset, maxRows * cols);
      current.sections.push({ label: sec.label, apps: sec.apps.slice(offset, offset + take) });
      used += headerH + Math.ceil(take / cols) * pitch;
      offset += take;
      if (offset < sec.apps.length) flush();
    }
  });
  if (current.sections.length) pages.push(current);
  return pages.length ? pages : [{ sections: [] }];
});

const pageIndex = ref(0);

function goPage(i: number) {
  pageIndex.value = Math.max(0, Math.min(i, launchpadPages.value.length - 1));
  const first = launchpadPages.value[pageIndex.value]?.sections[0]?.apps[0];
  activeAppKey.value = first?.key ?? null;
}

// 翻页手势：触控左右滑动、触控板/滚轮滑动（带 600ms 节流，防止惯性连续翻页）
let touchStartX = 0;
let wheelLockedAt = 0;

function onTouchStart(e: TouchEvent) {
  touchStartX = e.touches[0]?.clientX ?? 0;
}

function onTouchEnd(e: TouchEvent) {
  const dx = (e.changedTouches[0]?.clientX ?? 0) - touchStartX;
  if (Math.abs(dx) > 50) goPage(pageIndex.value + (dx < 0 ? 1 : -1));
}

function onWheel(e: WheelEvent) {
  const now = Date.now();
  if (now - wheelLockedAt < 600) return;
  const delta = Math.abs(e.deltaX) > Math.abs(e.deltaY) ? e.deltaX : e.deltaY;
  if (Math.abs(delta) < 16) return;
  wheelLockedAt = now;
  goPage(pageIndex.value + (delta > 0 ? 1 : -1));
}

function launchpadIconStyle(app: LaunchpadApp) {
  const idx = Math.max(0, launchpadApps.value.findIndex((x) => x.key === app.key));
  const color = PALETTE[idx % PALETTE.length];
  return {
    background: `linear-gradient(150deg, ${color}dd, ${color}99)`,
    boxShadow: `0 10px 24px ${color}44`
  };
}

// 图标依次入场（最多错开 12 项，避免大菜单等待过久）
function appDelay(app: LaunchpadApp) {
  const idx = Math.max(0, launchpadApps.value.findIndex((x) => x.key === app.key));
  return `${Math.min(idx, 12) * 18}ms`;
}

function resetActiveApp() {
  pageIndex.value = 0;
  activeAppKey.value = launchpadOrderedApps.value[0]?.key ?? null;
}

// 键盘上下移动选中项（按整列跨行），并自动切换到选中项所在页
function moveActive(step: number) {
  const list = launchpadOrderedApps.value;
  if (!list.length) return;
  const cur = list.findIndex((a) => a.key === activeAppKey.value);
  const next = Math.max(0, Math.min(cur < 0 ? 0 : cur + step, list.length - 1));
  activeAppKey.value = list[next].key;
  const page = launchpadPages.value.findIndex((p) =>
    p.sections.some((s) => s.apps.some((a) => a.key === list[next].key))
  );
  pageIndex.value = page < 0 ? 0 : page;
}

function onSearchKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.stopPropagation();
    closeLaunchpad();
    return;
  }
  if (e.key === 'Enter') {
    e.preventDefault();
    const target =
      launchpadOrderedApps.value.find((a) => a.key === activeAppKey.value) ||
      launchpadOrderedApps.value[0];
    if (target) openApp(target);
    return;
  }
  if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
    e.preventDefault();
    if (e.key === 'ArrowRight') goPage(pageIndex.value + 1);
    else moveActive(launchpadColumns.value);
    return;
  }
  if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
    e.preventDefault();
    if (e.key === 'ArrowLeft') goPage(pageIndex.value - 1);
    else moveActive(-launchpadColumns.value);
  }
}

// 搜索词变化后回到第一页并选中第一个结果
watch(launchpadKeyword, () => {
  resetActiveApp();
});

// 结果数变化时保证页码不越界
watch(
  () => launchpadPages.value.length,
  (len) => {
    if (pageIndex.value > len - 1) pageIndex.value = Math.max(0, len - 1);
  }
);

function toggleLaunchpad() {
  if (launchpadOpen.value) {
    closeLaunchpad();
    return;
  }
  launchpadKeyword.value = '';
  launchpadOpen.value = true;
  hidden.value = false;
  hoverKey.value = null;
  resetActiveApp();
  nextTick(() => launchpadInput.value?.focus());
}

function closeLaunchpad() {
  launchpadOpen.value = false;
  launchpadKeyword.value = '';
  activeAppKey.value = null;
  pageIndex.value = 0;
}

function openApp(app: LaunchpadApp) {
  closeLaunchpad();
  if (app.path) router.push(app.path);
}

// ESC 关闭全菜单（macOS 交互习惯）
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && launchpadOpen.value) closeLaunchpad();
}
</script>

<style scoped>
.dock-shell {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  z-index: 1200;
  pointer-events: none;
}

.dock-shell .dock-bar,
.dock-shell .dock-peek {
  pointer-events: auto;
}

/* 全菜单打开时 Dock 浮在遮罩之上：可再次点击入口关闭，其余 Dock 应用也可直接打开 */
.dock-shell-lp {
  z-index: 2700;
}

.dock-peek {
  width: 56px;
  height: 5px;
  border-radius: 3px;
  margin-bottom: 9px;
  background: rgba(255, 255, 255, 0.4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.dock-shell-hidden .dock-peek {
  opacity: 1;
}

.dock-bar {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 18px;
  background: rgba(30, 34, 42, 0.82);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: 0 10px 32px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(16px);
  transform: translateY(0);
  transition: transform 0.35s cubic-bezier(0.33, 1, 0.68, 1), opacity 0.3s ease;
}

.dock-shell-hidden .dock-bar {
  transform: translateY(calc(100% + 26px));
  opacity: 0;
}

.dock-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.18s ease;
}

.dock-item:hover {
  transform: translateY(-10px) scale(1.18);
}

.dock-icon {
  width: 52px;
  height: 52px;
  border-radius: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  transition: width 0.18s ease, height 0.18s ease, border-radius 0.18s ease;
}

.dock-item:hover .dock-icon {
  width: 60px;
  height: 60px;
  border-radius: 15px;
}

.dock-icon-plain {
  background: linear-gradient(150deg, #606a78cc, #3d4652cc);
}

/* 账号入口：只展示头像，点击弹出个人中心/改密/退出菜单 */
.dock-icon-account {
  background: linear-gradient(150deg, #4b5563cc, #1f2937cc);
  padding: 0;
}

.dock-item-account :deep(.user-info) {
  gap: 0;
}

.dock-item-account :deep(.user-name),
.dock-item-account :deep(.user-info > .el-icon) {
  display: none;
}

/* 全菜单入口：Launchpad 风格渐变 */
.dock-icon-launchpad {
  background: linear-gradient(150deg, #7c3aeddd, #4f46e5aa);
}

.dock-item-open .dock-icon {
  outline: 2px solid rgba(255, 255, 255, 0.75);
  outline-offset: 2px;
}

/* macOS 运行指示点：图标下方白点（指示该组有窗口打开） */
.dock-dot {
  position: absolute;
  bottom: -7px;
  left: 50%;
  transform: translateX(-50%);
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 0 6px rgba(255, 255, 255, 0.6);
}

.dock-tip {
  position: absolute;
  bottom: calc(100% + 12px);
  left: 50%;
  transform: translateX(-50%) translateY(4px);
  padding: 4px 10px;
  border-radius: 6px;
  background: rgba(20, 22, 26, 0.95);
  color: #e8ebf2;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s ease, transform 0.15s ease;
  border: 1px solid rgba(255, 255, 255, 0.14);
}

.dock-tip.show {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

.dock-sep {
  align-self: stretch;
  width: 1px;
  margin: 6px 2px;
  background: rgba(255, 255, 255, 0.2);
}

/* ===== 全菜单（Launchpad） ===== */
.lp-overlay {
  position: fixed;
  inset: 0;
  z-index: 2600;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5vh 4vw 124px;
  background: rgba(14, 16, 20, 0.45);
  backdrop-filter: blur(28px) saturate(150%);
  -webkit-backdrop-filter: blur(28px) saturate(150%);
}

.lp-panel {
  width: 85vw;
  height: 100%;
  display: flex;
  flex-direction: column;
  color: #fff;
}

.lp-head {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
  margin-bottom: 8px;
}

.lp-search {
  /* 默认收窄，聚焦/输入时展开到 1.5 倍 */
  width: 240px;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  color: rgba(255, 255, 255, 0.7);
  transition: width 0.25s ease, background 0.2s ease, border-color 0.2s ease;
}

.lp-search:focus-within {
  width: 360px;
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.3);
}

.lp-search-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: #fff;
}

.lp-search-input::placeholder {
  color: rgba(255, 255, 255, 0.45);
}

.lp-search-clear {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.lp-search-clear:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.16);
}

.lp-count {
  flex-shrink: 0;
  margin-left: auto;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.55);
  white-space: nowrap;
}

.lp-close {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  cursor: pointer;
  transition: background 0.18s, transform 0.18s;
}

.lp-close:hover {
  background: rgba(255, 255, 255, 0.24);
  transform: scale(1.06);
}

.lp-body {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

/* 横向分页轨道：整页平移，不出现滚动条 */
.lp-track {
  display: flex;
  width: 100%;
  height: 100%;
  transition: transform 0.38s cubic-bezier(0.22, 1, 0.36, 1);
}

.lp-page {
  flex: 0 0 100%;
  height: 100%;
  padding: 0 4px;
  box-sizing: border-box;
}

/* 分类纵向排列，整体在页面内上下居中 */
.lp-sections {
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.lp-section + .lp-section {
  margin-top: 14px;
}

.lp-group-title {
  margin: 0 0 14px;
  padding-left: 2px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 1px;
  color: rgba(255, 255, 255, 0.62);
}

.lp-grid {
  display: grid;
  gap: 26px 16px;
  justify-items: center;
  align-content: start;
}

.lp-app {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
  outline: none;
  animation: lp-pop 0.34s cubic-bezier(0.22, 1, 0.36, 1) both;
}

/* 键盘/悬停选中：白色光圈提示当前项，回车直接打开 */
.lp-app-active .lp-app-icon {
  outline: 3px solid rgba(255, 255, 255, 0.55);
  outline-offset: 3px;
  transform: translateY(-4px) scale(1.1);
}

.lp-app-icon {
  width: 76px;
  height: 76px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 18px;
  color: #fff;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.2s;
}

.lp-app:hover .lp-app-icon {
  transform: translateY(-4px) scale(1.1);
}

.lp-app-name {
  max-width: 112px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.86);
}

.lp-app:hover .lp-app-name,
.lp-app-active .lp-app-name {
  color: #fff;
}

.lp-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.55);
}

/* 分页圆点（macOS Launchpad 底部指示器） */
.lp-dots {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 26px;
}

.lp-dot {
  width: 8px;
  height: 8px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.32);
  cursor: pointer;
  transition: background 0.2s, transform 0.2s, box-shadow 0.2s;
}

.lp-dot:hover {
  background: rgba(255, 255, 255, 0.55);
}

.lp-dot-active {
  background: #fff;
  transform: scale(1.15);
  box-shadow: 0 0 8px rgba(255, 255, 255, 0.5);
}

.lp-foot {
  flex-shrink: 0;
  padding: 10px 0 2px;
  text-align: center;
  font-size: 12px;
  letter-spacing: 0.4px;
  color: rgba(255, 255, 255, 0.42);
}

@keyframes lp-pop {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.92);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

/* 窄屏：缩小图标与间距，为 Dock 留出空间 */
@media (max-width: 640px) {
  .lp-overlay {
    padding: 4vh 4vw 112px;
  }

  .lp-grid {
    gap: 20px 10px;
  }

  .lp-app-icon {
    width: 60px;
    height: 60px;
    border-radius: 15px;
  }

  .lp-app-name {
    max-width: 88px;
    font-size: 12px;
  }

  .lp-count,
  .lp-foot {
    display: none;
  }

  .lp-search {
    width: 170px;
  }

  .lp-search:focus-within {
    width: 255px;
  }

  .lp-group-title {
    margin-bottom: 10px;
    font-size: 12px;
  }
}

/* 打开/关闭：遮罩淡入 + 面板缩放（macOS Launchpad 展开感） */
.lp-fade-enter-active,
.lp-fade-leave-active {
  transition: opacity 0.26s ease;
}

.lp-fade-enter-active .lp-panel {
  transition: transform 0.32s cubic-bezier(0.22, 1, 0.36, 1);
}

.lp-fade-leave-active .lp-panel {
  transition: transform 0.2s ease;
}

.lp-fade-enter-from,
.lp-fade-leave-to {
  opacity: 0;
}

.lp-fade-enter-from .lp-panel {
  transform: scale(0.88);
}

.lp-fade-leave-to .lp-panel {
  transform: scale(0.94);
}
</style>
