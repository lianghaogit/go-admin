<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="接口名称">
          <el-input
            v-model="queryParams.apiName"
            placeholder="请输入接口名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="路径">
          <el-input
            v-model="queryParams.path"
            placeholder="如 /api/system/user/list"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="正常" value="0" />
            <el-option label="停用" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:api:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增接口
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="apiList" border stripe>
        <el-table-column prop="apiId" label="编号" width="70" />
        <el-table-column prop="apiName" label="接口名称" width="180" show-overflow-tooltip />
        <el-table-column prop="method" label="请求方式" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="methodType(row.method)">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="220" show-overflow-tooltip />
        <el-table-column prop="perms" label="权限字符" width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-if="row.apiId"
              v-hasPermi="['system:api:changeStatus']"
              :model-value="row.status === '0'"
              :loading="row.statusLoading"
              @change="(val) => handleStatusChange(row, val)"
            />
            <el-tag v-else :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" width="140" show-overflow-tooltip />
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="170"
        />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:api:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:api:remove']"
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

    <!-- 新增 / 修改接口 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.apiId ? '修改接口' : '新增接口'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="接口名称" prop="apiName">
          <el-input v-model="form.apiName" placeholder="如 用户列表" />
        </el-form-item>
        <el-form-item label="请求方式" prop="method">
          <el-select v-model="form.method" placeholder="请选择请求方式">
            <el-option v-for="m in methods" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="如 /api/system/user/list，参数段用 :id" />
        </el-form-item>
        <el-form-item label="权限字符" prop="perms">
          <el-input v-model="form.perms" placeholder="如 system:user:list，留空表示仅需登录" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">停用</el-radio>
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
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemApi' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue';
import { listApi, addApi, updateApi, delApi, changeApiStatus } from '@/api/api';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const apiList = ref([]);
const total = ref(0);
const formRef = ref();

const queryParams = reactive({ pageNum: 1, pageSize: 10, apiName: '', path: '', status: '' });
const methods = ['GET', 'POST', 'PUT', 'DELETE'];

const form = reactive({
  apiId: undefined,
  apiName: '',
  method: 'GET',
  path: '',
  perms: '',
  status: '0',
  remark: ''
});

const formRules = {
  apiName: [{ required: true, message: '请输入接口名称', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方式', trigger: 'change' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }]
};

// 请求方法类型映射
function methodType(m) {
  if (m === 'GET') return 'success';
  if (m === 'POST') return 'primary';
  if (m === 'PUT') return 'warning';
  if (m === 'DELETE') return 'danger';
  return 'info';
}

// 查询接口列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listApi(queryParams);
    apiList.value = (res.rows || []).map((r) => ({ ...r, statusLoading: false }));
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
  queryParams.apiName = '';
  queryParams.path = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
function openDialog(row = undefined) {
  Object.assign(form, {
    apiId: undefined,
    apiName: '',
    method: 'GET',
    path: '',
    perms: '',
    status: '0',
    remark: ''
  });
  if (row && row.apiId) {
    Object.assign(form, {
      apiId: row.apiId,
      apiName: row.apiName,
      method: row.method,
      path: row.path,
      perms: row.perms,
      status: row.status,
      remark: row.remark
    });
  }
  dialogVisible.value = true;
}

// 提交接口表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.apiId) {
      await updateApi(form);
      ElMessage.success('修改成功');
    } else {
      await addApi(form);
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

// 修改接口状态
async function handleStatusChange(row, val) {
  row.statusLoading = true;
  try {
    await changeApiStatus({ apiId: row.apiId, status: val ? '0' : '1' });
    row.status = val ? '0' : '1';
    ElMessage.success(val ? '启用成功' : '停用成功');
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    row.statusLoading = false;
  }
}

// 删除选中接口
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确认删除接口「${row.apiName}」吗？删除后该接口将不再校验权限。`,
      '提示',
      {
        type: 'warning'
      }
    );
    await delApi(row.apiId);
    ElMessage.success('删除成功');
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
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
