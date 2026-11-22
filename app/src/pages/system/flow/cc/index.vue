<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入申请标题搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>

    <view v-if="list.length" class="cc-list">
      <view v-for="t in list" :key="t.flowTaskId" class="cc-item" @tap="showDetail(t)">
        <view class="cc-line">
          <text class="cc-title">{{ t.title || '未命名申请' }}</text>
          <text class="status" :class="t.status === '0' ? 'unread' : 'read'">{{ t.status === '0' ? '未读' : '已读' }}</text>
        </view>
        <view class="cc-line sub">
          <text class="def">{{ t.defName || '—' }}</text>
          <text class="node">{{ t.nodeName || '—' }}</text>
        </view>
        <view class="cc-line sub">
          <text>发起人 {{ t.applicantName || '—' }}</text>
          <text class="time">{{ fmtTime(t.createTime) }}</text>
        </view>
        <view v-if="t.status === '0'" class="cc-ops">
          <text class="op" @tap.stop="onRead(t)">标记已读</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无抄送消息</view>
    <view v-else class="empty">加载中…</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listFlowCc, markCcRead, getFlowInstanceDetail } from '@/api/flow'

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

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listFlowCc({ pageNum: pageNum.value, pageSize, keyword: keyword.value || undefined })
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

async function onRead(t) {
  try {
    await markCcRead(t.flowTaskId)
    t.status = '6'
    uni.showToast({ title: '已标记已读', icon: 'none' })
  } catch (e) {}
}

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
  uni.setNavigationBarTitle({ title: '抄送我的' })
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
.cc-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.cc-item { padding: 26rpx 0; }
.cc-item + .cc-item { border-top: 1rpx solid #f0f3f7; }
.cc-line { display: flex; align-items: center; gap: 12rpx; }
.cc-title { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.unread { color: #e6a23c; background: rgba(230, 162, 60, 0.12); }
.status.read { color: #8a94a6; background: #f0f3f7; }
.cc-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.def { color: #409eff; }
.node::before { content: '·'; margin: 0 8rpx; color: #9aa6b8; }
.time { margin-left: auto; }
.cc-ops { display: flex; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 22rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
