<template>
  <view class="page">
    <view class="card">
      <view class="card-title">服务器</view>
      <view class="row"><text class="k">主机名</text><text class="v">{{ info.hostName || '—' }}</text></view>
      <view class="row"><text class="k">操作系统</text><text class="v">{{ info.os || '—' }}</text></view>
      <view class="row"><text class="k">内核架构</text><text class="v">{{ info.arch || '—' }}</text></view>
      <view class="row"><text class="k">内核版本</text><text class="v">{{ info.kernel || '—' }}</text></view>
      <view class="row"><text class="k">Go 版本</text><text class="v">{{ info.goVersion || '—' }}</text></view>
      <view class="row"><text class="k">CPU 核数</text><text class="v">{{ info.cpuNum || '—' }}</text></view>
    </view>

    <view class="card">
      <view class="card-title">运行状态</view>
      <view class="metric">
        <view class="metric-head"><text class="k">CPU 使用率</text><text class="metric-val">{{ info.cpuUsage || 0 }}%</text></view>
        <view class="bar"><view class="bar-inner" :style="{ width: pct(info.cpuUsage) + '%', background: color(info.cpuUsage) }"></view></view>
      </view>
      <view class="metric">
        <view class="metric-head"><text class="k">内存使用率</text><text class="metric-val">{{ info.memUsage || 0 }}%</text></view>
        <view class="bar"><view class="bar-inner" :style="{ width: pct(info.memUsage) + '%', background: color(info.memUsage) }"></view></view>
        <view class="metric-sub">{{ fmtBytes(info.memUsed) }} / {{ fmtBytes(info.memTotal) }}</view>
      </view>
      <view class="metric">
        <view class="metric-head"><text class="k">磁盘使用率</text><text class="metric-val">{{ info.diskUsage || 0 }}%</text></view>
        <view class="bar"><view class="bar-inner" :style="{ width: pct(info.diskUsage) + '%', background: color(info.diskUsage) }"></view></view>
        <view class="metric-sub">{{ fmtBytes(info.diskUsed) }} / {{ fmtBytes(info.diskTotal) }}</view>
      </view>
      <view class="row"><text class="k">已运行</text><text class="v">{{ info.runDays || 0 }} 天 {{ info.runHours || 0 }} 小时 {{ info.runMinutes || 0 }} 分</text></view>
      <view class="row"><text class="k">启动时间</text><text class="v">{{ info.startTime || '—' }}</text></view>
      <view class="row"><text class="k">进程内存</text><text class="v">{{ fmtBytes(info.procMemUsed) }} / {{ fmtBytes(info.procMemTotal) }}</text></view>
    </view>

    <view class="refresh" @tap="load">刷新数据</view>
    <view v-if="!info.cpuNum && loading" class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { serverInfo } from '@/api/monitor'

const info = ref({})
const loading = ref(false)

function fmtBytes(n) {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let v = Number(n)
  let i = 0
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i += 1 }
  return `${v.toFixed(1)} ${units[i]}`
}

function pct(v) {
  return Math.max(0, Math.min(100, Number(v) || 0))
}

function color(v) {
  const n = Number(v) || 0
  if (n >= 85) return '#f56c6c'
  if (n >= 60) return '#e6a23c'
  return '#409eff'
}

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const data = await serverInfo()
    info.value = data.data || {}
  } catch (e) {
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '服务监控' })
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
.card { background: #fff; border-radius: 18rpx; padding: 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.card + .card { margin-top: 24rpx; }
.card-title { font-size: 30rpx; font-weight: 600; color: #1f2d3d; margin-bottom: 16rpx; }
.row { display: flex; align-items: center; justify-content: space-between; gap: 24rpx; padding: 14rpx 0; border-bottom: 1rpx solid #f5f7fa; }
.row:last-child { border-bottom: none; }
.k { font-size: 26rpx; color: #8a94a6; flex-shrink: 0; }
.v { font-size: 26rpx; color: #1f2d3d; text-align: right; word-break: break-all; }
.metric { padding: 18rpx 0 12rpx; }
.metric + .metric { border-top: 1rpx solid #f5f7fa; }
.metric-head { display: flex; align-items: center; justify-content: space-between; }
.metric-val { font-size: 28rpx; font-weight: 600; color: #1f2d3d; }
.bar { height: 12rpx; background: #f0f3f7; border-radius: 6rpx; margin-top: 14rpx; overflow: hidden; }
.bar-inner { height: 100%; border-radius: 6rpx; transition: width 0.3s; }
.metric-sub { margin-top: 10rpx; font-size: 22rpx; color: #8a94a6; }
.refresh { margin-top: 32rpx; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.empty { margin-top: 24rpx; text-align: center; color: #9aa6b8; font-size: 26rpx; }
</style>
