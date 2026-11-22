<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="岗位名称">
          <el-input
            v-model="queryParams.postName"
            placeholder="请输入岗位名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="岗位编码">
          <el-input
            v-model="queryParams.postCode"
            placeholder="请输入岗位编码"
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
          v-hasPermi="['system:post:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增岗位
        </el-button>
        <el-button
          v-hasPermi="['system:post:remove']"
          type="danger"
          plain
          :icon="Delete"
          :disabled="selectedIds.length === 0"
          @click="handleBatchDelete"
        >
          删除
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="post" />
        </div>
      </div>

      <el-table
        v-table-height
        v-loading="loading"
        :data="postList"
        border
        stripe
        @selection-change="(s) => (selectedIds = s.map((r) => r.postId))"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column v-if="col('postId').visible" prop="postId" label="编号" width="70" />
        <el-table-column
          v-if="col('postCode').visible"
          prop="postCode"
          label="岗位编码"
          width="120"
        />
        <el-table-column
          v-if="col('postName').visible"
          prop="postName"
          label="岗位名称"
          width="140"
        />
        <el-table-column
          v-if="col('postSort').visible"
          prop="postSort"
          label="排序"
          width="80"
          align="center"
        />
        <el-table-column v-if="col('status').visible" label="状态" width="80" align="center">
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
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:post:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:post:remove']"
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

    <!-- 新增 / 修改岗位 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.postId ? '修改岗位' : '新增岗位'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="岗位编码" prop="postCode">
          <el-input v-model="form.postCode" placeholder="请输入岗位编码" />
        </el-form-item>
        <el-form-item label="岗位名称" prop="postName">
          <el-input v-model="form.postName" placeholder="请输入岗位名称" />
        </el-form-item>
        <el-form-item label="显示排序">
          <el-input-number v-model="form.postSort" :min="0" />
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
defineOptions({ name: 'SystemPost' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue';
import { listPost, addPost, updatePost, delPost } from '@/api/post';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'postId', label: '编号', visible: true, fixed: true },
  { prop: 'postCode', label: '岗位编码', visible: true },
  { prop: 'postName', label: '岗位名称', visible: true },
  { prop: 'postSort', label: '排序', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'remark', label: '备注', visible: true },
  { prop: 'createTime', label: '创建时间', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const postList = ref([]);
const total = ref(0);
const formRef = ref();
const selectedIds = ref([]);

const queryParams = reactive({ pageNum: 1, pageSize: 10, postName: '', postCode: '', status: '' });

const form = reactive({
  postId: undefined,
  postCode: '',
  postName: '',
  postSort: 0,
  status: '0',
  remark: ''
});

const formRules = {
  postCode: [{ required: true, message: '请输入岗位编码', trigger: 'blur' }],
  postName: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }]
};

// 查询岗位列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listPost(queryParams);
    postList.value = res.rows;
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
  queryParams.postName = '';
  queryParams.postCode = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
function openDialog(row = undefined) {
  Object.assign(form, {
    postId: undefined,
    postCode: '',
    postName: '',
    postSort: 0,
    status: '0',
    remark: ''
  });
  if (row && row.postId) {
    Object.assign(form, {
      postId: row.postId,
      postCode: row.postCode,
      postName: row.postName,
      postSort: row.postSort,
      status: row.status,
      remark: row.remark
    });
  }
  dialogVisible.value = true;
}

// 提交岗位表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.postId) {
      await updatePost(form);
      ElMessage.success('修改成功');
    } else {
      await addPost(form);
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

// 删除选中岗位
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除岗位「${row.postName}」吗？`, '提示', {
      type: 'warning'
    });
    await delPost(row.postId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 批量删除岗位
async function handleBatchDelete() {
  if (!selectedIds.value.length) return;
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selectedIds.value.length} 个岗位吗？`, '提示', {
      type: 'warning'
    });
    await delPost(selectedIds.value.join(','));
    ElMessage.success('删除成功');
    selectedIds.value = [];
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
</style>
