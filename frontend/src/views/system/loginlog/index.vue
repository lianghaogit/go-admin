<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form" @submit.prevent>
        <el-form-item label="用户名">
          <el-input
            v-model="queryParams.username"
            placeholder="请输入用户名"
            clearable
            style="width: 160px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="登录IP">
          <el-input
            v-model="queryParams.ipaddr"
            placeholder="请输入登录IP"
            clearable
            style="width: 160px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable style="width: 120px">
            <el-option label="成功" value="0" />
            <el-option label="失败" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="登录时间">
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
          v-hasPermi="['monitor:logininfor:remove']"
          type="danger"
          plain
          :icon="Delete"
          @click="handleClean"
        >
          清空
        </el-button>
        <el-button
          v-hasPermi="['monitor:logininfor:list']"
          type="warning"
          :icon="Download"
          :loading="exporting"
          @click="handleExport"
        >
          导出
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="loginlog" />
        </div>
      </div>
      <el-table v-table-height v-loading="loading" :data="logList" border stripe>
        <el-table-column v-if="col('infoId').visible" prop="infoId" label="编号" width="80" />
        <el-table-column
          v-if="col('userName').visible"
          prop="userName"
          label="用户名"
          width="140"
        />
        <el-table-column v-if="col('ipaddr').visible" prop="ipaddr" label="登录IP" width="150" />
        <el-table-column v-if="col('status').visible" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="col('msg').visible" prop="msg" label="消息" show-overflow-tooltip />
        <el-table-column
          v-if="col('loginTime').visible"
          :formatter="(row) => parseTime(row.loginTime)"
          prop="loginTime"
          label="登录时间"
          width="180"
        />
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemLoginLog' });

import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Delete, Download, Search, Refresh } from '@element-plus/icons-vue';
import { listLoginLog, delLoginLog, cleanLoginLog, exportLoginLog } from '@/api/loginlog';
import { parseTime } from '@/utils/format';
import { exportFile } from '@/utils/download';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';

const loading = ref(false);
const exporting = ref(false);
const logList = ref([]);
const total = ref(0);
const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  username: undefined,
  ipaddr: undefined,
  status: undefined,
  beginTime: undefined,
  endTime: undefined
});
const dateRange = ref<[string, string] | null>(null);

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'infoId', label: '编号', visible: true, fixed: true },
  { prop: 'userName', label: '用户名', visible: true },
  { prop: 'ipaddr', label: '登录IP', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'msg', label: '消息', visible: true },
  { prop: 'loginTime', label: '登录时间', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

// 查询登录日志列表
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
    const res = await listLoginLog(queryParams);
    logList.value = res.rows;
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
  queryParams.username = undefined;
  queryParams.ipaddr = undefined;
  queryParams.status = undefined;
  dateRange.value = null;
  handleQuery();
}

// 删除选中登录日志
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm('确认删除该条登录日志吗？', '提示', { type: 'warning' });
    await delLoginLog(row.infoId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 用户取消删除
  }
}

// 清空登录日志数据
async function handleClean() {
  try {
    await ElMessageBox.confirm('确认清空所有登录日志吗？', '提示', { type: 'warning' });
    await cleanLoginLog();
    ElMessage.success('清空成功');
    fetchList();
  } catch (e) {
    // 用户取消清空
  }
}

// 导出登录日志数据（携带当前过滤条件）
async function handleExport() {
  exporting.value = true;
  try {
    if (dateRange.value) {
      queryParams.beginTime = dateRange.value[0];
      queryParams.endTime = dateRange.value[1];
    } else {
      queryParams.beginTime = undefined;
      queryParams.endTime = undefined;
    }
    await exportFile(exportLoginLog(queryParams), '登录日志.xlsx');
  } finally {
    exporting.value = false;
  }
}

onMounted(fetchList);
</script>

<style scoped>
.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
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
</style>
