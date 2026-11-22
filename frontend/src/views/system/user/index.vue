<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="用户名">
          <el-input
            v-model="queryParams.username"
            placeholder="请输入用户名"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input
            v-model="queryParams.phone"
            placeholder="请输入手机号"
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
          v-hasPermi="['system:user:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增用户
        </el-button>
        <el-button
          v-hasPermi="['system:user:remove']"
          type="danger"
          plain
          :icon="Delete"
          :disabled="selectedIds.length === 0"
          @click="handleBatchDelete"
        >
          删除
        </el-button>
        <el-button
          v-hasPermi="['system:user:import']"
          type="success"
          :icon="Upload"
          @click="openImport"
        >
          导入
        </el-button>
        <el-button
          v-hasPermi="['system:user:export']"
          type="warning"
          :icon="Download"
          :loading="exporting"
          @click="handleExport"
        >
          导出
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="user" />
        </div>
      </div>

      <el-table
        ref="tableRef"
        v-table-height
        v-loading="loading"
        :data="userList"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column v-if="col('userId').visible" prop="userId" label="编号" width="70" />
        <el-table-column v-if="col('avatar').visible" label="头像" width="70" align="center">
          <template #default="{ row }">
            <el-avatar :size="32" :src="row.avatar || undefined">
              <el-icon><UserFilled /></el-icon>
            </el-avatar>
          </template>
        </el-table-column>
        <el-table-column
          v-if="col('userName').visible"
          prop="userName"
          label="用户名"
          width="120"
        />
        <el-table-column v-if="col('nickName').visible" prop="nickName" label="昵称" width="120" />
        <el-table-column v-if="col('dept').visible" label="部门" width="140">
          <template #default="{ row }">{{ row.dept?.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column v-if="col('posts').visible" label="岗位" width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row.posts || []).map((p) => p.postName).join('、') || '-' }}
          </template>
        </el-table-column>
        <el-table-column v-if="col('phone').visible" prop="phone" label="手机号" width="130" />
        <el-table-column v-if="col('status').visible" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-hasPermi="['system:user:edit']"
              :model-value="row.status === '0'"
              :disabled="row.userName === 'admin'"
              :before-change="() => handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column
          v-if="col('createTime').visible"
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="170"
        />
        <el-table-column
          v-if="col('remark').visible"
          prop="remark"
          label="备注"
          show-overflow-tooltip
        />
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:user:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:user:edit']"
              type="info"
              link
              :icon="UserFilled"
              @click="openAuthRole(row)"
            >
              分配角色
            </el-button>
            <el-button
              v-hasPermi="['system:user:resetPwd']"
              type="warning"
              link
              :icon="Key"
              @click="handleResetPwd(row)"
            >
              重置密码
            </el-button>
            <el-button
              v-hasPermi="['system:user:remove']"
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

    <!-- 新增 / 修改用户 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.userId ? '修改用户' : '新增用户'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!form.userId" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="!form.userId" label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickName">
          <el-input v-model="form.nickName" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="部门">
          <el-tree-select
            v-model="form.deptId"
            :data="deptOptions"
            :props="{ label: 'deptName', children: 'children', value: 'deptId' }"
            check-strictly
            clearable
            placeholder="请选择部门"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roleIds" multiple placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="role in roleOptions"
              :key="role.roleId"
              :label="role.roleName"
              :value="role.roleId"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="岗位">
          <el-select v-model="form.postIds" multiple placeholder="请选择岗位" style="width: 100%">
            <el-option
              v-for="post in postOptions"
              :key="post.postId"
              :label="post.postName"
              :value="post.postId"
            />
          </el-select>
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

    <!-- 重置密码 -->
    <el-dialog v-model="pwdVisible" title="重置密码" width="420px" destroy-on-close>
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="pwdForm.password"
            type="password"
            show-password
            placeholder="请输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleResetSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 用户导入 -->
    <el-dialog v-model="importVisible" title="导入用户" width="520px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="更新支持">
          <el-checkbox v-model="updateSupport">若用户名已存在，则覆盖更新该用户</el-checkbox>
        </el-form-item>
        <el-form-item label="导入文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".xlsx,.xls"
            :on-change="onFileChange"
            :on-remove="onFileRemove"
          >
            <el-button :icon="Upload">选择文件（xlsx）</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="模板">
          <el-button link type="primary" @click="handleDownloadTemplate">下载导入模板</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="handleImport">确定导入</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色 -->
    <el-dialog v-model="authRoleVisible" title="分配角色" width="520px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="用户名">
          <el-input :model-value="authRoleUser?.nickName" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="authRoleIds">
            <el-checkbox
              v-for="r in roleOptions"
              :key="r.roleId"
              :label="r.roleId"
              :disabled="r.roleKey === 'admin'"
            >
              {{ r.roleName }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="authRoleVisible = false">取消</el-button>
        <el-button type="primary" :loading="authRoleSubmitting" @click="handleAuthRole"
          >确定</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemUser' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete,
  Key,
  Download,
  Upload,
  UserFilled
} from '@element-plus/icons-vue';
import {
  listUser,
  getUser,
  addUser,
  updateUser,
  delUser,
  changeUserStatus,
  authRole,
  resetUserPwd,
  exportUser,
  importUserTemplate,
  importUser
} from '@/api/user';
import { exportFile } from '@/utils/download';
import { listAllRole } from '@/api/role';
import { listAllPost } from '@/api/post';
import { deptTreeSelect } from '@/api/dept';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';
import type { UserItem } from '@/api/user';

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'userId', label: '编号', visible: true, fixed: true },
  { prop: 'avatar', label: '头像', visible: true },
  { prop: 'userName', label: '用户名', visible: true },
  { prop: 'nickName', label: '昵称', visible: true },
  { prop: 'dept', label: '部门', visible: true },
  { prop: 'posts', label: '岗位', visible: true },
  { prop: 'phone', label: '手机号', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'createTime', label: '创建时间', visible: true },
  { prop: 'remark', label: '备注', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

const loading = ref(false);
const submitting = ref(false);
const exporting = ref(false);
const importing = ref(false);
const dialogVisible = ref(false);
const pwdVisible = ref(false);
const importVisible = ref(false);
const userList = ref<UserItem[]>([]);
const total = ref(0);
const roleOptions = ref([]);
const postOptions = ref([]);
const deptOptions = ref([]);
const formRef = ref();
const pwdFormRef = ref();
const uploadRef = ref();
const updateSupport = ref(false);
const importFile = ref(undefined);
const tableRef = ref();
const selectedIds = ref([]);
const authRoleVisible = ref(false);
const authRoleSubmitting = ref(false);
const authRoleUser = ref(null);
const authRoleIds = ref([]);

const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  username: '',
  phone: '',
  status: ''
});

const form = reactive({
  userId: undefined,
  deptId: undefined,
  username: '',
  password: '',
  nickName: '',
  email: '',
  phone: '',
  status: '0',
  remark: '',
  roleIds: [],
  postIds: []
});

const pwdForm = reactive({ userId: undefined, password: '' });

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  nickName: [{ required: true, message: '请输入昵称', trigger: 'blur' }]
};

const pwdRules = {
  password: [
    { required: true, min: 5, max: 20, message: '密码长度为 5-20 个字符', trigger: 'blur' }
  ]
};

// 查询用户列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listUser(queryParams);
    userList.value = res.rows;
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
  queryParams.username = '';
  queryParams.phone = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined) {
  Object.assign(form, {
    userId: undefined,
    deptId: undefined,
    username: '',
    password: '',
    nickName: '',
    email: '',
    phone: '',
    status: '0',
    remark: '',
    roleIds: [],
    postIds: []
  });
  if (row && row.userId) {
    const res = await getUser(row.userId);
    const u = res.data;
    Object.assign(form, {
      userId: u.userId,
      deptId: u.deptId,
      username: u.userName,
      nickName: u.nickName,
      email: u.email,
      phone: u.phone,
      status: u.status,
      remark: u.remark,
      roleIds: (u.roles || []).map((r) => r.roleId),
      postIds: (u.posts || []).map((p) => p.postId)
    });
  }
  dialogVisible.value = true;
}

// 提交用户表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.userId) {
      await updateUser(form);
      ElMessage.success('修改成功');
    } else {
      await addUser(form);
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

// 打开重置密码弹窗
function handleResetPwd(row) {
  pwdForm.userId = row.userId;
  pwdForm.password = '';
  pwdVisible.value = true;
}

// 提交重置密码
async function handleResetSubmit() {
  await pwdFormRef.value.validate();
  submitting.value = true;
  try {
    await resetUserPwd(pwdForm.userId, pwdForm.password);
    ElMessage.success('密码重置成功');
    pwdVisible.value = false;
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 删除选中用户
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除用户「${row.userName}」吗？`, '提示', {
      type: 'warning'
    });
    await delUser(row.userId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 多选
function handleSelectionChange(selection) {
  selectedIds.value = selection.map((s) => s.userId);
}

// 批量删除
async function handleBatchDelete() {
  if (!selectedIds.value.length) return;
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selectedIds.value.length} 个用户吗？`, '提示', {
      type: 'warning'
    });
    await delUser(selectedIds.value.join(','));
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 状态开关（before-change 返回 Promise，失败时切换不生效）
async function handleStatusChange(row) {
  const newStatus = row.status === '0' ? '1' : '0';
  try {
    await changeUserStatus(row.userId, newStatus);
    row.status = newStatus;
    ElMessage.success(newStatus === '0' ? '已启用' : '已停用');
    return true;
  } catch (e) {
    return false;
  }
}

// 分配角色
async function openAuthRole(row) {
  authRoleUser.value = row;
  // 用户当前角色（注意：响应体为 { data: {...} }，此前误用 detail.roles 导致勾选恒空）
  const detail = await getUser(row.userId);
  authRoleIds.value = (detail.data?.roles || []).map((r) => r.roleId);
  // 全量角色
  if (!roleOptions.value.length) {
    roleOptions.value = await listAllRole();
  }
  authRoleVisible.value = true;
}

// 打开分配角色弹窗
async function handleAuthRole() {
  authRoleSubmitting.value = true;
  try {
    await authRole({ userId: authRoleUser.value.userId, roleIds: authRoleIds.value });
    ElMessage.success('分配角色成功');
    authRoleVisible.value = false;
    fetchList();
  } finally {
    authRoleSubmitting.value = false;
  }
}

// ---------- 导入导出 ----------
function openImport() {
  importFile.value = undefined;
  updateSupport.value = false;
  uploadRef.value?.clearFiles();
  importVisible.value = true;
}

// 文件选择回调
function onFileChange(file) {
  importFile.value = file.raw;
}

// 文件移除回调
function onFileRemove() {
  importFile.value = undefined;
}

// 下载导入模板
async function handleDownloadTemplate() {
  const ok = await exportFile(importUserTemplate(), '用户导入模板.xlsx');
  if (ok) ElMessage.success('模板下载成功');
}

// 导出用户数据
async function handleExport() {
  exporting.value = true;
  try {
    const ok = await exportFile(exportUser(queryParams), '用户数据.xlsx');
    if (ok) ElMessage.success('导出成功');
  } finally {
    exporting.value = false;
  }
}

// 执行导入
async function handleImport() {
  if (!importFile.value) {
    ElMessage.warning('请先选择要导入的文件');
    return;
  }
  importing.value = true;
  try {
    const formData = new FormData();
    formData.append('file', importFile.value);
    formData.append('updateSupport', updateSupport.value ? 'true' : 'false');
    const res = await importUser(formData);
    const r = res.data || res;
    ElMessage.success(`导入成功 ${r.successNum} 条，失败 ${r.failureNum} 条`);
    if (r.failMessages && r.failMessages.length) {
      ElMessageBox.alert(r.failMessages.join('\n'), '导入失败明细', {
        type: 'error',
        confirmButtonText: '知道了'
      });
    }
    importVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    importing.value = false;
  }
}

onMounted(async () => {
  fetchList();
  try {
    const res = await listAllRole();
    roleOptions.value = res.data;
  } catch (e) {
    // 无角色权限时角色下拉为空
  }
  try {
    const res = await listAllPost();
    postOptions.value = res.data;
  } catch (e) {
    // 无岗位权限时岗位下拉为空
  }
  try {
    const res = await deptTreeSelect();
    deptOptions.value = res.data;
  } catch (e) {
    // 无权限时部门下拉为空
  }
});
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
