// 表格高度自适应指令
// 用法：<el-table v-table-height> —— 自动将表格高度撑满卡片剩余可用空间（表格内部滚动）
// 计算：滚动容器底部 - 表格顶部偏移 - 底部留白；监听容器尺寸变化（窗口拖拽/最大化/视口变化）自动重算
import { nextTick } from 'vue';

// 查找最近的滚动容器：优先 app-main-body（常规布局），跳过卡片类容器（高度随内容撑开，非滚动视口），
// 回退最近的 overflow 滚动祖先（桌面布局窗口 dw-body），最后 body
function findScrollContainer(el) {
  let container = el.parentElement;
  while (container && container !== document.body) {
    if (container.classList.contains('app-main-body')) return container;
    if (
      !container.classList.contains('el-card__body') &&
      !container.classList.contains('el-card')
    ) {
      const cs = getComputedStyle(container);
      if (/(auto|scroll|overlay)/.test(cs.overflowY)) return container;
    }
    container = container.parentElement;
  }
  return container === document.body ? document.body : null;
}

// 计算表格自适应高度
function computeHeight(el, binding) {
  // 指令挂在 el-table 组件根元素（本身即 .el-table）
  const table = el.classList.contains('el-table') ? el : el.querySelector('.el-table');
  if (!table) return;
  const container = findScrollContainer(el);
  const reserve = binding.value?.reserve ?? 78; // 底部留白（分页+间距）
  const containerBottom = container ? container.getBoundingClientRect().bottom : window.innerHeight;
  const top = table.getBoundingClientRect().top;
  const maxHeight = Math.max(containerBottom - top - reserve, 200);
  table.style.maxHeight = maxHeight + 'px';
}

export default {
  mounted(el, binding) {
    nextTick(() => computeHeight(el, binding));
    // 监听滚动容器尺寸变化（窗口拖拽/最大化/视口变化均触发），替代仅 window resize
    const container = findScrollContainer(el);
    const ro = new ResizeObserver(() => computeHeight(el, binding));
    if (container) ro.observe(container);
    el.__tableHeightRO = ro;
  },
  updated(el, binding) {
    // 数据变化后高度可能改变（搜索表单/工具栏高度），延迟重算
    nextTick(() => computeHeight(el, binding));
  },
  unmounted(el) {
    if (el.__tableHeightRO) {
      el.__tableHeightRO.disconnect();
    }
  }
};
