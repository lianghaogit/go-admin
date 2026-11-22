<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="560px"
    destroy-on-close
    @closed="reset"
  >
    <el-alert
      v-if="hint"
      :title="hint"
      type="info"
      :closable="false"
      show-icon
      class="import-hint"
    />
    <el-form label-width="90px">
      <el-form-item label="导入文件">
        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".xlsx"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :on-exceed="handleExceed"
        >
          <el-icon><UploadFilled /></el-icon>
          <div class="el-upload__text">将文件拖到此处，或<em>点击选择</em></div>
          <template #tip>
            <div class="el-upload__tip">仅支持 xlsx，最大 5MB、最多 5000 行</div>
          </template>
        </el-upload>
      </el-form-item>
      <el-form-item label="更新支持">
        <el-switch v-model="updateSupport" />
        <span class="switch-tip">主键已存在时覆盖更新，否则跳过并提示</span>
      </el-form-item>
      <el-form-item label="导入模板">
        <el-button link type="primary" :loading="templateLoading" @click="handleDownloadTemplate">
          下载导入模板
        </el-button>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="importing" @click="handleImport">确定导入</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { exportFile } from '@/utils/download'

const props = defineProps<{
  modelValue: boolean
  title: string
  templateApi: () => Promise<any>
  importApi: (data: FormData) => Promise<any>
  hint?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value)
})

const uploadRef = ref<any>()
const importFile = ref<File | undefined>()
const updateSupport = ref(false)
const importing = ref(false)
const templateLoading = ref(false)

const MAX_IMPORT_SIZE = 5 * 1024 * 1024

function validateFile(raw: File): boolean {
  if (!/\.xlsx$/i.test(raw.name)) {
    ElMessage.warning('仅支持 xlsx 格式文件')
    return false
  }
  if (raw.size > MAX_IMPORT_SIZE) {
    ElMessage.warning('文件不能超过 5MB')
    return false
  }
  return true
}

function handleFileChange(file: any) {
  const raw = file.raw as File | undefined
  if (!raw) return
  if (!validateFile(raw)) {
    uploadRef.value?.clearFiles()
    return
  }
  importFile.value = raw
}

function handleFileRemove() {
  importFile.value = undefined
}

function handleExceed(files: any[]) {
  uploadRef.value?.clearFiles()
  const raw = files[0] as File | undefined
  if (!raw || !validateFile(raw)) return
  importFile.value = raw
  uploadRef.value?.handleStart(raw)
}

function reset() {
  importFile.value = undefined
  updateSupport.value = false
  uploadRef.value?.clearFiles()
}

async function handleDownloadTemplate() {
  templateLoading.value = true
  try {
    const ok = await exportFile(props.templateApi(), `${props.title}导入模板.xlsx`)
    if (ok) ElMessage.success('模板下载成功')
  } finally {
    templateLoading.value = false
  }
}

async function handleImport() {
  if (!importFile.value) {
    ElMessage.warning('请先选择要导入的文件')
    return
  }
  importing.value = true
  try {
    const formData = new FormData()
    formData.append('file', importFile.value)
    formData.append('updateSupport', updateSupport.value ? 'true' : 'false')
    const res = await props.importApi(formData)
    const result = res?.data || res || {}
    const successNum = result.successNum || 0
    const failureNum = result.failureNum || 0
    if (successNum > 0) {
      ElMessage.success(`导入成功 ${successNum} 条，失败 ${failureNum} 条`)
    } else {
      ElMessage.warning(`未导入任何数据，失败 ${failureNum} 条`)
    }
    if (result.failMessages && result.failMessages.length) {
      ElMessageBox.alert(result.failMessages.join('\n'), '导入失败明细', {
        type: 'error',
        confirmButtonText: '知道了'
      })
    }
    visible.value = false
    emit('success')
  } catch (e) {
    // 错误已由请求拦截器提示
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.import-hint {
  margin-bottom: 14px;
}

.switch-tip {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}
</style>
