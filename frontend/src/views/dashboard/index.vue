<template>
  <div class="dash">
    <!-- 顶栏：标题 + 实时时间（无欢迎语） -->
    <div class="dash-bar">
      <div class="dash-bar-title">
        <span class="dash-bar-logo">Go Admin</span>
        <span class="dash-bar-sub">数据总览</span>
      </div>
      <div class="dash-bar-time">
        <el-icon :size="13"><Clock /></el-icon>
        <span>{{ nowText }}</span>
      </div>
    </div>

    <!-- 多彩统计卡 -->
    <div class="stat-row">
      <div
        v-for="item in stats"
        :key="item.title"
        class="stat-card"
        :style="{ background: item.bg }"
      >
        <div class="stat-icon">
          <el-icon :size="20" color="#fff"><component :is="item.icon" /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-num">{{ item.value }}</div>
          <div class="stat-title">{{ item.title }}</div>
        </div>
      </div>
    </div>

    <!-- 图表行 1：登录趋势曲线 + 登录热力图 -->
    <div class="chart-row">
      <el-card shadow="never" class="panel trend-panel">
        <template #header>
          <div class="panel-title"><i class="dot" style="background:#409EFF" />近 7 天登录趋势</div>
        </template>
        <div ref="trendRef" class="chart-box" />
      </el-card>
      <el-card shadow="never" class="panel heat-panel">
        <template #header>
          <div class="panel-title"><i class="dot" style="background:#9C27B0" />登录热力分布</div>
        </template>
        <div ref="heatRef" class="chart-box" />
      </el-card>
    </div>

    <!-- 图表行 2：操作玫瑰 + 系统资源仪表 + 最新公告 -->
    <div class="chart-row">
      <el-card shadow="never" class="panel">
        <template #header>
          <div class="panel-title"><i class="dot" style="background:#67C23A" />操作类型分布</div>
        </template>
        <div ref="pieRef" class="chart-box" />
      </el-card>
      <el-card shadow="never" class="panel">
        <template #header>
          <div class="panel-title"><i class="dot" style="background:#F56C6C" />系统资源</div>
        </template>
        <div ref="gaugeRef" class="chart-box gauge-box" />
        <div class="server-meta">Go {{ serverInfo.goVersion }} · 运行 {{ serverInfo.runDays }}</div>
      </el-card>
      <el-card shadow="never" class="panel">
        <template #header>
          <div class="panel-title"><i class="dot" style="background:#E6A23C" />最新公告</div>
        </template>
        <div v-if="notices.length" class="notice-list">
          <div v-for="n in notices" :key="n.noticeId" class="notice-item" @click="openNotice(n)">
            <el-tag size="small" :type="n.noticeType === '1' ? 'primary' : 'warning'" effect="dark">
              {{ n.noticeType === '1' ? '通知' : '公告' }}
            </el-tag>
            <span class="notice-title">{{ n.noticeTitle }}</span>
            <span class="notice-time">{{
              n.createTime ? parseTime(n.createTime, 'MM-DD') : ''
            }}</span>
          </div>
        </div>
        <el-empty v-else description="暂无公告" :image-size="48" />

        <el-dialog v-model="detailVisible" title="公告详情" width="600px" destroy-on-close>
          <div v-if="detail.noticeTitle" class="notice-detail">
            <h3 class="detail-title">{{ detail.noticeTitle }}</h3>
            <div class="detail-meta">
              <el-tag size="small" :type="detail.noticeType === '1' ? 'primary' : 'warning'">
                {{ detail.noticeType === '1' ? '通知' : '公告' }}
              </el-tag>
              <span
                >发布：{{ detail.createBy || '-' }} · {{ parseTime(detail.createTime) || '-' }}</span
              >
            </div>
            <div class="detail-content">{{ detail.noticeContent }}</div>
          </div>
          <template #footer>
            <el-button @click="detailVisible = false">关闭</el-button>
          </template>
        </el-dialog>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'Dashboard' });

import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useUserStore } from '@/store/user';
import { listActiveNotice, getNotice } from '@/api/notice';
import { parseTime } from '@/utils/format';
import { Clock } from '@element-plus/icons-vue';
// ECharts 按需引入（折线/饼图/热力图/仪表盘 + 组件）
import * as echarts from 'echarts/core';
import { LineChart, PieChart, HeatmapChart, GaugeChart } from 'echarts/charts';
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  VisualMapComponent
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
echarts.use([
  LineChart,
  PieChart,
  HeatmapChart,
  GaugeChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  VisualMapComponent,
  CanvasRenderer
]);
import { getDashStats, getLoginTrend, getOperTypeStats, getLoginHeatmap } from '@/api/dashboard';
import { getServerInfo } from '@/api/server';

const userStore = useUserStore();
const trendRef = ref();
const pieRef = ref();
const heatRef = ref();
const gaugeRef = ref();
let trendChart = null;
let pieChart = null;
let heatChart = null;
let gaugeChart = null;
let darkObserver = null;
// 图表容器尺寸观察器：桌面布局下窗口拖拽/缩放不触发 window resize，用 ResizeObserver 兜底
let chartObserver = null;

// 图表数据（ref 驱动：接口与图表初始化顺序无关，渲染函数幂等读取）
const trendData = ref([]);
const pieData = ref([]);
const heatData = ref([]);

// 暗黑模式适配：图表文字/分割线颜色随主题切换
const isDark = () => document.documentElement.classList.contains('dark');
const chartTextColor = () => (isDark() ? '#e5eaf3' : '#606266');
// 图表分割线颜色
const chartSplitColor = () => (isDark() ? '#2b2b2b' : '#e4e7ed');

// 图表基础文本样式
function baseTextStyle() {
  return { color: chartTextColor() };
}

// 实时时钟
const nowText = ref('');
let clockTimer = null;
function updateClock() {
  const d = new Date();
  const pad = (x: number) => String(x).padStart(2, '0');
  nowText.value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

// 首页统计卡片（多彩渐变 + 数字滚动）
interface StatCard {
  title: string;
  icon: string;
  bg: string;
  value: number | string;
}
const stats = ref<StatCard[]>([
  { title: '用户数', icon: 'User', bg: 'linear-gradient(135deg,#409EFF,#6cb6ff)', value: '-' },
  { title: '系统角色', icon: 'Avatar', bg: 'linear-gradient(135deg,#67C23A,#95d475)', value: '-' },
  { title: '菜单权限', icon: 'Menu', bg: 'linear-gradient(135deg,#E6A23C,#f0c26a)', value: '-' },
  { title: '部门数', icon: 'OfficeBuilding', bg: 'linear-gradient(135deg,#F56C6C,#f89898)', value: '-' },
  { title: '公告数', icon: 'Bell', bg: 'linear-gradient(135deg,#909399,#b1b3b8)', value: '-' },
  { title: '今日登录', icon: 'SwitchButton', bg: 'linear-gradient(135deg,#9C27B0,#ce93d8)', value: '-' },
  { title: '在线用户', icon: 'Connection', bg: 'linear-gradient(135deg,#00BCD4,#4dd0e1)', value: '-' }
]);

// 数字滚动动画（无依赖，自实现缓动）
function animateNum(el: HTMLElement, target: number) {
  if (typeof target !== 'number' || Number.isNaN(target)) return;
  const dur = 800;
  const start = performance.now();
  const tick = (t: number) => {
    const p = Math.min(1, (t - start) / dur);
    el.textContent = String(Math.round(target * (1 - Math.pow(1 - p, 3))));
    if (p < 1) requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
}
function animateStats() {
  const els = document.querySelectorAll('.stat-num');
  stats.value.forEach((s, i) => {
    if (els[i]) animateNum(els[i] as HTMLElement, s.value as number);
  });
}

// 系统信息（仪表盘 + 底部元信息）
const serverInfo = ref({
  goVersion: '-',
  runDays: '-',
  cpuUsage: 0,
  memUsage: 0,
  diskUsage: 0
});

const notices = ref([]);
const detailVisible = ref(false);

// 公告详情对象（接口数据宽松类型，后续可细化）
interface NoticeDetail {
  noticeId?: number;
  noticeTitle?: string;
  noticeType?: string;
  createBy?: string;
  createTime?: string;
  noticeContent?: string;
}
const detail = ref({} as NoticeDetail);

// 打开公告详情
async function openNotice(n) {
  try {
    const res = await getNotice(n.noticeId);
    detail.value = res.data;
    detailVisible.value = true;
  } catch (e) {
    // 无详情权限时忽略
  }
}

// 渲染图表（等容器宽高都就绪再初始化，避免 0 高 canvas 塌陷）
function renderCharts() {
  const ensureSize = (el, done, tries = 0) => {
    if ((el.clientWidth > 0 && el.clientHeight > 0) || tries > 40) return done();
    requestAnimationFrame(() => ensureSize(el, done, tries + 1));
  };

  // 安全初始化：若容器在 init 时刻内部尺寸仍为 0（flex 布局未稳定残留），销毁重建一次
  const safeInit = (el) => {
    let c = echarts.init(el);
    if (c.getWidth() === 0 || c.getHeight() === 0) {
      c.dispose();
      c = echarts.init(el);
    }
    return c;
  };

  // 登录趋势折线（渐变面积 + 发光）
  ensureSize(trendRef.value, () => {
    trendChart = safeInit(trendRef.value);
    trendChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 36, right: 16, top: 28, bottom: 24 },
      xAxis: {
        type: 'category',
        data: trendData.value.map((i) => i.date),
        boundaryGap: false,
        axisLabel: baseTextStyle(),
        axisLine: { lineStyle: { color: chartSplitColor() } }
      },
      yAxis: {
        type: 'value',
        minInterval: 1,
        axisLabel: baseTextStyle(),
        splitLine: { lineStyle: { color: chartSplitColor() } }
      },
      series: [
        {
          name: '登录次数',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 7,
          itemStyle: { color: '#409EFF', borderColor: '#fff', borderWidth: 1.5 },
          lineStyle: { width: 3, shadowBlur: 12, shadowColor: 'rgba(64,158,255,0.6)' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(64,158,255,0.55)' },
              { offset: 1, color: 'rgba(64,158,255,0.02)' }
            ])
          },
          data: trendData.value.map((i) => i.count)
        }
      ]
    });
  });

  // 登录热力图（天 × 24 小时）
  ensureSize(heatRef.value, () => {
    heatChart = safeInit(heatRef.value);
    const hm = buildHeatMatrix(heatData.value);
    heatChart.setOption({
      tooltip: {
        position: 'top',
        formatter: (p) => `${p.value[1]} 时 · ${p.value[2]} 次登录`
      },
      grid: { left: 30, right: 10, top: 24, bottom: 52 },
      xAxis: {
        type: 'category',
        data: hm.labels,
        axisLabel: { ...baseTextStyle(), fontSize: 11 },
        axisLine: { lineStyle: { color: chartSplitColor() } }
      },
      yAxis: {
        type: 'category',
        data: Array.from({ length: 24 }, (_, i) => i),
        axisLabel: { ...baseTextStyle(), fontSize: 10, interval: 5, formatter: (v) => (v === 0 ? '0时' : v) },
        splitArea: { show: true }
      },
      visualMap: {
        min: 0,
        max: hm.max,
        calculable: true,
        orient: 'horizontal',
        left: 'center',
        bottom: 4,
        textStyle: { color: chartTextColor(), fontSize: 10 },
        inRange: { color: ['#0b1a33', '#1e5aa8', '#7c3aed', '#db2777', '#f59e0b'] }
      },
      series: [
        {
          type: 'heatmap',
          data: hm.matrix,
          label: { show: false },
          itemStyle: { borderColor: isDark() ? '#141414' : '#fff', borderWidth: 1.5 }
        }
      ]
    });
  });

  // 操作类型玫瑰图
  ensureSize(pieRef.value, () => {
    pieChart = safeInit(pieRef.value);
    pieChart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} 次 ({d}%)' },
      legend: { bottom: 0, textStyle: baseTextStyle(), itemWidth: 10, itemHeight: 10 },
      color: ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#9C27B0', '#00BCD4'],
      series: [
        {
          name: '操作类型',
          type: 'pie',
          roseType: 'radius',
          radius: ['22%', '66%'],
          center: ['50%', '45%'],
          itemStyle: { borderRadius: 4, borderColor: 'transparent', borderWidth: 2 },
          label: { show: false },
          emphasis: { scaleSize: 6, label: { show: true, fontWeight: 'bold', formatter: '{b}\n{d}%' } },
          data: pieData.value
        }
      ]
    });
  });

  // 系统资源仪表盘（CPU / 内存 / 磁盘 三表合一）
  ensureSize(gaugeRef.value, () => {
    gaugeChart = safeInit(gaugeRef.value);
    const mkGauge = (center, color) => ({
      type: 'gauge',
      center,
      radius: '82%',
      startAngle: 220,
      endAngle: -40,
      min: 0,
      max: 100,
      splitNumber: 4,
      axisLine: { lineStyle: { width: 9, color: [[1, chartSplitColor()]] } },
      progress: { show: true, width: 9, itemStyle: { color } },
      pointer: { show: false },
      axisTick: { show: false },
      splitLine: { length: 6, lineStyle: { width: 1, color: chartSplitColor() } },
      axisLabel: { show: false },
      detail: {
        valueAnimation: true,
        formatter: (v) => Math.round(v) + '%',
        color: chartTextColor(),
        fontSize: 15,
        fontWeight: 700,
        offsetCenter: [0, '72%']
      },
      title: { show: true, offsetCenter: [0, '100%'], color: chartTextColor(), fontSize: 11 },
      data: [{ value: 0, name: '' }]
    });
    gaugeChart.setOption({
      series: [
        { ...mkGauge(['20%', '52%'], '#409EFF'), data: [{ value: 0, name: 'CPU' }] },
        { ...mkGauge(['50%', '52%'], '#67C23A'), data: [{ value: 0, name: '内存' }] },
        { ...mkGauge(['80%', '52%'], '#F56C6C'), data: [{ value: 0, name: '磁盘' }] }
      ]
    });
    // 数据可能已先到（接口快于图表初始化）：补应用
    applyGaugeData();
  });
}

// 应用系统资源数据到仪表盘（chart 未就绪时仅更新 ref，就绪后自动补齐）
function applyGaugeData() {
  if (!gaugeChart) return;
  gaugeChart.setOption({
    series: [
      { data: [{ value: serverInfo.value.cpuUsage, name: 'CPU' }] },
      { data: [{ value: serverInfo.value.memUsage, name: '内存' }] },
      { data: [{ value: serverInfo.value.diskUsage, name: '磁盘' }] }
    ]
  });
}

// 图表自适应缩放
function resizeCharts() {
  trendChart?.resize();
  pieChart?.resize();
  heatChart?.resize();
  gaugeChart?.resize();
}

// 组装热力图矩阵（7 天 × 24 小时，从接口原始数据）
function buildHeatMatrix(heat) {
  const days = 7;
  const fmtFull = (i: number) => {
    const d = new Date();
    d.setDate(d.getDate() - i);
    const pad = (x: number) => String(x).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  };
  const fullDates = Array.from({ length: days }, (_, i) => fmtFull(i));
  const labels = fullDates.map((f) => f.slice(5).replace('-', '-'));
  const map = {};
  const list = Array.isArray(heat) ? heat : [];
  list.forEach((item) => (map[item.date + ':' + item.hour] = item.count));
  const matrix = [];
  let max = 1;
  fullDates.forEach((fd, di) => {
    for (let hour = 0; hour < 24; hour++) {
      const v = map[fd + ':' + hour] || 0;
      if (v > max) max = v;
      matrix.push([di, hour, v]);
    }
  });
  return { labels, matrix, max };
}

// 接口响应可能是业务体 {code,data,...} 或直接数组，统一提取数组
function extractList(r) {
  if (Array.isArray(r)) return r;
  if (r && Array.isArray(r.data)) return r.data;
  return [];
}

// 幂等渲染：各图表实例存在时应用最新 ref 数据（接口完成或图表初始化后调用均安全）
function renderData() {
  if (trendChart && trendData.value.length) {
    trendChart.setOption({
      xAxis: { data: trendData.value.map((i) => i.date) },
      series: [{ data: trendData.value.map((i) => i.count) }]
    });
  }
  if (pieChart && pieData.value.length) {
    pieChart.setOption({ series: [{ data: pieData.value }] });
  }
  if (heatChart && heatData.value.length) {
    const hm = buildHeatMatrix(heatData.value);
    heatChart.setOption({
      xAxis: { data: hm.labels },
      visualMap: { max: hm.max },
      series: [{ data: hm.matrix }]
    });
  }
  applyGaugeData();
}

// 加载图表数据（各接口独立容错，结果存 ref 后幂等渲染）
async function loadChartData() {
  try {
    const [trend, pie, heat] = await Promise.all([
      getLoginTrend(),
      getOperTypeStats(),
      getLoginHeatmap()
    ]);
    trendData.value = extractList(trend);
    pieData.value = extractList(pie);
    heatData.value = extractList(heat);
  } catch (e) {
    // 任一接口失败仅跳过对应图表
    try {
      trendData.value = extractList(await getLoginTrend());
    } catch (e2) {
      /* 无权限时留空 */
    }
    try {
      pieData.value = extractList(await getOperTypeStats());
    } catch (e2) {
      /* 无权限时留空 */
    }
    try {
      heatData.value = extractList(await getLoginHeatmap());
    } catch (e2) {
      /* 无权限时留空 */
    }
  }
  renderData();
}

// 系统信息数据
async function loadServerInfo() {
  try {
    const res = await getServerInfo();
    const d = res.data;
    serverInfo.value = {
      goVersion: d.goVersion || '-',
      runDays: d.runDays !== undefined ? `${d.runDays}天 ${d.runHours ?? 0}小时` : '-',
      cpuUsage: d.cpuUsage !== undefined ? Math.round(d.cpuUsage) : 0,
      memUsage: d.memUsage !== undefined ? Math.round(d.memUsage) : 0,
      diskUsage: d.diskUsage !== undefined ? Math.round(d.diskUsage) : 0
    };
    renderData();
  } catch (e) {
    // 无监控权限时保留占位
  }
}

onMounted(async () => {
  updateClock();
  clockTimer = setInterval(updateClock, 1000);
  // 监听暗黑模式切换：重绘图表文字/分割线颜色
  darkObserver = new MutationObserver(() => {
    const refresh = () => {
      trendChart?.setOption({
        xAxis: { axisLabel: baseTextStyle(), axisLine: { lineStyle: { color: chartSplitColor() } } },
        yAxis: { axisLabel: baseTextStyle(), splitLine: { lineStyle: { color: chartSplitColor() } } }
      });
      pieChart?.setOption({ legend: { textStyle: baseTextStyle() } });
      heatChart?.setOption({
        xAxis: { axisLabel: baseTextStyle(), axisLine: { lineStyle: { color: chartSplitColor() } } },
        yAxis: { axisLabel: baseTextStyle() },
        visualMap: { textStyle: { color: chartTextColor() } },
        series: [{ itemStyle: { borderColor: isDark() ? '#141414' : '#fff' } }]
      });
      gaugeChart?.setOption({
        series: [
          { axisLine: { lineStyle: { color: [[1, chartSplitColor()]] } } },
          { axisLine: { lineStyle: { color: [[1, chartSplitColor()]] } } },
          { axisLine: { lineStyle: { color: [[1, chartSplitColor()]] } } }
        ]
      });
    };
    refresh();
    requestAnimationFrame(refresh);
  });
  darkObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
  // 统计卡片（后端汇总接口，失败时回退到逐页统计）
  try {
    const res = await getDashStats();
    const s = res.data;
    stats.value = [
      { title: '用户数', icon: 'User', bg: 'linear-gradient(135deg,#409EFF,#6cb6ff)', value: s.userCount },
      { title: '系统角色', icon: 'Avatar', bg: 'linear-gradient(135deg,#67C23A,#95d475)', value: s.roleCount },
      { title: '菜单权限', icon: 'Menu', bg: 'linear-gradient(135deg,#E6A23C,#f0c26a)', value: s.menuCount },
      { title: '部门数', icon: 'OfficeBuilding', bg: 'linear-gradient(135deg,#F56C6C,#f89898)', value: s.deptCount },
      { title: '公告数', icon: 'Bell', bg: 'linear-gradient(135deg,#909399,#b1b3b8)', value: s.noticeCount },
      { title: '今日登录', icon: 'SwitchButton', bg: 'linear-gradient(135deg,#9C27B0,#ce93d8)', value: s.todayLogin },
      { title: '在线用户', icon: 'Connection', bg: 'linear-gradient(135deg,#00BCD4,#4dd0e1)', value: s.onlineCount }
    ];
    nextTick(animateStats);
  } catch (e) {
    try {
      const [u, r, m] = await Promise.allSettled([
        import('@/api/user').then((mod) => mod.listUser({ pageNum: 1, pageSize: 1 })),
        import('@/api/role').then((mod) => mod.listRole({ pageNum: 1, pageSize: 1 })),
        import('@/api/menu').then((mod) => mod.listMenu())
      ]);
      if (u.status === 'fulfilled') stats.value[0].value = u.value.total;
      if (r.status === 'fulfilled') stats.value[1].value = r.value.total;
      if (m.status === 'fulfilled') {
        const count = (nodes) =>
          nodes.reduce((s, n) => s + 1 + (n.children ? count(n.children) : 0), 0);
        stats.value[2].value = count(m.value.data);
      }
      nextTick(animateStats);
    } catch (e2) {
      // 无权限时保持占位符
    }
  }
  try {
    const res = await listActiveNotice();
    notices.value = res.data || [];
  } catch (e) {
    // 无公告权限时留空
  }
  renderCharts();
  loadChartData();
  loadServerInfo();
  window.addEventListener('resize', resizeCharts);
  // 桌面布局：首页运行在可拖拽缩放的桌面窗口内，窗口尺寸变化不触发 window resize，
  // 用 ResizeObserver 观察各图表容器，尺寸变化即重算图表
  const boxes = [trendRef.value, heatRef.value, pieRef.value, gaugeRef.value].filter(Boolean);
  if (boxes.length && typeof ResizeObserver !== 'undefined') {
    chartObserver = new ResizeObserver(() => resizeCharts());
    boxes.forEach((el) => chartObserver.observe(el));
  }
  nextTick(() => resizeCharts());
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts);
  chartObserver?.disconnect();
  if (clockTimer) clearInterval(clockTimer);
  darkObserver?.disconnect();
  trendChart?.dispose();
  pieChart?.dispose();
  heatChart?.dispose();
  gaugeChart?.dispose();
});
</script>

<style scoped>
/* 大屏容器：一屏撑满，无滚动条 */
.dash {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
  /* 容器查询：桌面布局下页面在可缩放的窗口内，按窗口宽度（而非视口）做响应式 */
  container-type: inline-size;
  container-name: dash;
}

/* 顶栏：标题 + 实时时间 */
.dash-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 0 14px;
  height: 38px;
  border-radius: 2px;
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.14), rgba(64, 158, 255, 0.04));
  border: 1px solid var(--go-border-color, #e4e7ed);
}

.dash-bar-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.dash-bar-logo {
  font-size: 15px;
  font-weight: 700;
  background: linear-gradient(90deg, #409eff, #9c27b0);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.dash-bar-sub {
  font-size: 12px;
  color: var(--go-text-secondary);
}

.dash-bar-time {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--go-text-secondary);
  font-variant-numeric: tabular-nums;
}

/* 多彩统计卡 */
.stat-row {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 10px;
}

@container dash (max-width: 760px) {
  .stat-row {
    grid-template-columns: repeat(4, 1fr);
  }
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 2px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(31, 45, 61, 0.14);
  transition: transform 0.2s, box-shadow 0.2s;
  min-width: 0;
}

.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 20px rgba(31, 45, 61, 0.2);
}

.stat-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.22);
}

.stat-body {
  min-width: 0;
}

.stat-num {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
}

.stat-title {
  font-size: 11px;
  opacity: 0.92;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 图表行 */
.chart-row {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 12px;
}

.trend-panel {
  grid-column: span 2;
}

@container dash (max-width: 760px) {
  .chart-row {
    grid-template-columns: 1fr;
    /* 窄容器下不再平分高度：面板按最小高度撑开，整页滚动，图表完整可读 */
    flex: 0 0 auto;
  }
  .chart-row .panel {
    min-height: 220px;
  }
  /* 窄容器下整页纵向滚动，保证各图表完整显示 */
  .dash {
    overflow-y: auto;
  }
}

.panel {
  min-width: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  border-radius: 2px;
}

.panel :deep(.el-card__header) {
  padding: 10px 14px;
}

.panel :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  padding: 8px 12px 12px;
  display: flex;
  flex-direction: column;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 13px;
  color: var(--go-text-primary);
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 6px currentColor;
}

.chart-box {
  flex: 1;
  min-height: 0;
}

.gauge-box {
  flex: 1;
}

.server-meta {
  flex-shrink: 0;
  text-align: center;
  font-size: 11px;
  color: var(--go-text-secondary);
  padding-top: 2px;
}

/* 公告列表（表格风格） */
.notice-list {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.notice-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 2px;
  cursor: pointer;
  border-bottom: 1px solid var(--go-border-regular);
}

.notice-item:last-child {
  border-bottom: none;
}

.notice-item:hover {
  background: var(--go-fill-light);
}

.notice-title {
  flex: 1;
  min-width: 0;
  color: var(--go-text-primary);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-time {
  font-size: 11px;
  color: var(--go-text-secondary);
  white-space: nowrap;
  flex-shrink: 0;
}

.notice-detail .detail-title {
  margin: 0 0 10px;
  font-size: 18px;
  color: var(--go-text-primary);
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  font-size: 12px;
  color: var(--go-text-secondary);
}

.detail-content {
  line-height: 1.8;
  color: var(--go-text-regular);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
