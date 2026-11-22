<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="字典名称">
          <el-input
            v-model="queryParams.dictName"
            placeholder="请输入字典名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="字典类型">
          <el-input
            v-model="queryParams.dictType"
            placeholder="请输入字典类型"
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
        <el-button
          v-hasPermi="['system:dict:add']"
          type="primary"
          :icon="Plus"
          @click="openTypeDialog()"
        >
          新增字典
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="dict" />
        </div>
      </div>

      <el-table v-table-height v-loading="loading" :data="typeList" border stripe>
        <el-table-column v-if="col('dictId').visible" prop="dictId" label="编号" width="70" />
        <el-table-column
          v-if="col('dictName').visible"
          prop="dictName"
          label="字典名称"
          width="160"
        />
        <el-table-column
          v-if="col('dictType').visible"
          prop="dictType"
          label="字典类型"
          width="180"
        />
        <el-table-column
          v-if="col('status').visible"
          prop="status"
          label="状态"
          width="80"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          v-if="col('remark').visible"
          prop="remark"
          label="备注"
          show-overflow-tooltip
        />
        <el-table-column
          v-if="col('createTime').visible"
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="170"
        />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:dict:list']"
              type="primary"
              link
              :icon="Tickets"
              @click="openDataDialog(row)"
            >
              字典数据
            </el-button>
            <el-button
              v-hasPermi="['system:dict:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openTypeDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:dict:remove']"
              type="danger"
              link
              :icon="Delete"
              @click="handleDeleteType(row)"
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

    <!-- 新增 / 修改字典类型 -->
    <el-dialog
      v-model="typeDialogVisible"
      :title="typeForm.dictId ? '修改字典类型' : '新增字典类型'"
      width="480px"
      destroy-on-close
    >
      <el-form ref="typeFormRef" :model="typeForm" :rules="typeRules" label-width="90px">
        <el-form-item label="字典名称" prop="dictName">
          <el-input v-model="typeForm.dictName" placeholder="请输入字典名称" />
        </el-form-item>
        <el-form-item label="字典类型" prop="dictType">
          <el-input v-model="typeForm.dictType" placeholder="请输入字典类型，如 sys_user_status" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="typeForm.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="typeForm.remark" type="textarea" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmitType"> 确定 </el-button>
      </template>
    </el-dialog>

    <!-- 字典数据管理 -->
    <el-dialog
      v-model="dataDialogVisible"
      :title="'字典数据 - ' + dataType.dictType"
      width="760px"
      destroy-on-close
    >
      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:dict:add']"
          type="primary"
          :icon="Plus"
          @click="openDataForm()"
        >
          新增数据
        </el-button>
        <el-button
          v-hasPermi="['system:dict:remove']"
          type="danger"
          plain
          :icon="Delete"
          :disabled="dataSelectedIds.length === 0"
          @click="handleBatchDeleteData"
        >
          删除
        </el-button>
      </div>
      <el-table
        v-table-height
        v-loading="dataLoading"
        :data="dataList"
        border
        stripe
        @selection-change="(s) => (dataSelectedIds = s.map((r) => r.dictCode))"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="dictCode" label="编号" width="70" />
        <el-table-column prop="dictLabel" label="标签" width="140" />
        <el-table-column prop="dictValue" label="键值" width="140" />
        <el-table-column prop="dictSort" label="排序" width="70" align="center" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:dict:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDataForm(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:dict:remove']"
              type="danger"
              link
              :icon="Delete"
              @click="handleDeleteData(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="dataQuery.pageNum"
        v-model:page-size="dataQuery.pageSize"
        class="pagination"
        :total="dataTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchDataList"
        @current-change="fetchDataList"
      />

      <!-- 新增 / 修改字典数据 -->
      <el-dialog
        v-model="dataFormVisible"
        :title="dataForm.dictCode ? '修改字典数据' : '新增字典数据'"
        width="460px"
        append-to-body
        destroy-on-close
      >
        <el-form ref="dataFormRef" :model="dataForm" :rules="dataRules" label-width="90px">
          <el-form-item label="标签" prop="dictLabel">
            <el-input v-model="dataForm.dictLabel" placeholder="请输入标签，如：正常" />
          </el-form-item>
          <el-form-item label="键值" prop="dictValue">
            <el-input v-model="dataForm.dictValue" placeholder="请输入键值，如：0" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="dataForm.dictSort" :min="0" :max="99" />
          </el-form-item>
          <el-form-item label="状态">
            <el-radio-group v-model="dataForm.status">
              <el-radio value="0">正常</el-radio>
              <el-radio value="1">停用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dataFormVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmitData">
            确定
          </el-button>
        </template>
      </el-dialog>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemDict' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete, Tickets } from '@element-plus/icons-vue';
import {
  listType,
  getType,
  addType,
  updateType,
  delType,
  listData,
  getData,
  addData,
  updateData,
  delData
} from '@/api/dict';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'dictId', label: '编号', visible: true, fixed: true },
  { prop: 'dictName', label: '字典名称', visible: true },
  { prop: 'dictType', label: '字典类型', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'remark', label: '备注', visible: true },
  { prop: 'createTime', label: '创建时间', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

const loading = ref(false);
const dataLoading = ref(false);
const submitting = ref(false);
const typeDialogVisible = ref(false);
const dataDialogVisible = ref(false);
const dataFormVisible = ref(false);
const typeList = ref([]);
const dataList = ref([]);
const total = ref(0);
const dataTotal = ref(0);
const typeFormRef = ref();
const dataFormRef = ref();
const dataSelectedIds = ref([]);

const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  dictName: '',
  dictType: ''
});

const dataQuery = reactive({ pageNum: 1, pageSize: 10, dictType: '' });
const dataType = reactive({ dictType: '' });

const typeForm = reactive({
  dictId: undefined,
  dictName: '',
  dictType: '',
  status: '0',
  remark: ''
});

const dataForm = reactive({
  dictCode: undefined,
  dictSort: 1,
  dictLabel: '',
  dictValue: '',
  dictType: '',
  status: '0'
});

const typeRules = {
  dictName: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  dictType: [{ required: true, message: '请输入字典类型', trigger: 'blur' }]
};

const dataRules = {
  dictLabel: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  dictValue: [{ required: true, message: '请输入键值', trigger: 'blur' }]
};

// 查询字典列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listType(queryParams);
    typeList.value = res.rows;
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
  queryParams.dictName = '';
  queryParams.dictType = '';
  handleQuery();
}

// 打开字典类型弹窗
async function openTypeDialog(row = undefined) {
  Object.assign(typeForm, {
    dictId: undefined,
    dictName: '',
    dictType: '',
    status: '0',
    remark: ''
  });
  if (row && row.dictId) {
    const res = await getType(row.dictId);
    Object.assign(typeForm, res.data);
  }
  typeDialogVisible.value = true;
}

// 提交字典类型表单
async function handleSubmitType() {
  await typeFormRef.value.validate();
  submitting.value = true;
  try {
    if (typeForm.dictId) {
      await updateType(typeForm);
      ElMessage.success('修改成功');
    } else {
      await addType(typeForm);
      ElMessage.success('新增成功');
    }
    typeDialogVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 删除字典类型
async function handleDeleteType(row) {
  try {
    await ElMessageBox.confirm(`确认删除字典类型「${row.dictName}」吗？`, '提示', {
      type: 'warning'
    });
    await delType(row.dictId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 打开字典数据弹窗
async function openDataDialog(row) {
  dataType.dictType = row.dictType;
  dataQuery.dictType = row.dictType;
  dataQuery.pageNum = 1;
  dataDialogVisible.value = true;
  fetchDataList();
}

// 查询字典数据列表
async function fetchDataList() {
  dataLoading.value = true;
  try {
    const res = await listData(dataQuery);
    dataList.value = res.rows;
    dataTotal.value = res.total;
  } finally {
    dataLoading.value = false;
  }
}

// 打开字典数据表单
async function openDataForm(row = undefined) {
  Object.assign(dataForm, {
    dictCode: undefined,
    dictSort: 1,
    dictLabel: '',
    dictValue: '',
    dictType: dataType.dictType,
    status: '0'
  });
  if (row && row.dictCode) {
    const res = await getData(row.dictCode);
    Object.assign(dataForm, res.data);
  }
  dataFormVisible.value = true;
}

// 提交字典数据表单
async function handleSubmitData() {
  await dataFormRef.value.validate();
  submitting.value = true;
  try {
    if (dataForm.dictCode) {
      await updateData(dataForm);
      ElMessage.success('修改成功');
    } else {
      await addData(dataForm);
      ElMessage.success('新增成功');
    }
    dataFormVisible.value = false;
    fetchDataList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 删除字典数据
async function handleDeleteData(row) {
  try {
    await ElMessageBox.confirm(`确认删除「${row.dictLabel}」吗？`, '提示', {
      type: 'warning'
    });
    await delData(row.dictCode);
    ElMessage.success('删除成功');
    fetchDataList();
  } catch (e) {
    // 取消或失败
  }
}

// 批量删除字典数据
async function handleBatchDeleteData() {
  if (!dataSelectedIds.value.length) return;
  try {
    await ElMessageBox.confirm(
      `确认删除选中的 ${dataSelectedIds.value.length} 条字典数据吗？`,
      '提示',
      {
        type: 'warning'
      }
    );
    await delData(dataSelectedIds.value.join(','));
    ElMessage.success('删除成功');
    dataSelectedIds.value = [];
    fetchDataList();
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
</style>
