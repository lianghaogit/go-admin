<template>
  <div class="app-container">
    <el-card>
      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:dept:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增部门
        </el-button>
      </div>

      <el-table
        v-table-height
        v-loading="loading"
        :data="deptList"
        border
        row-key="deptId"
        :tree-props="{ children: 'children' }"
        default-expand-all
      >
        <el-table-column prop="deptName" label="部门名称" width="220" />
        <el-table-column prop="orderNum" label="排序" width="70" align="center" />
        <el-table-column prop="leader" label="负责人" width="120" />
        <el-table-column prop="phone" label="联系电话" width="140" />
        <el-table-column prop="email" label="邮箱" show-overflow-tooltip />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:dept:add']"
              type="primary"
              link
              :icon="Plus"
              @click="openDialog(row, true)"
            >
              新增
            </el-button>
            <el-button
              v-hasPermi="['system:dept:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:dept:remove']"
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
    </el-card>

    <!-- 新增 / 修改部门 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.deptId ? '修改部门' : '新增部门'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="form.parentId"
            :data="deptOptions"
            :props="{ label: 'deptName', children: 'children', value: 'deptId' }"
            check-strictly
            clearable
            placeholder="选择上级部门（默认总公司）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="deptName">
          <el-input v-model="form.deptName" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="显示排序">
          <el-input-number v-model="form.orderNum" :min="0" :max="99" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.leader" placeholder="请输入负责人" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">停用</el-radio>
          </el-radio-group>
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
defineOptions({ name: 'SystemDept' });

import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';
import { listDept, deptTreeSelect, getDept, addDept, updateDept, delDept } from '@/api/dept';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const deptList = ref([]);
const deptOptions = ref([]);
const formRef = ref();

const form = reactive({
  deptId: undefined,
  parentId: 100,
  deptName: '',
  orderNum: 1,
  leader: '',
  phone: '',
  email: '',
  status: '0'
});

const formRules = {
  deptName: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
};

// 查询部门列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listDept();
    deptList.value = res.data;
  } finally {
    loading.value = false;
  }
}

// 获取部门下拉选项
async function fetchOptions() {
  const res = await deptTreeSelect();
  deptOptions.value = res.data;
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined, asChild = undefined) {
  Object.assign(form, {
    deptId: undefined,
    parentId: 100,
    deptName: '',
    orderNum: 1,
    leader: '',
    phone: '',
    email: '',
    status: '0'
  });
  if (row && row.deptId) {
    if (asChild) {
      // 新增下级部门：上级为当前部门
      form.parentId = row.deptId;
    } else {
      const res = await getDept(row.deptId);
      Object.assign(form, res.data);
    }
  }
  dialogVisible.value = true;
}

// 提交部门表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.deptId) {
      await updateDept(form);
      ElMessage.success('修改成功');
    } else {
      await addDept(form);
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

// 删除选中部门
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除部门「${row.deptName}」吗？`, '提示', {
      type: 'warning'
    });
    await delDept(row.deptId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

onMounted(() => {
  fetchList();
  fetchOptions();
});
</script>

<style scoped>
.table-toolbar {
  margin-bottom: 12px;
}
</style>
