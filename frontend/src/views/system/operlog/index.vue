<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="模块标题">
          <el-input
            v-model="queryParams.title"
            placeholder="请输入模块标题"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="操作人员">
          <el-input
            v-model="queryParams.operName"
            placeholder="请输入操作人员"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="成功" value="0" />
            <el-option label="失败" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['monitor:operlog:remove']"
          type="danger"
          plain
          :icon="Delete"
          @click="handleClean"
        >
          清空
        </el-button>
        <el-button
          v-hasPermi="['monitor:operlog:list']"
          type="warning"
          :icon="Download"
          :loading="exporting"
          @click="handleExport"
        >
          导出
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="operlog" />
        </div>
      </div>

      <el-table v-table-height v-loading="loading" :data="logList" border stripe>
        <el-table-column v-if="col('operId').visible" prop="operId" label="编号" width="70" />
        <el-table-column
          v-if="col('title').visible"
          prop="title"
          label="模块标题"
          show-overflow-tooltip
        />
        <el-table-column
          v-if="col('requestMethod').visible"
          prop="requestMethod"
          label="请求方式"
          width="90"
          align="center"
        />
        <el-table-column
          v-if="col('operName').visible"
          prop="operName"
          label="操作人员"
          width="110"
        />
        <el-table-column v-if="col('operIp').visible" prop="operIp" label="IP 地址" width="130" />
        <el-table-column v-if="col('status').visible" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          v-if="col('costTime').visible"
          prop="costTime"
          label="耗时(ms)"
          width="90"
          align="center"
        />
        <el-table-column
          v-if="col('operTime').visible"
          :formatter="(row) => parseTime(row.operTime)"
          prop="operTime"
          label="操作时间"
          width="170"
        />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="View" @click="handleView(row)"> 详情 </el-button>
            <el-button
              v-hasPermi="['monitor:operlog:remove']"
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

    <!-- 操作日志详情 -->
    <el-dialog v-model="detailVisible" title="操作日志详情" width="680px" destroy-on-close>
      <el-descriptions v-if="detail.operId" :column="2" border>
        <el-descriptions-item label="模块标题" :span="2">{{
          detail.title || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          {{ bizTypeText(detail.businessType) }}
        </el-descriptions-item>
        <el-descriptions-item label="请求方式">{{
          detail.requestMethod || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="操作人员" :span="2">{{
          detail.operName || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="操作IP" :span="2">{{
          detail.operIp || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="操作地址" :span="2">{{
          detail.operUrl || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="请求参数" :span="2">
          <pre class="param-pre">{{ detail.operParam || '-' }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="操作状态">
          <el-tag :type="detail.status === '0' ? 'success' : 'danger'">
            {{ detail.status === '0' ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="耗时">{{ detail.costTime }} ms</el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">
          {{ parseTime(detail.operTime) || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.status !== '0'" label="错误信息" :span="2">
          {{ detail.errorMsg || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemOperLog' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Delete, View, Download } from '@element-plus/icons-vue';
import { listOperLog, delOperLog, cleanOperLog, exportOperLog } from '@/api/operlog';
import { exportFile } from '@/utils/download';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';

const loading = ref(false);
const exporting = ref(false);
const logList = ref([]);
const total = ref(0);
const detailVisible = ref(false);

// 操作日志详情对象（接口数据宽松类型，后续可细化）
interface OperLogDetail {
  operId?: number;
  title?: string;
  businessType?: number;
  requestMethod?: string;
  operName?: string;
  operIp?: string;
  operUrl?: string;
  operParam?: string;
  status?: number | string;
  costTime?: number;
  operTime?: string;
  errorMsg?: string;
}
const detail = ref({} as OperLogDetail);

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'operId', label: '编号', visible: true, fixed: true },
  { prop: 'title', label: '模块标题', visible: true },
  { prop: 'requestMethod', label: '请求方式', visible: true },
  { prop: 'operName', label: '操作人员', visible: true },
  { prop: 'operIp', label: 'IP 地址', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'costTime', label: '耗时(ms)', visible: true },
  { prop: 'operTime', label: '操作时间', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

// 导出操作日志数据
async function handleExport() {
  exporting.value = true;
  try {
    await exportFile(exportOperLog(queryParams), '操作日志.xlsx');
  } finally {
    exporting.value = false;
  }
}

const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  title: '',
  operName: '',
  status: '',
  beginTime: undefined,
  endTime: undefined
});
const dateRange = ref<[string, string] | null>(null);

// 查询操作日志列表
async function fetchList() {
  loading.value = true;
  try {
    // 日期范围拆分为起止时间
    if (dateRange.value) {
      queryParams.beginTime = dateRange.value[0];
      queryParams.endTime = dateRange.value[1];
    } else {
      queryParams.beginTime = undefined;
      queryParams.endTime = undefined;
    }
    const res = await listOperLog(queryParams);
    logList.value = res.rows;
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
  queryParams.title = '';
  queryParams.operName = '';
  queryParams.status = '';
  dateRange.value = null;
  handleQuery();
}

// 业务类型文案映射
function bizTypeText(type) {
  const map = { 0: '其它', 1: '新增', 2: '修改', 3: '删除', 4: '查询' };
  return map[type] || '其它';
}

// 查看操作日志详情
function handleView(row) {
  detail.value = row;
  detailVisible.value = true;
}

// 删除选中操作日志
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除该条操作日志吗？`, '提示', { type: 'warning' });
    await delOperLog(row.operId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 清空操作日志数据
async function handleClean() {
  try {
    await ElMessageBox.confirm('确认清空所有操作日志吗？', '提示', { type: 'warning' });
    await cleanOperLog();
    ElMessage.success('清空成功');
    fetchList();
  } catch (e) {
    // 取消或失败
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
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.toolbar-right {
  margin-left: auto;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}

.param-pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: Consolas, monospace;
  font-size: 12px;
  color: var(--go-text-regular);
  max-height: 180px;
  overflow: auto;
}
</style>
