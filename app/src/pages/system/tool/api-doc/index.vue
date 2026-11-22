<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="搜索接口路径 / 名称 / 权限标识" placeholder-class="ph" />
      <view class="search-btn plain" @tap="keyword = ''">清空</view>
    </view>
    <view class="summary">共 {{ total }} 个接口{{ keyword ? '，匹配 ' + filtered.length + ' 个' : '' }}</view>

    <view v-if="filtered.length" class="doc-list">
      <view v-for="m in filtered" :key="m.name" class="module">
        <view class="module-head">
          <text class="module-name">{{ m.name }}</text>
          <text class="module-count">{{ m.items.length }}</text>
        </view>
        <view v-for="it in m.items" :key="it.method + it.path" class="api-item">
          <text class="method" :class="it.method.toLowerCase()">{{ it.method }}</text>
          <view class="api-main">
            <text class="api-name">{{ it.name || '未命名接口' }}</text>
            <text class="api-path">{{ it.path }}</text>
            <text v-if="it.perms" class="api-perms">{{ it.perms }}</text>
          </view>
        </view>
      </view>
    </view>
    <view v-else-if="!loading" class="empty">{{ keyword ? '没有匹配的接口' : '暂无接口文档' }}</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { apiDocList } from '@/api/tool'

const keyword = ref('')
const modules = ref([])
const total = ref(0)
const loading = ref(false)

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return modules.value
  return modules.value
    .map((m) => ({ name: m.name, items: (m.items || []).filter((it) => (`${it.method} ${it.path} ${it.name || ''} ${it.perms || ''}`).toLowerCase().includes(kw)) }))
    .filter((m) => m.items.length)
})

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const data = await apiDocList()
    const d = data.data || {}
    modules.value = d.modules || []
    total.value = d.total || 0
  } catch (e) {
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '接口文档' })
  load()
})

onPullDownRefresh(async () => {
  try {
    await load()
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
  } finally {
    uni.stopPullDownRefresh()
  }
})
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 12rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; min-width: 0; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.search-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.summary { margin: 18rpx 4rpx 0; font-size: 24rpx; color: #8a94a6; }
.doc-list { margin-top: 16rpx; background: #fff; border-radius: 18rpx; padding: 8rpx 28rpx 24rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.module { padding-top: 20rpx; }
.module-head { display: flex; align-items: center; gap: 12rpx; padding: 12rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.module-name { font-size: 28rpx; font-weight: 600; color: #1f2d3d; }
.module-count { font-size: 20rpx; color: #409eff; background: rgba(64, 158, 255, 0.1); padding: 2rpx 14rpx; border-radius: 18rpx; }
.api-item { display: flex; align-items: flex-start; gap: 14rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f7f9fc; }
.api-item:last-child { border-bottom: none; }
.method { width: 92rpx; text-align: center; flex-shrink: 0; font-size: 20rpx; padding: 6rpx 0; border-radius: 18rpx; color: #fff; background: #909399; }
.method.get { background: #67c23a; }
.method.post { background: #409eff; }
.method.put { background: #e6a23c; }
.method.delete { background: #f56c6c; }
.api-main { flex: 1; min-width: 0; }
.api-name { display: block; font-size: 26rpx; color: #1f2d3d; }
.api-path { display: block; margin-top: 6rpx; font-size: 22rpx; color: #8a94a6; word-break: break-all; }
.api-perms { display: inline-block; margin-top: 8rpx; font-size: 20rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); padding: 2rpx 12rpx; border-radius: 18rpx; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
</style>
