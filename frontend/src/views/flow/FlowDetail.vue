<template>
  <div v-if="detail" class="flow-detail">
    <!-- 实例基本信息 -->
    <el-descriptions :column="2" border size="small">
      <el-descriptions-item label="申请标题">{{ detail.instance.title }}</el-descriptions-item>
      <el-descriptions-item label="流程">{{ detail.instance.defName }}</el-descriptions-item>
      <el-descriptions-item label="申请人">{{ detail.instance.applicantName }}</el-descriptions-item>
      <el-descriptions-item label="申请时间">{{ parseTime(detail.instance.createTime) }}</el-descriptions-item>
      <el-descriptions-item label="当前状态" :span="2">
        <el-tag :type="statusTagType(detail.instance.status)">{{ statusText(detail.instance.status) }}</el-tag>
        <el-tag
          v-if="detail.instance.status === '3'"
          type="danger"
          size="small"
          style="margin-left: 6px"
        >
          驳回节点：{{ rejectFromName }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="申请内容" :span="2">
        <div class="flow-content">{{ detail.instance.content || '无' }}</div>
      </el-descriptions-item>
    </el-descriptions>

    <!-- 节点进度 -->
    <el-steps :active="activeStep" align-center finish-status="success" style="margin: 16px 0">
      <el-step
        v-for="n in detail.nodes"
        :key="n.flowNodeId"
        :title="n.nodeName"
        :description="nodeDesc(n)"
      />
    </el-steps>

    <!-- 任务列表 -->
    <template v-if="detail.tasks.length">
      <div class="flow-section-title">任务记录</div>
      <el-table :data="detail.tasks" border stripe size="small">
        <el-table-column prop="nodeName" label="节点" width="140" />
        <el-table-column prop="assigneeName" label="处理人" width="120" />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            {{ row.taskType === '1' ? '审批' : row.taskType === '2' ? '抄送' : '加签' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="taskTagType(row.status)">{{ taskStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="意见" show-overflow-tooltip />
        <el-table-column label="处理时间" width="160">
          <template #default="{ row }">{{ parseTime(row.passedAt) || parseTime(row.updateTime) }}</template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 审批时间线 -->
    <div class="flow-section-title">审批时间线</div>
    <el-timeline v-if="detail.records.length" style="padding-left: 4px">
      <el-timeline-item
        v-for="r in detail.records"
        :key="r.flowRecordId"
        :type="recordTagType(r.action)"
        :timestamp="parseTime(r.createTime)"
      >
        <b>{{ r.operatorName }}</b> {{ recordActionText(r) }}
        <span v-if="r.targetName">→ {{ r.targetName }}</span>
        <span v-if="r.comment">：{{ r.comment }}</span>
      </el-timeline-item>
    </el-timeline>
    <el-empty v-else description="暂无记录" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { parseTime } from '@/utils/format';

const props = defineProps<{
  detail: any;
}>();

const detail = computed(() => props.detail);

// 当前激活步骤（通过的节点数，节点顺序从 1 起）
const activeStep = computed(() => {
  if (!detail.value) return 0;
  const cur = detail.value.instance.currentOrder;
  return cur > 0 ? cur - 1 : 0;
});

// 驳回来源节点名
const rejectFromName = computed(() => {
  const d = detail.value;
  if (!d) return '';
  const n = d.nodes.find((x: any) => x.flowNodeId === d.instance.rejectFromNodeId);
  return n ? n.nodeName : '-';
});

function nodeDesc(n: any): string {
  const mode = n.nodeMode === 'any' ? '会签' : '或签';
  let who = '';
  if (n.approverType === 'user') who = '指定人员';
  else if (n.approverType === 'role') who = '按角色';
  else who = '部门负责人';
  return `${mode}·${who}`;
}

function statusText(s: string): string {
  const m: Record<string, string> = {
    '1': '运行中',
    '2': '已通过',
    '3': '已驳回',
    '4': '已撤销',
    '5': '已终止'
  };
  return m[s] || s;
}

function statusTagType(s: string): string {
  const m: Record<string, string> = {
    '1': 'primary',
    '2': 'success',
    '3': 'danger',
    '4': 'info',
    '5': 'warning'
  };
  return m[s] || 'info';
}

function taskStatusText(s: string): string {
  const m: Record<string, string> = {
    '0': '待办',
    '1': '已通过',
    '2': '已驳回',
    '3': '已转出',
    '4': '已加签',
    '5': '已撤销',
    '6': '已读'
  };
  return m[s] || s;
}

function taskTagType(s: string): string {
  const m: Record<string, string> = {
    '0': 'warning',
    '1': 'success',
    '2': 'danger',
    '3': 'info',
    '4': 'info',
    '5': 'info',
    '6': 'info'
  };
  return m[s] || 'info';
}

function recordTagType(action: string): string {
  const m: Record<string, string> = {
    '1': 'primary',
    '2': 'success',
    '3': 'danger',
    '4': 'warning',
    '5': 'warning',
    '6': 'info',
    '7': 'info',
    '8': 'warning',
    '9': 'warning',
    '10': 'primary',
    '11': 'info',
    '12': 'warning'
  };
  return m[action] || 'info';
}

function recordActionText(r: any): string {
  const m: Record<string, string> = {
    '1': '发起申请',
    '2': '审批通过',
    '3': '审批驳回',
    '4': '转办给',
    '5': '加签给',
    '6': '撤回申请',
    '7': '撤销',
    '8': '终止流程',
    '9': '发起催办',
    '10': '重新提交',
    '11': '抄送给',
    '12': '退回至'
  };
  return m[r.action] || r.action;
}
</script>

<style scoped>
.flow-content {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
}

.flow-section-title {
  font-weight: 600;
  margin: 12px 0 8px;
}
</style>
