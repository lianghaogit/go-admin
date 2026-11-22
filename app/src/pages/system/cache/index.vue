<template>
  <view class="page">
    <view class="stats">
      <view class="stat">
        <text class="stat-num">{{ info.keyCount || 0 }}</text>
        <text class="stat-label">缓存键</text>
      </view>
      <view class="stat">
        <text class="stat-num">{{ fmtRate(info.hitRate) }}</text>
        <text class="stat-label">命中率 %</text>
      </view>
      <view class="stat">
        <text class="stat-num">{{ fmtBytes(info.memBytes) }}</text>
        <text class="stat-label">占用内存</text>
      </view>
    </view>

    <view class="card">
      <view class="row"><text class="k">缓存类型</text><text class="v">{{ info.type || '—' }}</text></view>
      <view class="row"><text class="k">命中 / 未命中</text><text class="v">{{ info.hitCount || 0 }} / {{ info.missCount || 0 }}</text></view>
      <view class="row"><text class="k">启动时间</text><text class="v">{{ info.startTime || '—' }}</text></view>
    </view>

    <view class="card" v-if="groupList.length">
      <view class="card-title">
        <text>键分组</text>
        <input v-model="keyFilter" class="key-filter" placeholder="搜索键名…" placeholder-class="ph" confirm-type="search" />
      </view>
      <view v-for="g in filteredGroups" :key="g.name" class="group">
        <view class="group-head">
          <text class="group-name">{{ g.name }}</text>
          <text class="group-count">{{ g.keys.length }}</text>
        </view>
        <view v-for="k in g.keys" :key="k" class="key-item" @tap="showKey(k)">
          <text class="key">{{ k }}</text>
          <view class="key-del" @tap.stop="onRemoveKey(k)">×</view>
        </view>
      </view>
      <view v-if="!filteredGroups.length" class="empty-hint">没有匹配的缓存键</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无缓存数据</view>

    <view class="actions">
      <view class="btn" @tap="load">刷新</view>
      <view v-if="$hasPerm('monitor:cache:clear')" class="btn danger" @tap="onClear">清空缓存</view>
    </view>
    <view class="tip">提示：清空缓存会清除登录会话与验证码，在线用户需重新登录。</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { cacheInfo, cacheKeyDetail, cacheClear, cacheRemoveKey } from '@/api/monitor'

const info = ref({})
const loading = ref(false)
const keyFilter = ref('')

const groupList = computed(() => {
  const groups = info.value.groups || {}
  return Object.keys(groups).sort().map((name) => ({ name, keys: groups[name] || [] }))
})

const filteredGroups = computed(() => {
  const kw = keyFilter.value.trim()
  if (!kw) return groupList.value
  return groupList.value
    .map((g) => ({ ...g, keys: g.keys.filter((k) => k.toLowerCase().includes(kw.toLowerCase())) }))
    .filter((g) => g.keys.length > 0)
})

function fmtRate(v) {
  return Number(v || 0).toFixed(1)
}

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
    const data = await cacheInfo()
    info.value = data.data || {}
  } catch (e) {
  } finally {
    loading.value = false
  }
}

async function showKey(key) {
  try {
    const data = await cacheKeyDetail(key)
    const d = data.data || {}
    uni.showModal({
      title: key,
      content: `值：${d.value ?? '—'}\n创建时间：${d.createAt || '—'}\n过期时间：${d.expireAt || '—'}\n有效期：${d.ttl ?? 0} 秒`,
      showCancel: false,
      confirmText: '关闭'
    })
  } catch (e) {}
}

function onRemoveKey(key) {
  uni.showModal({
    title: '提示',
    content: `确定删除缓存键「${key}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await cacheRemoveKey(key)
        uni.showToast({ title: '已删除', icon: 'success' })
        load()
      } catch (e) {}
    }
  })
}

function onClear() {
  uni.showModal({
    title: '提示',
    content: '确定清空全部缓存吗？当前登录可能失效',
    confirmText: '清空',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await cacheClear()
        uni.showToast({ title: '已清空', icon: 'success' })
        load()
      } catch (e) {}
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '缓存监控' })
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
.stat-num { display: block; font-size: 34rpx; font-weight: 600; color: #409eff; }
.stat-label { display: block; margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); margin-top: 24rpx; }
.card-title { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; font-size: 30rpx; font-weight: 600; color: #1f2d3d; margin-bottom: 16rpx; }
.key-filter { flex: 1; height: 56rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 14rpx; font-size: 24rpx; min-width: 0; }
.ph { color: #9aa6b8; }
.row { display: flex; align-items: center; justify-content: space-between; gap: 24rpx; padding: 14rpx 0; border-bottom: 1rpx solid #f5f7fa; }
.row:last-child { border-bottom: none; }
.k { font-size: 26rpx; color: #8a94a6; }
.v { font-size: 26rpx; color: #1f2d3d; text-align: right; word-break: break-all; }
.group + .group { margin-top: 20rpx; }
.group-head { display: flex; align-items: center; gap: 12rpx; padding: 10rpx 0; }
.group-name { font-size: 26rpx; font-weight: 600; color: #1f2d3d; }
.group-count { font-size: 20rpx; color: #409eff; background: rgba(64, 158, 255, 0.1); padding: 2rpx 14rpx; border-radius: 18rpx; }
.key-item { display: flex; align-items: center; justify-content: space-between; gap: 20rpx; padding: 16rpx 20rpx; background: #f7f9fc; border-radius: 14rpx; margin-top: 10rpx; }
.key { font-size: 24rpx; color: #5a6678; word-break: break-all; flex: 1; min-width: 0; }
.key-del { width: 44rpx; height: 44rpx; line-height: 44rpx; text-align: center; color: #f56c6c; font-size: 28rpx; border-radius: 50%; background: rgba(245, 108, 108, 0.08); flex-shrink: 0; }
.empty-hint { padding: 40rpx 0; text-align: center; color: #9aa6b8; font-size: 24rpx; }
.empty { margin-top: 24rpx; padding: 80rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.actions { display: flex; gap: 20rpx; margin-top: 32rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: linear-gradient(135deg, #4a7df0, #3457c9); color: #fff; font-size: 28rpx; font-weight: 600; border-radius: 18rpx; box-shadow: 0 12rpx 26rpx rgba(52, 87, 201, 0.24); }
.btn.danger { background: linear-gradient(135deg, #ff8a8a, #f56c6c); box-shadow: 0 12rpx 26rpx rgba(245, 108, 108, 0.24); }
.tip { margin-top: 18rpx; font-size: 22rpx; color: #9aa6b8; text-align: center; }
</style>
