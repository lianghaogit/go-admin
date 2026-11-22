<template>
  <div class="app-container">
    <!-- 数据库概览 -->
    <el-row :gutter="12" class="stat-row">
      <el-col v-for="s in stats" :key="s.label" :xs="12" :sm="8" :md="4">
        <el-card class="stat-card">
          <div class="stat-label">{{ s.label }}</div>
          <div class="stat-value" :style="{ color: s.color }">{{ s.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <div class="toolbar">
        <span class="title">数据表统计（共 {{ tableCount }} 张表）</span>
        <el-button type="primary" :icon="Refresh" @click="fetchInfo">刷新</el-button>
      </div>
      <el-table v-table-height v-loading="loading" :data="tables" border stripe size="small">
        <el-table-column type="index" label="#" width="60" />
        <el-table-column prop="tableName" label="表名" min-width="200" />
        <el-table-column prop="rowCount" label="行数" width="140" align="right" />
        <el-table-column label="说明" min-width="220">
          <template #default="{ row }">
            <span class="table-desc">{{ tableDesc(row.tableName) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorDb' });

import { computed, onMounted, ref } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { getDbInfo } from '@/api/db';

const loading = ref(false);

// 数据监控信息对象（接口数据宽松类型，后续可细化）
interface DbTableRow {
  tableName: string;
  rowCount: number;
}
interface DbInfo {
  dbType?: string;
  version?: string;
  versionRaw?: string;
  fileSize?: number;
  dbSize?: number;
  journalMode?: string;
  connections?: number;
  maxConnections?: number;
  cacheHitRatio?: number;
  idleInTransaction?: number;
  totalRows?: number;
  tables?: DbTableRow[];
  tableCount?: number;
}
const info = ref({} as DbInfo);

// 概览卡片：按方言展示对应指标（PostgreSQL 展示连接数/缓存命中率，SQLite 展示文件大小/日志模式）
const stats = computed(() => {
  const d = info.value;
  const isPg = d.dbType === 'postgres';
  const list = [
    { label: '数据库类型', value: d.dbType ? d.dbType.toUpperCase() : '-', color: '#409EFF' },
    { label: '版本', value: d.version || '-', color: '#67C23A' },
    { label: '数据库大小', value: fmtBytes(d.dbSize ?? d.fileSize), color: '#E6A23C' }
  ];
  if (isPg) {
    list.push({
      label: '连接数 / 上限',
      value: `${d.connections ?? '-'} / ${d.maxConnections ?? '-'}`,
      color: '#909399'
    });
    list.push({
      label: '缓存命中率',
      value: d.cacheHitRatio != null ? `${d.cacheHitRatio}%` : '-',
      color: '#909399'
    });
    list.push({ label: '总行数', value: d.totalRows != null ? String(d.totalRows) : '-', color: '#F56C6C' });
  } else {
    list.push({ label: '日志模式', value: d.journalMode || '-', color: '#909399' });
    list.push({ label: '总行数', value: d.totalRows != null ? String(d.totalRows) : '-', color: '#F56C6C' });
  }
  return list;
});

const tables = computed(() => info.value.tables || []);
const tableCount = computed(() => info.value.tableCount ?? 0);

const DESCS = {
  sys_user: '用户表',
  sys_role: '角色表',
  sys_menu: '菜单表',
  sys_dept: '部门表',
  sys_dict_type: '字典类型',
  sys_dict_data: '字典数据',
  sys_post: '岗位表',
  sys_config: '参数配置',
  sys_notice: '通知公告',
  sys_online: '在线会话',
  sys_user_role: '用户角色关联',
  sys_user_post: '用户岗位关联',
  sys_role_menu: '角色菜单关联',
  sys_logininfor: '登录日志',
  sys_oper_log: '操作日志',
  sys_job: '定时任务',
  sys_job_log: '任务执行日志',
  sys_api: '接口权限表',
  sys_gen_table: '代码生成表',
  flow_definition: '工作流定义',
  flow_instance: '工作流实例',
  flow_task: '工作流任务',
  flow_task_log: '任务处理日志',
  flow_cc: '抄送记录',
  flow_comment: '审批意见'
};

// 数据库表描述格式化
function tableDesc(name) {
  return DESCS[name] || '业务表';
}

// 字节数格式化
function fmtBytes(b) {
  if (b === undefined || b === null) return '-';
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  if (b < 1024 * 1024 * 1024) return `${(b / 1024 / 1024).toFixed(1)} MB`;
  return `${(b / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

// 获取监控信息
async function fetchInfo() {
  loading.value = true;
  try {
    const res = await getDbInfo();
    info.value = res.data;
  } catch (e) {
    // 拦截器已提示
  } finally {
    loading.value = false;
  }
}

onMounted(fetchInfo);
</script>

<style scoped>
.stat-row {
  margin-bottom: 14px;
}

.stat-card {
  text-align: center;
  padding: 4px 0;
}

.stat-label {
  font-size: 12px;
  color: var(--go-text-secondary);
  margin-bottom: 6px;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.title {
  flex: 1;
  font-weight: 600;
  color: var(--go-text-primary);
}

.table-desc {
  color: var(--go-text-secondary);
  font-size: 12px;
}
</style>
