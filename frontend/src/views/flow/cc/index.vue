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
        <span class="page-desc">抄送给我查看的申请</span>
      </div>

      <el-table v-table-height v-loading="loading" :data="ccList" border stripe>
        <el-table-column prop="flowTaskId" label="编号" width="70" />
        <el-table-column prop="title" label="申请标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="defName" label="流程" width="120" />
        <el-table-column prop="applicantName" label="申请人" width="100" />
        <el-table-column label="阅读状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === '0' ? 'danger' : 'info'">
              {{ row.status === '0' ? '未读' : '已读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="抄送时间"
          width="165"
        />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === '0'"
              type="success"
              link
              :icon="Check"
              @click="handleRead(row)"
            >
              标记已读
            </el-button>
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
    <el-dialog v-model="detailVisible" title="抄送详情" width="820px" destroy-on-close top="6vh">
      <FlowDetail :detail="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'FlowCc' });

import { onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { View, Check, Search, Refresh } from '@element-plus/icons-vue';
import { listFlowCc, readFlowCc, getFlowInstance } from '@/api/flow';
import FlowDetail from '@/views/flow/FlowDetail.vue';
import { parseTime } from '@/utils/format';

const loading = ref(false);
const ccList = ref([]);
const total = ref(0);
const detailVisible = ref(false);
const detailData = ref(null);

const queryParams = reactive({ pageNum: 1, pageSize: 10, keyword: '' });

async function fetchList() {
  loading.value = true;
  try {
    const res = await listFlowCc(queryParams);
    ccList.value = res.rows;
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

// 标记已读
async function handleRead(row) {
  await readFlowCc(row.flowTaskId);
  ElMessage.success('已读');
  fetchList();
}

// 详情（查看时自动标记已读）
async function openDetail(row) {
  const res = await getFlowInstance(row.instanceId);
  detailData.value = res.data;
  detailVisible.value = true;
  if (row.status === '0') {
    try {
      await readFlowCc(row.flowTaskId);
      fetchList();
    } catch (e) {
      // 忽略标记失败
    }
  }
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
