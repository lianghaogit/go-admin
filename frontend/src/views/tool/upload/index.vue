<template>
  <div class="app-container">
    <el-card>
      <div class="page-title">文件上传（通用上传接口演示）</div>
      <el-alert
        title="支持图片（jpg/png/gif/bmp/webp/svg）、PDF、Office 文档（xls/xlsx/doc/docx）、文本（txt/md）、压缩包（zip/rar/tar/gz）、音视频（mp3/mp4），单个文件不超过 10MB"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-upload
        drag
        :auto-upload="false"
        :limit="1"
        accept=".jpg,.jpeg,.png,.gif,.bmp,.webp,.svg,.pdf,.xlsx,.xls,.docx,.doc,.txt,.md,.zip,.rar,.tar,.gz,.mp3,.mp4"
        :on-change="onFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或<em>点击选择文件</em></div>
      </el-upload>

      <div class="actions">
        <el-button type="primary" :loading="uploading" :disabled="!file" @click="handleUpload">
          开始上传
        </el-button>
      </div>

      <el-divider v-if="result" />

      <div v-if="result" class="result">
        <div class="result-item">
          原文件名：<code>{{ result.originalName }}</code>
        </div>
        <div class="result-item">
          存储文件名：<code>{{ result.fileName }}</code>
        </div>
        <div class="result-item">
          文件大小：<code>{{ (result.size / 1024).toFixed(1) }} KB</code>
        </div>
        <div class="result-item">
          访问地址：
          <el-link type="primary" :href="result.url" target="_blank">{{ result.url }}</el-link>
        </div>
        <div v-if="isImage" class="preview">
          <img
            :src="result.url"
            alt="预览"
            style="
              max-width: 320px;
              max-height: 240px;
              border: 1px solid var(--go-border-regular);
              border-radius: 2px;
            "
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'ToolUpload' });

import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { UploadFilled } from '@element-plus/icons-vue';
import { uploadFile } from '@/api/tool/apiDoc';

const file = ref(undefined);
const uploading = ref(false);
const result = ref(undefined);

const isImage = computed(() => {
  if (!result.value) return false;
  return /\.(jpg|jpeg|png|gif|bmp|webp|svg)$/i.test(result.value.url);
});

// 文件选择回调
function onFileChange(f) {
  file.value = f.raw;
  result.value = undefined;
}

// 执行上传
async function handleUpload() {
  if (!file.value) {
    ElMessage.warning('请先选择文件');
    return;
  }
  uploading.value = true;
  try {
    const res = await uploadFile(file.value);
    result.value = res.data;
    ElMessage.success('上传成功');
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    uploading.value = false;
  }
}
</script>

<style scoped>
.page-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 14px;
}

.actions {
  margin-top: 16px;
}

.result {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
}

.result-item code {
  color: #409eff;
}

.preview {
  margin-top: 6px;
}
</style>
