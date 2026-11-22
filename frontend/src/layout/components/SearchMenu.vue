<template>
  <!-- 全局命令面板（导航栏放大镜 / Ctrl+K / ⌘K）：菜单直达 + 快捷命令 -->
  <el-dialog
    v-model="visible"
    :title="t('layout.searchMenu')"
    width="520px"
    append-to-body
    :close-on-click-modal="true"
    @opened="focusInput"
  >
    <el-input
      ref="inputRef"
      v-model="keyword"
      :placeholder="t('layout.searchMenuPlaceholder')"
      clearable
      :prefix-icon="Search"
      @keyup.enter="handleEnter"
      @keyup.down.prevent="moveActive(1)"
      @keyup.up.prevent="moveActive(-1)"
    />
    <!-- 最近使用（空输入时显示） -->
    <div v-if="historyItems.length" class="palette-group">
      <div class="palette-group-title">{{ t('layout.paletteRecent') }}</div>
      <div
        v-for="(h, i) in historyItems"
        :key="'h-' + i"
        class="palette-item"
        :class="{ active: i === activeIndex }"
        @mouseenter="setActiveHistory(i)"
        @click="runHistory(h)"
      >
        <el-icon :size="14" :class="h.kind === 'cmd' ? 'palette-cmd-icon' : ''">
          <component :is="h.kind === 'cmd' ? 'Lightning' : h.icon || 'Menu'" />
        </el-icon>
        <span class="palette-title">
          {{ h.kind === 'cmd' ? h.title : h.titleParts.map((s) => menuTitle(s)).join(' / ') }}
        </span>
        <span v-if="h.kind === 'menu'" class="palette-path">{{ h.path }}</span>
        <span v-else class="palette-key">{{ t('layout.paletteRecent') }}</span>
      </div>
    </div>
    <!-- 命令区 -->
    <div v-if="matchedCommands.length" class="palette-group">
      <div class="palette-group-title">{{ t('layout.paletteCommands') }}</div>
      <div
        v-for="(c, i) in matchedCommands"
        :key="c.id"
        class="palette-item"
        :class="{ active: cmdOffset + i === activeIndex }"
        @mouseenter="setActiveFromGroup(0, i)"
        @click="runCommand(c)"
      >
        <el-icon :size="14" class="palette-cmd-icon"><Lightning /></el-icon>
        <span class="palette-title">{{ c.title }}</span>
        <span class="palette-key">{{ c.kbd }}</span>
      </div>
    </div>
    <!-- 菜单区 -->
    <div v-if="filtered.length" class="palette-group">
      <div class="palette-group-title">{{ t('layout.palettePages') }}</div>
      <div
        v-for="(m, i) in filtered"
        :key="m.path"
        class="palette-item"
        :class="{ active: menuOffset + i === activeIndex }"
        @mouseenter="setActiveFromGroup(1, i)"
        @click="goMenu(m)"
      >
        <el-icon :size="14"><component :is="m.icon || 'Menu'" /></el-icon>
        <span class="palette-title">{{ m.titleParts.map((s) => menuTitle(s)).join(' / ') }}</span>
        <span class="palette-path">{{ m.path }}</span>
      </div>
    </div>
    <div v-if="!filtered.length && !matchedCommands.length && keyword" class="search-empty">
      {{ t('layout.searchMenuEmpty') }}
    </div>
    <div v-if="!keyword" class="palette-hint">
      <span class="palette-hint-row">
        <el-icon :size="13"><Lightning /></el-icon>
        {{ t('layout.paletteHint') }}
      </span>
    </div>
    <div class="palette-footer">
      <span>↑↓ {{ t('layout.paletteNav') }}</span>
      <span>Enter {{ t('layout.paletteOpen') }}</span>
      <span>Esc {{ t('layout.paletteClose') }}</span>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onBeforeUnmount, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Search, Lightning, HomeFilled, Setting, FullScreen, Refresh, Grid } from '@element-plus/icons-vue';
import { useUserStore } from '@/store/user';
import { useSettingsStore } from '@/store/settings';
import { useMenuTitle } from '@/i18n/translate';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();
const settings = useSettingsStore();
const menuTitle = useMenuTitle();

const visible = ref(false);
const keyword = ref('');
const activeIndex = ref(0);
const inputRef = ref();

const HISTORY_KEY = 'go-admin-palette-history';
// 最近使用（命令 id / 菜单 path，最多 5 条）
const history = ref<string[]>([]);
try {
  history.value = JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]');
} catch {
  history.value = [];
}

function recordKey(k: string) {
  history.value = [k, ...history.value.filter((x) => x !== k)].slice(0, 5);
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history.value));
  } catch {}
}

// 扁平化菜单：只保留可跳转的 C 型菜单（含父级目录标题与完整路径）
const flatMenus = computed(() => {
  const result = [];
  const walk = (menus, parent, base) => {
    (menus || []).forEach((m) => {
      if (m.menuType === 'M') {
        walk(m.children, m, (base || '') + '/' + m.path);
      } else if (m.menuType === 'C') {
        const fullPath = base ? base + '/' + m.path : '/' + m.path;
        result.push({
          path: fullPath,
          title: parent ? parent.menuName + ' / ' + m.menuName : m.menuName,
          titleParts: parent ? [parent.menuName, m.menuName] : [m.menuName],
          icon: m.icon
        });
      }
    });
  };
  walk(userStore.menus, null, '');
  return result;
});

interface PaletteCommand {
  id: string;
  title: string;
  icon?: string;
  kbd: string;
  run: () => void;
}

// 快捷命令
const commands: PaletteCommand[] = [
  { id: 'home', title: t('layout.cmdGoHome'), kbd: '', run: () => { visible.value = false; router.push('/index'); } },
  { id: 'settings', title: t('layout.cmdSettings'), kbd: '', run: () => { visible.value = false; settings.drawerVisible = true; } },
  { id: 'fullscreen', title: t('layout.cmdFullscreen'), kbd: '', run: () => { visible.value = false; toggleFullscreen(); } },
  { id: 'reload', title: t('layout.cmdReload'), kbd: '', run: () => { visible.value = false; location.reload(); } },
  {
    id: 'lang',
    title: t('layout.cmdLang') + '：' + (settings.language === 'zh' ? 'English' : '中文'),
    kbd: '',
    run: () => {
      visible.value = false;
      settings.setLanguage(settings.language === 'zh' ? 'en' : 'zh');
    }
  },
  {
    id: 'dark',
    title: t('layout.cmdDark') + '：' + (settings.darkMode ? '关' : '开'),
    kbd: '',
    run: () => {
      visible.value = false;
      settings.darkMode = !settings.darkMode;
      settings.applyTheme();
      settings.save();
    }
  },
  {
    id: 'clear-history',
    title: t('layout.cmdClearRecent'),
    kbd: '',
    run: () => {
      visible.value = false;
      history.value = [];
      try {
        localStorage.removeItem(HISTORY_KEY);
      } catch {}
    }
  }
];

// 布局切换命令（动态生成，含当前态标记）
const layoutCommands = computed<PaletteCommand[]>(() => {
  const L = [
    ['side', t('layout.laySide')],
    ['rightSide', t('layout.layRightSide')],
    ['top', t('layout.layTop')],
    ['bottom', t('layout.layBottom')],
    ['mix', t('layout.layMix')],
    ['dial', t('layout.layDial')],
    ['desktop', t('layout.layDesktop')],
    ['dock', t('layout.layDock')],
    ['vtab', t('layout.layVtab')]
  ] as const;
  return L.map(([mode, title]) => ({
    id: 'layout-' + mode,
    title: t('layout.cmdLayout') + '：' + title + (settings.layoutMode === mode ? ' ✓' : ''),
    kbd: settings.layoutMode === mode ? t('layout.cmdCurrent') : '',
    run: () => {
      visible.value = false;
      settings.layoutMode = mode;
      settings.save();
    }
  }));
});

const allCommands = computed(() => [...commands, ...layoutCommands.value]);

const filtered = computed(() => {
  if (!keyword.value) return [];
  const kw = keyword.value.toLowerCase();
  return flatMenus.value.filter((m) => m.title.toLowerCase().includes(kw));
});

const matchedCommands = computed(() => {
  if (!keyword.value) return allCommands.value;
  const kw = keyword.value.toLowerCase();
  return allCommands.value.filter((c) => c.title.toLowerCase().includes(kw));
});

// 最近使用项（空输入时显示）
const historyItems = computed(() => {
  if (keyword.value) return [];
  return history.value
    .map((k) => {
      const cmd = allCommands.value.find((c) => c.id === k);
      if (cmd) return { kind: 'cmd', title: cmd.title, run: cmd.run };
      const menu = flatMenus.value.find((m) => m.path === k);
      if (menu) return { kind: 'menu', path: menu.path, titleParts: menu.titleParts, icon: menu.icon };
      return null;
    })
    .filter(Boolean);
});

// 分组索引：历史（空输入时）→ 命令 → 菜单
const cmdOffset = computed(() => historyItems.value.length);
const menuOffset = computed(() => cmdOffset.value + matchedCommands.value.length);
const totalLen = computed(
  () => historyItems.value.length + matchedCommands.value.length + filtered.value.length
);

function moveActive(d: number) {
  if (!totalLen.value) return;
  activeIndex.value = (activeIndex.value + d + totalLen.value) % totalLen.value;
}

function setActiveFromGroup(group: number, i: number) {
  activeIndex.value = group === 0 ? cmdOffset.value + i : menuOffset.value + i;
}

function setActiveHistory(i: number) {
  activeIndex.value = i;
}

function open() {
  keyword.value = '';
  activeIndex.value = 0;
  visible.value = true;
}

function focusInput() {
  nextTick(() => inputRef.value?.focus());
}

function goMenu(m) {
  visible.value = false;
  router.push(m.path);
  recordKey(m.path);
}

function runCommand(c: PaletteCommand) {
  c.run();
  // 清空最近使用不应把自己记进历史
  if (c.id !== 'clear-history') recordKey(c.id);
}

function runHistory(h) {
  if (h.kind === 'cmd') runCommand(h);
  else goMenu(h);
}

function handleEnter() {
  const total = totalLen.value;
  if (!total) return;
  const idx = activeIndex.value;
  if (idx < cmdOffset.value) {
    runHistory(historyItems.value[idx]);
  } else if (idx < menuOffset.value) {
    runCommand(matchedCommands.value[idx - cmdOffset.value]);
  } else {
    const mi = idx - menuOffset.value;
    if (filtered.value[mi]) goMenu(filtered.value[mi]);
  }
}

function toggleFullscreen() {
  const d = document as any;
  if (!d.fullscreenElement && !d.webkitFullscreenElement) {
    const el = document.documentElement as any;
    (el.requestFullscreen || el.webkitRequestFullscreen)?.call(el);
  } else {
    (d.exitFullscreen || d.webkitExitFullscreen)?.call(d);
  }
}

function onKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    open();
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown);
  window.addEventListener('go-open-search', open);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown);
  window.removeEventListener('go-open-search', open);
});

defineExpose({ open });
</script>

<style scoped>
.palette-group {
  margin-top: 10px;
}

.palette-group-title {
  font-size: 11px;
  color: var(--go-text-secondary);
  padding: 4px 8px;
  letter-spacing: 0.5px;
}

.palette-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--go-text-primary);
}

.palette-item.active {
  background: var(--el-color-primary);
  color: #fff;
}

.palette-item.active .palette-path,
.palette-item.active .palette-key {
  color: rgba(255, 255, 255, 0.75);
}

.palette-cmd-icon {
  color: #e6a23c;
}

.palette-item.active .palette-cmd-icon {
  color: #fff;
}

.palette-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.palette-path {
  font-size: 12px;
  color: var(--go-text-secondary);
}

.palette-key {
  font-size: 11px;
  color: var(--go-text-secondary);
  background: var(--go-fill-light);
  padding: 1px 6px;
  border-radius: 4px;
}

.palette-hint {
  margin-top: 14px;
  padding: 12px;
  border-radius: 8px;
  background: var(--go-fill-light);
  color: var(--go-text-secondary);
  font-size: 12px;
}

.palette-hint-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.palette-footer {
  display: flex;
  gap: 16px;
  margin-top: 14px;
  padding-top: 10px;
  border-top: 1px solid var(--el-border-color);
  font-size: 11px;
  color: var(--go-text-secondary);
}

.search-empty {
  margin-top: 12px;
  text-align: center;
  color: var(--go-text-secondary);
  font-size: 13px;
  padding: 20px 0;
}
</style>
