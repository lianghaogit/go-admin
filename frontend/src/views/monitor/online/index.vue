<template>
  <div class="app-container">
    <el-card>
      <!-- 搜索条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="登录账号">
          <el-input
            v-model="queryParams.userName"
            placeholder="请输入登录账号"
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
          v-hasPermi="['monitor:online:forceLogout']"
          type="danger"
          :icon="Delete"
          @click="handleForceAll"
        >
          强退全部
        </el-button>
      </div>

      <el-table v-table-height v-loading="loading" :data="onlineList" border stripe>
        <el-table-column prop="tokenHash" label="会话编号" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.tokenHash.slice(0, 10) }}…</template>
        </el-table-column>
        <el-table-column prop="userName" label="登录账号" width="120" />
        <el-table-column prop="ip" label="登录IP" width="140" />
        <el-table-column prop="browser" label="浏览器" width="110" />
        <el-table-column prop="os" label="操作系统" width="110" />
        <el-table-column
          :formatter="(row) => parseTime(row.loginTime)"
          prop="loginTime"
          label="登录时间"
          width="170"
        />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-hasPermi="['monitor:online:forceLogout']"
              type="danger"
              link
              :icon="Delete"
              @click="handleForce(row)"
            >
              强退
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
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorOnline' });

import { parseTime } from '@/utils/format';
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Refresh, Delete } from '@element-plus/icons-vue';
import { listOnline, forceLogout, forceLogoutAll } from '@/api/online';

const loading = ref(false);
const onlineList = ref([]);
const total = ref(0);

const queryParams = reactive({ pageNum: 1, pageSize: 10, userName: '' });

// 查询在线用户列表
async function fetchList() {
  loading.value = true;
  try {
    const res = await listOnline(queryParams);
    onlineList.value = res.rows;
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
  queryParams.userName = '';
  handleQuery();
}

// 强制选中的在线用户下线
async function handleForce(row) {
  try {
    await ElMessageBox.confirm(
      `确认强退用户「${row.userName}」的登录会话吗？该用户将被强制下线。`,
      '提示',
      { type: 'warning' }
    );
    await forceLogout(row.tokenHash);
    ElMessage.success('强退成功');
    fetchList();
  } catch (e) {
    // 取消或失败
  }
}

// 强制所有在线用户下线
async function handleForceAll() {
  try {
    await ElMessageBox.confirm(
      '确认强退全部在线用户吗？所有登录会话将被清除（包括自己）。',
      '提示',
      {
        type: 'warning'
      }
    );
    await forceLogoutAll();
    ElMessage.success('已强退全部在线用户');
    // 自己的会话也被清除，触发重新登录
    setTimeout(() => {
      window.location.href = '/login';
    }, 800);
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
