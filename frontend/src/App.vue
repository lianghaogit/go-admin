<template>
  <el-config-provider :size="settings.size" :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { ElConfigProvider } from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import en from 'element-plus/es/locale/lang/en';
import { useSettingsStore } from '@/store/settings';

const settings = useSettingsStore();

// Element Plus 组件语言随界面语言切换
const elementLocale = computed(() => (settings.language === 'en' ? en : zhCn));

onMounted(() => {
  settings.init();
});
</script>
