<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="标题">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入申请标题"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="运行中" value="1" />
            <el-option label="已通过" value="2" />
            <el-option label="已驳回" value="3" />
            <el-option label="已撤销" value="4" />
            <el-option label="已终止" value="5" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['flow:apply:add']"
          type="primary"
          :icon="Plus"
          @click="openApplyDialog"
        >
          发起申请
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="applyList" border stripe>
        <el-table-column prop="flowInstanceId" label="编号" width="70" />
        <el-table-column prop="title" label="申请标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="defName" label="流程" width="120" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前节点" width="120">
          <template #default="{ row }">{{ currentNodeName(row) }}</template>
        </el-table-column>
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="申请时间"
          width="165"
        />
        <el-table-column label="操作" width="270" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['flow:instance:query']"
              type="primary"
              link
              :icon="View"
              @click="openDetail(row)"
            >
              详情
            </el-button>
            <el-button
              v-if="row.status === '1'"
              v-hasPermi="['flow:apply:withdraw']"
              type="warning"
              link
              @click="handleWithdraw(row)"
            >
              撤回
            </el-button>
            <el-button
              v-if="row.status === '1'"
              v-hasPermi="['flow:apply:urge']"
              type="warning"
              link
              @click="handleUrge(row)"
            >
              催办
            </el-button>
            <el-button
              v-if="row.status === '1'"
              v-hasPermi="['flow:apply:terminate']"
              type="danger"
              link
              @click="handleTerminate(row)"
            >
              终止
            </el-button>
            <el-button
              v-if="row.status === '3'"
              v-hasPermi="['flow:apply:reapply']"
              type="success"
              link
              :icon="RefreshRight"
              @click="openReapply(row)"
            >
              重新提交
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
        @size-change="handleQuery"
        @current-change="fetchList"
      />
    </el-card>

    <!-- 发起申请 -->
    <el-dialog v-model="applyVisible" title="发起申请" width="600px" destroy-on-close>
      <el-form ref="applyFormRef" :model="applyForm" :rules="applyRules" label-width="90px">
        <el-form-item label="审批流程" prop="flowDefId">
          <el-select v-model="applyForm.flowDefId" placeholder="请选择流程" style="width: 100%">
            <el-option v-for="d in enabledDefs" :key="d.flowDefId" :label="d.defName" :value="d.flowDefId" />
          </el-select>
        </el-form-item>
        <el-form-item label="申请标题" prop="title">
          <el-input v-model="applyForm.title" placeholder="请填写申请标题" maxlength="128" />
        </el-form-item>
        <el-form-item label="申请内容" prop="content">
          <el-input
            v-model="applyForm.content"
            type="textarea"
            :rows="4"
            placeholder="请填写申请内容/事由"
          />
        </el-form-item>
        <el-form-item label="抄送人">
          <el-select
            v-model="applyForm.ccUserIds"
            multiple
            filterable
            placeholder="选填，抄送给相关人"
            style="width: 100%"
          >
            <el-option v-for="u in userOptions" :key="u.userId" :label="u.nickName" :value="u.userId" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmitApply"> 提交 </el-button>
      </template>
    </el-dialog>

    <!-- 重新提交 -->
    <el-dialog v-model="reapplyVisible" title="重新提交" width="600px" destroy-on-close>
      <el-form ref="reapplyFormRef" :model="reapplyForm" :rules="applyRules" label-width="90px">
        <el-form-item label="申请标题" prop="title">
          <el-input v-model="reapplyForm.title" maxlength="128" />
        </el-form-item>
        <el-form-item label="申请内容" prop="content">
          <el-input v-model="reapplyForm.content" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reapplyVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmitReapply"> 提交 </el-button>
      </template>
    </el-dialog>

    <!-- 实例详情 -->
    <el-dialog v-model="detailVisible" title="申请详情" width="820px" destroy-on-close top="6vh">
      <FlowDetail :detail="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'FlowApply' });

import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, View, RefreshRight } from '@element-plus/icons-vue';
import {
  listFlowApply,
  addFlowApply,
  withdrawFlowApply,
  reapplyFlowApply,
  urgeFlowApply,
  terminateFlowApply,
  getFlowInstance,
  listEnabledFlowDef
} from '@/api/flow';
import { listUser } from '@/api/user';
import FlowDetail from '@/views/flow/FlowDetail.vue';
import { parseTime } from '@/utils/format';

const loading = ref(false);
const submitting = ref(false);
const applyList = ref([]);
const total = ref(0);
const enabledDefs = ref([]);
const userOptions = ref([]);
const applyVisible = ref(false);
const reapplyVisible = ref(false);
const detailVisible = ref(false);
const detailData = ref(null);
const applyFormRef = ref();
const reapplyFormRef = ref();

const queryParams = reactive({ pageNum: 1, pageSize: 10, keyword: '', status: '' });

const applyForm = reactive({
  flowDefId: undefined,
  title: '',
  content: '',
  ccUserIds: [] as number[]
});

const reapplyForm = reactive({
  title: '',
  content: ''
});

const applyRules = {
  flowDefId: [{ required: true, message: '请选择审批流程', trigger: 'change' }],
  title: [{ required: true, message: '请填写申请标题', trigger: 'blur' }],
  content: [{ required: true, message: '请填写申请内容', trigger: 'blur' }]
};

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

async function fetchList() {
  loading.value = true;
  try {
    const res = await listFlowApply(queryParams);
    applyList.value = res.rows;
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
  queryParams.status = '';
  handleQuery();
}

// 打开发起申请弹窗
async function openApplyDialog() {
  const res = await listEnabledFlowDef();
  enabledDefs.value = res.data || [];
  applyForm.flowDefId = undefined;
  applyForm.title = '';
  applyForm.content = '';
  applyForm.ccUserIds = [];
  applyVisible.value = true;
}

// 提交发起
async function handleSubmitApply() {
  await applyFormRef.value.validate();
  submitting.value = true;
  try {
    await addFlowApply(applyForm);
    ElMessage.success('发起成功');
    applyVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 撤回
async function handleWithdraw(row) {
  try {
    await ElMessageBox.confirm(`确认撤回申请「${row.title}」吗？`, '提示', { type: 'warning' });
    await withdrawFlowApply(row.flowInstanceId);
    ElMessage.success('撤回成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 催办
async function handleUrge(row) {
  try {
    await ElMessageBox.confirm(`确认催办申请「${row.title}」吗？`, '提示', { type: 'warning' });
    await urgeFlowApply(row.flowInstanceId);
    ElMessage.success('已催办');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 终止（管理员）
async function handleTerminate(row) {
  try {
    const { value } = await ElMessageBox.prompt(`确认终止申请「${row.title}」吗？`, '终止申请', {
      inputPlaceholder: '请填写终止原因（可选）',
      inputValidator: () => true,
      type: 'warning',
      confirmButtonText: '确认终止'
    });
    await terminateFlowApply(row.flowInstanceId, { comment: value || '' });
    ElMessage.success('已终止');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 重新提交
function openReapply(row) {
  reapplyForm.title = row.title;
  reapplyForm.content = row.content;
  reapplyVisible.value = true;
  // 记录目标实例
  (reapplyForm as any)._instanceId = row.flowInstanceId;
}

async function handleSubmitReapply() {
  await reapplyFormRef.value.validate();
  submitting.value = true;
  try {
    await reapplyFlowApply((reapplyForm as any)._instanceId, {
      title: reapplyForm.title,
      content: reapplyForm.content
    });
    ElMessage.success('已重新提交');
    reapplyVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 详情
async function openDetail(row) {
  const res = await getFlowInstance(row.flowInstanceId);
  detailData.value = res.data;
  detailVisible.value = true;
}

// 当前节点名（从详情获取过重，列表用 currentOrder 映射不可靠，回退显示序号）
function currentNodeName(row): string {
  return row.currentOrder > 0 ? `第 ${row.currentOrder} 步` : '-';
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

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
