<template>
  <view class="page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="登录账号" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <input v-model="ipaddr" class="search-input" placeholder="IP" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="toggleFilter">{{ filterOpen ? '收起' : '筛选' }}</view>
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('monitor:logininfor:remove')" class="search-btn danger" @tap="onClean">清空</view>
    </view>

    <!-- 展开的更多筛选 -->
    <view v-if="filterOpen" class="filter-panel">
      <view class="filter-row">
        <text class="filter-label">登录状态</text>
        <view class="chip-wrap">
          <view v-for="o in statusOptions" :key="o.value" class="chip" :class="{ on: status === o.value }" @tap="status = o.value">{{ o.label }}</view>
        </view>
      </view>
      <view class="filter-row">
        <text class="filter-label">登录时间</text>
        <picker mode="date" :value="beginTime" @change="beginTime = $event.detail.value">
          <view class="date-cell">{{ beginTime || '开始日期' }}</view>
        </picker>
        <text class="filter-sep">—</text>
        <picker mode="date" :value="endTime" @change="endTime = $event.detail.value">
          <view class="date-cell">{{ endTime || '结束日期' }}</view>
        </picker>
      </view>
    </view>

    <!-- 工具栏 -->
    <view v-if="$hasPerm('monitor:logininfor:list') || $hasPerm('monitor:logininfor:remove')" class="tool-bar">
      <view v-if="$hasPerm('monitor:logininfor:list')" class="tool-btn" @tap="onExport">导出 Excel</view>
      <view v-if="$hasPerm('monitor:logininfor:remove') && selectedIds.length" class="tool-btn danger" @tap="onBatchDelete">批量删除（{{ selectedIds.length }}）</view>
      <view v-if="selectedIds.length" class="tool-btn plain" @tap="toggleSelectAll">
        {{ allSelected ? '取消全选' : '全选' }}
      </view>
    </view>

    <!-- 列表 -->
    <view v-if="list.length" class="log-list">
      <view v-for="l in list" :key="l.infoId" class="log-item" @tap="toggleSelect(l)">
        <view v-if="$hasPerm('monitor:logininfor:remove')" class="check" :class="{ on: selectedIds.includes(l.infoId) }">
          <text v-if="selectedIds.includes(l.infoId)">✓</text>
        </view>
        <view class="avatar" :style="{ background: l.status === '0' ? '#67c23a' : '#f56c6c' }">{{ (l.userName || 'U').charAt(0).toUpperCase() }}</view>
        <view class="log-main">
          <view class="log-line">
            <text class="log-name">{{ l.userName }}</text>
            <text class="status" :class="l.status === '0' ? 'on' : 'off'">{{ l.status === '0' ? '成功' : '失败' }}</text>
          </view>
          <view class="log-line sub">
            <text class="log-ip">{{ l.ipaddr || '—' }}</text>
            <text class="log-time">{{ fmtTime(l.loginTime) }}</text>
          </view>
          <view v-if="l.msg" class="log-line sub">
            <text class="log-msg">{{ l.msg }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('monitor:logininfor:remove')" class="op danger" @tap.stop="onDelete(l)">✕</view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无登录日志</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listLoginLog, deleteLoginLog, deleteLoginLogs, cleanLoginLog } from '@/api/log'
import { exportExcel, fileStamp } from '@/utils/download'

const keyword = ref('')
const ipaddr = ref('')
const status = ref('')
const beginTime = ref('')
const endTime = ref('')
const filterOpen = ref(false)
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const selectedIds = ref([])

const statusOptions = [
  { value: '', label: '全部' },
  { value: '0', label: '成功' },
  { value: '1', label: '失败' }
]

const allSelected = computed(() => list.value.length > 0 && list.value.every((l) => selectedIds.value.includes(l.infoId)))

function fmtTime(t) {
  if (!t) return '—'
  return String(t).replace('T', ' ').slice(0, 19)
}

function buildParams() {
  const p = { pageNum: pageNum.value, pageSize }
  if (keyword.value) p.username = keyword.value
  if (ipaddr.value) p.ipaddr = ipaddr.value
  if (status.value) p.status = status.value
  if (beginTime.value || endTime.value) {
    p.params = {}
    if (beginTime.value) p.params.beginTime = beginTime.value + ' 00:00:00'
    if (endTime.value) p.params.endTime = endTime.value + ' 23:59:59'
  }
  return p
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false; selectedIds.value = [] }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listLoginLog(buildParams())
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

function toggleFilter() { filterOpen.value = !filterOpen.value }
function onSearch() { fetchList(true) }
function onReset() {
  keyword.value = ''
  ipaddr.value = ''
  status.value = ''
  beginTime.value = ''
  endTime.value = ''
  fetchList(true)
}

function onExport() {
  const p = {}
  if (keyword.value) p.username = keyword.value
  if (ipaddr.value) p.ipaddr = ipaddr.value
  if (status.value) p.status = status.value
  if (beginTime.value || endTime.value) {
    p.params = {}
    if (beginTime.value) p.params.beginTime = beginTime.value + ' 00:00:00'
    if (endTime.value) p.params.endTime = endTime.value + ' 23:59:59'
  }
  exportExcel('/api/monitor/logininfor/export', p, `登录日志_${fileStamp()}.xlsx`)
}

function toggleSelect(l) {
  const idx = selectedIds.value.indexOf(l.infoId)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(l.infoId)
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = selectedIds.value.filter((id) => !list.value.some((l) => l.infoId === id))
  } else {
    const current = list.value.map((l) => l.infoId)
    current.forEach((id) => {
      if (!selectedIds.value.includes(id)) selectedIds.value.push(id)
    })
  }
}

function onBatchDelete() {
  if (!selectedIds.value.length) return
  uni.showModal({
    title: '提示',
    content: `确定删除选中的 ${selectedIds.value.length} 条登录日志吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteLoginLogs(selectedIds.value.join(','))
        if (data.code === 200) {
          uni.showToast({ title: '批量删除成功', icon: 'success' })
          fetchList(true)
        }
      } catch (e) {}
    }
  })
}

function onDelete(l) {
  uni.showModal({
    title: '提示',
    content: '确定删除该条登录日志吗？',
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteLoginLog(l.infoId)
        if (data.code === 200) {
          uni.showToast({ title: '删除成功', icon: 'success' })
          list.value = list.value.filter((x) => x.infoId !== l.infoId)
          total.value = Math.max(0, total.value - 1)
          selectedIds.value = selectedIds.value.filter((id) => id !== l.infoId)
          if (!list.value.length) fetchList(true)
        }
      } catch (e) {}
    }
  })
}

function onClean() {
  uni.showModal({
    title: '提示',
    content: '确定清空全部登录日志吗？此操作不可恢复',
    confirmText: '清空',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await cleanLoginLog()
        if (data.code === 200) {
          uni.showToast({ title: '已清空', icon: 'success' })
          fetchList(true)
        }
      } catch (e) {}
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '登录日志' })
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

onReachBottom(() => { fetchList(false) })
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 12rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); flex-wrap: wrap; }
.search-input { flex: 1; min-width: 160rpx; height: 64rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 14rpx; font-size: 24rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 24rpx; height: 64rpx; line-height: 64rpx; background: #409eff; color: #fff; font-size: 24rpx; border-radius: 14rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.search-btn.danger { background: #fef0f0; color: #f56c6c; }

.filter-panel { margin-top: 16rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.filter-row { display: flex; align-items: center; gap: 16rpx; padding: 10rpx 0; }
.filter-label { font-size: 24rpx; color: #6b7a90; width: 120rpx; flex-shrink: 0; }
.filter-sep { color: #c3ccd8; }
.chip-wrap { display: flex; flex-wrap: wrap; gap: 12rpx; flex: 1; }
.chip { font-size: 22rpx; padding: 8rpx 22rpx; border-radius: 20rpx; color: #5a6678; background: #f5f7fa; }
.chip.on { color: #fff; background: linear-gradient(135deg, #4a7df0, #3457c9); }
.date-cell { padding: 0 24rpx; height: 64rpx; line-height: 64rpx; background: #f5f7fa; border-radius: 14rpx; font-size: 24rpx; color: #1c2738; }

.tool-bar { display: flex; justify-content: flex-end; gap: 16rpx; margin-top: 16rpx; flex-wrap: wrap; }
.tool-btn { padding: 0 22rpx; height: 64rpx; line-height: 64rpx; background: #e6a23c; color: #fff; font-size: 24rpx; border-radius: 14rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.tool-btn.plain { background: #fff; color: #5a6678; border: 1rpx solid #e4e9f1; box-shadow: none; }
.tool-btn.danger { background: #f56c6c; box-shadow: 0 12rpx 30rpx rgba(245, 108, 108, 0.24); }

.log-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.log-item { display: flex; align-items: center; gap: 20rpx; padding: 26rpx 0; }
.log-item + .log-item { border-top: 1rpx solid #f0f3f7; }
.check { width: 40rpx; height: 40rpx; border-radius: 10rpx; border: 2rpx solid #d0d7e2; display: flex; align-items: center; justify-content: center; font-size: 24rpx; color: #fff; flex-shrink: 0; }
.check.on { background: linear-gradient(135deg, #4a7df0, #3457c9); border-color: #3457c9; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.log-main { flex: 1; min-width: 0; }
.log-line { display: flex; align-items: center; gap: 12rpx; }
.log-name { font-size: 28rpx; color: #1f2d3d; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.log-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.log-time::before { content: '·'; margin: 0 8rpx; }
.log-msg { color: #5a6678; }
.op { width: 56rpx; height: 56rpx; line-height: 56rpx; text-align: center; font-size: 26rpx; border-radius: 50%; flex-shrink: 0; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
