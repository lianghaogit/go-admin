<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入用户名搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('monitor:online:forceLogout')" class="search-btn danger" @tap="onForceAll">全部强退</view>
    </view>

    <view v-if="list.length" class="online-list">
      <view v-for="o in list" :key="o.tokenHash" class="online-item">
        <view class="avatar">{{ (o.userName || 'U').charAt(0).toUpperCase() }}</view>
        <view class="online-main">
          <view class="online-line">
            <text class="online-name">{{ o.userName || '—' }}</text>
            <text v-if="$hasPerm('monitor:online:forceLogout')" class="op danger" @tap="onForce(o)">强退</text>
          </view>
          <view class="online-line sub">
            <text class="online-meta">{{ o.ip || '—' }}</text>
            <text class="online-meta">{{ o.browser || '—' }}</text>
            <text class="online-meta">{{ o.os || '—' }}</text>
          </view>
          <view class="online-line sub">
            <text class="online-time">{{ fmtTime(o.loginTime) }}</text>
          </view>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无在线用户</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listOnline, forceLogout, forceLogoutAll } from '@/api/online'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

function fmtTime(t) {
  if (!t) return '—'
  return String(t).replace('T', ' ').slice(0, 19)
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listOnline({ pageNum: pageNum.value, pageSize, userName: keyword.value || undefined })
    if (data.code === 200) {
      total.value = data.total
      const rows = data.rows || []
      list.value = reset ? rows : list.value.concat(rows)
      finished.value = list.value.length >= data.total
      if (rows.length) pageNum.value += 1
    } else {
      uni.showToast({ title: data.msg || '查询失败', icon: 'none' })
    }
  } catch (e) {
  } finally {
    loading.value = false
  }
}

function onSearch() { fetchList(true) }
function onReset() { keyword.value = ''; fetchList(true) }

function onForce(o) {
  uni.showModal({
    title: '提示',
    content: `确定强退用户「${o.userName}」吗？该会话将立即失效`,
    confirmText: '强退',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await forceLogout(o.tokenHash)
        list.value = list.value.filter((it) => it.tokenHash !== o.tokenHash)
        total.value = Math.max(0, total.value - 1)
        uni.showToast({ title: '已强退', icon: 'none' })
        if (!list.value.length) fetchList(true)
      } catch (e) {}
    }
  })
}

function onForceAll() {
  uni.showModal({
    title: '提示',
    content: '确定强退全部在线用户吗？当前登录也会失效',
    confirmText: '全部强退',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await forceLogoutAll()
        list.value = []
        total.value = 0
        finished.value = true
        uni.showToast({ title: '已全部强退，请重新登录', icon: 'none' })
        // 全部会话已清除，当前登录同样失效，直接回登录页
        userStore.reset()
        setTimeout(() => {
          uni.reLaunch({ url: '/pages/login/index' })
        }, 800)
      } catch (e) {}
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '在线用户' })
  fetchList(true)
})

onPullDownRefresh(async () => {
  try {
    await fetchList(true)
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
    uni.showToast({ title: '刷新失败', icon: 'none' })
  } finally {
    uni.stopPullDownRefresh()
  }
})

onReachBottom(() => {
  fetchList(false)
})
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 28rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.search-btn.danger { background: #fef0f0; color: #f56c6c; }
.online-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.online-item { display: flex; align-items: center; gap: 20rpx; padding: 26rpx 0; }
.online-item + .online-item { border-top: 1rpx solid #f0f3f7; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; background: #409eff; flex-shrink: 0; }
.online-main { flex: 1; min-width: 0; }
.online-line { display: flex; align-items: center; gap: 12rpx; }
.online-name { font-size: 28rpx; color: #1f2d3d; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.online-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.online-meta + .online-meta::before { content: '·'; margin: 0 8rpx; }
.online-time { color: #8a94a6; }
.op { font-size: 24rpx; flex-shrink: 0; padding: 4rpx 18rpx; border-radius: 18rpx; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
