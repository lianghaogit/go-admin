<template>
  <div class="icon-select" @click="visible = true">
    <el-input :model-value="modelValue" placeholder="点击选择图标" readonly>
      <template #prefix>
        <el-icon v-if="modelValue" class="icon-preview"><component :is="modelValue" /></el-icon>
      </template>
    </el-input>
    <el-dialog v-model="visible" title="选择图标" width="640px" append-to-body destroy-on-close>
      <el-input v-model="keyword" placeholder="搜索图标名称" clearable class="icon-search" />
      <div class="icon-grid">
        <div
          v-for="icon in filteredIcons"
          :key="icon"
          class="icon-item"
          :class="{ active: icon === modelValue }"
          :title="icon"
          @click="selectIcon(icon)"
        >
          <el-icon :size="22"><component :is="icon" /></el-icon>
          <span class="icon-name">{{ icon }}</span>
        </div>
        <el-empty v-if="!filteredIcons.length" description="没有匹配的图标" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'IconSelect' });

import { computed, ref } from 'vue';

defineProps({
  modelValue: { type: String, default: '' }
});
const emit = defineEmits(['update:modelValue']);

const visible = ref(false);
const keyword = ref('');

// 常用 Element Plus 图标（菜单/按钮场景）
const iconList = [
  'Menu',
  'HomeFilled',
  'Setting',
  'User',
  'UserFilled',
  'Avatar',
  'UserFilled',
  'Goods',
  'ShoppingCart',
  'Monitor',
  'Document',
  'Folder',
  'FolderOpened',
  'Files',
  'Tickets',
  'DataLine',
  'PieChart',
  'Histogram',
  'TrendCharts',
  'DataAnalysis',
  'Coin',
  'Money',
  'Lock',
  'Unlock',
  'Key',
  'Edit',
  'Delete',
  'Plus',
  'CirclePlus',
  'Search',
  'Refresh',
  'Upload',
  'Download',
  'UploadFilled',
  'Bell',
  'ChatDotRound',
  'Message',
  'Phone',
  'Email',
  'Calendar',
  'Clock',
  'Timer',
  'Link',
  'Star',
  'StarFilled',
  'View',
  'Hide',
  'Warning',
  'InfoFilled',
  'SuccessFilled',
  'QuestionFilled',
  'Guide',
  'Aim',
  'Compass',
  'Position',
  'House',
  'OfficeBuilding',
  'School',
  'Reading',
  'Notebook',
  'Collection',
  'Grid',
  'List',
  'Operation',
  'Odometer',
  'Cpu',
  'Headset',
  'Microphone',
  'VideoCamera',
  'Picture',
  'PictureFilled',
  'Camera',
  'Brush',
  'Tools',
  'Setting',
  'SwitchButton',
  'FullScreen',
  'Rank',
  'Sort',
  'Filter',
  'Flag',
  'Location',
  'MapLocation',
  'Medal',
  'Trophy',
  'Present',
  'Box',
  'TakeawayBox'
];

const filteredIcons = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return iconList;
  return iconList.filter((i) => i.toLowerCase().includes(kw));
});

// 选择图标并关闭面板
function selectIcon(icon) {
  emit('update:modelValue', icon);
  visible.value = false;
}
</script>

<style scoped>
.icon-search {
  margin-bottom: 12px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(76px, 1fr));
  gap: 8px;
  max-height: 360px;
  overflow: auto;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 4px;
  border: 1px solid var(--go-border-regular);
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-item:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.icon-item.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.icon-name {
  font-size: 11px;
  color: var(--go-text-secondary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-preview {
  font-size: 16px;
}
</style>
