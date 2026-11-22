<template>
  <div class="app-container">
    <!-- 缓存统计 -->
    <el-row :gutter="12" class="stat-row">
      <el-col v-for="s in stats" :key="s.label" :xs="12" :sm="8" :md="4">
        <el-card class="stat-card">
          <div class="stat-label">{{ s.label }}</div>
          <div class="stat-value" :style="{ color: s.color }">{{ s.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <div class="toolbar">
        <span class="title">缓存键列表（按前缀分组）</span>
        <el-button
          v-hasPermi="['monitor:cache:clear']"
          type="danger"
          :icon="Delete"
          @click="handleClear"
        >
          清空缓存
        </el-button>
        <el-button type="primary" :icon="Refresh" @click="fetchInfo">刷新</el-button>
      </div>

      <el-row :gutter="12">
        <el-col v-for="(keys, group) in groups" :key="group" :xs="24" :md="8">
          <div class="group-box">
            <div class="group-title">
              <el-tag size="small" type="info">{{ group }}</el-tag>
              <span class="group-count">{{ keys.length }} 个键</span>
            </div>
            <div class="key-list">
              <div v-for="k in keys" :key="k" class="key-item" @click="showKeyDetail(k)">
                <span class="key-name">{{ k }}</span>
                <el-icon class="key-icon"><Search /></el-icon>
              </div>
              <el-empty v-if="!keys.length" description="该分组暂无键" :image-size="60" />
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 键详情 -->
    <el-dialog
      v-model="detailVisible"
      :title="`缓存键：${detail.key}`"
      width="520px"
      destroy-on-close
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="键名">{{ detail.key }}</el-descriptions-item>
        <el-descriptions-item label="值">
          <span class="value-text">{{ detail.value }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.createAt }}</el-descriptions-item>
        <el-descriptions-item label="过期时间">{{ detail.expireAt }}</el-descriptions-item>
        <el-descriptions-item label="有效期">{{ detail.ttl }} 秒</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorCache' });

import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Delete, Refresh, Search } from '@element-plus/icons-vue';
import { getCacheInfo, getCacheKey, clearCache } from '@/api/cache';

// 缓存统计对象（接口数据宽松类型，后续可细化）
interface CacheInfo {
  type?: string;
  keyCount?: number;
  hitCount?: number;
  missCount?: number;
  hitRate?: number;
  memBytes?: number;
  groups?: Record<string, string[]>;
}
const info = ref({} as CacheInfo);
const detailVisible = ref(false);
const detail = reactive({ key: '', value: '', createAt: '', expireAt: '', ttl: 0 });

const stats = computed(() => [
  { label: '缓存类型', value: info.value.type || '-', color: '#409EFF' },
  { label: '键数量', value: info.value.keyCount ?? '-', color: '#67C23A' },
  { label: '命中次数', value: info.value.hitCount ?? '-', color: '#E6A23C' },
  { label: '未命中次数', value: info.value.missCount ?? '-', color: '#F56C6C' },
  {
    label: '命中率',
    value: info.value.hitRate !== undefined ? `${info.value.hitRate}%` : '-',
    color: '#409EFF'
  },
  { label: '内存占用', value: fmtBytes(info.value.memBytes), color: '#909399' }
]);

const groups = computed(() => info.value.groups || {});

// 字节数格式化
function fmtBytes(b) {
  if (b === undefined || b === null) return '-';
  if (b < 1024) return `${b} B`;
  return `${(b / 1024).toFixed(1)} KB`;
}

// 获取监控信息
async function fetchInfo() {
  try {
    const res = await getCacheInfo();
    info.value = res.data;
  } catch (e) {
    // 拦截器已提示
  }
}

// 查看缓存键详情
async function showKeyDetail(key) {
  try {
    const res = await getCacheKey(key);
    Object.assign(detail, res.data);
    detailVisible.value = true;
  } catch (e) {
    // 键可能已过期
  }
}

// 清空全部缓存（需二次确认）
async function handleClear() {
  try {
    await ElMessageBox.confirm('确认清空全部缓存吗？清空后验证码等临时数据将失效。', '提示', {
      type: 'warning'
    });
    await clearCache();
    ElMessage.success('清空成功');
    fetchInfo();
  } catch (e) {
    // 取消或失败
  }
}

onMounted(fetchInfo);
</script>

<style scoped>
.stat-row {
  margin-bottom: 14px;
}

.stat-card {
  text-align: center;
  padding: 4px 0;
}

.stat-label {
  font-size: 12px;
  color: var(--go-text-secondary);
  margin-bottom: 6px;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.title {
  flex: 1;
  font-weight: 600;
  color: var(--go-text-primary);
}

.group-box {
  border: 1px solid var(--go-border-regular);
  border-radius: 2px;
  padding: 10px;
  margin-bottom: 12px;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.group-count {
  font-size: 12px;
  color: var(--go-text-secondary);
}

.key-list {
  max-height: 320px;
  overflow: auto;
}

.key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  border-radius: 2px;
  cursor: pointer;
  font-size: 13px;
}

.key-item:hover {
  background: var(--go-fill-light);
  color: #409eff;
}

.key-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-icon {
  color: #c0c4cc;
}

.value-text {
  word-break: break-all;
  font-family: Consolas, monospace;
}
</style>
