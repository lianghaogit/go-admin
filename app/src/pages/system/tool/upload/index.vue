<template>
  <view class="page">
    <view class="tip">上传的图片/文件会保存到服务器 uploads 目录，并返回访问地址。</view>
    <view class="actions">
      <view class="btn" @tap="chooseImage">选择图片上传</view>
      <view v-if="canChooseFile" class="btn plain" @tap="chooseFile">选择文件上传</view>
    </view>

    <view v-if="uploading" class="progress">正在上传 {{ doneCount }}/{{ totalCount }} …</view>

    <view v-if="files.length" class="file-list">
      <view v-for="f in files" :key="f.url" class="file-item">
        <image v-if="f.isImage" :src="f.url" class="thumb" mode="aspectFill" @tap="preview(f)" />
        <view v-else class="thumb file-icon">FILE</view>
        <view class="file-main">
          <text class="file-name">{{ f.name || '未命名文件' }}</text>
          <text class="file-url">{{ f.url }}</text>
        </view>
        <text class="op" @tap="copy(f.url)">复制</text>
      </view>
    </view>
    <view v-else-if="!uploading" class="empty">暂无上传记录</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { uploadFile } from '@/api/tool'
import { getBaseUrl } from '@/utils/config'

const files = ref([])
const uploading = ref(false)
const doneCount = ref(0)
const totalCount = ref(0)
const canChooseFile = ref(typeof uni !== 'undefined' && typeof uni.chooseFile === 'function')

function fullUrl(url) {
  if (!url) return ''
  return /^http(s)?:\/\//.test(url) ? url : getBaseUrl() + url
}

function isImagePath(p) {
  return /\.(png|jpg|jpeg|gif|webp|bmp|svg)$/i.test(p || '')
}

async function doUpload(paths) {
  uploading.value = true
  totalCount.value = paths.length
  doneCount.value = 0
  for (const p of paths) {
    try {
      const res = await uploadFile(p)
      const data = res.data || {}
      const url = data.url || data.fileName || ''
      if (url) {
        const full = fullUrl(url)
        files.value.unshift({ name: data.fileName || p.split('/').pop(), url: full, isImage: isImagePath(url) })
      }
    } catch (e) {
    }
    doneCount.value += 1
  }
  uploading.value = false
  if (doneCount.value) uni.showToast({ title: '上传完成', icon: 'success' })
}

function chooseImage() {
  uni.chooseImage({
    count: 9,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      doUpload(res.tempFilePaths || [])
    }
  })
}

function chooseFile() {
  if (!canChooseFile.value) return
  uni.chooseFile({
    count: 1,
    success: (res) => {
      const paths = res.tempFilePaths || (res.tempFiles || []).map((f) => f.path)
      doUpload(paths)
    }
  })
}

function copy(url) {
  uni.setClipboardData({
    data: url,
    success: () => uni.showToast({ title: '已复制地址', icon: 'none' })
  })
}

function preview(f) {
  uni.previewImage({ urls: [f.url], current: f.url })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '文件上传' })
})
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.tip { font-size: 24rpx; color: #8a94a6; line-height: 1.6; }
.actions { display: flex; gap: 20rpx; margin-top: 24rpx; }
.btn { flex: 1; height: 120rpx; line-height: 120rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 8rpx; }
.btn.plain { background: #fff; color: #409eff; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.progress { margin-top: 24rpx; text-align: center; font-size: 24rpx; color: #409eff; }
.file-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.file-item { display: flex; align-items: center; gap: 20rpx; padding: 24rpx 0; }
.file-item + .file-item { border-top: 1rpx solid #f0f3f7; }
.thumb { width: 88rpx; height: 88rpx; border-radius: 18rpx; flex-shrink: 0; background: #f5f7fa; }
.file-icon { display: flex; align-items: center; justify-content: center; font-size: 20rpx; color: #8a94a6; }
.file-main { flex: 1; min-width: 0; }
.file-name { display: block; font-size: 26rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-url { display: block; margin-top: 6rpx; font-size: 22rpx; color: #8a94a6; word-break: break-all; }
.op { font-size: 24rpx; color: #409eff; padding: 6rpx 18rpx; border-radius: 18rpx; background: rgba(64, 158, 255, 0.08); flex-shrink: 0; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
</style>
