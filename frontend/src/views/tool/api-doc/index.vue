<template>
  <div class="app-container">
    <el-card>
      <div class="doc-toolbar">
        <div class="doc-title">
          系统接口文档
          <el-tag size="small" type="info" style="margin-left: 8px">共 {{ total }} 个接口</el-tag>
        </div>
        <el-input
          v-model="keyword"
          placeholder="搜索接口路径 / 名称 / 权限标识"
          clearable
          :prefix-icon="Search"
          style="width: 320px"
        />
      </div>

      <el-collapse v-model="activeNames">
        <el-collapse-item v-for="mod in filteredModules" :key="mod.name" :name="mod.name">
          <template #title>
            <div class="module-title">
              <span class="module-name">{{ moduleLabel(mod.name) }}</span>
              <el-tag size="small" type="info">{{ mod.items.length }} 个</el-tag>
            </div>
          </template>
          <el-table v-table-height :data="mod.items" border size="small">
            <el-table-column label="方法" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="methodType(row.method)" size="small">{{ row.method }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="接口地址" min-width="320" show-overflow-tooltip>
              <template #default="{ row }">
                <code class="api-path">{{ row.path }}</code>
              </template>
            </el-table-column>
            <el-table-column label="接口名称" min-width="160" prop="name" show-overflow-tooltip />
            <el-table-column label="权限标识" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag v-if="row.perms" type="warning" size="small">{{ row.perms }}</el-tag>
                <span v-else class="no-perm">仅需登录</span>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'ToolApiDoc' });

import { computed, onMounted, ref } from 'vue';
import { Search } from '@element-plus/icons-vue';
import { listApiDoc } from '@/api/tool/apiDoc';

const modules = ref([]);
const total = ref(0);
const keyword = ref('');
const activeNames = ref([]);

const moduleLabelMap = {
  base: '基础接口',
  system: '系统管理',
  monitor: '系统监控',
  tool: '系统工具',
  common: '通用功能'
};

// 模块标签映射
function moduleLabel(name) {
  return moduleLabelMap[name] || name;
}

// 请求方法类型映射
function methodType(method) {
  const map = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger'
  };
  return map[method] || 'info';
}

const filteredModules = computed(() => {
  if (!keyword.value) return modules.value;
  const kw = keyword.value.toLowerCase();
  return modules.value
    .map((m) => ({
      ...m,
      items: m.items.filter(
        (i) =>
          i.path.toLowerCase().includes(kw) ||
          (i.name || '').toLowerCase().includes(kw) ||
          (i.perms || '').toLowerCase().includes(kw)
      )
    }))
    .filter((m) => m.items.length > 0);
});

onMounted(async () => {
  const res = await listApiDoc();
  const d = res.data;
  modules.value = d.modules;
  total.value = d.total;
  activeNames.value = d.modules.map((m) => m.name);
});
</script>

<style scoped>
.doc-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  flex-wrap: wrap;
  gap: 10px;
}

.doc-title {
  font-size: 15px;
  font-weight: 600;
}

.module-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.module-name {
  font-size: 14px;
  font-weight: 600;
}

.api-path {
  font-size: 12px;
  color: var(--go-text-primary);
}

.no-perm {
  color: var(--go-text-secondary);
  font-size: 12px;
}
</style>
