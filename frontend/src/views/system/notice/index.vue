<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="公告标题">
          <el-input
            v-model="queryParams.noticeTitle"
            placeholder="请输入公告标题"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="公告类型">
          <el-select v-model="queryParams.noticeType" placeholder="请选择类型" clearable>
            <el-option label="通知" value="1" />
            <el-option label="公告" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="正常" value="0" />
            <el-option label="关闭" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">搜索</el-button>
          <el-button :icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="table-toolbar">
        <el-button
          v-hasPermi="['system:notice:add']"
          type="primary"
          :icon="Plus"
          @click="openDialog()"
        >
          新增公告
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="noticeList" border stripe>
        <el-table-column prop="noticeId" label="编号" width="70" />
        <el-table-column
          prop="noticeTitle"
          label="公告标题"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.noticeType === '1' ? 'primary' : 'warning'">
              {{ row.noticeType === '1' ? '通知' : '公告' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '0' ? 'success' : 'info'">
              {{ row.status === '0' ? '正常' : '关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createBy" label="创建者" width="100" />
        <el-table-column
          :formatter="(row) => parseTime(row.createTime)"
          prop="createTime"
          label="创建时间"
          width="170"
        />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['system:notice:query']"
              type="info"
              link
              :icon="View"
              @click="openDetail(row)"
            >
              查看
            </el-button>
            <el-button
              v-hasPermi="['system:notice:edit']"
              type="primary"
              link
              :icon="Edit"
              @click="openDialog(row)"
            >
              修改
            </el-button>
            <el-button
              v-hasPermi="['system:notice:remove']"
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

    <!-- 新增 / 修改公告 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.noticeId ? '修改公告' : '新增公告'"
      width="640px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="公告标题" prop="noticeTitle">
          <el-input v-model="form.noticeTitle" placeholder="请输入公告标题" />
        </el-form-item>
        <el-form-item label="公告类型">
          <el-radio-group v-model="form.noticeType">
            <el-radio value="1">通知</el-radio>
            <el-radio value="2">公告</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="公告内容" prop="noticeContent">
          <el-input
            v-model="form.noticeContent"
            type="textarea"
            :rows="6"
            placeholder="请输入公告内容"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="0">正常</el-radio>
            <el-radio value="1">关闭</el-radio>
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

    <!-- 查看公告 -->
    <el-dialog v-model="detailVisible" title="公告详情" width="640px" destroy-on-close>
      <div class="detail-box">
        <h3 class="detail-title">{{ detail.noticeTitle }}</h3>
        <div class="detail-meta">
          <el-tag size="small" :type="detail.noticeType === '1' ? 'primary' : 'warning'">
            {{ detail.noticeType === '1' ? '通知' : '公告' }}
          </el-tag>
          <span>创建者：{{ detail.createBy || '-' }}</span>
          <span>创建时间：{{ parseTime(detail.createTime) || '-' }}</span>
        </div>
        <div class="detail-content">{{ detail.noticeContent }}</div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemNotice' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Plus, Edit, Delete, View } from '@element-plus/icons-vue';
import { listNotice, getNotice, addNotice, updateNotice, delNotice } from '@/api/notice';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const detailVisible = ref(false);
const noticeList = ref([]);
const total = ref(0);
const formRef = ref();

// 公告详情对象（接口数据宽松类型，后续可细化）
interface NoticeDetail {
  noticeId?: number;
  noticeTitle?: string;
  noticeType?: string;
  createBy?: string;
  createTime?: string;
  noticeContent?: string;
}
const detail = ref({} as NoticeDetail);

const queryParams = reactive({
  pageNum: 1,
  pageSize: 10,
  noticeTitle: '',
  noticeType: '',
  status: ''
});

const form = reactive({
  noticeId: undefined,
  noticeTitle: '',
  noticeType: '1',
  noticeContent: '',
  status: '0',
  remark: ''
});

const formRules = {
  noticeTitle: [{ required: true, message: '请输入公告标题', trigger: 'blur' }],
  noticeContent: [{ required: true, message: '请输入公告内容', trigger: 'blur' }]
};

// 查询公告列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listNotice(queryParams);
    noticeList.value = res.rows;
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
  queryParams.noticeTitle = '';
  queryParams.noticeType = '';
  queryParams.status = '';
  handleQuery();
}

// 打开新增/编辑弹窗
function openDialog(row = undefined) {
  Object.assign(form, {
    noticeId: undefined,
    noticeTitle: '',
    noticeType: '1',
    noticeContent: '',
    status: '0',
    remark: ''
  });
  if (row && row.noticeId) {
    Object.assign(form, {
      noticeId: row.noticeId,
      noticeTitle: row.noticeTitle,
      noticeType: row.noticeType,
      noticeContent: row.noticeContent,
      status: row.status,
      remark: row.remark
    });
  }
  dialogVisible.value = true;
}

// 查看详情
async function openDetail(row) {
  const res = await getNotice(row.noticeId);
  detail.value = res.data;
  detailVisible.value = true;
}

// 提交公告表单（新增/修改）
async function handleSubmit() {
  await formRef.value.validate();
  submitting.value = true;
  try {
    if (form.noticeId) {
      await updateNotice(form);
      ElMessage.success('修改成功');
    } else {
      await addNotice(form);
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

// 删除选中公告
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除公告「${row.noticeTitle}」吗？`, '提示', {
      type: 'warning'
    });
    await delNotice(row.noticeId);
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

.detail-box {
  padding: 4px 8px;
}

.detail-title {
  margin: 0 0 10px;
  font-size: 18px;
  color: var(--go-text-primary);
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
  font-size: 12px;
  color: var(--go-text-secondary);
}

.detail-content {
  line-height: 1.8;
  color: var(--go-text-regular);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
