<template>
  <div
    ref="wrapRef"
    class="dial-wrap"
    :class="{ expanded }"
    :style="wrapStyle"
    @mouseenter="cancelCollapse"
    @mouseleave="scheduleCollapse"
    @wheel.prevent="onWheel"
    @click="onWrapClick"
  >
    <!-- 最小态：右上角悬浮小圆钮（透明度 ~0.4），hover/点击放大 -->
    <button
      v-if="!expanded"
      class="dial-mini"
      title="转盘菜单"
      @mouseenter="expand"
      @click="expand"
    >
      <el-icon :size="16"><Menu /></el-icon>
    </button>

    <!-- 展开态：1/4 圆转盘（圆心在右上角） -->
    <transition name="dial-pop">
      <div v-if="expanded" class="dial-panel">
        <!-- 扇形底（1/4 圆） -->
        <div class="dial-sector" :style="sectorStyle" />
        <!-- 圆心设计按钮：展开态替代小圆钮，点击打开界面设计（布局/配色等） -->
        <button class="dial-design" title="界面设计" @click.stop="openSettings">
          <el-icon :size="16"><Setting /></el-icon>
        </button>
        <!-- 各层级圆环：点击目录项后，其子菜单在更外圈展开 -->
        <div
          v-for="(lv, li) in levels"
          :key="li"
          class="dial-ring"
          :style="{ '--r': lv.radius + 'px' }"
        >
          <div
            v-for="item in lv.items"
            :key="item.menuId"
            class="dial-item"
            :class="{ visible: itemVisible(lv, item), active: isActive(lv, item) }"
            :style="itemStyle(lv, item)"
            @click="onItemClick(li, lv, item)"
          >
            <span class="dial-item-inner">
              <el-icon v-if="settings.menuIcon && item.icon" :size="13"
                ><component :is="safeIcon(item.icon)"
              /></el-icon>
              {{ menuTitle(item.menuName) }}
            </span>
          </div>
        </div>
        <!-- 可滚动提示（当前层菜单项超出可见窗口时） -->
        <div v-if="lastScrollable" class="dial-hint">
          <el-icon :size="12"><RefreshRight /></el-icon>滚轮浏览
        </div>
      </div>
    </transition>
  </div>

  <!-- 展开态全屏透明遮罩：点击光圈（转盘）以外任意位置即整圈收起；
       遮罩 z-index 低于转盘 wrap，转盘滚轮与交互不受影响 -->
  <div v-if="expanded" class="dial-backdrop" @click="collapse"></div>

  <!-- 右上角工具条：命令面板 / 全屏 / 账号（dial 布局无顶栏，补齐常用入口）；
       转盘展开时隐藏，避免与转盘交互区叠放 -->
  <div v-show="!expanded" class="dial-tools">
    <button class="dial-tool" :title="t('layout.searchMenu') + ' (Ctrl+K)'" @click="openSearch">
      <el-icon :size="15"><Search /></el-icon>
    </button>
    <button class="dial-tool" title="全屏" @click="toggleFullscreen">
      <el-icon :size="15"><FullScreen /></el-icon>
    </button>
    <div class="dial-tool dial-tool-user" title="账号">
      <UserDropdown />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { FullScreen, Menu, RefreshRight, Search, Setting } from '@element-plus/icons-vue';
import { useUserStore } from '@/store/user';
import { useSettingsStore } from '@/store/settings';
import { safeIcon } from '@/utils/iconNames';
import { useMenuTitle } from '@/i18n/translate';
import { useI18n } from 'vue-i18n';
import UserDropdown from './UserDropdown.vue';

const { t } = useI18n();

// 单个 1/4 圆内最多同时可见的菜单项数（超出需滚轮滑动）
const VISIBLE = 8;
// 各层级圆环半径（圆心右上角，逐级向外）
const RADIUS_STEP = 110;
const RADIUS_BASE = 168;

const router = useRouter();
const userStore = useUserStore();
const settings = useSettingsStore();
const menuTitle = useMenuTitle();

// 打开全局命令面板（与 layout 顶层 SearchMenu 联动）
function openSearch() {
  window.dispatchEvent(new Event('go-open-search'));
}

// 切换全屏（无顶栏布局的常用入口）
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {});
  } else {
    document.exitFullscreen().catch(() => {});
  }
}

const expanded = ref(false);
// 转盘容器（用于判断点击是否落在 1/4 圆光圈内）
const wrapRef = ref<HTMLElement | null>(null);
// 折叠计时器（鼠标移出转盘区域后收起）
let collapseTimer: number | null = null;

// 交互区尺寸：展开态覆盖扇形区域（最深圈半径矩形），最小态仅小圆钮
const wrapStyle = computed(() => ({
  width: expanded.value ? maxRadius.value * 2 + 'px' : '0px',
  height: expanded.value ? maxRadius.value * 2 + 'px' : '0px'
}));

// 一级菜单：固定首页入口 + 后端菜单
const rootMenus = computed(() => [
  {
    menuId: -1,
    path: 'index',
    menuType: 'C',
    menuName: '首页',
    icon: 'HomeFilled',
    children: []
  },
  ...userStore.menus
]);

// 各级圆环：{ items, base(路径前缀), radius, offset(滚动偏移) }
const levels = reactive<any[]>([]);
// 当前点击路径链（用于高亮）
const activePathChain = ref<string[]>([]);

// 扇形底尺寸 = 最深圈菜单弧半径（光圈不超出菜单显示区域）
const maxRadius = computed(() => {
  const deep = levels.length ? levels[levels.length - 1].radius : RADIUS_BASE;
  return deep;
});
const sectorStyle = computed(() => ({
  width: maxRadius.value * 2 + 'px',
  height: maxRadius.value * 2 + 'px'
}));

// 最后一级是否可滚动：项数超过可见窗口即支持滚轮（一级/二级通用，固定间隔下超出窗口的部分滚动浏览）
const lastScrollable = computed(() => {
  if (!levels.length) return false;
  const lv = levels[levels.length - 1];
  return lv.items.length > VISIBLE;
});

function clamp(v: number, min: number, max: number) {
  return Math.min(Math.max(v, min), max);
}

// 二级展开的固定角间隔（度）：围绕所点一级的角度向两侧展开，项间距恒定（不随项数变大变小）
const CHILD_STEP_DEG = 12;

// 该项角度：一级按实际项数均分 1/4 圆；二级以所点一级的角度为中心、固定间隔向两侧展开
// （项少时距离恒定不再拉大；项多时向两边延伸，超出 1/4 圆的部分整体平移保证在视口内，无需滚轮）
function itemDeg(lv: any, item: any) {
  if (lv.parentDeg != null) return childDeg(lv, item);
  const idx = lv.items.indexOf(item);
  return ((idx + 0.5 - (lv.offset || 0)) * 90) / lv.items.length;
}

// 二级角度：以所点一级的角度为中心、固定间隔向两侧展开（间距恒定，不随项数变化）；
// 项数超出可见窗口（跨度 > 90°）时启用 offset 滚轮窗口——窗口内仍固定 12° 间隔，超出部分滚动浏览；
// 圆心在视口右上角，越界保护按窗口整体跨度平移（任意 offset 下窗口都落在 [0, 90]，末项不会越界隐藏）
function childDeg(lv: any, item: any) {
  const idx = lv.items.indexOf(item);
  const n = lv.items.length;
  const off = lv.offset || 0;
  const winN = Math.min(n, VISIBLE);
  const base = lv.parentDeg;
  const rel = idx - off - (winN - 1) / 2; // 窗口内相对位置（恒跨 (winN-1)*STEP°）
  let deg = base + rel * CHILD_STEP_DEG;
  // 窗口整体越界平移：min/max 只与窗口跨度有关，与 offset 无关
  const halfSpan = ((winN - 1) / 2) * CHILD_STEP_DEG;
  if (base - halfSpan < 0) deg += halfSpan - base;
  if (base + halfSpan > 90) deg += 90 - (base + halfSpan);
  return deg;
}

// 可见窗口：当前滚动 offset 起的连续 VISIBLE 项（同时保证角度不越界）。
// 一级与二级统一：项数 ≤ VISIBLE 时 offset 恒为 0 → 全部可见；项数多时窗口滚动浏览
function itemVisible(lv: any, item: any) {
  const idx = lv.items.indexOf(item);
  const inWindow = idx >= (lv.offset || 0) && idx < (lv.offset || 0) + VISIBLE;
  if (!inWindow) return false;
  const deg = itemDeg(lv, item);
  return deg >= -4 && deg <= 94;
}

// 该项样式：圆心在视口右上角，项中心精确落在弧线上（left/top 三角函数定位），
// 仅用 rotate 让项沿径向倾斜（扇形姿态），文字随项同角度；位移不用 transform 平移，避免浏览器变换偏差
function itemStyle(lv: any, item: any) {
  const deg = itemDeg(lv, item);
  const rad = (deg * Math.PI) / 180;
  const x = -lv.radius * Math.cos(rad); // 相对圆心水平偏移（向左）
  const y = lv.radius * Math.sin(rad); // 相对圆心垂直偏移（向下）
  return {
    left: x - 33 + 'px', // 项中心对准弧线（项宽约 66px → 半宽 33）
    top: y - 14 + 'px', // 项高约 28px → 半高 14
    transform: 'rotate(' + -deg + 'deg)',
    opacity: itemVisible(lv, item) ? 1 : 0,
    pointerEvents: (itemVisible(lv, item) ? 'auto' : 'none') as 'auto' | 'none'
  };
}

// 高亮：当前选中路径链
function isActive(lv: any, item: any) {
  const path = joinPath(lv.base, item.path);
  return activePathChain.value.includes(path);
}

function joinPath(base: string, p: string) {
  const seg = '/' + String(p).replace(/^\/+/, '');
  return base === '' ? seg : base.replace(/\/$/, '') + seg;
}

// 点击菜单项：目录 → 外圈展开子菜单；叶子 → 跳转 + 转盘回到最小状态
function onItemClick(levelIdx: number, lv: any, item: any) {
  const path = joinPath(lv.base, item.path);
  const children = (item.children || []).filter((c: any) => c.menuType !== 'F');
  if (!children.length) {
    // 叶子：跳转并收起
    activePathChain.value = [];
    collapse();
    router.push(path);
    return;
  }
  // 目录：截断到当前圈再向外展开子菜单（记录所点一级的角度，二级围绕它固定间隔展开）
  levels.splice(levelIdx + 1);
  activePathChain.value = [...activePathChain.value.slice(0, levelIdx), path];
  levels.push({
    items: children,
    base: path,
    radius: lv.radius + RADIUS_STEP,
    offset: 0,
    parentDeg: itemDeg(lv, item)
  });
}

// 展开转盘：初始化一级菜单
function expand() {
  expanded.value = true;
  activePathChain.value = [];
  levels.splice(0, levels.length, {
    items: rootMenus.value,
    base: '',
    radius: RADIUS_BASE,
    offset: 0
  });
}

// 收起转盘
function collapse() {
  expanded.value = false;
  levels.splice(0);
  if (collapseTimer) {
    window.clearTimeout(collapseTimer);
    collapseTimer = null;
  }
}

function cancelCollapse() {
  if (collapseTimer) {
    window.clearTimeout(collapseTimer);
    collapseTimer = null;
  }
}

// 鼠标移出转盘区域后自动收起（点击/操作后留有充分阅读时间）
function scheduleCollapse() {
  if (!expanded.value) return;
  cancelCollapse();
  collapseTimer = window.setTimeout(() => {
    collapse();
  }, 1200);
}

// 滚轮：最后一级项数超过可见窗口时滚动浏览（固定间隔展开，超出窗口的部分滚轮查看）
function onWheel(e: WheelEvent) {
  if (!levels.length) return;
  const lv = levels[levels.length - 1];
  const maxOff = Math.max(0, lv.items.length - VISIBLE);
  if (maxOff <= 0) return;
  lv.offset = clamp((lv.offset || 0) + (e.deltaY > 0 ? 1 : -1), 0, maxOff);
}

// 点击转盘 wrap 内空白（扇形外，如右下直角区）也整圈收起：
// 转盘交互元素（小圆钮/设计按钮/菜单项/面板）直接放行，其余落在 1/4 圆扇形外的点击收起。
// 点击 wrap 以外位置由全屏透明遮罩（.dial-backdrop）处理，无需 document 全局监听。
function onWrapClick(e: MouseEvent) {
  if (!expanded.value) return;
  const target = e.target as HTMLElement;
  if (target.closest('.dial-mini, .dial-design, .dial-item, .dial-ring, .dial-panel')) return;
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  const relX = rect.right - e.clientX; // 相对圆心向左为正
  const relY = e.clientY - rect.top; // 相对圆心向下为正
  if (relX >= 0 && relY >= 0 && Math.hypot(relX, relY) <= maxRadius.value) return;
  collapse();
}

// 展开态圆心设计按钮：先收起转盘（全屏遮罩 z-index 高于抽屉，不收起会拦截抽屉内点击），再打开界面设计抽屉
function openSettings() {
  collapse();
  settings.drawerVisible = true;
}

onBeforeUnmount(() => {
  if (collapseTimer) window.clearTimeout(collapseTimer);
});
</script>

<style scoped>
/* 展开态全屏透明遮罩：点击转盘以外区域即收起（置于转盘 wrap 之下，不影响滚轮与交互） */
.dial-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2999;
  background: transparent;
}

.dial-wrap {
  position: fixed;
  top: 0;
  right: 0;
  z-index: 3000;
  /* 交互区覆盖扇形区域（最深圈半径矩形） */
  width: 0;
  height: 0;
}

/* 最小态小圆钮：右上角，透明度 ~0.4，hover 放大 */
.dial-mini {
  position: absolute;
  right: 8px;
  top: 8px;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(135deg, var(--go-topnav-active, #409eff), #8b5cf6);
  opacity: 0.4;
  box-shadow: 0 2px 8px rgba(31, 45, 61, 0.25);
  transition: opacity 0.25s, transform 0.25s;
}

.dial-mini:hover {
  opacity: 1;
  transform: scale(1.15);
}

/* 右上角工具条：命令面板 / 全屏 / 账号（dial 无顶栏补齐常用入口），与转盘同风格 */
.dial-tools {
  position: fixed;
  top: 46px;
  right: 8px;
  z-index: 2998;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.dial-tool {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(135deg, var(--go-topnav-active, #409eff), #8b5cf6);
  opacity: 0.5;
  box-shadow: 0 2px 8px rgba(31, 45, 61, 0.25);
  transition: opacity 0.25s, transform 0.25s;
  padding: 0;
}

.dial-tool:hover {
  opacity: 1;
  transform: scale(1.1);
}

/* 账号按钮：只展示头像（复用 UserDropdown 内头像） */
.dial-tool-user {
  overflow: hidden;
  background: linear-gradient(135deg, #4b5563, #1f2937);
}

.dial-tool-user :deep(.user-info) {
  gap: 0;
}

.dial-tool-user :deep(.user-name),
.dial-tool-user :deep(.user-info > .el-icon) {
  display: none;
}

/* 展开态圆心设计按钮：替代小圆钮，点击打开界面设计（布局/配色等） */
.dial-design {
  position: absolute;
  right: 8px;
  top: 8px;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(135deg, var(--go-topnav-active, #409eff), #8b5cf6);
  opacity: 0.85;
  box-shadow: 0 2px 8px rgba(31, 45, 61, 0.25);
  transition: opacity 0.25s, transform 0.25s;
  z-index: 2;
}

.dial-design:hover {
  opacity: 1;
  transform: scale(1.15);
}

/* 展开态转盘（圆心在右上角 0,0） */
.dial-panel {
  position: absolute;
  top: 0;
  right: 0;
  width: 0;
  height: 0;
}

.dial-pop-enter-active,
.dial-pop-leave-active {
  transition: opacity 0.22s, transform 0.22s;
}

.dial-pop-enter-from,
.dial-pop-leave-to {
  opacity: 0;
  transform: scale(0.6);
  transform-origin: top right;
}

/* 扇形底：1/4 圆（圆心右上角，圆向右下展开） */
.dial-sector {
  position: absolute;
  right: 0;
  top: 0;
  border-radius: 0 0 0 100%;
  background: conic-gradient(
    from 0deg at 0 0,
    rgba(64, 158, 255, 0.16),
    rgba(139, 92, 246, 0.1) 90deg,
    transparent 90deg
  );
  box-shadow: inset 0 0 40px rgba(64, 158, 255, 0.08);
  pointer-events: none;
}

/* 圆环：承载一圈菜单项 */
.dial-ring {
  position: absolute;
  top: 0;
  right: 0;
  width: 0;
  height: 0;
}

/* 菜单项：位置由 JS 三角函数直接定位（圆心右上角），旋转仅用于径向倾斜 */
.dial-item {
  position: absolute;
  left: 0;
  top: 0;
  transition: left 0.28s ease, top 0.28s ease, transform 0.28s ease, opacity 0.2s;
  will-change: transform;
}

.dial-item-inner {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 96px;
  padding: 5px 12px;
  border-radius: 14px;
  font-size: 12px;
  /* 文字/背景跟随 Element 主题变量：浅色主题深字白底，深色主题自动反色，避免与侧栏主题变量混用导致白底浅字看不清 */
  color: var(--el-text-color-primary, #303133);
  background: var(--el-bg-color-overlay, #ffffff);
  border: 1px solid var(--el-border-color, #e4e7ed);
  box-shadow: 0 2px 8px rgba(31, 45, 61, 0.12);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  /* 不额外旋转：跟随 .dial-item 的径向倾斜，呈现扇形姿态 */
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}

.dial-item:hover .dial-item-inner {
  color: #fff;
  background: linear-gradient(135deg, var(--go-topnav-active, #409eff), #8b5cf6);
  border-color: transparent;
}

/* 当前选中路径上的项 */
.dial-item.active .dial-item-inner {
  color: var(--go-topnav-active, #409eff);
  border-color: var(--go-topnav-active, #409eff);
  background: rgba(64, 158, 255, 0.08);
}

/* 滚动提示 */
.dial-hint {
  position: absolute;
  right: 14px;
  top: 46px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--go-text-secondary, #909399);
  background: rgba(255, 255, 255, 0.85);
  padding: 3px 8px;
  border-radius: 10px;
  box-shadow: 0 1px 6px rgba(31, 45, 61, 0.12);
  pointer-events: none;
}
</style>
