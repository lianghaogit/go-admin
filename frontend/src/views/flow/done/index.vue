<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="申请标题">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入申请标题"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <span class="page-desc">我处理过的任务（通过/驳回/转出）</span>
      </div>

      <el-table v-table-height v-loading="loading" :data="doneList" border stripe>
        <el-table-column prop="flowTaskId" label="编号" width="70" />
        <el-table-column prop="title" label="申请标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="defName" label="流程" width="120" />
        <el-table-column prop="applicantName" label="申请人" width="100" />
        <el-table-column prop="nodeName" label="节点" width="130" />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            {{ row.taskType === '3' ? '加签' : '审批' }}
          </template>
        </el-table-column>
        <el-table-column label="处理结果" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="resultTagType(row.status)">
              {{ resultText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="我的意见" min-width="120" show-overflow-tooltip />
        <el-table-column label="处理时间" width="165">
          <template #default="{ row }">{{ parseTime(row.passedAt) || parseTime(row.updateTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="View" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="queryParams.pageNum"
        v-model:page-size="queryParams.pageSize"
        class="pagination"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </el-card>

    <!-- 实例详情 -->
    <el-dialog v-model="detailVisible" title="任务详情" width="820px" destroy-on-close top="6vh">
      <FlowDetail :detail="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'FlowDone' });

import { onMounted, reactive, ref } from 'vue';
import { View, Search, Refresh } from '@element-plus/icons-vue';
import { listFlowDone, getFlowInstance } from '@/api/flow';
import FlowDetail from '@/views/flow/FlowDetail.vue';
import { parseTime } from '@/utils/format';

const loading = ref(false);
const doneList = ref([]);
const total = ref(0);
const detailVisible = ref(false);
const detailData = ref(null);

const queryParams = reactive({ pageNum: 1, pageSize: 10, keyword: '' });

function resultText(s: string): string {
  const m: Record<string, string> = {
    '1': '已通过',
    '2': '已驳回',
    '3': '已转出'
  };
  return m[s] || s;
}

function resultTagType(s: string): string {
  const m: Record<string, string> = { '1': 'success', '2': 'danger', '3': 'info' };
  return m[s] || 'info';
}

async function fetchList() {
  loading.value = true;
  try {
    const res = await listFlowDone(queryParams);
    doneList.value = res.rows;
    total.value = res.total;
  } finally {
    loading.value = false;
  }
}

function handleQuery() {
  queryParams.pageNum = 1;
  fetchList();
}

function resetQuery() {
  queryParams.keyword = '';
  handleQuery();
}

async function openDetail(row) {
  const res = await getFlowInstance(row.instanceId);
  detailData.value = res.data;
  detailVisible.value = true;
}

onMounted(fetchList);
</script>

<style scoped>
.search-form {
  margin-bottom: 12px;
}

.table-toolbar {
  margin-bottom: 12px;
}

.page-desc {
  color: var(--go-text-secondary);
  font-size: 13px;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
