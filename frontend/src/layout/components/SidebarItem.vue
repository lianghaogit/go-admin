<template>
  <template v-if="menu.menuType === 'M'">
    <el-sub-menu :index="base + '/' + menu.path">
      <template #title>
        <el-icon v-if="showIcon"><component :is="safeIcon(menu.icon)" /></el-icon>
        <span>{{ menuTitle(menu.menuName) }}</span>
      </template>
      <sidebar-item
        v-for="child in visibleChildren"
        :key="child.menuId"
        :menu="child"
        :base="base + '/' + menu.path"
      />
    </el-sub-menu>
  </template>
  <template v-else-if="menu.menuType === 'C'">
    <el-menu-item :index="base + '/' + menu.path">
      <el-icon v-if="showIcon"><component :is="safeIcon(menu.icon)" /></el-icon>
      <template #title>{{ menuTitle(menu.menuName) }}</template>
    </el-menu-item>
  </template>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useSettingsStore } from '@/store/settings';
import { safeIcon } from '@/utils/iconNames';
import { useMenuTitle } from '@/i18n/translate';

const props = defineProps({
  menu: { type: Object, required: true },
  base: { type: String, default: '' }
});

const settings = useSettingsStore();
const showIcon = computed(() => settings.menuIcon);
const menuTitle = useMenuTitle();

// 只展示目录与菜单，按钮类型不进入侧边栏
const visibleChildren = computed(() =>
  (props.menu.children || []).filter((ch) => ch.menuType !== 'F')
);
</script>
