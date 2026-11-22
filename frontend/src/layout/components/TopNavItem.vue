<template>
  <!-- 混合布局：一级目录平铺为菜单项（index=目录段，点击由 TopNav @select 映射跳转其子菜单） -->
  <template v-if="menu.menuType === 'M' && isMix">
    <el-menu-item :index="base + '/' + menu.path">
      <el-icon v-if="showIcon"><component :is="safeIcon(menu.icon)" /></el-icon>
      <template #title>{{ menuTitle(menu.menuName) }}</template>
    </el-menu-item>
  </template>
  <!-- 顶部导航布局：一级目录保持下拉子菜单 -->
  <template v-else-if="menu.menuType === 'M'">
    <el-sub-menu :index="base + '/' + menu.path">
      <template #title>
        <el-icon v-if="showIcon"><component :is="safeIcon(menu.icon)" /></el-icon>
        <span>{{ menuTitle(menu.menuName) }}</span>
      </template>
      <top-nav-item
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
const isMix = computed(() => settings.layoutMode === 'mix');
const menuTitle = useMenuTitle();

const visibleChildren = computed(() =>
  (props.menu.children || []).filter((ch) => ch.menuType !== 'F')
);
</script>
