<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="角色名称">
          <el-input
            v-model="queryParams.roleName"
            placeholder="请输入角色名称"
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
          v-hasPermi="['system:role:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增角色
        </el-button>
        <div class="toolbar-right">
          <column-setting v-model:columns="columns" storage-key="role" />
        </div>
      </div>

      <el-table v-table-height v-loading="loading" :data="roleList" border stripe>
        <el-table-column v-if="col('roleId').visible" prop="roleId" label="编号" width="70" />
        <el-table-column
          v-if="col('roleName').visible"
          prop="roleName"
          label="角色名称"
          width="160"
        />
        <el-table-column
          v-if="col('roleKey').visible"
          prop="roleKey"
          label="权限字符"
          width="140"
        />
        <el-table-column
          v-if="col('roleSort').visible"
          prop="roleSort"
          label="排序"
          width="80"
          align="center"
        />
        <el-table-column
          v-if="col('dataScope').visible"
          label="数据范围"
          width="130"
          align="center"
        >
          <template #default="{ row }">
            <el-tag size="small">{{ dataScopeName(row.dataScope) }}</el-tag>
          </template>
        </el-table-column>
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
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:role:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:role:edit']"
              type="info"
              link
              :icon="UserFilled"
              @click="openAuthUser(row)"
            >
              分配用户
            </el-button>
            <el-button
              v-hasPermi="['system:role:remove']"
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

    <!-- 新增 / 修改角色 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.roleId ? '修改角色' : '新增角色'"
      width="620px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="角色名称" prop="roleName">
          <el-input v-model="form.roleName" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="权限字符" prop="roleKey">
          <el-input v-model="form.roleKey" placeholder="如 system:user" />
        </el-form-item>
        <el-form-item label="显示排序">
          <el-input-number v-model="form.roleSort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="数据权限">
          <el-radio-group v-model="form.dataScope">
            <el-radio value="1">全部数据权限</el-radio>
            <el-radio value="2">自定义数据权限</el-radio>
            <el-radio value="3">本部门数据权限</el-radio>
            <el-radio value="4">本部门及以下数据权限</el-radio>
            <el-radio value="5">仅本人数据权限</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.dataScope === '2'" label="部门权限">
          <div class="menu-tree-box">
            <el-tree
              ref="deptTreeRef"
              :data="deptOptions"
              show-checkbox
              node-key="deptId"
              :props="{ label: 'deptName', children: 'children' }"
              default-expand-all
            />
          </div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" />
        </el-form-item>
        <el-form-item label="菜单权限">
          <div class="menu-tree-box">
            <el-tree
              ref="menuTreeRef"
              :data="menuOptions"
              show-checkbox
              node-key="menuId"
              :props="{ label: 'menuName', children: 'children' }"
              default-expand-all
            />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit"> 确定 </el-button>
      </template>
    </el-dialog>

    <!-- 分配用户 -->
    <el-dialog v-model="authUserVisible" title="分配用户" width="760px" destroy-on-close>
      <div class="auth-user-head">
        <span>
          角色：<strong>{{ authRole?.roleName }}</strong>
          <el-tag size="small" class="auth-role-key">{{ authRole?.roleKey }}</el-tag>
        </span>
        <el-button
          v-hasPermi="['system:role:edit']"
          type="primary"
          :icon="Plus"
          @click="openSelectUser"
        >
          选择用户
        </el-button>
      </div>
      <el-table
        v-loading="allocLoading"
        :data="allocatedUsers"
        border
        stripe
        size="small"
        @selection-change="(s) => (allocSelected = s)"
      >
        <el-table-column type="selection" width="40" align="center" />
        <el-table-column prop="userId" label="编号" width="70" />
        <el-table-column prop="userName" label="用户名" width="130" />
        <el-table-column prop="nickName" label="昵称" width="130" />
        <el-table-column label="部门" width="140">
          <template #default="{ row }">{{ row.dept?.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="160"
        />
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:role:edit']"
              type="danger"
              link
              @click="handleCancelUser(row)"
            >
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="allocQuery.pageNum"
        v-model:page-size="allocQuery.pageSize"
        class="pagination"
        :total="allocTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchAllocated"
        @current-change="fetchAllocated"
      />
      <template #footer>
        <el-button
          type="danger"
          plain
          :disabled="!allocSelected.length"
          @click="handleCancelSelected"
        >
          批量移除
        </el-button>
        <el-button @click="authUserVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 选择用户（未分配） -->
    <el-dialog v-model="selectUserVisible" title="选择用户" width="680px" destroy-on-close>
      <el-form :inline="true" class="search-form">
        <el-form-item label="用户名">
          <el-input
            v-model="unallocQuery.username"
            placeholder="请输入用户名"
            clearable
            @keyup.enter="fetchUnallocated"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchUnallocated">搜索</el-button>
        </el-form-item>
      </el-form>
      <el-table
        v-loading="unallocLoading"
        :data="unallocatedUsers"
        border
        stripe
        size="small"
        @selection-change="(s) => (unallocSelected = s)"
      >
        <el-table-column type="selection" width="40" align="center" />
        <el-table-column prop="userId" label="编号" width="70" />
        <el-table-column prop="userName" label="用户名" width="130" />
        <el-table-column prop="nickName" label="昵称" width="130" />
        <el-table-column label="部门" width="140">
          <template #default="{ row }">{{ row.dept?.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="unallocQuery.pageNum"
        v-model:page-size="unallocQuery.pageSize"
        class="pagination"
        :total="unallocTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchUnallocated"
        @current-change="fetchUnallocated"
      />
      <template #footer>
        <el-button @click="selectUserVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!unallocSelected.length" @click="handleAssignSelected">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemRole' });

import { parseTime } from '@/utils/format';
import { nextTick, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete, UserFilled } from '@element-plus/icons-vue';
import {
  listRole,
  addRole,
  updateRole,
  delRole,
  getRole,
  roleMenuTreeselect,
  allocatedUserList,
  unallocatedUserList,
  assignAuthUsers,
  cancelAuthUsers
} from '@/api/role';
import { menuTreeselect } from '@/api/menu';
import { deptTreeSelect } from '@/api/dept';
import ColumnSetting from '@/components/ColumnSetting/index.vue';
import type { ColumnItem } from '@/types/column';

// 表格列配置（ColumnSetting 显隐 + localStorage 持久化）
const columns = ref<ColumnItem[]>([
  { prop: 'roleId', label: '编号', visible: true, fixed: true },
  { prop: 'roleName', label: '角色名称', visible: true },
  { prop: 'roleKey', label: '权限字符', visible: true },
  { prop: 'roleSort', label: '排序', visible: true },
  { prop: 'dataScope', label: '数据范围', visible: true },
  { prop: 'status', label: '状态', visible: true },
  { prop: 'remark', label: '备注', visible: true },
  { prop: 'createTime', label: '创建时间', visible: true }
]);

// 获取列配置（未配置列默认显示）
const col = (prop) => columns.value.find((c) => c.prop === prop) || { visible: true };

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const roleList = ref([]);
const total = ref(0);
const menuOptions = ref([]);
const deptOptions = ref([]);
const menuTreeRef = ref();
const deptTreeRef = ref();
const formRef = ref();

const queryParams = reactive({ pageNum: 1, pageSize: 10, roleName: '', status: '' });

const form = reactive({
  roleId: undefined,
  roleName: '',
  roleKey: '',
  roleSort: 0,
  status: '0',
  dataScope: '1',
  deptIds: [],
  remark: '',
  menuIds: []
});

const dataScopeMap = {
  1: '全部',
  2: '自定义',
  3: '本部门',
  4: '本部门及以下',
  5: '仅本人'
};

// 数据权限范围名称
function dataScopeName(v) {
  return dataScopeMap[v] || '全部';
}

const formRules = {
  roleName: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  roleKey: [{ required: true, message: '请输入权限字符', trigger: 'blur' }]
};

// 查询角色列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listRole(queryParams);
    roleList.value = res.rows;
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
  queryParams.roleName = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined) {
  Object.assign(form, {
    roleId: undefined,
    roleName: '',
    roleKey: '',
    roleSort: 0,
    status: '0',
    dataScope: '1',
    deptIds: [],
    remark: '',
    menuIds: []
  });
  const [menuRes, deptRes] = await Promise.all([menuTreeselect(), deptTreeSelect()]);
  menuOptions.value = menuRes.data;
  deptOptions.value = deptRes.data;
  dialogVisible.value = true;
  await nextTick();
  if (row && row.roleId) {
    Object.assign(form, {
      roleId: row.roleId,
      roleName: row.roleName,
      roleKey: row.roleKey,
      roleSort: row.roleSort,
      status: row.status,
      remark: row.remark
    });
    const [roleRes, roleDetail] = await Promise.all([
      roleMenuTreeselect(row.roleId),
      getRole(row.roleId)
    ]);
    menuTreeRef.value.setCheckedKeys(roleRes.data.checkedKeys);
    if (roleDetail.data && roleDetail.data.dataScope) {
      form.dataScope = roleDetail.data.dataScope;
    }
    if (roleDetail.data && roleDetail.data.deptIds) {
      form.deptIds = roleDetail.data.deptIds;
      nextTick(() => {
        deptTreeRef.value?.setCheckedKeys(roleDetail.data.deptIds);
      });
    }
  }
}

// 提交角色表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  // 收集勾选与半选节点（父级目录需要半选键保证权限链完整）
  const checkedKeys = menuTreeRef.value.getCheckedKeys();
  const halfCheckedKeys = menuTreeRef.value.getHalfCheckedKeys();
  form.menuIds = [...checkedKeys, ...halfCheckedKeys];
  // 自定义数据权限时收集勾选的部门（精确部门集合）
  if (form.dataScope === '2') {
    form.deptIds = deptTreeRef.value.getCheckedKeys();
  } else {
    form.deptIds = [];
  }
  submitting.value = true;
  try {
    if (form.roleId) {
      await updateRole(form);
      ElMessage.success('修改成功');
    } else {
      await addRole(form);
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

// 删除选中角色
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除角色「${row.roleName}」吗？`, '提示', {
      type: 'warning'
    });
    await delRole(row.roleId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// ---------- 分配用户 ----------
const authUserVisible = ref(false);
const authRole = ref(null);
const allocLoading = ref(false);
const unallocLoading = ref(false);
const allocatedUsers = ref([]);
const unallocatedUsers = ref([]);
const allocTotal = ref(0);
const unallocTotal = ref(0);
const allocQuery = reactive({ pageNum: 1, pageSize: 10 });
const unallocQuery = reactive({ pageNum: 1, pageSize: 10, username: '' });
const allocSelected = ref([]);
const unallocSelected = ref([]);
const selectUserVisible = ref(false);

// 打开分配用户弹窗
async function openAuthUser(row) {
  authRole.value = row;
  allocQuery.pageNum = 1;
  authUserVisible.value = true;
  await fetchAllocated();
}

// 查询已分配用户列表
async function fetchAllocated() {
  allocLoading.value = true;
  try {
    const res = await allocatedUserList(authRole.value.roleId, allocQuery);
    allocatedUsers.value = res.rows;
    allocTotal.value = res.total;
  } finally {
    allocLoading.value = false;
  }
}

// 打开选择用户弹窗
async function openSelectUser() {
  unallocQuery.pageNum = 1;
  unallocQuery.username = '';
  unallocSelected.value = [];
  selectUserVisible.value = true;
  await fetchUnallocated();
}

// 查询未分配用户列表
async function fetchUnallocated() {
  unallocLoading.value = true;
  try {
    const res = await unallocatedUserList(authRole.value.roleId, unallocQuery);
    unallocatedUsers.value = res.rows;
    unallocTotal.value = res.total;
  } finally {
    unallocLoading.value = false;
  }
}

// 确认分配选中的用户
async function handleAssignSelected() {
  const ids = unallocSelected.value.map((u) => u.userId);
  if (!ids.length) return;
  try {
    await assignAuthUsers(authRole.value.roleId, ids);
    ElMessage.success('分配成功');
    selectUserVisible.value = false;
    fetchAllocated();
  } catch (e) {
    // 拦截器提示
  }
}

// 确认取消选中的用户
async function handleCancelSelected() {
  const ids = allocSelected.value.map((u) => u.userId);
  if (!ids.length) return;
  try {
    await ElMessageBox.confirm(`确认移除选中的 ${ids.length} 个用户吗？`, '提示', {
      type: 'warning'
    });
    await cancelAuthUsers(authRole.value.roleId, ids);
    ElMessage.success('移除成功');
    allocSelected.value = [];
    fetchAllocated();
  } catch (e) {
    // 取消或失败
  }
}

// 取消单个用户的角色
async function handleCancelUser(row) {
  try {
    await ElMessageBox.confirm(`确认移除用户「${row.userName}」吗？`, '提示', { type: 'warning' });
    await cancelAuthUsers(authRole.value.roleId, [row.userId]);
    ElMessage.success('移除成功');
    fetchAllocated();
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

.menu-tree-box {
  width: 100%;
  max-height: 300px;
  overflow: auto;
  border: 1px solid var(--go-border-regular);
  border-radius: 2px;
  padding: 8px;
}

.auth-user-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.auth-role-key {
  margin-left: 8px;
}
</style>
