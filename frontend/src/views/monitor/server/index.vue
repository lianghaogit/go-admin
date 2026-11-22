<template>
  <div class="app-container">
    <!-- 使用率仪表盘 -->
    <el-row :gutter="16">
      <el-col v-for="item in gauges" :key="item.title" :xs="24" :sm="8">
        <el-card class="gauge-card">
          <el-progress type="dashboard" :percentage="item.value" :color="item.color" :width="140">
            <template #default="{ percentage }">
              <span class="gauge-num">{{ percentage }}%</span>
              <div class="gauge-title">{{ item.title }}</div>
            </template>
          </el-progress>
          <div class="gauge-detail">
            <span>已用：{{ fmtSize(item.used) }}</span>
            <span>总量：{{ fmtSize(item.total) }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统信息 -->
    <el-card class="info-card">
      <template #header>
        <span class="info-header">服务器信息</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="操作系统">{{ info.os || '-' }}</el-descriptions-item>
        <el-descriptions-item label="主机名">{{ info.hostName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="CPU 核数">{{ info.cpuNum || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Go 版本">{{ info.goVersion || '-' }}</el-descriptions-item>
        <el-descriptions-item label="系统架构">{{ info.arch || '-' }}</el-descriptions-item>
        <el-descriptions-item label="内核版本">{{ info.kernel || '-' }}</el-descriptions-item>
        <el-descriptions-item label="服务启动时间">{{
          parseTime(info.startTime) || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="运行时长">
          {{
            info.runDays !== undefined
              ? `${info.runDays}天 ${info.runHours}小时 ${info.runMinutes}分`
              : '-'
          }}
        </el-descriptions-item>
        <el-descriptions-item label="进程内存" :span="2">
          {{
            info.procMemUsed !== undefined
              ? `${fmtSize(info.procMemUsed)}（已分配） / ${fmtSize(info.procMemTotal)}（系统）`
              : '-'
          }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorServer' });

import { computed, onMounted, onUnmounted, ref } from 'vue';
import { getServerInfo } from '@/api/server';
import { parseTime } from '@/utils/format';

// 服务监控信息对象（接口数据宽松类型，后续可细化）
interface ServerInfo {
  os?: string;
  hostName?: string;
  cpuNum?: number;
  goVersion?: string;
  arch?: string;
  kernel?: string;
  startTime?: string;
  runDays?: number;
  runHours?: number;
  runMinutes?: number;
  procMemUsed?: number;
  procMemTotal?: number;
  cpuUsage?: number;
  memUsage?: number;
  memUsed?: number;
  memTotal?: number;
  diskUsage?: number;
  diskUsed?: number;
  diskTotal?: number;
}
const info = ref({} as ServerInfo);
let timer = null;

const gauges = computed(() => [
  {
    title: 'CPU 使用率',
    value: Math.round(info.value.cpuUsage ?? 0),
    used: 0,
    total: 0,
    color: '#409EFF'
  },
  {
    title: '内存使用率',
    value: Math.round(info.value.memUsage ?? 0),
    used: info.value.memUsed ?? 0,
    total: info.value.memTotal ?? 0,
    color: '#67C23A'
  },
  {
    title: '磁盘使用率',
    value: Math.round(info.value.diskUsage ?? 0),
    used: info.value.diskUsed ?? 0,
    total: info.value.diskTotal ?? 0,
    color: '#E6A23C'
  }
]);

// 字节数格式化
function fmtSize(bytes) {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let v = bytes;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(1)} ${units[i]}`;
}

// 获取监控信息
async function fetchInfo() {
  try {
    const res = await getServerInfo();
    info.value = res.data;
  } catch (e) {
    // 无权限时保持空
  }
}

onMounted(() => {
  fetchInfo();
  timer = setInterval(fetchInfo, 30000);
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<style scoped>
.gauge-card {
  text-align: center;
}

.gauge-num {
  font-size: 22px;
  font-weight: 600;
  display: block;
}

.gauge-title {
  font-size: 13px;
  color: var(--go-text-secondary);
  margin-top: 4px;
}

.gauge-detail {
  margin-top: 12px;
  display: flex;
  justify-content: center;
  gap: 18px;
  font-size: 12px;
  color: var(--go-text-regular);
}

.info-card {
  margin-top: 16px;
}

.info-header {
  font-weight: 600;
  color: var(--go-text-primary);
}
</style>
