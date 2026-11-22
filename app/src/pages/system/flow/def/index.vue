<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入流程名称 / 标识搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>
    <view class="filter-bar">
      <text v-for="f in FILTERS" :key="f.value" class="filter" :class="{ active: status === f.value }" @tap="onFilter(f.value)">{{ f.label }}</text>
    </view>

    <view v-if="list.length" class="def-list">
      <view v-for="d in list" :key="d.flowDefId" class="def-item">
        <view class="def-line" @tap="showDetail(d)">
          <text class="def-name">{{ d.defName }}</text>
          <text class="status" :class="d.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(d)">{{ d.status === '0' ? '启用' : '停用' }}</text>
        </view>
        <view class="def-line sub" @tap="showDetail(d)">
          <text class="key">{{ d.defKey }}</text>
          <text v-if="d.category" class="cat">{{ d.category }}</text>
          <text class="nodes">{{ (d.nodes || []).length }} 个节点</text>
        </view>
        <view v-if="d.remark" class="def-remark">{{ d.remark }}</view>
        <view class="def-ops">
          <text class="op" @tap="showDetail(d)">查看节点</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无流程定义</view>
    <view v-else class="empty">加载中…</view>

    <!-- 节点详情 -->
    <view v-if="showDetailSheet" class="mask" @tap="showDetailSheet = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (current.defName || 'F').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ current.defName }}</text>
            <text class="sheet-sub">{{ current.defKey }} · {{ current.category || '未分类' }}</text>
          </view>
          <view class="sheet-close" @tap="showDetailSheet = false">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">流程节点（{{ (current.nodes || []).length }}）</text>
          </view>
          <view v-if="(current.nodes || []).length" class="node-list">
            <view v-for="n in current.nodes || []" :key="n.flowNodeId" class="node-item">
              <text class="node-index">{{ n.nodeOrder }}</text>
              <view class="node-main">
                <view class="node-line">
                  <text class="node-name">{{ n.nodeName }}</text>
                  <text class="node-mode" :class="n.nodeMode === 'any' ? 'any' : ''">{{ n.nodeMode === 'any' ? '会签' : '或签' }}</text>
                </view>
                <text class="node-desc">{{ approverText(n.approverType) }}{{ n.timeoutHours ? ' · 超时 ' + n.timeoutHours + ' 小时' : '' }}{{ n.canAddSign === '1' ? ' · 可加签' : '' }}</text>
              </view>
            </view>
          </view>
          <view v-else class="hint">该流程未配置节点</view>
        </view>

        <view class="form-actions">
          <view class="btn plain" @tap="showDetailSheet = false">关闭</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listFlowDefs, changeFlowDefStatus } from '@/api/flow'

const FILTERS = [
  { label: '全部', value: '' },
  { label: '启用', value: '0' },
  { label: '停用', value: '1' }
]

const keyword = ref('')
const status = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showDetailSheet = ref(false)
const current = ref({})

function approverText(type) {
  if (type === 'role') return '按角色审批'
  if (type === 'user') return '指定人员审批'
  if (type === 'dept_leader') return '部门负责人审批'
  return type || '—'
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listFlowDefs({ pageNum: pageNum.value, pageSize, keyword: keyword.value || undefined, status: status.value || undefined })
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

function showDetail(d) {
  current.value = d
  showDetailSheet.value = true
}

async function toggleStatus(d) {
  if (!$hasPerm('flow:def:changeStatus')) {
    uni.showToast({ title: '没有启停流程的权限', icon: 'none' })
    return
  }
  const next = d.status === '0' ? '1' : '0'
  try {
    await changeFlowDefStatus({ flowDefId: d.flowDefId, status: next })
    d.status = next
    uni.showToast({ title: next === '0' ? '已启用' : '已停用', icon: 'none' })
  } catch (e) {}
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '流程定义' })
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
.filter-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 12rpx; margin-top: 16rpx; }
.filter { padding: 10rpx 24rpx; font-size: 24rpx; color: #5a6678; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.filter.active { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.def-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.def-item { padding: 26rpx 0; }
.def-item + .def-item { border-top: 1rpx solid #f0f3f7; }
.def-line { display: flex; align-items: center; gap: 12rpx; }
.def-name { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.def-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.key { color: #409eff; }
.cat::before { content: '·'; margin: 0 8rpx; color: #9aa6b8; }
.nodes { margin-left: auto; }
.def-remark { margin-top: 10rpx; font-size: 24rpx; color: #8a94a6; }
.def-ops { display: flex; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 22rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 86vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; }
.sheet-sub { margin-top: 10rpx; font-size: 24rpx; color: #8a94a6; text-align: center; }
.node-list { margin-top: 24rpx; }
.node-item { display: flex; gap: 18rpx; padding: 20rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.node-index { width: 44rpx; height: 44rpx; line-height: 44rpx; text-align: center; border-radius: 50%; background: rgba(64, 158, 255, 0.1); color: #409eff; font-size: 24rpx; flex-shrink: 0; }
.node-main { flex: 1; min-width: 0; }
.node-line { display: flex; align-items: center; gap: 12rpx; }
.node-name { font-size: 26rpx; color: #1f2d3d; }
.node-mode { font-size: 20rpx; padding: 2rpx 12rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.1); }
.node-mode.any { color: #e6a23c; background: rgba(230, 162, 60, 0.12); }
.node-desc { display: block; margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.hint { padding: 30rpx 0; text-align: center; font-size: 24rpx; color: #9aa6b8; }
.form-actions { display: flex; gap: 20rpx; margin-top: 32rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 弹层精修（新版） ===== */
.sheet { padding: 0; max-height: 90vh; border-radius: 28rpx 28rpx 0 0; overflow-y: auto; background: #fff; }
.sheet-head { position: sticky; top: 0; z-index: 5; display: flex; align-items: center; gap: 20rpx; padding: 30rpx 32rpx 24rpx; background: #fff; border-bottom: 1rpx solid #f2f5fb; }
.sheet-avatar { width: 88rpx; height: 88rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 700; color: #fff; flex-shrink: 0; background: linear-gradient(135deg, #a99bff, #7b61ff); box-shadow: 0 10rpx 24rpx rgba(123, 97, 255, 0.26); }
.sheet-head-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8rpx; }
.sheet-head .sheet-title { font-size: 32rpx; font-weight: 700; color: #1c2738; text-align: left; margin: 0; }
.sheet-sub { margin: 0; font-size: 22rpx; color: #9aa6b8; text-align: left; }
.sheet-close { width: 56rpx; height: 56rpx; line-height: 56rpx; text-align: center; font-size: 32rpx; color: #9aa6b8; background: #f3f6fb; border-radius: 50%; flex-shrink: 0; }
.group { padding: 28rpx 32rpx 0; }
.group-head { display: flex; align-items: center; gap: 12rpx; margin-bottom: 20rpx; }
.group-bar { width: 6rpx; height: 26rpx; border-radius: 4rpx; background: linear-gradient(180deg, #a99bff, #7b61ff); }
.group-title { font-size: 26rpx; font-weight: 600; color: #1c2738; }
.node-list { margin-top: 0; }
.node-item { display: flex; gap: 18rpx; padding: 22rpx; margin-bottom: 16rpx; background: #f5f7fb; border-radius: 16rpx; border-bottom: none; }
.node-index { width: 48rpx; height: 48rpx; line-height: 48rpx; text-align: center; border-radius: 50%; background: linear-gradient(135deg, #a99bff, #7b61ff); color: #fff; font-size: 24rpx; font-weight: 600; flex-shrink: 0; }
.node-name { font-size: 27rpx; font-weight: 500; color: #1c2738; }
.node-desc { display: block; margin-top: 8rpx; font-size: 21rpx; color: #8a94a6; }
.hint { padding: 40rpx 0; text-align: center; font-size: 24rpx; color: #9aa6b8; background: #f5f7fb; border-radius: 16rpx; }
.form-actions { position: sticky; bottom: 0; z-index: 5; display: flex; gap: 20rpx; padding: 20rpx 32rpx 40rpx; margin-top: 12rpx; background: #fff; border-top: 1rpx solid #f2f5fb; }
.btn { flex: 1; height: 92rpx; line-height: 92rpx; text-align: center; color: #fff; font-size: 28rpx; font-weight: 600; border-radius: 18rpx; background: linear-gradient(135deg, #7b61ff, #5b45d6); box-shadow: 0 12rpx 26rpx rgba(91, 69, 214, 0.24); }
.btn.plain { background: #f3f6fb; color: #6b7a90; box-shadow: none; }
</style>
