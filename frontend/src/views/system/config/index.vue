<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="参数名称">
          <el-input
            v-model="queryParams.configName"
            placeholder="请输入参数名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="参数键名">
          <el-input
            v-model="queryParams.configKey"
            placeholder="请输入参数键名"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="系统内置">
          <el-select v-model="queryParams.configType" placeholder="请选择" clearable>
            <el-option label="是" value="Y" />
            <el-option label="否" value="N" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:config:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增参数
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="configList" border stripe>
        <el-table-column prop="configId" label="编号" width="70" />
        <el-table-column prop="configName" label="参数名称" width="180" show-overflow-tooltip />
        <el-table-column prop="configKey" label="参数键名" width="200" show-overflow-tooltip />
        <el-table-column prop="configValue" label="参数键值" width="150" show-overflow-tooltip />
        <el-table-column label="系统内置" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.configType === 'Y' ? 'primary' : 'info'">
              {{ row.configType === 'Y' ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:config:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-tooltip
              :disabled="row.configType !== 'Y'"
              content="内置参数不允许删除"
              placement="top"
            >
              <span>
                <el-button
                  v-hasPermi="['system:config:remove']"
                  type="danger"
                  link
                  :icon="Delete"
                  :disabled="row.configType === 'Y'"
                  @click="handleDelete(row)"
                >
                  删除
                </el-button>
              </span>
            </el-tooltip>
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

    <!-- 新增 / 修改参数 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.configId ? '修改参数' : '新增参数'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="参数名称" prop="configName">
          <el-input v-model="form.configName" placeholder="请输入参数名称" />
        </el-form-item>
        <el-form-item label="参数键名" prop="configKey">
          <el-input v-model="form.configKey" placeholder="如 sys.user.initPassword" />
        </el-form-item>
        <el-form-item label="参数键值" prop="configValue">
          <el-input v-model="form.configValue" placeholder="请输入参数键值" />
        </el-form-item>
        <el-form-item label="系统内置">
          <el-radio-group v-model="form.configType" :disabled="!!form.configId && isBuiltin">
            <el-radio value="Y">是</el-radio>
            <el-radio value="N">否</el-radio>
          </el-radio-group>
          <div v-if="!!form.configId && isBuiltin" class="builtin-tip">内置参数不允许修改类型</div>
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
defineOptions({ name: 'SystemConfig' });

import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue';
import { listConfig, addConfig, updateConfig, delConfig } from '@/api/config';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const configList = ref([]);
const total = ref(0);
const formRef = ref();

const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  configName: '',
  configKey: '',
  configType: ''
});

const form = reactive({
  configId: undefined,
  configName: '',
  configKey: '',
  configValue: '',
  configType: 'N',
  remark: ''
});

const isBuiltin = computed(() => form.configType === 'Y');

const formRules = {
  configName: [{ required: true, message: '请输入参数名称', trigger: 'blur' }],
  configKey: [{ required: true, message: '请输入参数键名', trigger: 'blur' }],
  configValue: [{ required: true, message: '请输入参数键值', trigger: 'blur' }]
};

// 查询参数配置列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listConfig(queryParams);
    configList.value = res.rows;
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
  queryParams.configName = '';
  queryParams.configKey = '';
  queryParams.configType = '';
  handleQuery();
}

// 打开新增/编辑弹窗
function openDialog(row = undefined) {
  Object.assign(form, {
    configId: undefined,
    configName: '',
    configKey: '',
    configValue: '',
    configType: 'N',
    remark: ''
  });
  if (row && row.configId) {
    Object.assign(form, {
      configId: row.configId,
      configName: row.configName,
      configKey: row.configKey,
      configValue: row.configValue,
      configType: row.configType,
      remark: row.remark
    });
  }
  dialogVisible.value = true;
}

// 提交参数配置表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.configId) {
      await updateConfig(form);
      ElMessage.success('修改成功');
    } else {
      await addConfig(form);
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

// 删除选中参数配置
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除参数「${row.configName}」吗？`, '提示', {
      type: 'warning'
    });
    await delConfig(row.configId);
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

.builtin-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #e6a23c;
}
</style>
