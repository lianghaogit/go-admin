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
        <span class="page-desc">当前共 {{ total }} 条待办</span>
      </div>

      <el-table v-table-height v-loading="loading" :data="todoList" border stripe>
        <el-table-column prop="flowTaskId" label="编号" width="70" />
        <el-table-column prop="title" label="申请标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="defName" label="流程" width="120" />
        <el-table-column prop="applicantName" label="申请人" width="100" />
        <el-table-column prop="nodeName" label="当前节点" width="130" />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.taskType === '3' ? 'warning' : 'primary'">
              {{ row.taskType === '3' ? '加签' : '审批' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="超时" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.overdueAt && isOverdue(row.overdueAt)" type="danger" size="small">已超时</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="到达时间"
          width="165"
        />
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="View" @click="openDetail(row)">详情</el-button>
            <el-button
              v-hasPermi="['flow:todo:approve']"
              type="success"
              link
              :icon="Check"
              @click="openAction(row, 'approve')"
            >
              通过
            </el-button>
            <el-button
              v-hasPermi="['flow:todo:reject']"
              type="danger"
              link
              :icon="Close"
              @click="openAction(row, 'reject')"
            >
              驳回
            </el-button>
            <el-button
              v-hasPermi="['flow:todo:transfer']"
              type="warning"
              link
              :icon="Right"
              @click="openAction(row, 'transfer')"
            >
              转办
            </el-button>
            <el-button
              v-hasPermi="['flow:todo:addSign']"
              type="warning"
              link
              :icon="Plus"
              @click="openAction(row, 'addSign')"
            >
              加签
            </el-button>
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

    <!-- 审批操作弹窗 -->
    <el-dialog
      v-model="actionVisible"
      :title="actionTitle"
      width="520px"
      destroy-on-close
    >
      <el-form ref="actionFormRef" :model="actionForm" :rules="actionRules" label-width="90px">
        <el-form-item label="申请">
          <div class="action-title">{{ actionTask?.title }}</div>
        </el-form-item>
        <el-form-item v-if="actionType === 'transfer' || actionType === 'addSign'" label="目标人" prop="toUserId">
          <el-select
            v-model="actionForm.toUserId"
            filterable
            placeholder="请选择用户"
            style="width: 100%"
          >
            <el-option
              v-for="u in userOptions"
              :key="u.userId"
              :label="u.nickName + (u.deptName ? '（' + u.deptName + '）' : '')"
              :value="u.userId"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="审批意见" prop="comment">
          <el-input
            v-model="actionForm.comment"
            type="textarea"
            :rows="3"
            :placeholder="actionType === 'approve' ? '同意（可填写意见）' : '请填写意见'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="actionVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmitAction"> 确定 </el-button>
      </template>
    </el-dialog>

    <!-- 实例详情 -->
    <el-dialog v-model="detailVisible" title="待办详情" width="820px" destroy-on-close top="6vh">
      <FlowDetail :detail="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'FlowTodo' });

import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { View, Check, Close, Right, Plus, Search, Refresh } from '@element-plus/icons-vue';
import {
  listFlowTodo,
  approveFlowTask,
  rejectFlowTask,
  transferFlowTask,
  addSignFlowTask,
  getFlowInstance
} from '@/api/flow';
import { listUser } from '@/api/user';
import FlowDetail from '@/views/flow/FlowDetail.vue';
import { parseTime } from '@/utils/format';

const loading = ref(false);
const submitting = ref(false);
const todoList = ref([]);
const total = ref(0);
const userOptions = ref([]);
const actionVisible = ref(false);
const detailVisible = ref(false);
const detailData = ref(null);
const actionFormRef = ref();

const queryParams = reactive({ pageNum: 1, pageSize: 10, keyword: '' });

const actionType = ref('approve');
const actionTask = ref(null);
const actionForm = reactive({
  taskId: undefined,
  toUserId: undefined,
  comment: ''
});

const actionRules: Record<string, any> = {
  comment: [{ required: true, message: '请填写审批意见', trigger: 'blur' }]
};

const actionTitle = computed(() => {
  const m: Record<string, string> = {
    approve: '审批通过',
    reject: '审批驳回',
    transfer: '转办',
    addSign: '加签'
  };
  return m[actionType.value] || '';
});

async function fetchList() {
  loading.value = true;
  try {
    const res = await listFlowTodo(queryParams);
    todoList.value = res.rows;
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

function isOverdue(t: string): boolean {
  return new Date(t).getTime() < Date.now();
}

// 打开操作弹窗
function openAction(row, type: string) {
  actionType.value = type;
  actionTask.value = row;
  actionForm.taskId = row.flowTaskId;
  actionForm.toUserId = undefined;
  actionForm.comment = '';
  if (type === 'approve') {
    // 通过时意见可选
    delete actionRules.comment[0].required;
  } else {
    actionRules.comment[0].required = true;
  }
  actionVisible.value = true;
}

// 提交操作
async function handleSubmitAction() {
  const valid = await actionFormRef.value.validate().catch(() => false);
  if (!valid) return;
  submitting.value = true;
  try {
    const payload = {
      taskId: actionForm.taskId,
      comment: actionForm.comment,
      toUserId: actionForm.toUserId
    };
    if (actionType.value === 'approve') {
      await approveFlowTask(payload);
      ElMessage.success('已通过');
    } else if (actionType.value === 'reject') {
      await rejectFlowTask(payload);
      ElMessage.success('已驳回');
    } else if (actionType.value === 'transfer') {
      await transferFlowTask(payload);
      ElMessage.success('转办成功');
    } else {
      await addSignFlowTask(payload);
      ElMessage.success('加签成功');
    }
    actionVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 详情
async function openDetail(row) {
  const res = await getFlowInstance(row.instanceId);
  detailData.value = res.data;
  detailVisible.value = true;
}

onMounted(() => {
  fetchList();
  listUser({ pageNum: 1, pageSize: 200 }).then((res) => {
    userOptions.value = res.rows || [];
  });
});
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

.action-title {
  width: 100%;
  word-break: break-all;
}
</style>
