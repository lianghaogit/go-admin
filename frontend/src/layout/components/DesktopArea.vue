<template>
  <div class="desktop-area">
    <!-- 全屏桌面画布：窗口层叠区（右键弹出快捷菜单） -->
    <div
      class="desktop-canvas"
      :style="canvasStyle"
      @contextmenu.prevent="openCtx"
      @click="closeCtx"
    >
      <!-- 桌面图标墙：顶层应用图标，单击选中 / 双击打开 -->
      <div class="desktop-icons" @click.stop>
        <div
          v-for="(it, i) in iconWall"
          :key="it.key"
          class="desktop-icon"
          :class="{ selected: selectedIcon === i }"
          @click="selectedIcon = i"
          @dblclick="openIcon(it)"
        >
          <div class="di-box">
            <el-icon :size="26"><component :is="it.icon" /></el-icon>
          </div>
          <span class="di-name">{{ it.label }}</span>
        </div>
      </div>
      <DesktopWindow
        v-for="w in windows"
        :key="w.path"
        :tag="w"
        :active="w.path === activePath"
        :minimized="getState(w.path).minimized"
        :maximized="getState(w.path).maximized"
        :pos="getState(w.path)"
        :z="getState(w.path).z"
        @activate="activateWindow(w)"
        @close="closeWindow(w)"
        @minimize="setMinimized(w, true)"
        @maximize="toggleMaximized(w)"
        @move="moveWindow(w, $event)"
        @resize="resizeWindow(w, $event)"
      />

      <!-- 桌面右键快捷菜单 -->
      <div
        v-if="ctx.visible"
        class="desktop-ctx"
        :style="{ left: ctx.x + 'px', top: ctx.y + 'px' }"
        @click.stop
      >
        <div class="ctx-title">{{ t('layout.ctxTitle') }}</div>
        <div class="ctx-item" @click="ctxAction('settings')">
          <el-icon :size="14"><Setting /></el-icon>{{ t('layout.settings') }}
        </div>
        <div class="ctx-item" @click="ctxAction('home')">
          <el-icon :size="14"><HomeFilled /></el-icon>{{ t('layout.goHome') }}
        </div>
        <div class="ctx-item" @click="ctxAction('minAll')">
          <el-icon :size="14"><Minus /></el-icon>{{ t('layout.minimizeAll') }}
        </div>
        <div class="ctx-item" @click="ctxAction('arrange')">
          <el-icon :size="14"><Rank /></el-icon>{{ t('layout.rearrange') }}
        </div>
        <div class="ctx-sep"></div>
        <div class="ctx-item" @click="ctxAction('bg')">
          <el-icon :size="14"><Picture /></el-icon>{{ t('layout.changeBg') }}
        </div>
        <div class="ctx-item" @click="ctxAction('refresh')">
          <el-icon :size="14"><Refresh /></el-icon>{{ t('layout.refresh') }}
        </div>
      </div>

      <!-- 更换桌面背景面板 -->
      <div v-if="bgPanel.visible" class="desktop-bg-panel" @click.stop>
        <div class="bg-title">{{ t('layout.changeBg') }}</div>
        <div class="bg-grid">
          <div
            v-for="b in presets"
            :key="b.id"
            class="bg-cell"
            :class="{ selected: settings.desktopBg === b.value }"
            :style="{ background: b.preview }"
            @click="applyBg(b.value)"
          >
            <span class="bg-name">{{ b.name }}</span>
          </div>
        </div>
        <div class="bg-upload">
          <el-button size="small" type="primary" plain :loading="uploading" @click="pickFile">
            上传图片
          </el-button>
          <input
            ref="fileInput"
            type="file"
            accept="image/*"
            style="display: none"
            @change="handleFile"
          />
        </div>
      </div>
    </div>

    <!-- 开始菜单（左下角弹出，级联导航） -->
    <StartMenu v-if="startVisible" :menus="topMenus" @open="openFromStart" />

    <!-- Windows 任务栏：开始按钮 + 窗口任务项 + 右下角系统托盘（时间/设置/用户） -->
    <div class="desktop-taskbar">
      <div class="tb-left">
        <button
          class="start-btn"
          :class="{ active: startVisible }"
          :title="t('layout.start')"
          @click="toggleStart"
        >
          <el-icon :size="17" class="start-logo"><Grid /></el-icon>
        </button>
      </div>
      <div class="tb-windows">
        <template v-for="g in windowGroups" :key="g.key">
          <!-- 组内单窗口：直接显示窗口项 -->
          <div
            v-if="g.items.length === 1"
            class="taskbar-item"
            :class="{ active: g.items[0].path === activePath && !getState(g.items[0].path).minimized }"
            @click="onTaskbarClick(g.items[0])"
          >
            <el-icon :size="14"
              ><component :is="g.items[0].affix ? 'HomeFilled' : g.items[0].meta?.icon || 'Document'"
            /></el-icon>
            <span class="taskbar-title">{{ menuTitle(g.items[0].title) }}</span>
          </div>
          <!-- 组内多窗口：分组按钮 + 浮层窗口列表（预览/切换） -->
          <div
            v-else
            class="taskbar-item taskbar-group"
            :class="{ 'group-open': groupOpen === g.key }"
            @click="toggleGroup(g)"
            @mouseenter="groupHover = g.key"
            @mouseleave="groupHover = null"
          >
            <el-icon :size="14"
              ><component :is="g.items[0].meta?.icon || 'Document'"
            /></el-icon>
            <span class="taskbar-title">{{ g.label }}</span>
            <span class="tb-count">{{ g.items.length }}</span>
            <div
              v-if="groupOpen === g.key || (groupHover === g.key && g.items.length <= 4)"
              class="tb-flyout"
              @click.stop
            >
              <div
                v-for="w in g.items"
                :key="w.path"
                class="tb-fly-item"
                :class="{ active: w.path === activePath && !getState(w.path).minimized }"
                @click="activateWindow(w)"
              >
                <el-icon :size="14"
                  ><component :is="w.affix ? 'HomeFilled' : w.meta?.icon || 'Document'"
                /></el-icon>
                <span class="tb-fly-title">{{ menuTitle(w.title) }}</span>
                <span v-if="w.path === activePath && !getState(w.path).minimized" class="tb-fly-dot"></span>
              </div>
            </div>
          </div>
        </template>
      </div>
      <div class="tb-tray">
        <span class="tb-time">{{ now }}</span>
        <div class="tray-icon" :title="t('layout.settings')" @click="settings.drawerVisible = true">
          <el-icon :size="17"><Setting /></el-icon>
        </div>
        <UserDropdown />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Setting, Grid, Minus, Rank, Picture, Refresh, HomeFilled } from '@element-plus/icons-vue';
import DesktopWindow from './DesktopWindow.vue';
import StartMenu from './StartMenu.vue';
import UserDropdown from './UserDropdown.vue';
import { useTagsViewStore } from '@/store/tagsView';
import { useSettingsStore } from '@/store/settings';
import { useUserStore } from '@/store/user';
import { uploadFile } from '@/api/tool/apiDoc';
import { ElMessage } from 'element-plus';
import { useMenuTitle } from '@/i18n/translate';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const tagsStore = useTagsViewStore();
const settings = useSettingsStore();
const userStore = useUserStore();
const menuTitle = useMenuTitle();

// 开始菜单数据 = 动态菜单顶层目录（登录后）
const topMenus = computed(() => userStore.menus || []);
const startVisible = ref(false);
const now = ref('');
let clockTimer = null;

// 窗口列表 = 页签（打开过的页面）
const windows = computed(() => tagsStore.visitedViews);
const activePath = computed(() => route.path);

// 窗口状态（位置 / 尺寸 / 层级 / 最小化 / 最大化），按 path 管理
const states = reactive({});
let maxZ = 100;

// 桌面背景：settings.desktopBg（''=默认渐变；'url(...)'=图片；其他=CSS background）
const canvasStyle = computed(() => {
  if (settings.desktopBg && settings.desktopBg.startsWith('url(')) {
    return {
      backgroundImage: settings.desktopBg,
      backgroundSize: 'cover',
      backgroundPosition: 'center'
    };
  }
  if (settings.desktopBg) {
    return { background: settings.desktopBg };
  }
  return {};
});

// 桌面右键菜单状态
const ctx = reactive({ visible: false, x: 0, y: 0 });
// 更换背景面板
const bgPanel = reactive({ visible: false });
const fileInput = ref();
const uploading = ref(false);
const presets = [
  { id: 'default', name: '默认', value: '', preview: 'linear-gradient(160deg,#e8edf3,#d7e0ec)' },
  {
    id: 'deepblue',
    name: '深蓝',
    value: 'linear-gradient(160deg,#1e3a5f,#0f2036)',
    preview: 'linear-gradient(160deg,#1e3a5f,#0f2036)'
  },
  {
    id: 'purple',
    name: '暮紫',
    value: 'linear-gradient(160deg,#4a3f7a,#241f44)',
    preview: 'linear-gradient(160deg,#4a3f7a,#241f44)'
  },
  {
    id: 'forest',
    name: '森绿',
    value: 'linear-gradient(160deg,#2f5d50,#14281f)',
    preview: 'linear-gradient(160deg,#2f5d50,#14281f)'
  },
  {
    id: 'sunset',
    name: '落日',
    value: 'linear-gradient(160deg,#b4543a,#5c1f2e)',
    preview: 'linear-gradient(160deg,#b4543a,#5c1f2e)'
  },
  {
    id: 'ink',
    name: '墨黑',
    value: 'linear-gradient(160deg,#23262b,#101114)',
    preview: 'linear-gradient(160deg,#23262b,#101114)'
  }
];

// 桌面窗口位置/尺寸持久化：刷新、切换布局后恢复用户上次排布的窗口（z 与最小化不持久化，按会话重建）
const DESKTOP_STORE_KEY = 'go-admin-desktop-windows';
function loadWindowStates() {
  try {
    const raw = JSON.parse(localStorage.getItem(DESKTOP_STORE_KEY) || '{}');
    return raw && typeof raw === 'object' ? raw : {};
  } catch {
    return {};
  }
}
function saveWindowStates() {
  try {
    const pick = {};
    Object.keys(states).forEach((p) => {
      const s = states[p];
      pick[p] = { x: s.x, y: s.y, w: s.w, h: s.h };
    });
    localStorage.setItem(DESKTOP_STORE_KEY, JSON.stringify(pick));
  } catch {
    // 存储失败（隐私模式等）静默忽略
  }
}

// 获取窗口状态（位置/尺寸/层级）
function getState(path) {
  if (!states[path]) {
    const idx = windows.value.findIndex((w) => w.path === path);
    const n = idx >= 0 ? idx : windows.value.length;
    const vw = window.innerWidth;
    const vh = window.innerHeight - 42;
    // 优先恢复持久化排布；新视口下越界则夹回画布内
    const saved = loadWindowStates()[path] || {};
    const maxW = Math.max(320, vw);
    const maxH = Math.max(220, vh);
    states[path] = {
      x: Math.max(0, Math.min(saved.x ?? 60 + (n % 8) * 30, Math.max(0, vw - 320))),
      y: Math.max(0, Math.min(saved.y ?? 30 + (n % 6) * 26, Math.max(0, vh - 220))),
      z: 100 + n,
      w: Math.max(320, Math.min(saved.w ?? Math.round(vw * 0.72), maxW)),
      h: Math.max(220, Math.min(saved.h ?? Math.round(vh * 0.68), maxH)),
      minimized: false,
      maximized: false
    };
  }
  return states[path];
}

// 激活窗口并置顶
function activateWindow(w) {
  const s = getState(w.path);
  s.minimized = false;
  s.z = ++maxZ;
  if (route.path !== w.path) {
    router.push(w.path);
  }
}

// 关闭窗口并移除页签
function closeWindow(w) {
  tagsStore.delView(w);
  delete states[w.path];
  saveWindowStates();
  if (route.path === w.path) {
    const views = tagsStore.visitedViews;
    const last = views[views.length - 1];
    router.push(last ? last.path : '/index');
  }
}

// 最小化窗口
function setMinimized(w, v) {
  getState(w.path).minimized = v;
}

// 切换最大化/还原
function toggleMaximized(w) {
  const s = getState(w.path);
  s.maximized = !s.maximized;
  saveWindowStates();
}

// 开始拖动移动窗口（Windows 风格半屏吸附：拖到画布左/右边缘吸附半屏，顶部边缘最大化）
function moveWindow(w, pos) {
  const s = getState(w.path);
  const canvas = document.querySelector('.desktop-canvas');
  const cr = canvas ? canvas.getBoundingClientRect() : null;
  const vw = cr ? cr.width : window.innerWidth;
  const vh = cr ? cr.height : Math.max(0, window.innerHeight - 42);
  const SNAP = 14;
  // 顶部边缘 → 最大化
  if (pos.y <= SNAP && !s.maximized) {
    s.maximized = true;
    saveWindowStates();
    return;
  }
  // 左边缘 → 左半屏
  if (pos.x <= SNAP && !s.maximized) {
    s.x = 0;
    s.y = 0;
    s.w = Math.round(vw / 2);
    s.h = vh;
    s.maximized = false;
    saveWindowStates();
    return;
  }
  // 右边缘 → 右半屏
  if (pos.x + s.w >= vw - SNAP && !s.maximized) {
    s.x = Math.round(vw / 2);
    s.y = 0;
    s.w = vw - s.x;
    s.h = vh;
    s.maximized = false;
    saveWindowStates();
    return;
  }
  s.x = pos.x;
  s.y = pos.y;
  saveWindowStates();
}

// 开始拖动调整窗口大小
function resizeWindow(w, size) {
  const s = getState(w.path);
  s.w = size.w;
  s.h = size.h;
  saveWindowStates();
}

// 任务栏单击（Windows 11 行为）：已激活 → 最小化；最小化/未激活 → 激活还原
function onTaskbarClick(w) {
  const s = getState(w.path);
  if (route.path === w.path && !s.minimized) {
    s.minimized = true;
  } else {
    activateWindow(w);
  }
}

// ===== 任务栏分组（按顶层菜单分组，多窗口合并为一个组按钮） =====
const groupOpen = ref<string | null>(null);
const groupHover = ref<string | null>(null);

// 顶层菜单名：按路径首段匹配动态菜单树
function groupLabel(seg: string) {
  if (seg === 'index') return t('layout.goHome');
  const top = (userStore.menus || []).find(
    (m) => (m.path || '') === seg || (m.path || '').startsWith(seg + '/')
  );
  return top ? menuTitle(top.menuName) : seg;
}

const windowGroups = computed(() => {
  const map: Record<string, Array<any>> = {};
  windows.value.forEach((w) => {
    const seg = (w.path || '').split('/')[1] || 'index';
    (map[seg] || (map[seg] = [])).push(w);
  });
  return Object.keys(map).map((k) => ({ key: k, items: map[k], label: groupLabel(k) }));
});

function toggleGroup(g) {
  groupOpen.value = groupOpen.value === g.key ? null : g.key;
}

// 切换窗口/组时关闭浮层
watch(activePath, () => {
  groupOpen.value = null;
  groupHover.value = null;
});

// ===== 桌面图标墙：顶层应用图标（单击选中 / 双击打开） =====
const selectedIcon = ref(-1);
const iconWall = computed(() => {
  const list = [{ key: 'index', label: menuTitle(t('layout.goHome')), icon: 'HomeFilled', menu: null }];
  (topMenus.value || []).forEach((m) => {
    if (m.menuType === 'F') return;
    list.push({ key: String(m.menuId), label: menuTitle(m.menuName), icon: m.icon || 'Folder', menu: m });
  });
  return list;
});

// 顶层菜单第一个可访问叶子路径（目录 → 首个子页面）
function firstLeaf(menu) {
  const kids = (menu.children || []).filter((c) => c.menuType !== 'F');
  if (!kids.length) return '/' + menu.path;
  for (const k of kids) {
    const sub = firstLeaf(k);
    if (sub) return '/' + menu.path + sub;
  }
  return '';
}

function openIcon(it) {
  if (!it.menu) {
    router.push('/index');
    return;
  }
  const p = firstLeaf(it.menu);
  if (p) router.push(p);
}

// ===== 桌面右键快捷菜单 =====
function openCtx(e) {
  const canvas = document.querySelector('.desktop-canvas');
  const cr = canvas ? canvas.getBoundingClientRect() : null;
  const mw = 170;
  const mh = 260;
  ctx.x = cr ? Math.min(e.clientX - cr.left, Math.max(0, cr.width - mw)) : 0;
  ctx.y = cr ? Math.min(e.clientY - cr.top, Math.max(0, cr.height - mh)) : 0;
  ctx.visible = true;
  bgPanel.visible = false;
}

// 关闭右键菜单
function closeCtx() {
  ctx.visible = false;
  bgPanel.visible = false;
}

// 执行右键菜单操作
function ctxAction(action) {
  ctx.visible = false;
  if (action === 'settings') {
    settings.drawerVisible = true;
  } else if (action === 'home') {
    router.push('/index');
  } else if (action === 'minAll') {
    for (const w of windows.value) {
      getState(w.path).minimized = true;
    }
  } else if (action === 'arrange') {
    // 重置所有窗口位置为默认层叠
    let n = 0;
    for (const w of windows.value) {
      const s = getState(w.path);
      s.x = 60 + (n % 8) * 30;
      s.y = 30 + (n % 6) * 26;
      s.minimized = false;
      n++;
    }
  } else if (action === 'bg') {
    bgPanel.visible = !bgPanel.visible;
  } else if (action === 'refresh') {
    // 刷新当前激活窗口（重新挂载当前页）
    const p = route.path;
    router.replace('/index').then(() => {
      setTimeout(() => router.replace(p), 60);
    });
  }
}

// 应用桌面背景
function applyBg(value) {
  settings.desktopBg = value;
  settings.save();
  ElMessage.success('桌面背景已更新');
}

// 触发文件选择
function pickFile() {
  fileInput.value?.click();
}

// 背景图片选择回调
async function handleFile(e) {
  const file = e.target.files && e.target.files[0];
  if (!file) return;
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件');
    return;
  }
  uploading.value = true;
  try {
    const res = await uploadFile(file);
    const url = res && (res.url || (res.data && res.data.url));
    if (url) {
      settings.desktopBg = 'url(' + url + ')';
      settings.save();
      ElMessage.success('桌面背景已更换');
      bgPanel.visible = false;
    } else {
      ElMessage.error('上传失败：未返回图片地址');
    }
  } catch (err) {
    ElMessage.error('上传失败，请重试');
  } finally {
    uploading.value = false;
    e.target.value = '';
  }
}

// 开始菜单
function toggleStart() {
  startVisible.value = !startVisible.value;
}

// 开始菜单选中菜单项：打开窗口并收起开始菜单
function openFromStart(path) {
  startVisible.value = false;
  router.push(path);
}

// 更新桌面时钟
function updateClock() {
  const d = new Date();
  now.value = String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0');
}

// 点击开始菜单外部关闭 / 右键菜单外部关闭 / Esc 关闭（Windows 行为）
function onDocClick(e) {
  if (startVisible.value && !e.target.closest('.sm-flyout') && !e.target.closest('.start-btn')) {
    startVisible.value = false;
  }
  if (
    (ctx.visible || bgPanel.visible) &&
    !e.target.closest('.desktop-ctx') &&
    !e.target.closest('.desktop-bg-panel')
  ) {
    ctx.visible = false;
    bgPanel.visible = false;
  }
}

// 桌面全局键盘事件（Alt+Tab 切换窗口等）
function onKeydown(e) {
  if (e.key === 'Escape') {
    if (startVisible.value) startVisible.value = false;
    if (ctx.visible || bgPanel.visible) {
      ctx.visible = false;
      bgPanel.visible = false;
    }
  }
}

// 路由变化：登记窗口并激活置顶
watch(
  () => route.path,
  (path) => {
    tagsStore.addView(route);
    const s = getState(path);
    s.minimized = false;
    s.z = ++maxZ;
  }
);

onMounted(() => {
  tagsStore.addView({ path: '/index', meta: { title: '首页' } });
  tagsStore.addView(route);
  getState(route.path).z = ++maxZ;
  updateClock();
  clockTimer = setInterval(updateClock, 30000);
  document.addEventListener('click', onDocClick);
  window.addEventListener('keydown', onKeydown);
});

onBeforeUnmount(() => {
  if (clockTimer) clearInterval(clockTimer);
  document.removeEventListener('click', onDocClick);
  window.removeEventListener('keydown', onKeydown);
});
</script>

<style scoped>
.desktop-area {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
}

.desktop-canvas {
  flex: 1;
  min-height: 0;
  position: relative;
  overflow: hidden;
  background: linear-gradient(
    160deg,
    var(--go-desktop-bg1, #e8edf3) 0%,
    var(--go-desktop-bg2, #d7e0ec) 100%
  );
}

/* ===== 桌面图标墙 ===== */
.desktop-icons {
  position: absolute;
  left: 16px;
  top: 56px;
  z-index: 50;
  display: flex;
  flex-direction: column;
  gap: 4px;
  pointer-events: auto;
}

.desktop-icon {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  width: 72px;
  padding: 8px 4px 6px;
  border-radius: 4px;
  cursor: default;
  user-select: none;
  border: 1px solid transparent;
}

.desktop-icon:hover {
  background: rgba(64, 158, 255, 0.14);
}

.desktop-icon.selected {
  background: rgba(64, 158, 255, 0.22);
  border-color: rgba(64, 158, 255, 0.55);
}

.di-box {
  width: 46px;
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: linear-gradient(150deg, rgba(255, 255, 255, 0.85), rgba(255, 255, 255, 0.55));
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 3px 8px rgba(31, 45, 61, 0.18);
  color: var(--el-color-primary);
}

html.dark .di-box {
  background: linear-gradient(150deg, rgba(64, 70, 82, 0.9), rgba(48, 53, 64, 0.8));
  border-color: rgba(255, 255, 255, 0.16);
  color: #79bbff;
}

.di-name {
  max-width: 68px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #303133;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.7);
}

html.dark .di-name {
  color: #e5e9f0;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.6);
}

/* ===== 桌面右键快捷菜单 ===== */
.desktop-ctx {
  position: absolute;
  width: 170px;
  padding: 6px;
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  box-shadow: 0 8px 24px rgba(0, 21, 41, 0.2);
  z-index: 2900;
}

html.dark .desktop-ctx {
  background: rgba(29, 30, 31, 0.97);
}

.ctx-title {
  font-size: 12px;
  color: #909399;
  padding: 4px 10px;
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 10px;
  border-radius: 2px;
  font-size: 13px;
  color: var(--go-navbar-text, #303133);
  cursor: pointer;
}

.ctx-item:hover {
  background: rgba(64, 158, 255, 0.12);
  color: var(--go-topnav-active, #409eff);
}

.ctx-item .el-icon {
  color: var(--go-topnav-active, #409eff);
}

.ctx-sep {
  height: 1px;
  margin: 4px 8px;
  background: var(--go-border-color, #e4e7ed);
}

/* ===== 更换桌面背景面板 ===== */
.desktop-bg-panel {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 420px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  box-shadow: 0 12px 36px rgba(0, 21, 41, 0.26);
  z-index: 2900;
}

html.dark .desktop-bg-panel {
  background: rgba(29, 30, 31, 0.97);
}

.bg-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--go-navbar-text, #303133);
  margin-bottom: 10px;
}

.bg-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.bg-cell {
  height: 64px;
  border-radius: 2px;
  position: relative;
  cursor: pointer;
  border: 2px solid transparent;
  overflow: hidden;
}

.bg-cell:hover {
  border-color: rgba(64, 158, 255, 0.5);
}

.bg-cell.selected {
  border-color: var(--el-color-primary);
}

.bg-name {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  font-size: 11px;
  color: #fff;
  text-align: center;
  padding: 2px 0;
  background: rgba(0, 0, 0, 0.35);
}

.bg-upload {
  margin-top: 12px;
  display: flex;
  justify-content: center;
}

/* ===== Windows 任务栏 ===== */
.desktop-taskbar {
  height: 42px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  background: rgba(24, 33, 47, 0.94);
  position: relative;
  z-index: 2500;
}

.tb-left {
  flex-shrink: 0;
}

.start-btn {
  width: 38px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 2px;
  background: transparent;
  cursor: pointer;
}

.start-btn:hover,
.start-btn.active {
  background: rgba(255, 255, 255, 0.14);
}

.start-logo {
  color: #4aa3ff;
}

.tb-windows {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  scrollbar-width: none;
}

.tb-windows::-webkit-scrollbar {
  display: none;
}

.taskbar-item {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  border-radius: 2px;
  color: #cfd3dc;
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.08);
  transition: background 0.15s;
}

.taskbar-item:hover {
  background: rgba(255, 255, 255, 0.16);
}

.taskbar-item.active {
  background: var(--el-color-primary);
  color: #fff;
}

.taskbar-title {
  white-space: nowrap;
}

/* 任务栏分组 */
.taskbar-group {
  position: relative;
}

.tb-count {
  flex-shrink: 0;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  text-align: center;
  padding: 0 4px;
  border-radius: 8px;
  font-size: 10px;
  background: rgba(255, 255, 255, 0.22);
  color: #e5e9f0;
}

.taskbar-group.group-open {
  background: rgba(255, 255, 255, 0.22);
}

/* 分组浮层：窗口列表预览 */
.tb-flyout {
  position: absolute;
  bottom: 44px;
  left: 0;
  min-width: 180px;
  max-height: 300px;
  overflow-y: auto;
  padding: 6px;
  border-radius: 2px;
  background: rgba(32, 36, 44, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
  z-index: 3200;
  backdrop-filter: blur(10px);
}

.tb-fly-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 10px;
  border-radius: 2px;
  color: #cfd3dc;
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
}

.tb-fly-item:hover {
  background: rgba(64, 158, 255, 0.25);
}

.tb-fly-item.active {
  color: #fff;
  background: rgba(64, 158, 255, 0.35);
}

.tb-fly-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tb-fly-dot {
  flex-shrink: 0;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #67c23a;
}

.tb-tray {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.tb-time {
  color: #e5e9f0;
  font-size: 12px;
  margin-right: 4px;
  font-variant-numeric: tabular-nums;
}

.tray-icon {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  color: #e5e9f0;
  cursor: pointer;
}

.tray-icon:hover {
  background: rgba(255, 255, 255, 0.14);
}

/* 深色任务栏托盘内用户区适配（覆盖顶栏浅色文字） */
.desktop-area :deep(.user-info) {
  color: #e5e9f0;
}

.desktop-area :deep(.user-name) {
  color: #e5e9f0;
}

.desktop-area :deep(.user-avatar) {
  border: 1px solid rgba(255, 255, 255, 0.25);
}
</style>
