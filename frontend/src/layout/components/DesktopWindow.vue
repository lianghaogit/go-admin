<template>
  <div
    class="desktop-window"
    :class="{ active, minimized, maximized }"
    :style="winStyle"
    @mousedown="onFocus"
  >
    <!-- 标题栏：可拖拽，双击最大化/还原 -->
    <div class="dw-titlebar" @mousedown.prevent="startDrag" @dblclick="toggleMaximize">
      <el-icon :size="14" class="dw-ico"><component :is="winIcon" /></el-icon>
      <span class="dw-title">{{ menuTitle(tag.title) }}</span>
      <div class="dw-controls" @mousedown.stop>
        <button class="dw-btn" title="最小化" @click="onMinimize">
          <el-icon :size="12"><Minus /></el-icon>
        </button>
        <button
          class="dw-btn dw-btn-max"
          :title="maximized ? '还原' : '最大化'"
          @click="toggleMaximize"
        >
          <span class="dw-max-sym" :class="{ restore: maximized }"></span>
        </button>
        <button v-if="!tag.affix" class="dw-btn dw-btn-close" title="关闭" @click="onClose">
          <el-icon :size="12"><Close /></el-icon>
        </button>
      </div>
    </div>
    <!-- 窗口内容：统一渲染自身路径组件，单一 keep-alive 常驻（所有窗口内容可见、层叠显示，激活窗口置顶） -->
    <div class="dw-body">
      <keep-alive :include="tagsStore.cachedViews">
        <component :is="idleComp" :key="tag.path" />
      </keep-alive>
    </div>
    <!-- 调整大小手柄：右下角 / 右缘 / 下缘（最大化时禁用） -->
    <div
      v-if="!maximized"
      class="dw-resize dw-resize-se"
      title="调整大小"
      @mousedown.prevent="startResize('se', $event)"
    ></div>
    <div
      v-if="!maximized"
      class="dw-resize dw-resize-e"
      @mousedown.prevent="startResize('e', $event)"
    ></div>
    <div
      v-if="!maximized"
      class="dw-resize dw-resize-s"
      @mousedown.prevent="startResize('s', $event)"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Minus, Close } from '@element-plus/icons-vue';
import { useTagsViewStore } from '@/store/tagsView';
import { resolveView } from '@/utils/viewMap';
import { useMenuTitle } from '@/i18n/translate';

const menuTitle = useMenuTitle();

const props = defineProps({
  tag: { type: Object, required: true },
  active: { type: Boolean, default: false },
  minimized: { type: Boolean, default: false },
  maximized: { type: Boolean, default: false },
  pos: { type: Object, default: () => ({ x: 60, y: 40 }) },
  z: { type: Number, default: 100 }
});

const emit = defineEmits(['activate', 'close', 'minimize', 'maximize', 'move', 'resize']);
const tagsStore = useTagsViewStore();

// 窗口组件：按自身路径解析页面组件（独立实例；激活与否只切换 display，实例常驻保状态）
const idleComp = computed(() => resolveView(props.tag.path));
// 窗口图标：首页固定 HomeFilled，其余取菜单图标，兜底 Document
const winIcon = computed(() =>
  props.tag.affix ? 'HomeFilled' : props.tag.meta?.icon || 'Document'
);

const winStyle = computed(() => {
  if (props.maximized) {
    return { zIndex: props.z };
  }
  return {
    left: props.pos.x + 'px',
    top: props.pos.y + 'px',
    width: props.pos.w + 'px',
    height: props.pos.h + 'px',
    zIndex: props.z
  };
});

// 窗口获得焦点
function onFocus() {
  if (!props.active) emit('activate');
}

// 最小化窗口
function onMinimize() {
  emit('minimize');
}

// 切换最大化/还原
function toggleMaximize() {
  emit('maximize');
}

// 关闭窗口
function onClose() {
  emit('close');
}

// 标题栏拖拽（最大化状态下不拖动）
function startDrag(e) {
  if (props.maximized || props.minimized) return;
  const startX = e.clientX - props.pos.x;
  const startY = e.clientY - props.pos.y;
  // 拖动移动回调：边界 clamp 在画布内
  const onMove = (ev) => {
    const canvas = document.querySelector('.desktop-canvas');
    if (!canvas) return;
    const cr = canvas.getBoundingClientRect();
    const x = Math.max(0, Math.min(ev.clientX - startX, cr.width - props.pos.w));
    const y = Math.max(0, Math.min(ev.clientY - startY, cr.height - props.pos.h));
    emit('move', { x, y });
  };
  // 拖动结束回调：移除监听
  const onUp = () => {
    document.removeEventListener('mousemove', onMove);
    document.removeEventListener('mouseup', onUp);
  };
  document.addEventListener('mousemove', onMove);
  document.addEventListener('mouseup', onUp);
}

// 调整窗口大小（右下角/右缘/下缘拖拽，画布内 clamp）
function startResize(dir, e) {
  if (props.maximized || props.minimized) return;
  const startX = e.clientX;
  const startY = e.clientY;
  const sw = props.pos.w;
  const sh = props.pos.h;
  // 缩放回调：按拖拽方向计算新宽高并 clamp
  const onMove = (ev) => {
    const canvas = document.querySelector('.desktop-canvas');
    if (!canvas) return;
    const cr = canvas.getBoundingClientRect();
    let w = sw;
    let h = sh;
    if (dir === 'se' || dir === 'e') {
      w = Math.max(Math.min(320, cr.width), Math.min(sw + (ev.clientX - startX), cr.width));
    }
    if (dir === 'se' || dir === 's') {
      h = Math.max(Math.min(220, cr.height), Math.min(sh + (ev.clientY - startY), cr.height));
    }
    emit('resize', { w, h });
  };
  // 缩放结束回调：移除监听
  const onUp = () => {
    document.removeEventListener('mousemove', onMove);
    document.removeEventListener('mouseup', onUp);
  };
  document.addEventListener('mousemove', onMove);
  document.addEventListener('mouseup', onUp);
}
</script>

<style scoped>
.desktop-window {
  position: absolute;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  min-width: min(320px, 100%);
  max-width: 100%;
  min-height: 220px;
  max-height: 100%;
  background: var(--go-main-bg, #fff);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 2px;
  box-shadow: 0 6px 24px rgba(0, 21, 41, 0.18);
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.desktop-window.active {
  box-shadow: 0 10px 32px rgba(0, 21, 41, 0.28);
}

.desktop-window.minimized {
  display: none;
}

.desktop-window.maximized {
  left: 0 !important;
  top: 0 !important;
  width: 100% !important;
  height: 100% !important;
  border-radius: 0;
}

.dw-titlebar {
  height: 34px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 8px 0 12px;
  background: var(--go-navbar-bg, #fff);
  border-bottom: 1px solid var(--go-border-color, #e4e7ed);
  cursor: move;
  user-select: none;
}

.dw-ico {
  color: var(--go-topnav-active, #409eff);
  flex-shrink: 0;
}

.dw-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--go-navbar-text, #303133);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dw-controls {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.dw-btn {
  width: 26px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 2px;
  background: transparent;
  color: var(--go-navbar-text, #303133);
  cursor: pointer;
}

.dw-btn:hover {
  background: rgba(0, 0, 0, 0.08);
}

.dw-btn-close:hover {
  background: #f56c6c;
  color: #fff;
}

/* 最大化/还原符号（Windows 风格） */
.dw-max-sym {
  width: 10px;
  height: 10px;
  border: 1.5px solid currentColor;
  border-radius: 1px;
}

.dw-max-sym.restore {
  position: relative;
  transform: translate(2px, -2px);
  border: none;
}

.dw-max-sym.restore::before {
  content: '';
  position: absolute;
  left: -4px;
  top: 1px;
  width: 8px;
  height: 8px;
  border: 1.5px solid currentColor;
  border-radius: 1px;
}

.dw-max-sym.restore::after {
  content: '';
  position: absolute;
  left: 0;
  top: -2px;
  width: 8px;
  height: 8px;
  background: var(--go-navbar-bg, #fff);
  border: 1.5px solid currentColor;
  border-radius: 1px;
}

.dw-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 12px;
  background: var(--go-main-bg, #f0f2f5);
}

/* ===== 调整大小手柄 ===== */
.dw-resize {
  position: absolute;
  z-index: 10;
}

.dw-resize-e {
  top: 0;
  right: 0;
  width: 6px;
  height: 100%;
  cursor: e-resize;
}

.dw-resize-s {
  bottom: 0;
  left: 0;
  width: 100%;
  height: 6px;
  cursor: s-resize;
}

.dw-resize-se {
  right: 0;
  bottom: 0;
  width: 16px;
  height: 16px;
  cursor: nwse-resize;
  background: linear-gradient(135deg, transparent 50%, rgba(64, 158, 255, 0.45) 50%);
  border-radius: 0 0 2px 0;
}

.dw-resize-se:hover {
  background: linear-gradient(135deg, transparent 50%, rgba(64, 158, 255, 0.8) 50%);
}
</style>
