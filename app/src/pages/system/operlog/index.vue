<template>
  <view class="page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="操作人" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <input v-model="operIp" class="search-input" placeholder="IP" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="toggleFilter">{{ filterOpen ? '收起' : '筛选' }}</view>
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('monitor:operlog:remove')" class="search-btn danger" @tap="onClean">清空</view>
    </view>

    <!-- 展开的更多筛选 -->
    <view v-if="filterOpen" class="filter-panel">
      <view class="filter-row">
        <text class="filter-label">模块标题</text>
        <input v-model="title" class="filter-input" placeholder="如 用户管理" placeholder-class="ph" />
      </view>
      <view class="filter-row">
        <text class="filter-label">操作类型</text>
        <view class="chip-wrap">
          <view v-for="o in businessTypes" :key="o.value" class="chip" :class="{ on: businessType === o.value }" @tap="businessType = o.value">{{ o.label }}</view>
        </view>
      </view>
      <view class="filter-row">
        <text class="filter-label">操作状态</text>
        <view class="chip-wrap">
          <view v-for="o in statusOptions" :key="o.value" class="chip" :class="{ on: status === o.value }" @tap="status = o.value">{{ o.label }}</view>
        </view>
      </view>
      <view class="filter-row">
        <text class="filter-label">操作时间</text>
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
    <view class="tool-bar">
      <view class="tool-btn" @tap="onExport">导出 Excel</view>
      <view v-if="$hasPerm('monitor:operlog:remove') && selectedIds.length" class="tool-btn danger" @tap="onBatchDelete">批量删除（{{ selectedIds.length }}）</view>
      <view v-if="selectedIds.length" class="tool-btn plain" @tap="toggleSelectAll">{{ allSelected ? '取消全选' : '全选' }}</view>
    </view>

    <!-- 列表 -->
    <view v-if="list.length" class="log-list">
      <view v-for="l in list" :key="l.operId" class="log-item" @tap="toggleSelect(l)">
        <view v-if="$hasPerm('monitor:operlog:remove')" class="check" :class="{ on: selectedIds.includes(l.operId) }">
          <text v-if="selectedIds.includes(l.operId)">✓</text>
        </view>
        <view class="avatar" :style="{ background: l.status === '0' ? '#409eff' : '#f56c6c' }">{{ (l.title || 'O').charAt(0).toUpperCase() }}</view>
        <view class="log-main">
          <view class="log-line">
            <text class="log-title">{{ l.title || '—' }}</text>
            <text class="status" :class="l.status === '0' ? 'on' : 'off'">{{ l.status === '0' ? '成功' : '失败' }}</text>
            <text class="biz-tag" :class="bizClass(l.businessType)">{{ bizLabel(l.businessType) }}</text>
            <text v-if="$hasPerm('monitor:operlog:remove')" class="op danger" @tap.stop="onDelete(l)">✕</text>
          </view>
          <view class="log-line sub">
            <text class="log-meta">{{ l.operName || '—' }}</text>
            <text class="log-meta">{{ l.requestMethod || '—' }}</text>
            <text class="log-meta">{{ l.operIp || '—' }}</text>
            <text class="log-meta">{{ l.costTime }}ms</text>
          </view>
          <view class="log-line sub">
            <text class="log-time">{{ fmtTime(l.operTime) }}</text>
          </view>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无操作日志</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listOperLog, deleteOperLog, deleteOperLogs, cleanOperLog } from '@/api/log'
import { exportExcel, fileStamp } from '@/utils/download'

const keyword = ref('')
const title = ref('')
const operIp = ref('')
const businessType = ref('')
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

const businessTypes = [
  { value: '', label: '全部' },
  { value: '1', label: '新增' },
  { value: '2', label: '修改' },
  { value: '3', label: '删除' },
  { value: '4', label: '授权' },
  { value: '5', label: '导出' },
  { value: '6', label: '导入' },
  { value: '7', label: '强退' },
  { value: '8', label: '清空数据' }
]
const statusOptions = [
  { value: '', label: '全部' },
  { value: '0', label: '成功' },
  { value: '1', label: '失败' }
]

const allSelected = computed(() => list.value.length > 0 && list.value.every((l) => selectedIds.value.includes(l.operId)))

function bizLabel(v) {
  const o = businessTypes.find((b) => b.value === String(v))
  return o ? o.label : '其他'
}
function bizClass(v) {
  const map = { '1': 'biz-add', '2': 'biz-edit', '3': 'biz-del', '4': 'biz-auth', '5': 'biz-exp', '6': 'biz-imp', '7': 'biz-kick', '8': 'biz-clean' }
  return map[String(v)] || ''
}
function fmtTime(t) {
  if (!t) return '—'
  return String(t).replace('T', ' ').slice(0, 19)
}

function buildParams() {
  const p = { pageNum: pageNum.value, pageSize }
  if (keyword.value) p.operName = keyword.value
  if (title.value) p.title = title.value
  if (operIp.value) p.operIp = operIp.value
  if (businessType.value) p.businessType = businessType.value
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
    const data = await listOperLog(buildParams())
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
  title.value = ''
  operIp.value = ''
  businessType.value = ''
  status.value = ''
  beginTime.value = ''
  endTime.value = ''
  fetchList(true)
}

function onExport() {
  const p = {}
  if (keyword.value) p.operName = keyword.value
  if (title.value) p.title = title.value
  if (operIp.value) p.operIp = operIp.value
  if (businessType.value) p.businessType = businessType.value
  if (status.value) p.status = status.value
  if (beginTime.value || endTime.value) {
    p.params = {}
    if (beginTime.value) p.params.beginTime = beginTime.value + ' 00:00:00'
    if (endTime.value) p.params.endTime = endTime.value + ' 23:59:59'
  }
  exportExcel('/api/monitor/operlog/export', p, `操作日志_${fileStamp()}.xlsx`)
}

function toggleSelect(l) {
  const idx = selectedIds.value.indexOf(l.operId)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(l.operId)
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = selectedIds.value.filter((id) => !list.value.some((l) => l.operId === id))
  } else {
    const current = list.value.map((l) => l.operId)
    current.forEach((id) => { if (!selectedIds.value.includes(id)) selectedIds.value.push(id) })
  }
}

function onBatchDelete() {
  if (!selectedIds.value.length) return
  uni.showModal({
    title: '提示',
    content: `确定删除选中的 ${selectedIds.value.length} 条操作日志吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteOperLogs(selectedIds.value.join(','))
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
    content: '确定删除该条操作日志吗？',
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteOperLog(l.operId)
        list.value = list.value.filter((it) => it.operId !== l.operId)
        selectedIds.value = selectedIds.value.filter((id) => id !== l.operId)
        total.value = Math.max(0, total.value - 1)
        if (!list.value.length) fetchList(true)
      } catch (e) {}
    }
  })
}

function onClean() {
  uni.showModal({
    title: '提示',
    content: '确定清空全部操作日志吗？该操作不可恢复',
    confirmText: '清空',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await cleanOperLog()
        list.value = []
        total.value = 0
        finished.value = true
        selectedIds.value = []
        uni.showToast({ title: '已清空', icon: 'none' })
      } catch (e) {}
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '操作日志' })
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
.filter-row { display: flex; align-items: center; gap: 16rpx; padding: 10rpx 0; flex-wrap: wrap; }
.filter-label { font-size: 24rpx; color: #6b7a90; width: 140rpx; flex-shrink: 0; }
.filter-input { flex: 1; min-width: 180rpx; height: 60rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 14rpx; font-size: 24rpx; }
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
.avatar { width: 72rpx; height: 72rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.log-main { flex: 1; min-width: 0; }
.log-line { display: flex; align-items: center; gap: 10rpx; flex-wrap: wrap; }
.log-title { font-size: 28rpx; color: #1f2d3d; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.biz-tag { font-size: 18rpx; padding: 2rpx 12rpx; border-radius: 16rpx; background: #f5f7fa; color: #6b7a90; flex-shrink: 0; }
.biz-add { background: rgba(63, 174, 122, 0.12); color: #3fae7a; }
.biz-edit { background: rgba(58, 111, 224, 0.12); color: #3a6fe0; }
.biz-del { background: rgba(245, 108, 108, 0.12); color: #f56c6c; }
.biz-auth { background: rgba(123, 97, 255, 0.12); color: #7b61ff; }
.biz-exp, .biz-imp { background: rgba(230, 162, 60, 0.14); color: #d98b21; }
.biz-kick { background: rgba(245, 108, 108, 0.12); color: #f56c6c; }
.biz-clean { background: rgba(154, 166, 184, 0.18); color: #6b7a90; }
.log-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.log-meta + .log-meta::before { content: '·'; margin: 0 8rpx; }
.log-time { color: #8a94a6; }
.op { width: 44rpx; height: 44rpx; line-height: 44rpx; text-align: center; font-size: 24rpx; border-radius: 50%; flex-shrink: 0; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
