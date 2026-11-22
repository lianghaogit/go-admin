<template>
  <div class="app-container">
    <el-card>
      <div class="table-toolbar">
        <el-button
          v-hasPermi="['tool:gen:import']"
          type="primary"
          :icon="Download"
          @click="openImport"
        >
          导入
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="tableList" border stripe>
        <el-table-column prop="tableId" label="编号" width="70" />
        <el-table-column prop="tableName" label="表名称" width="180" show-overflow-tooltip />
        <el-table-column prop="tableComment" label="表描述" width="160" show-overflow-tooltip />
        <el-table-column prop="className" label="实体类名称" width="150" show-overflow-tooltip />
        <el-table-column label="生成类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="categoryTag(row.tplCategory)">{{ categoryLabel(row.tplCategory) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="moduleName" label="前端模块" width="110" align="center" />
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="170"
        />
        <el-table-column
          :formatter="(row) => parseTime(row.updateTime)"
          prop="updateTime"
          label="更新时间"
          width="170"
        />
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['tool:gen:preview']"
              type="primary"
              link
              :icon="View"
              @click="openPreview(row)"
            >
              预览
            </el-button>
            <el-button
              v-hasPermi="['tool:gen:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openEdit(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['tool:gen:code']"
              type="success"
              link
              :icon="Download"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
            <el-button
              v-hasPermi="['tool:gen:remove']"
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

    <!-- 导入表 -->
    <el-dialog v-model="importVisible" title="导入表" width="620px" destroy-on-close>
      <el-form :inline="true" class="search-form">
        <el-form-item label="表名称">
          <el-input
            v-model="dbQuery.tableName"
            placeholder="请输入表名称"
            clearable
            @keyup.enter="fetchDbTables"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchDbTables">搜索</el-button>
        </el-form-item>
      </el-form>
      <el-table
        ref="dbTableRef"
        v-table-height
        v-loading="dbLoading"
        :data="dbTables"
        border
        max-height="420"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="tableName" label="表名称" width="200" show-overflow-tooltip />
        <el-table-column prop="tableComment" label="表描述" min-width="160" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="importing"
          :disabled="selectedTables.length === 0"
          @click="handleImport"
        >
          导入
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改配置 -->
    <el-dialog
      v-model="editVisible"
      :title="'修改生成配置 - ' + form.tableName"
      width="960px"
      destroy-on-close
    >
      <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="base">
          <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
            <el-form-item label="表名称" prop="tableName">
              <el-input v-model="form.tableName" disabled />
            </el-form-item>
            <el-form-item label="表描述" prop="tableComment">
              <el-input v-model="form.tableComment" placeholder="如 测试客户表" />
            </el-form-item>
            <el-form-item label="实体类名称" prop="className">
              <el-input v-model="form.className" placeholder="如 TestCustomer" />
            </el-form-item>
            <el-form-item label="前端模块" prop="moduleName">
              <el-input v-model="form.moduleName" placeholder="如 system" />
            </el-form-item>
            <el-form-item label="后端包名" prop="packageName">
              <el-input v-model="form.packageName" placeholder="如 system" />
            </el-form-item>
            <el-form-item label="功能名称" prop="functionName">
              <el-input v-model="form.functionName" placeholder="如 测试客户" />
            </el-form-item>
            <el-form-item label="作者" prop="author">
              <el-input v-model="form.author" placeholder="如 go-admin" />
            </el-form-item>
            <el-form-item label="生成类型" prop="tplCategory">
              <el-radio-group v-model="form.tplCategory">
                <el-radio value="crud">单表</el-radio>
                <el-radio value="tree">树形结构</el-radio>
                <el-radio value="sub">主子表</el-radio>
              </el-radio-group>
            </el-form-item>
            <template v-if="form.tplCategory === 'tree'">
              <el-form-item label="父级字段">
                <el-select v-model="form.treeParent" placeholder="如 parent_id" filterable allow-create>
                  <el-option v-for="c in form.columns" :key="c.columnName" :label="c.columnName + '（' + (c.columnComment || '') + '）'" :value="c.columnName" />
                </el-select>
              </el-form-item>
              <el-form-item label="名称字段">
                <el-select v-model="form.treeName" placeholder="树节点显示名称" filterable allow-create>
                  <el-option v-for="c in form.columns" :key="c.columnName" :label="c.columnName + '（' + (c.columnComment || '') + '）'" :value="c.columnName" />
                </el-select>
              </el-form-item>
              <el-form-item label="编码字段">
                <el-select v-model="form.treeCode" placeholder="默认同名称字段" filterable allow-create clearable>
                  <el-option v-for="c in form.columns" :key="c.columnName" :label="c.columnName + '（' + (c.columnComment || '') + '）'" :value="c.columnName" />
                </el-select>
              </el-form-item>
            </template>
            <template v-if="form.tplCategory === 'sub'">
              <el-form-item label="子表名称">
                <el-input v-model="form.subTableName" placeholder="如 demo_order_item（需已存在于数据库）" />
              </el-form-item>
              <el-form-item label="子表外键">
                <el-input v-model="form.subTableFk" placeholder="如 order_id（子表中关联主表的列）" />
              </el-form-item>
            </template>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="字段配置" name="columns">
          <el-table v-table-height :data="form.columns" border size="small" max-height="420">
            <el-table-column prop="columnName" label="列名" width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <span>{{ row.columnName }}</span>
              </template>
            </el-table-column>
            <el-table-column label="列注释" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.columnComment" size="small" placeholder="列注释" />
              </template>
            </el-table-column>
            <el-table-column label="Go 类型" width="90">
              <template #default="{ row }">
                <span>{{ row.goType }}</span>
              </template>
            </el-table-column>
            <el-table-column label="查询方式" width="100">
              <template #default="{ row }">
                <el-select v-model="row.queryType" size="small">
                  <el-option v-for="t in queryTypes" :key="t" :label="t" :value="t" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="表单类型" width="110">
              <template #default="{ row }">
                <el-select v-model="row.formType" size="small">
                  <el-option v-for="t in formTypes" :key="t" :label="t" :value="t" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="必填" width="65" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.isRequired" :active-value="1" :inactive-value="0" />
              </template>
            </el-table-column>
            <el-table-column label="列表" width="65" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.isList" :active-value="1" :inactive-value="0" />
              </template>
            </el-table-column>
            <el-table-column label="查询" width="65" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.isQuery" :active-value="1" :inactive-value="0" />
              </template>
            </el-table-column>
            <el-table-column label="表单" width="65" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.isForm" :active-value="1" :inactive-value="0" />
              </template>
            </el-table-column>
            <el-table-column label="排序" width="80">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.sort"
                  size="small"
                  :min="0"
                  controls-position="right"
                />
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 预览代码 -->
    <el-dialog v-model="previewVisible" title="预览代码" width="860px" destroy-on-close>
      <el-select
        v-model="previewFile"
        class="preview-select"
        placeholder="请选择文件"
        @change="switchPreviewFile"
      >
        <el-option v-for="f in previewFiles" :key="f" :label="f" :value="f" />
      </el-select>
      <pre v-loading="previewLoading" class="code-pre">{{ previewCode }}</pre>
      <template #footer>
        <el-button type="primary" @click="previewVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'ToolGen' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Edit, Delete, View, Download } from '@element-plus/icons-vue';
import {
  listDbTables,
  importGenTables,
  listGenTable,
  getGenTable,
  updateGenTable,
  delGenTable,
  previewGenTable,
  downloadGenCode
} from '@/api/tool/gen';

const loading = ref(false);
const submitting = ref(false);
const importing = ref(false);
const dbLoading = ref(false);
const previewLoading = ref(false);
const tableList = ref([]);
const total = ref(0);
const dbTables = ref([]);
const selectedTables = ref([]);
const importVisible = ref(false);
const editVisible = ref(false);
const previewVisible = ref(false);
const activeTab = ref('base');
const formRef = ref();
const dbTableRef = ref();

const queryParams = reactive({ pageNum: 1, pageSize: 10, tableName: '' });
const dbQuery = reactive({ tableName: '' });
const queryTypes = ['EQ', 'NE', 'GT', 'LT', 'LIKE'];
const formTypes = ['input', 'textarea', 'select', 'date'];

const form = reactive({
  tableId: undefined,
  tableName: '',
  tableComment: '',
  className: '',
  moduleName: 'system',
  packageName: 'system',
  functionName: '',
  author: 'go-admin',
  tplCategory: 'crud',
  treeCode: '',
  treeName: '',
  treeParent: '',
  subTableName: '',
  subTableFk: '',
  columns: []
});

const categoryLabel = (v?: string) => (v === 'tree' ? '树形' : v === 'sub' ? '主子表' : '单表');
const categoryTag = (v?: string) => (v === 'tree' ? 'success' : v === 'sub' ? 'warning' : 'info');

const formRules = {
  tableComment: [{ required: true, message: '请输入表描述', trigger: 'blur' }],
  className: [{ required: true, message: '请输入实体类名称', trigger: 'blur' }],
  moduleName: [{ required: true, message: '请输入前端模块', trigger: 'blur' }],
  functionName: [{ required: true, message: '请输入功能名称', trigger: 'blur' }]
};

const previewFiles = ref([]);
const previewFile = ref('');
const previewCode = ref('');

// 代码预览数据：模板文件路径 -> 代码内容（接口数据宽松类型）
const previewData = ref({} as Record<string, string>);

// 查询代码生成列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listGenTable(queryParams);
    tableList.value = res.rows || [];
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

// 获取数据库表列表
async function fetchDbTables() {
  dbLoading.value = true;
  try {
    const res = await listDbTables(dbQuery);
    dbTables.value = res.data || [];
  } finally {
    dbLoading.value = false;
  }
}

// 打开导入弹窗
function openImport() {
  selectedTables.value = [];
  fetchDbTables();
  importVisible.value = true;
}

// 表格选择变化回调
function handleSelectionChange(rows) {
  selectedTables.value = rows;
}

// 执行导入
async function handleImport() {
  if (selectedTables.value.length === 0) return;
  importing.value = true;
  try {
    const names = selectedTables.value.map((t) => t.tableName).join(',');
    await importGenTables(names);
    ElMessage.success('导入成功');
    importVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    importing.value = false;
  }
}

// 打开编辑弹窗
async function openEdit(row) {
  activeTab.value = 'base';
  const res = await getGenTable(row.tableId);
  const data = res.data || {};
  Object.assign(form, {
    tableId: data.tableId,
    tableName: data.tableName,
    tableComment: data.tableComment,
    className: data.className,
    moduleName: data.moduleName,
    packageName: data.packageName,
    functionName: data.functionName,
    author: data.author,
    tplCategory: data.tplCategory || 'crud',
    treeCode: data.treeCode || '',
    treeName: data.treeName || '',
    treeParent: data.treeParent || '',
    subTableName: data.subTableName || '',
    subTableFk: data.subTableFk || '',
    columns: (data.columns || []).map((c) => ({ ...c }))
  });
  editVisible.value = true;
}

// 保存生成配置
async function handleSave() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    await updateGenTable({
      tableId: form.tableId,
      tableName: form.tableName,
      tableComment: form.tableComment,
      className: form.className,
      moduleName: form.moduleName,
      packageName: form.packageName,
      functionName: form.functionName,
      author: form.author,
      tplCategory: form.tplCategory,
      treeCode: form.treeCode,
      treeName: form.treeName,
      treeParent: form.treeParent,
      subTableName: form.subTableName,
      subTableFk: form.subTableFk,
      columns: form.columns
    });
    ElMessage.success('保存成功');
    editVisible.value = false;
    fetchList();
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 打开预览弹窗
async function openPreview(row) {
  previewVisible.value = true;
  previewLoading.value = true;
  try {
    const res = await previewGenTable(row.tableId);
    const files = res.data || {};
    previewData.value = files;
    previewFiles.value = Object.keys(files);
    if (previewFiles.value.length > 0) {
      previewFile.value = previewFiles.value[0];
      previewCode.value = files[previewFile.value];
    }
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    previewLoading.value = false;
  }
}

// 切换预览文件
function switchPreviewFile(f) {
  previewCode.value = (previewData.value && previewData.value[f]) || '';
}

// 下载生成的代码
async function handleDownload(row) {
  try {
    const res = await downloadGenCode(row.tableId);
    if (res.data instanceof Blob && res.data.type.includes('zip')) {
      const blob = new Blob([res.data], { type: 'application/zip' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = row.className + '.zip';
      link.click();
      URL.revokeObjectURL(link.href);
      ElMessage.success('下载成功');
    } else {
      // blob 内是 JSON 错误
      const text = await res.data.text();
      let msg = '下载失败';
      try {
        msg = JSON.parse(text).msg || msg;
      } catch (e2) {
        // ignore
      }
      ElMessage.error(msg);
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

// 删除选中代码生成
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除表「${row.tableName}」的生成配置吗？`, '提示', {
      type: 'warning'
    });
    await delGenTable(row.tableId);
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

.preview-select {
  width: 100%;
  margin-bottom: 12px;
}

.code-pre {
  margin: 0;
  padding: 12px;
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 2px;
  font-size: 12px;
  line-height: 1.6;
  max-height: 480px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
