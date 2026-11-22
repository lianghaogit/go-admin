<template>
  <el-popover
    placement="bottom-end"
    :width="200"
    trigger="click"
    popper-class="column-setting-popper"
  >
    <template #reference>
      <el-tooltip content="列设置" placement="top">
        <el-button :icon="Setting" circle />
      </el-tooltip>
    </template>
    <div class="column-setting">
      <div class="column-setting-title">表格列设置</div>
      <el-checkbox-group v-model="checkedKeys" class="column-list">
        <el-checkbox v-for="c in columnList" :key="c.prop" :value="c.prop" :disabled="c.disabled">
          {{ c.label }}
        </el-checkbox>
      </el-checkbox-group>
      <div class="column-setting-footer">
        <el-button link type="primary" size="small" @click="reset">恢复默认</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
defineOptions({ name: 'ColumnSetting' });

import { computed, ref, watch } from 'vue';
import type { PropType } from 'vue';
import { Setting } from '@element-plus/icons-vue';
import type { ColumnItem } from '@/types/column';

const props = defineProps({
  columns: { type: Array as PropType<ColumnItem[]>, required: true },
  storageKey: { type: String, required: true }
});
const emit = defineEmits(['update:columns']);

const columnList = computed(() => props.columns.filter((c) => !c.fixed));
const checkedKeys = ref([] as string[]);

// 初始化：从 localStorage 读取用户保存的显隐，合并默认值
function loadSaved(): ColumnItem[] {
  try {
    const saved = JSON.parse(localStorage.getItem(`column-setting:${props.storageKey}`));
    if (Array.isArray(saved)) {
      const mapped = props.columns
        .filter((c) => !c.fixed)
        .map((c) => (saved.includes(c.prop) ? { ...c, visible: true } : { ...c, visible: false }));
      return [...mapped, ...props.columns.filter((c) => c.fixed)] as ColumnItem[];
    }
  } catch (e) {
    // 解析失败用默认
  }
  return props.columns.map((c) => ({ ...c, visible: c.visible !== false }));
}

// 同步 columns：把勾选状态写回父组件并持久化
function syncColumns(keys: string[]) {
  const next = props.columns.map((c) =>
    c.fixed ? { ...c } : { ...c, visible: keys.includes(c.prop) }
  ) as ColumnItem[];
  emit('update:columns', next);
  localStorage.setItem(`column-setting:${props.storageKey}`, JSON.stringify(keys));
}

// 恢复默认列设置（回到初始显隐并持久化）
function reset() {
  const def = props.columns.filter((c) => !c.fixed && c.visible !== false).map((c) => c.prop);
  checkedKeys.value = def;
  syncColumns(def);
}

watch(checkedKeys, (keys) => syncColumns(keys), { deep: true });

// 首次加载：用保存值初始化 checkedKeys
checkedKeys.value = loadSaved()
  .filter((c) => !c.fixed && c.visible)
  .map((c) => c.prop);
</script>

<style scoped>
.column-setting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--go-text-primary);
  margin-bottom: 8px;
}

.column-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 320px;
  overflow: auto;
}

.column-setting-footer {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #ebeef5;
  text-align: right;
}
</style>
