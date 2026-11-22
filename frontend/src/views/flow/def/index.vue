<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="流程名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入流程名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="启用" value="0" />
            <el-option label="停用" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-input
            v-model="queryParams.category"
            placeholder="如：审批/请假"
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
          v-hasPermi="['flow:def:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增流程
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="defList" border stripe>
        <el-table-column prop="flowDefId" label="编号" width="70" />
        <el-table-column prop="defName" label="流程名称" min-width="130" />
        <el-table-column prop="defKey" label="标识" width="110" />
        <el-table-column prop="category" label="分类" width="90" />
        <el-table-column label="节点数" width="80" align="center">
          <template #default="{ row }">{{ row.nodes?.length || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'info'">
              {{ row.status === '0' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="timeoutHours" label="超时(小时)" width="100" align="center">
          <template #default="{ row }">
            {{ row.timeoutHours > 0 ? row.timeoutHours + 'h' : '不限' }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['flow:def:changeStatus']"
              link
              :type="row.status === '0' ? 'warning' : 'primary'"
              @click="handleChangeStatus(row)"
            >
              {{ row.status === '0' ? '停用' : '启用' }}
            </el-button>
            <el-button
              v-hasPermi="['flow:def:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['flow:def:remove']"
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

    <!-- 新增 / 修改流程定义 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.flowDefId ? '修改流程' : '新增流程'"
      width="1280px"
      class="flow-def-dialog"
      destroy-on-close
      top="2vh"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="80px"
        class="flow-form"
      >
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="流程名称" prop="defName">
              <el-input v-model="form.defName" placeholder="如：请假审批" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="流程标识" prop="defKey">
              <el-input
                v-model="form.defKey"
                placeholder="英文唯一标识，如 leave"
                :disabled="!!form.flowDefId"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="分类">
              <el-input v-model="form.category" placeholder="如：审批/请假" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-radio-group v-model="form.status">
                <el-radio value="0">启用</el-radio>
                <el-radio value="1">停用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="超时(小时)">
              <el-input-number v-model="form.timeoutHours" :min="0" :max="720" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="1" placeholder="流程用途说明" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 审批节点编辑器：X6 可视化画布 -->
        <div class="node-editor-title">流程编排</div>
        <FlowCanvas ref="canvasRef" :init-nodes="dialogNodes" @edit="openNodeDialog" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit"> 确定 </el-button>
      </template>
    </el-dialog>

    <!-- 审批节点属性编辑 -->
    <el-dialog
      v-model="nodeDialogVisible"
      title="节点配置"
      width="560px"
      destroy-on-close
      top="18vh"
    >
      <el-form ref="nodeFormRef" :model="currentNode" :rules="nodeFormRules" label-width="90px">
        <el-form-item label="节点名称" prop="nodeName">
          <el-input v-model="currentNode.nodeName" placeholder="如：部门负责人审批" />
        </el-form-item>
        <el-form-item label="审批人类型">
          <el-select v-model="currentNode.approverType" style="width: 100%" @change="onApproverTypeChange">
            <el-option label="指定人员" value="user" />
            <el-option label="按角色" value="role" />
            <el-option label="部门负责人" value="dept_leader" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="currentNode.approverType === 'user'" label="审批人">
          <el-select
            v-model="currentNode.approverIds"
            multiple
            filterable
            placeholder="选择审批人"
            style="width: 100%"
          >
            <el-option v-for="u in userOptions" :key="u.userId" :label="u.nickName" :value="String(u.userId)" />
          </el-select>
        </el-form-item>
        <el-form-item v-else-if="currentNode.approverType === 'role'" label="审批角色">
          <el-select
            v-model="currentNode.approverIds"
            multiple
            filterable
            placeholder="选择角色"
            style="width: 100%"
          >
            <el-option v-for="r in roleOptions" :key="r.roleId" :label="r.roleName" :value="String(r.roleId)" />
          </el-select>
        </el-form-item>
        <el-form-item v-else label="审批人">
          <el-input :model-value="deptLeaderText" disabled />
        </el-form-item>
        <el-form-item label="模式">
          <el-radio-group v-model="currentNode.nodeMode">
            <el-radio value="one">或签（任一通过即推进）</el-radio>
            <el-radio value="any">会签（全部通过才推进）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="驳回策略">
          <el-radio-group v-model="currentNode.rejectTo">
            <el-radio value="0">驳回回发起人</el-radio>
            <el-radio value="1">驳回退上一节点</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="可加签">
          <el-switch v-model="currentNode.canAddSign" active-value="1" inactive-value="0" />
        </el-form-item>
        <el-form-item label="超时(小时)">
          <el-input-number v-model="currentNode.timeoutHours" :min="0" :max="720" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="currentNode.remark" type="textarea" :rows="2" placeholder="节点说明（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nodeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveNodeDialog">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'FlowDef' });

import { onMounted, reactive, ref, nextTick } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue';
import {
  listFlowDef,
  getFlowDef,
  addFlowDef,
  updateFlowDef,
  delFlowDef,
  changeFlowDefStatus
} from '@/api/flow';
import { listUser } from '@/api/user';
import { listAllRole } from '@/api/role';
import FlowCanvas, { type FlowNode } from './FlowCanvas.vue';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
// 画布初始化数据（打开弹窗时设置，FlowCanvas 挂载后自动渲染）
const dialogNodes = ref<FlowNode[]>([]);
const defList = ref([]);
const total = ref(0);
const formRef = ref();
const userOptions = ref([]);
const roleOptions = ref([]);
const canvasRef = ref<InstanceType<typeof FlowCanvas>>();

const queryParams = reactive({ pageNum: 1, pageSize: 10, keyword: '', status: '', category: '' });

// 节点属性编辑弹窗
const nodeDialogVisible = ref(false);
const nodeFormRef = ref();
const currentNode = ref<FlowNode>({
  nodeName: '',
  approverType: 'user',
  approverIds: [],
  nodeMode: 'one',
  canAddSign: '0',
  rejectTo: '0',
  timeoutHours: 0,
  remark: ''
});
const nodeFormRules = {
  nodeName: [{ required: true, message: '请填写节点名称', trigger: 'blur' }]
};

const form = reactive<any>({
  flowDefId: undefined,
  defName: '',
  defKey: '',
  category: '',
  status: '0',
  timeoutHours: 0,
  remark: '',
  nodes: []
});

const formRules = {
  defName: [{ required: true, message: '请输入流程名称', trigger: 'blur' }],
  defKey: [{ required: true, message: '请输入流程标识', trigger: 'blur' }]
};

const deptLeaderText = '发起人所在部门负责人';

// 查询流程定义列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listFlowDef(queryParams);
    defList.value = res.rows;
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
  queryParams.keyword = '';
  queryParams.status = '';
  queryParams.category = '';
  handleQuery();
}

// 打开新增/编辑弹窗
async function openDialog(row = undefined) {
  Object.assign(form, {
    flowDefId: undefined,
    defName: '',
    defKey: '',
    category: '',
    status: '0',
    timeoutHours: 0,
    remark: '',
    nodes: []
  });
  let nodes: FlowNode[] = [];
  if (row && row.flowDefId) {
    const res = await getFlowDef(row.flowDefId);
    const d = res.data;
    Object.assign(form, {
      flowDefId: d.flowDefId,
      defName: d.defName,
      defKey: d.defKey,
      category: d.category,
      status: d.status,
      timeoutHours: d.timeoutHours,
      remark: d.remark
    });
    nodes = (d.nodes || []).map((n: any) => ({
      nodeName: n.nodeName || '未命名节点',
      approverType: n.approverType || 'user',
      approverIds: n.approverIds ? n.approverIds.split(',') : [],
      nodeMode: n.nodeMode || 'one',
      canAddSign: n.canAddSign || '0',
      rejectTo: n.rejectTo || '0',
      timeoutHours: n.timeoutHours || 0,
      remark: n.remark || ''
    }));
  } else {
    nodes = [
      {
        nodeName: '审批节点1',
        approverType: 'user',
        approverIds: [],
        nodeMode: 'one',
        canAddSign: '0',
        rejectTo: '0',
        timeoutHours: 0,
        remark: ''
      }
    ];
  }
  dialogNodes.value = nodes;
  dialogVisible.value = true;
  await nextTick();
  // 渲染由 FlowCanvas 挂载后自动完成（initNodes 驱动），无需再等待 canvasRef
}

// 节点属性编辑（深拷贝快照：取消不污染画布节点数据）
function openNodeDialog(node: FlowNode) {
  currentNode.value = { ...node, approverIds: [...(node.approverIds || [])] };
  nodeDialogVisible.value = true;
}

function onApproverTypeChange() {
  currentNode.value.approverIds = [];
}

async function saveNodeDialog() {
  const ok = await nodeFormRef.value.validate().catch(() => false);
  if (!ok) return;
  nodeDialogVisible.value = false;
  canvasRef.value?.updateNode(currentNode.value);
}

// 提交（画布节点按 y 顺序收集，转后端存储格式）
async function handleSubmit() {
  const ok = await formRef.value.validate().catch(() => false);
  if (!ok) return;
  const nodes = canvasRef.value?.collect() || [];
  if (!nodes.length) {
    ElMessage.warning('至少配置一个审批节点');
    return;
  }
  for (const n of nodes) {
    if (!n.nodeName) {
      ElMessage.warning('请为所有节点填写名称');
      return;
    }
    if (n.approverType === 'user' || n.approverType === 'role') {
      const ids = Array.isArray(n.approverIds) ? n.approverIds : String(n.approverIds || '');
      if (!ids.length) {
        ElMessage.warning(`节点「${n.nodeName}」未配置审批人`);
        return;
      }
    }
  }
  const payload = {
    ...form,
    nodes: nodes.map((n, i) => ({
      nodeOrder: i + 1,
      nodeName: n.nodeName,
      approverType: n.approverType,
      approverIds: Array.isArray(n.approverIds) ? n.approverIds.join(',') : n.approverIds,
      nodeMode: n.nodeMode,
      canAddSign: n.canAddSign,
      rejectTo: n.rejectTo,
      timeoutHours: n.timeoutHours || 0,
      remark: n.remark || ''
    }))
  };
  submitting.value = true;
  try {
    if (form.flowDefId) {
      await updateFlowDef(payload);
      ElMessage.success('修改成功');
    } else {
      await addFlowDef(payload);
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

// 删除流程定义
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除流程「${row.defName}」吗？`, '提示', { type: 'warning' });
    await delFlowDef(row.flowDefId);
    ElMessage.success('删除成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 启停流程定义
async function handleChangeStatus(row) {
  const next = row.status === '0' ? '1' : '0';
  const label = next === '1' ? '停用' : '启用';
  try {
    await ElMessageBox.confirm(`确认${label}流程「${row.defName}」吗？`, '提示', { type: 'warning' });
    await changeFlowDefStatus(row.flowDefId, next);
    ElMessage.success(`${label}成功`);
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

onMounted(() => {
  fetchList();
  // 审批人下拉数据（用户/角色）
  listUser({ pageNum: 1, pageSize: 200 }).then((res) => {
    userOptions.value = res.rows || [];
  });
  listAllRole().then((res) => {
    roleOptions.value = res.data || [];
  });
});
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

.node-editor-title {
  font-weight: 600;
  margin-bottom: 8px;
  border-top: 1px dashed var(--go-border-color, #ebeef5);
  padding-top: 10px;
  flex-shrink: 0;
}

/* ===== 流程编排弹窗：整体加宽 + 紧凑 + 弹窗自身不出现滚动条 ===== */
:deep(.flow-def-dialog) {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 4vh);
  max-width: calc(100vw - 32px);
  overflow: hidden;
}

:deep(.flow-def-dialog .el-dialog__header) {
  flex-shrink: 0;
  padding-bottom: 10px;
}

:deep(.flow-def-dialog .el-dialog__body) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0 20px 12px;
}

:deep(.flow-def-dialog .el-dialog__footer) {
  flex-shrink: 0;
  padding-top: 10px;
}

/* 表单纵向布局：字段区紧凑固定，画布撑满剩余空间 */
:deep(.flow-def-dialog .flow-form) {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

:deep(.flow-def-dialog .flow-form .el-form-item) {
  margin-bottom: 10px;
}
</style>
