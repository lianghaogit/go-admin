<template>
  <div class="flow-canvas-wrap">
    <div class="canvas-toolbar">
      <el-button size="small" type="primary" plain :icon="Plus" @click="addNode">添加节点</el-button>
      <el-button
        size="small"
        type="danger"
        plain
        :icon="Delete"
        :disabled="!selectedId"
        @click="removeSelected"
      >
        删除选中
      </el-button>
      <el-button size="small" plain :icon="Sort" @click="autoLayout">自动排列</el-button>
      <span class="canvas-tip">拖拽节点可调整审批顺序（自上而下），点击节点编辑配置</span>
    </div>
    <div ref="containerRef" class="canvas-container" />
  </div>
</template>

<script lang="ts">
export type FlowNode = {
  nodeName: string;
  approverType: string;
  approverIds: string[];
  nodeMode: string;
  canAddSign: string;
  rejectTo: string;
  timeoutHours: number;
  remark: string;
};
</script>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { Graph, Node } from '@antv/x6';
import { content as x6Style } from '@antv/x6/es/style/raw';
import { ElMessage } from 'element-plus';
import { Plus, Delete, Sort } from '@element-plus/icons-vue';

// X6 3.x 样式为 CSS-in-JS（raw 导出字符串），全局注入一次
(function injectX6Style() {
  if (document.getElementById('x6-base-style')) return;
  const styleEl = document.createElement('style');
  styleEl.id = 'x6-base-style';
  styleEl.textContent = x6Style;
  document.head.appendChild(styleEl);
})();

const props = defineProps<{ initNodes?: FlowNode[] }>();

const emit = defineEmits<{ edit: [node: FlowNode] }>();

const containerRef = ref<HTMLDivElement>();
let graph: Graph | null = null;
let nodesData: FlowNode[] = [];
let selectedId = ref<string | null>(null);

const START_ID = 'flow-start';
const END_ID = 'flow-end';
const NODE_W = 220;
const NODE_H = 72;
const GAP = 44;
const START_Y = 16;

const emptyNode = (i: number): FlowNode => ({
  nodeName: `审批节点${i}`,
  approverType: 'user',
  approverIds: [],
  nodeMode: 'one',
  canAddSign: '0',
  rejectTo: '0',
  timeoutHours: 0,
  remark: ''
});

/** 明暗主题色板（跟随 html.dark，与 element-plus dark 变量协调） */
function palette() {
  const dark = document.documentElement.classList.contains('dark');
  return {
    nodeFill: dark ? '#1f2d3d' : '#ffffff',
    nodeStroke: '#409eff',
    labelFill: dark ? '#e5eaf3' : '#303133',
    summaryFill: dark ? '#a8abb2' : '#909399',
    startFill: dark ? '#16301f' : '#f0f9eb',
    endFill: dark ? '#3a1f1f' : '#fef0f0',
    startText: dark ? '#5cb87a' : '#67c23a',
    endText: dark ? '#e08080' : '#f56c6c',
    edgeStroke: dark ? '#5c5c66' : '#909399'
  };
}

function summaryOf(n: FlowNode): string {
  const who =
    n.approverType === 'role' ? '按角色' : n.approverType === 'dept_leader' ? '部门负责人' : `指定人员${n.approverIds?.length || 0}人`;
  const mode = n.nodeMode === 'any' ? '会签' : '或签';
  const sign = n.canAddSign === '1' ? '·可加签' : '';
  const timeout = n.timeoutHours > 0 ? `·超时${n.timeoutHours}h` : '';
  return `${who}·${mode}${sign}${timeout}`;
}

function buildGraph(container: HTMLElement, height: number) {
  if (graph) {
    graph.dispose();
    graph = null;
  }
  graph = new Graph({
    container,
    width: container.clientWidth,
    height,
    grid: true,
    virtual: false,
    // X6 3.x 默认 async:true（渲染走 rAF 队列）；页面不可见时 rAF 不触发会导致视图不渲染，故用同步渲染
    async: false,
    panning: false,
    mousewheel: false,
    connecting: { allowBlank: false, allowLoop: false, allowMulti: false, allowNode: false },
    interacting: {
      edgeMovable: false,
      // 起止节点固定位置，仅审批节点可拖拽
      nodeMovable: (view: any) => view?.cell?.id !== START_ID && view?.cell?.id !== END_ID
    }
  });

  graph.on('node:click', ({ node }) => {
    selectNode(node);
    if (node.id !== START_ID && node.id !== END_ID) {
      emit('edit', node.getData() as FlowNode);
    }
  });
  graph.on('blank:click', () => selectNode(null));
}

function selectNode(node: Node | null) {
  if (selectedId.value && graph) {
    const prev = graph.getCellById(selectedId.value);
    if (prev && prev.isNode()) applySelectStyle(prev as Node, false);
  }
  selectedId.value = node ? node.id : null;
  if (node) applySelectStyle(node, true);
}

/** 主题切换时刷新节点配色（html.dark 类变化触发） */
function applyThemeToNodes() {
  if (!graph) return;
  const p = palette();
  graph.getNodes().forEach((n) => {
    if (n.id === START_ID) {
      n.attr({ body: { fill: p.startFill }, label: { fill: p.startText } });
    } else if (n.id === END_ID) {
      n.attr({ body: { fill: p.endFill }, label: { fill: p.endText } });
    } else {
      n.attr({
        body: { fill: p.nodeFill },
        label: { fill: p.labelFill },
        summary: { fill: p.summaryFill }
      });
    }
  });
  graph.getEdges().forEach((e) => e.attr({ line: { stroke: p.edgeStroke } }));
}

function applySelectStyle(node: Node, selected: boolean) {
  const stroke = selected ? '#409eff' : '#dcdfe6';
  const width = selected ? 2 : 1;
  node.attr('body/stroke', stroke);
  node.attr('body/strokeWidth', width);
}

function heightFor(n: number) {
  return START_Y + n * (NODE_H + GAP) + 40 + 40 + 16;
}

function addStartEnd() {
  if (!graph) return;
  const p = palette();
  graph.addNode({
    id: START_ID,
    x: 40,
    y: START_Y,
    width: NODE_W,
    height: 40,
    shape: 'rect',
    attrs: {
      body: { rx: 20, fill: p.startFill, stroke: '#67c23a', strokeWidth: 1 },
      label: { text: '发起申请', fill: p.startText, fontSize: 13, fontWeight: 600 }
    }
  });
  const endY = START_Y + nodesData.length * (NODE_H + GAP) + 40;
  graph.addNode({
    id: END_ID,
    x: 40,
    y: endY,
    width: NODE_W,
    height: 40,
    shape: 'rect',
    attrs: {
      body: { rx: 20, fill: p.endFill, stroke: '#f56c6c', strokeWidth: 1 },
      label: { text: '流程结束', fill: p.endText, fontSize: 13, fontWeight: 600 }
    }
  });
}

function addApproveNodes() {
  if (!graph) return;
  const p = palette();
  nodesData.forEach((n, i) => {
    const y = START_Y + 40 + GAP + i * (NODE_H + GAP);
    graph!.addNode({
      id: `flow-node-${i}`,
      x: 40,
      y,
      width: NODE_W,
      height: NODE_H,
      shape: 'rect',
      data: n,
      markup: [
        { tagName: 'rect', selector: 'body' },
        { tagName: 'text', selector: 'label' },
        { tagName: 'text', selector: 'summary' }
      ],
      attrs: {
        body: { fill: p.nodeFill, stroke: p.nodeStroke, strokeWidth: 1, rx: 6, ry: 6 },
        label: {
          text: n.nodeName || '未命名节点',
          fill: p.labelFill,
          fontSize: 14,
          fontWeight: 600,
          textAnchor: 'middle',
          refX: 0.5,
          refY: 0.42
        },
        summary: {
          text: summaryOf(n),
          fill: p.summaryFill,
          fontSize: 11,
          textAnchor: 'middle',
          refX: 0.5,
          refY: 0.85
        }
      }
    });
  });
}

function addEdges() {
  if (!graph) return;
  const p = palette();
  const ids = [START_ID, ...nodesData.map((_, i) => `flow-node-${i}`), END_ID];
  for (let i = 0; i < ids.length - 1; i++) {
    graph.addEdge({
      source: ids[i],
      target: ids[i + 1],
      // manhattan 避障路由（拖拽实时重算，路径不穿节点）+ rounded 圆角连接 + smooth 平滑
      router: { name: 'manhattan', args: { padding: 12 } },
      connector: { name: 'rounded', args: { radius: 6 } },
      smooth: true,
      attrs: {
        line: { stroke: p.edgeStroke, strokeWidth: 1.5, targetMarker: { name: 'block', size: 6 } }
      }
    });
  }
}

function relayout(container: HTMLElement) {
  // 画布高度 = 至少撑满弹窗剩余可视区（操作空间最大化），节点超出时画布内滚动
  const h = Math.max(heightFor(nodesData.length), container.clientHeight || 0);
  if (graph) {
    graph.dispose();
    graph = null;
  }
  buildGraph(container, h);
  addStartEnd();
  addApproveNodes();
  addEdges();
}

function render(nodes: FlowNode[]) {
  nodesData = nodes.map((n) => ({ ...n, approverIds: [...(n.approverIds || [])] }));
  selectedId.value = null;
  if (containerRef.value) {
    try {
      relayout(containerRef.value);
    } catch (e) {
      // 渲染异常不打断交互，交由画布后续操作兜底
    }
  }
}

/** 按画布 y 坐标收集审批节点（自上而下 = 审批顺序） */
function collect(): FlowNode[] {
  if (!graph) return nodesData;
  const cells = graph
    .getNodes()
    .filter((n) => n.id !== START_ID && n.id !== END_ID)
    .sort((a, b) => a.getBBox().y - b.getBBox().y);
  return cells.map((c) => c.getData() as FlowNode);
}

/** 属性编辑保存后刷新画布文本 */
function updateNode(node: FlowNode) {
  if (!graph) return;
  const cell = selectedId.value ? graph.getCellById(selectedId.value) : null;
  if (!cell || !cell.isNode()) return;
  cell.setData({ ...node });
  cell.attr({
    label: { text: node.nodeName || '未命名节点' },
    summary: { text: summaryOf(node) }
  });
}

function addNode() {
  if (!graph || !containerRef.value) return;
  const idx = nodesData.length + 1;
  const node = emptyNode(idx);
  let insertAfterId = START_ID;
  if (selectedId.value) {
    const sel = graph.getCellById(selectedId.value);
    if (sel && sel.isNode() && sel.id !== END_ID) {
      insertAfterId = sel.id;
    }
  }
  // 插入 nodesData（按当前位置：选中节点之后，未选中则追加到末尾）
  const pos = collect().findIndex((n) => n === (graph.getCellById(insertAfterId) as Node)?.getData());
  nodesData = collect();
  const idxArr = pos >= 0 ? pos + 1 : nodesData.length;
  nodesData.splice(idxArr, 0, node);
  relayout(containerRef.value);
  // 找到新节点并选中
  const newCell = graph?.getNodes().find((n) => n.getData() === node);
  if (newCell) selectNode(newCell);
}

function removeSelected() {
  if (!graph || !selectedId.value || !containerRef.value) return;
  const cell = graph.getCellById(selectedId.value);
  if (!cell || !cell.isNode()) return;
  const node = cell as Node;
  if (node.id === START_ID || node.id === END_ID) return;
  if (nodesData.length <= 1) {
    ElMessage.warning('至少保留一个审批节点');
    return;
  }
  nodesData = collect().filter((n) => n !== node.getData());
  selectedId.value = null;
  relayout(containerRef.value);
}

function autoLayout() {
  if (!graph || !containerRef.value) return;
  const sorted = collect();
  sorted.forEach((n, i) => {
    const cell = graph!.getNodes().find((c) => c.getData() === n);
    if (cell) cell.position(40, START_Y + 40 + GAP + i * (NODE_H + GAP), { absolute: true });
  });
}

defineExpose({ render, collect, updateNode });

let themeObserver: MutationObserver | null = null;

// 数据驱动渲染：挂载后自动用 initNodes 绘制（消除 openDialog 等待 canvasRef 的竞态）
onMounted(async () => {
  // 明暗主题切换时实时刷新画布节点配色
  themeObserver = new MutationObserver(() => applyThemeToNodes());
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
  if (!props.initNodes || !props.initNodes.length) return;
  // 等待弹窗过渡动画结束、容器具备实际尺寸（宽 + 高）后再初始化 Graph
  for (let i = 0; i < 30 && !containerRef.value?.clientWidth; i++) {
    await new Promise((r) => setTimeout(r, 100));
  }
  // 高度可能比宽度晚就绪（弹窗纵向动画），再等高度就绪以确保画布撑满
  for (let i = 0; i < 30 && !containerRef.value?.clientHeight; i++) {
    await new Promise((r) => setTimeout(r, 100));
  }
  render(props.initNodes);
});

onBeforeUnmount(() => {
  themeObserver?.disconnect();
  themeObserver = null;
  if (graph) {
    graph.dispose();
    graph = null;
  }
});
</script>

<style scoped>
.flow-canvas-wrap {
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 6px;
  padding: 10px;
  background: var(--el-fill-color-lighter, #fafafa);
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.canvas-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  row-gap: 4px;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.canvas-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  margin-left: 8px;
}

.canvas-container {
  width: 100%;
  flex: 1;
  min-height: 120px;
  background: var(--el-bg-color, #fff);
  border: 1px solid var(--go-border-color, #e4e7ed);
  border-radius: 4px;
  overflow: auto;
}

.canvas-container :deep(.x6-node) {
  transition: filter 0.2s ease;
}

.canvas-container :deep(.x6-node:hover) {
  filter: brightness(1.02);
}

html.dark .canvas-container :deep(.x6-node:hover) {
  filter: brightness(1.2);
}
</style>
