<template>
  <div class="error-page">
    <div class="error-code">403</div>
    <div class="error-title">抱歉，您没有访问权限</div>
    <div class="error-desc">请联系管理员分配权限后重试</div>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'Error403' });
</script>

<style scoped>
.error-page {
  height: 100%;
  min-height: 60vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  background: var(--go-main-bg, #f0f2f5);
}

.error-code {
  font-size: 96px;
  font-weight: 700;
  color: var(--el-color-primary, #409eff);
  line-height: 1;
  letter-spacing: 8px;
}

.error-title {
  font-size: 18px;
  color: var(--go-text-primary);
  font-weight: 600;
}

.error-desc {
  font-size: 13px;
  color: var(--go-text-secondary);
  margin-bottom: 8px;
}
</style>
