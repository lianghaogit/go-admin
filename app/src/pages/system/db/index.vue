<template>
  <view class="page">
    <view class="stats">
      <view class="stat">
        <text class="stat-num">{{ info.tableCount || 0 }}</text>
        <text class="stat-label">数据表</text>
      </view>
      <view class="stat">
        <text class="stat-num">{{ info.totalRows || 0 }}</text>
        <text class="stat-label">总行数</text>
      </view>
      <view class="stat">
        <text class="stat-num">{{ fmtBytes(info.fileSize) }}</text>
        <text class="stat-label">文件大小</text>
      </view>
    </view>

    <view class="card">
      <view class="row"><text class="k">数据库版本</text><text class="v">{{ info.version || '—' }}</text></view>
      <view class="row"><text class="k">日志模式</text><text class="v">{{ info.journalMode || '—' }}</text></view>
    </view>

    <view class="card" v-if="tables.length">
      <view class="card-title">数据表行数</view>
      <view v-for="t in tables" :key="t.tableName" class="table-row">
        <text class="table-name">{{ t.tableName }}</text>
        <text class="table-count" :class="{ empty: t.rowCount < 0 }">{{ t.rowCount < 0 ? '统计失败' : t.rowCount }}</text>
      </view>
    </view>
    <view v-else-if="!loading" class="empty">暂无数据表信息</view>

    <view class="refresh" @tap="load">刷新数据</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { dbInfo } from '@/api/monitor'

const info = ref({})
const loading = ref(false)

const tables = computed(() => info.value.tables || [])

function fmtBytes(n) {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let v = Number(n)
  let i = 0
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i += 1 }
  return `${v.toFixed(1)} ${units[i]}`
}

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const data = await dbInfo()
    info.value = data.data || {}
  } catch (e) {
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '数据监控' })
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
.stats { display: flex; gap: 20rpx; }
.stat { flex: 1; background: #fff; border-radius: 18rpx; padding: 28rpx 16rpx; text-align: center; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.stat-num { display: block; font-size: 32rpx; font-weight: 600; color: #409eff; }
.stat-label { display: block; margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); margin-top: 24rpx; }
.card-title { font-size: 30rpx; font-weight: 600; color: #1f2d3d; margin-bottom: 16rpx; }
.row { display: flex; align-items: center; justify-content: space-between; gap: 24rpx; padding: 14rpx 0; border-bottom: 1rpx solid #f5f7fa; }
.row:last-child { border-bottom: none; }
.k { font-size: 26rpx; color: #8a94a6; flex-shrink: 0; }
.v { font-size: 26rpx; color: #1f2d3d; text-align: right; word-break: break-all; }
.table-row { display: flex; align-items: center; justify-content: space-between; gap: 20rpx; padding: 16rpx 0; border-bottom: 1rpx solid #f5f7fa; }
.table-row:last-child { border-bottom: none; }
.table-name { font-size: 24rpx; color: #5a6678; word-break: break-all; }
.table-count { font-size: 24rpx; color: #409eff; flex-shrink: 0; }
.table-count.empty { color: #9aa6b8; }
.empty { margin-top: 24rpx; padding: 80rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.refresh { margin-top: 32rpx; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
</style>
