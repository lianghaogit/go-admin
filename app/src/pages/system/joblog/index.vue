<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入任务名称搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>
    <view class="filter-bar">
      <text v-for="f in FILTERS" :key="f.value" class="filter" :class="{ active: status === f.value }" @tap="onFilter(f.value)">{{ f.label }}</text>
      <text v-if="$hasPerm('monitor:job:remove')" class="filter danger" @tap="onClean">清空</text>
    </view>

    <view v-if="list.length" class="log-list">
      <view v-for="l in list" :key="l.jobLogId" class="log-item" @tap="showDetail(l)">
        <view class="log-line">
          <text class="log-name">{{ l.jobName || '—' }}</text>
          <text class="status" :class="l.status === '0' ? 'on' : 'off'">{{ l.status === '0' ? '成功' : '失败' }}</text>
          <text v-if="$hasPerm('monitor:job:remove')" class="del" @tap.stop="onDelete(l)">删除</text>
        </view>
        <view class="log-line sub">
          <text>{{ l.jobGroup || 'DEFAULT' }}</text>
          <text class="time">{{ fmtTime(l.createTime) }}</text>
        </view>
        <view class="log-msg">{{ l.jobMessage || (l.status === '0' ? '执行完成' : '执行失败') }}</view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无执行日志</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listJobLog, deleteJobLog, cleanJobLog } from '@/api/job'

const FILTERS = [
  { label: '全部', value: '' },
  { label: '成功', value: '0' },
  { label: '失败', value: '1' }
]

const keyword = ref('')
const status = ref('')
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
    const data = await listJobLog({ pageNum: pageNum.value, pageSize, jobName: keyword.value || undefined, status: status.value || undefined })
    total.value = data.total
    const rows = data.rows || []
    list.value = reset ? rows : list.value.concat(rows)
    finished.value = list.value.length >= data.total
    if (rows.length) pageNum.value += 1
  } catch (e) {
  } finally {
    loading.value = false
  }
}

function onSearch() { fetchList(true) }
function onReset() { keyword.value = ''; status.value = ''; fetchList(true) }
function onFilter(v) { status.value = v; fetchList(true) }

function showDetail(l) {
  const msg = [l.jobMessage, l.exceptionInfo].filter(Boolean).join('\n\n')
  uni.showModal({ title: `${l.jobName || '任务'} · ${fmtTime(l.createTime)}`, content: msg || '无执行信息', showCancel: false, confirmText: '关闭' })
}

function onDelete(l) {
  uni.showModal({
    title: '提示',
    content: '确定删除这条执行日志吗？',
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteJobLog(l.jobLogId)
        list.value = list.value.filter((it) => it.jobLogId !== l.jobLogId)
        total.value = Math.max(0, total.value - 1)
        uni.showToast({ title: '删除成功', icon: 'success' })
        if (!list.value.length) fetchList(true)
      } catch (e) {}
    }
  })
}

function onClean() {
  uni.showModal({
    title: '提示',
    content: '确定清空全部执行日志吗？此操作不可恢复',
    confirmText: '清空',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await cleanJobLog()
        list.value = []
        total.value = 0
        finished.value = true
        uni.showToast({ title: '已清空', icon: 'success' })
      } catch (e) {}
    }
  })
}

onLoad((options) => {
  uni.setNavigationBarTitle({ title: '任务执行日志' })
  if (options && options.jobName) keyword.value = options.jobName
  fetchList(true)
})

onPullDownRefresh(async () => {
  try {
    await fetchList(true)
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
  } finally {
    uni.stopPullDownRefresh()
  }
})

onReachBottom(() => { fetchList(false) })
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 12rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; min-width: 0; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.filter-bar { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; }
.filter { padding: 10rpx 24rpx; font-size: 24rpx; color: #5a6678; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.filter.active { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.filter.danger { margin-left: auto; color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.log-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.log-item { padding: 26rpx 0; }
.log-item + .log-item { border-top: 1rpx solid #f0f3f7; }
.log-line { display: flex; align-items: center; gap: 12rpx; }
.log-name { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.del { font-size: 24rpx; color: #f56c6c; padding: 4rpx 16rpx; border-radius: 18rpx; background: rgba(245, 108, 108, 0.08); flex-shrink: 0; }
.log-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.time { margin-left: auto; }
.log-msg { margin-top: 10rpx; font-size: 24rpx; color: #5a6678; word-break: break-all; }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
