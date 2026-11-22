<template>
  <el-drawer v-model="settings.drawerVisible" :title="t('layout.settings')" size="300px" append-to-body>
    <div class="setting-group">
      <div class="group-title">{{ t('layout.language') }}</div>
      <el-radio-group v-model="settings.language" size="small" @change="onLanguageChange">
        <el-radio-button value="zh">中文</el-radio-button>
        <el-radio-button value="en">English</el-radio-button>
      </el-radio-group>
    </div>

    <el-divider />

    <div class="setting-group">
      <div class="group-title">主题色</div>
      <div class="theme-colors">
        <div
          v-for="c in presetColors"
          :key="c"
          class="theme-color-item"
          :class="{ active: settings.themeColor === c }"
          :style="{ background: c }"
          @click="setThemeColor(c)"
        />
        <el-color-picker v-model="settings.themeColor" size="small" @change="onThemeColorChange" />
      </div>
    </div>

    <el-divider />

    <div class="setting-group">
      <div class="group-title">布局模式</div>
      <el-radio-group v-model="settings.layoutMode" size="small">
        <el-radio-button value="side">左侧边栏</el-radio-button>
        <el-radio-button value="rightSide">右侧边栏</el-radio-button>
        <el-radio-button value="top">顶部导航</el-radio-button>
        <el-radio-button value="bottom">底部导航</el-radio-button>
        <el-radio-button value="mix">混合</el-radio-button>
        <el-radio-button value="dial">转盘</el-radio-button>
        <el-radio-button value="desktop">桌面</el-radio-button>
        <el-radio-button value="dock">Dock 坞</el-radio-button>
        <el-radio-button value="vtab">垂直页签</el-radio-button>
      </el-radio-group>
    </div>

    <el-divider />

    <div class="setting-group">
      <div class="group-title">侧边栏主题</div>
      <el-radio-group v-model="settings.sidebarTheme" size="small">
        <el-radio-button value="dark">深色</el-radio-button>
        <el-radio-button value="light">浅色</el-radio-button>
      </el-radio-group>
    </div>

    <el-divider />

    <div class="setting-group">
      <div class="group-title">界面设置</div>
      <el-form label-width="90px" size="small">
        <el-form-item label="暗黑模式">
          <el-switch v-model="settings.darkMode" />
        </el-form-item>
        <el-form-item label="显示 Logo">
          <el-switch v-model="settings.showLogo" />
        </el-form-item>
        <el-form-item label="面包屑">
          <el-switch v-model="settings.breadcrumb" />
        </el-form-item>
        <el-form-item label="菜单图标">
          <el-switch v-model="settings.menuIcon" />
        </el-form-item>
        <el-form-item label="多页签">
          <el-switch v-model="settings.showTagsView" />
        </el-form-item>
        <el-form-item label="页脚">
          <el-switch v-model="settings.showFooter" />
        </el-form-item>
        <el-form-item label="组件尺寸">
          <el-select v-model="settings.size" style="width: 140px">
            <el-option label="大" value="large" />
            <el-option label="默认" value="default" />
            <el-option label="小" value="small" />
          </el-select>
        </el-form-item>
      </el-form>
    </div>

    <el-divider />

    <div class="setting-group">
      <el-button type="primary" plain size="small" @click="handleReset">恢复默认</el-button>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useSettingsStore } from '@/store/settings';
import { ElMessage } from 'element-plus';

const { t } = useI18n();
const settings = useSettingsStore();

const presetColors = [
  '#409EFF',
  '#1890FF',
  '#36C6C6',
  '#009688',
  '#536DFE',
  '#FF5C93',
  '#EE4D12',
  '#F5222D',
  '#8B5CF6'
];

// 设置主题色
function setThemeColor(c) {
  settings.themeColor = c;
  settings.applyTheme();
  settings.save();
}

// 主题色变化回调
function onThemeColorChange() {
  settings.applyTheme();
  settings.save();
}

// 切换语言
function onLanguageChange() {
  settings.setLanguage(settings.language);
}

// 恢复默认布局配置
function handleReset() {
  settings.reset();
  ElMessage.success('已恢复默认布局配置');
}

// 其他配置项变更即应用并持久化
watch(
  () => [
    settings.darkMode,
    settings.sidebarTheme,
    settings.layoutMode,
    settings.showLogo,
    settings.breadcrumb,
    settings.showTagsView,
    settings.showFooter,
    settings.menuIcon,
    settings.size
  ],
  () => {
    settings.applyTheme();
    settings.save();
  }
);
</script>

<style scoped>
.setting-group {
  padding: 0 4px;
}

.group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--go-text-primary);
  margin-bottom: 10px;
}

.theme-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

/* 布局模式 6 个选项自动换行，避免溢出 300px 抽屉 */
.setting-group :deep(.el-radio-group) {
  display: flex;
  flex-wrap: wrap;
  row-gap: 8px;
}

.theme-color-item {
  width: 22px;
  height: 22px;
  border-radius: 2px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: transform 0.15s;
}

.theme-color-item:hover {
  transform: scale(1.1);
}

.theme-color-item.active {
  border-color: #fff;
  box-shadow: 0 0 0 2px #409eff;
}
</style>
