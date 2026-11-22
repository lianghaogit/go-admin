<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="任务名称">
          <el-input
            v-model="queryParams.jobName"
            placeholder="请输入任务名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="任务分组">
          <el-input
            v-model="queryParams.jobGroup"
            placeholder="如 DEFAULT"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="正常" value="0" />
            <el-option label="暂停" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['monitor:job:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增任务
        </el-button>
        <el-button v-hasPermi="['monitor:job:list']" :icon="List" @click="openLogDialog()">
          执行日志
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="jobList" border stripe>
        <el-table-column prop="jobId" label="编号" width="70" />
        <el-table-column prop="jobName" label="任务名称" min-width="140" />
        <el-table-column prop="jobGroup" label="任务分组" width="100" />
        <el-table-column prop="invokeTarget" label="调用目标" width="150" show-overflow-tooltip />
        <el-table-column prop="cronExpression" label="cron 表达式" width="130" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'info'">
              {{ row.status === '0' ? '正常' : '暂停' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['monitor:job:run']"
              type="success"
              link
              :icon="VideoPlay"
              @click="handleRun(row)"
            >
              执行一次
            </el-button>
            <el-button
              v-hasPermi="['monitor:job:changeStatus']"
              :type="row.status === '0' ? 'warning' : 'primary'"
              link
              :icon="row.status === '0' ? VideoPause : VideoPlay"
              @click="handleChangeStatus(row)"
            >
              {{ row.status === '0' ? '暂停' : '恢复' }}
            </el-button>
            <el-button
              v-hasPermi="['monitor:job:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['monitor:job:remove']"
              type="danger"
              link
              :icon="Delete"
              @click="handleDelete(row)"
            >
              删除
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

    <!-- 新增 / 修改任务 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.jobId ? '修改任务' : '新增任务'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px">
        <el-form-item label="任务名称" prop="jobName">
          <el-input v-model="form.jobName" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="任务分组" prop="jobGroup">
          <el-input v-model="form.jobGroup" placeholder="如 DEFAULT / SYSTEM" />
        </el-form-item>
        <el-form-item label="调用目标" prop="invokeTarget">
          <el-select
            v-model="form.invokeTarget"
            placeholder="请选择内置任务，或输入 http(s) 地址"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <el-option
              v-for="ex in executors"
              :key="ex.name"
              :label="ex.name + '：' + ex.desc"
              :value="ex.name"
            />
          </el-select>
          <div class="invoke-hint">
            内置任务可选；直接输入以 http:// 或 https:// 开头的地址，将发起 HTTP GET 请求
          </div>
        </el-form-item>
        <el-form-item label="cron 表达式" prop="cronExpression">
          <el-input v-model="form.cronExpression" placeholder="秒 分 时 日 月 周，如 0 * * * * *" />
        </el-form-item>
        <el-form-item label="是否并发">
          <el-radio-group v-model="form.concurrent">
            <el-radio value="0">允许</el-radio>
            <el-radio value="1">禁止</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">暂停</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit"> 确定 </el-button>
      </template>
    </el-dialog>

    <!-- 执行日志 -->
    <el-dialog v-model="logVisible" title="执行日志" width="760px" destroy-on-close>
      <div class="log-toolbar">
        <el-button
          v-hasPermi="['monitor:job:remove']"
          type="danger"
          :icon="Delete"
          @click="handleCleanLog"
        >
          清空日志
        </el-button>
      </div>
      <el-table v-table-height v-loading="logLoading" :data="logList" border stripe size="small">
        <el-table-column prop="jobName" label="任务名称" min-width="120" />
        <el-table-column prop="invokeTarget" label="调用目标" width="150" show-overflow-tooltip />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="jobMessage" label="消息" min-width="160" show-overflow-tooltip />
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="执行时间"
          width="170"
        />
      </el-table>
      <el-pagination
        v-model:current-page="logParams.pageNum"
        v-model:page-size="logParams.pageSize"
        class="pagination"
        :total="logTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorJob' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete,
  List,
  VideoPlay,
  VideoPause
} from '@element-plus/icons-vue';
import {
  listJob,
  getJob,
  addJob,
  updateJob,
  delJob,
  changeJobStatus,
  runJob,
  listJobLog,
  cleanJobLog,
  listExecutors
} from '@/api/job';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const logVisible = ref(false);
const jobList = ref([]);
const total = ref(0);
const formRef = ref();
// 内置任务执行器（调用目标下拉）
const executors = ref([]);

const queryParams = reactive({ pageNum: 1, pageSize: 10, jobName: '', jobGroup: '', status: '' });

const form = reactive({
  jobId: undefined,
  jobName: '',
  jobGroup: 'DEFAULT',
  invokeTarget: '',
  cronExpression: '',
  misfirePolicy: '3',
  concurrent: '1',
  status: '0',
  remark: ''
});

const formRules = {
  jobName: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  jobGroup: [{ required: true, message: '请输入任务分组', trigger: 'blur' }],
  invokeTarget: [{ required: true, message: '请选择调用目标', trigger: 'change' }],
  cronExpression: [{ required: true, message: '请输入 cron 表达式', trigger: 'blur' }]
};

// 执行日志
const logLoading = ref(false);
const logList = ref([]);
const logTotal = ref(0);
const logParams = reactive({ pageNum: 1, pageSize: 10 });

// 查询定时任务列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listJob(queryParams);
    jobList.value = res.rows;
    total.value = res.total;
  } finally {
    loading.value = false;
  }
}

// 处理查询
function handleQuery() {
  queryParams.pageNum = 1;
  fetchList();
}

// 重置查询条件
function resetQuery() {
  queryParams.jobName = '';
  queryParams.jobGroup = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined) {
  Object.assign(form, {
    jobId: undefined,
    jobName: '',
    jobGroup: 'DEFAULT',
    invokeTarget: '',
    cronExpression: '',
    misfirePolicy: '3',
    concurrent: '1',
    status: '0',
    remark: ''
  });
  if (row && row.jobId) {
    const res = await getJob(row.jobId);
    const j = res.data;
    Object.assign(form, {
      jobId: j.jobId,
      jobName: j.jobName,
      jobGroup: j.jobGroup,
      invokeTarget: j.invokeTarget,
      cronExpression: j.cronExpression,
      misfirePolicy: j.misfirePolicy,
      concurrent: j.concurrent,
      status: j.status,
      remark: j.remark
    });
  }
  dialogVisible.value = true;
}

// 提交定时任务表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.jobId) {
      await updateJob(form);
      ElMessage.success('修改成功');
    } else {
      await addJob(form);
      ElMessage.success('新增成功');
    }
    dialogVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 删除选中定时任务
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除任务「${row.jobName}」吗？`, '提示', { type: 'warning' });
    await delJob(row.jobId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 立即执行定时任务
async function handleRun(row) {
  try {
    await ElMessageBox.confirm(`确认立即执行一次任务「${row.jobName}」吗？`, '提示', {
      type: 'warning'
    });
    await runJob(row.jobId);
    ElMessage.success('已触发执行，请查看执行日志');
  } catch (e) {
    // 取消或失败
  }
}

// 修改定时任务状态
async function handleChangeStatus(row) {
  const next = row.status === '0' ? '1' : '0';
  const label = next === '1' ? '暂停' : '恢复';
  try {
    await ElMessageBox.confirm(`确认${label}任务「${row.jobName}」吗？`, '提示', {
      type: 'warning'
    });
    await changeJobStatus(row.jobId, next);
    ElMessage.success(`${label}成功`);
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 查询执行日志
async function fetchLogs() {
  logLoading.value = true;
  try {
    const res = await listJobLog(logParams);
    logList.value = res.rows;
    logTotal.value = res.total;
  } finally {
    logLoading.value = false;
  }
}

// 打开执行日志弹窗
async function openLogDialog() {
  logVisible.value = true;
  logParams.pageNum = 1;
  fetchLogs();
}

// 清空执行日志
async function handleCleanLog() {
  try {
    await ElMessageBox.confirm('确认清空全部执行日志吗？', '提示', { type: 'warning' });
    await cleanJobLog();
    ElMessage.success('清空成功');
    fetchLogs();
  } catch (e) {
    // 取消或失败
  }
}

onMounted(() => {
  fetchList();
  listExecutors()
    .then((res) => {
      executors.value = res.data || [];
    })
    .catch(() => {});
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

.log-toolbar {
  margin-bottom: 10px;
}

.invoke-hint {
  font-size: 12px;
  color: var(--go-text-secondary);
  line-height: 1.5;
  margin-top: 4px;
  width: 100%;
}
</style>
