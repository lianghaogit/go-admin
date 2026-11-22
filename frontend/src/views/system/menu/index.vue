<template>
  <div class="app-container">
    <el-card>
      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:menu:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增菜单
        </el-button>
      </div>

      <el-table
        v-table-height
        v-loading="loading"
        :data="menuList"
        border
        row-key="menuId"
        default-expand-all
      >
        <el-table-column prop="menuName" label="菜单名称" width="180" />
        <el-table-column label="图标" width="70" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="orderNum" label="排序" width="70" align="center" />
        <el-table-column prop="perms" label="权限标识" width="180" show-overflow-tooltip />
        <el-table-column prop="component" label="组件路径" show-overflow-tooltip />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.menuType === 'M'" type="warning">目录</el-tag>
            <el-tag v-else-if="row.menuType === 'C'" type="success">菜单</el-tag>
            <el-tag v-else type="info">按钮</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'danger'">
              {{ row.status === '0' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:menu:add']"
              type="primary"
              link
              :icon="Plus"
              @click="openDialog(row)"
            >
              新增
            </el-button>
            <el-button
              v-hasPermi="['system:menu:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row, true)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:menu:remove']"
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

    <!-- 新增 / 修改菜单 -->
    <el-dialog v-model="dialogVisible" :title="title" width="620px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parentId"
            :data="menuOptions"
            :props="{ label: 'menuName', value: 'menuId', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单类型">
          <el-radio-group v-model="form.menuType">
            <el-radio-button value="M">目录</el-radio-button>
            <el-radio-button value="C">菜单</el-radio-button>
            <el-radio-button value="F">按钮</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="menuName">
          <el-input v-model="form.menuName" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item v-if="form.menuType !== 'F'" label="显示排序">
          <el-input-number v-model="form.orderNum" :min="0" />
        </el-form-item>
        <el-form-item v-if="form.menuType !== 'F'" label="路由地址">
          <el-input v-model="form.path" placeholder="请输入路由地址" />
        </el-form-item>
        <el-form-item v-if="form.menuType === 'C'" label="组件路径">
          <el-input v-model="form.component" placeholder="如 system/user/index" />
        </el-form-item>
        <el-form-item v-if="form.menuType !== 'M'" label="权限标识">
          <el-input v-model="form.perms" placeholder="如 system:user:list" />
        </el-form-item>
        <el-form-item v-if="form.menuType !== 'F'" label="菜单图标">
          <icon-select v-model="form.icon" />
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
defineOptions({ name: 'SystemMenu' });

import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';
import { listMenu, menuTreeselect, getMenu, addMenu, updateMenu, delMenu } from '@/api/menu';
import IconSelect from '@/components/IconSelect/index.vue';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const title = ref('新增菜单');
const menuList = ref([]);
const menuOptions = ref([]);
const formRef = ref();

const defaultForm = {
  menuId: undefined,
  menuName: '',
  parentId: 0,
  orderNum: 0,
  path: '',
  component: '',
  menuType: 'C',
  perms: '',
  icon: '',
  visible: '0',
  status: '0'
};

const form = reactive({ ...defaultForm });

const formRules = {
  menuName: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }]
};

// 查询菜单列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listMenu();
    menuList.value = res.data;
  } finally {
    loading.value = false;
  }
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined, isEdit = undefined) {
  Object.assign(form, defaultForm);
  const res = await menuTreeselect();
  // 根节点「主目录」便于回到顶层
  menuOptions.value = [{ menuId: 0, menuName: '主目录', children: res.data }];
  if (row && row.menuId) {
    if (isEdit) {
      const detail = await getMenu(row.menuId);
      Object.assign(form, detail.data);
      title.value = '修改菜单';
    } else {
      Object.assign(form, { parentId: row.menuId });
      title.value = '新增子菜单';
    }
  } else {
    title.value = '新增菜单';
  }
  dialogVisible.value = true;
}

// 提交菜单表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.menuId) {
      await updateMenu(form);
      ElMessage.success('修改成功');
    } else {
      await addMenu(form);
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

// 删除选中菜单
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除菜单「${row.menuName}」吗？`, '提示', {
      type: 'warning'
    });
    await delMenu(row.menuId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

onMounted(fetchList);
</script>

<style scoped>
.table-toolbar {
  margin-bottom: 12px;
}
</style>
