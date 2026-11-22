<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入申请标题搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>

    <view v-if="list.length" class="done-list">
      <view v-for="t in list" :key="t.flowTaskId" class="done-item" @tap="showDetail(t)">
        <view class="done-line">
          <text class="done-title">{{ t.title || '未命名申请' }}</text>
          <text class="status" :class="t.status === '1' ? 'ok' : (t.status === '2' ? 'bad' : 'gray')">{{ statusText(t.status) }}</text>
        </view>
        <view class="done-line sub">
          <text class="def">{{ t.defName || '—' }}</text>
          <text class="node">{{ t.nodeName || '—' }}</text>
        </view>
        <view class="done-line sub">
          <text>发起人 {{ t.applicantName || '—' }}</text>
          <text class="time">{{ fmtTime(t.passedAt || t.createTime) }}</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无已办任务</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listFlowDone, getFlowInstanceDetail } from '@/api/flow'

const STATUS_TEXT = { 1: '已通过', 2: '已驳回', 3: '已转出' }
const ACTION_TEXT = { 1: '发起', 2: '通过', 3: '驳回', 4: '转办', 5: '加签', 6: '撤回', 7: '撤销', 8: '终止', 9: '催办', 10: '重新提交', 11: '抄送' }

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

function statusText(s) { return STATUS_TEXT[s] || '—' }

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listFlowDone({ pageNum: pageNum.value, pageSize, keyword: keyword.value || undefined })
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
function onReset() { keyword.value = ''; fetchList(true) }

async function showDetail(t) {
  if (!$hasPerm('flow:instance:query')) return
  try {
    const data = await getFlowInstanceDetail(t.instanceId)
    const d = data.data || {}
    const inst = d.instance || {}
    const records = (d.records || []).map((r) => `${r.operatorName || '—'} ${ACTION_TEXT[r.action] || r.action || ''}${r.comment ? '：' + r.comment : ''}`)
    const content = [
      `流程：${inst.defName || t.defName || '—'}`,
      `发起人：${inst.applicantName || t.applicantName || '—'}`,
      inst.content ? `申请内容：${inst.content}` : '',
      records.length ? `审批记录：\n${records.slice(-6).join('\n')}` : ''
    ].filter(Boolean).join('\n')
    uni.showModal({ title: inst.title || t.title || '申请详情', content, showCancel: false, confirmText: '关闭' })
  } catch (e) {}
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '已办任务' })
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
.done-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.done-item { padding: 26rpx 0; }
.done-item + .done-item { border-top: 1rpx solid #f0f3f7; }
.done-line { display: flex; align-items: center; gap: 12rpx; }
.done-title { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.ok { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.bad { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.status.gray { color: #8a94a6; background: #f0f3f7; }
.done-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.def { color: #409eff; }
.node::before { content: '·'; margin: 0 8rpx; color: #9aa6b8; }
.time { margin-left: auto; }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
