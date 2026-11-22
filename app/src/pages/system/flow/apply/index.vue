<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入申请标题搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>
    <view class="tool-bar">
      <view v-if="$hasPerm('flow:apply:add')" class="tool-btn add" @tap="openStart">＋ 发起申请</view>
    </view>
    <view class="filter-bar">
      <text v-for="f in FILTERS" :key="f.value" class="filter" :class="{ active: status === f.value }" @tap="onFilter(f.value)">{{ f.label }}</text>
    </view>

    <view v-if="list.length" class="apply-list">
      <view v-for="a in list" :key="a.flowInstanceId" class="apply-item">
        <view class="apply-line" @tap="showDetail(a)">
          <text class="apply-title">{{ a.title || '未命名申请' }}</text>
          <text class="status" :class="statusClass(a.status)">{{ statusText(a.status) }}</text>
        </view>
        <view class="apply-line sub" @tap="showDetail(a)">
          <text class="def">{{ a.defName || '—' }}</text>
          <text class="time">{{ fmtTime(a.createTime) }}</text>
        </view>
        <view class="apply-ops">
          <text v-if="a.status === '1' && $hasPerm('flow:apply:withdraw')" class="op" @tap="onWithdraw(a)">撤回</text>
          <text v-if="a.status === '1' && $hasPerm('flow:apply:urge')" class="op" @tap="onUrge(a)">催办</text>
          <text v-if="a.status === '1' && $hasPerm('flow:apply:terminate')" class="op danger" @tap="openTerminate(a)">终止</text>
          <text v-if="a.status === '3' && $hasPerm('flow:apply:reapply')" class="op" @tap="openReapply(a)">重新提交</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无申请记录</view>
    <view v-else class="empty">加载中…</view>

    <!-- 发起申请 -->
    <view v-if="showStart" class="mask" @tap="closeStart">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">发起申请</view>
        <view class="form-item column">
          <text class="label">选择流程</text>
          <view v-if="defs.length" class="chips">
            <text v-for="d in defs" :key="d.flowDefId" class="chip" :class="{ active: startForm.flowDefId === d.flowDefId }" @tap="startForm.flowDefId = d.flowDefId">{{ d.defName }}</text>
          </view>
          <text v-else class="hint">暂无启用的流程，请联系管理员在 Web 端配置</text>
        </view>
        <view class="form-item">
          <text class="label">申请标题</text>
          <input v-model="startForm.title" class="input" placeholder="必填" placeholder-class="ph" />
        </view>
        <view class="form-item column">
          <text class="label">申请内容</text>
          <textarea v-model="startForm.content" class="textarea" placeholder="填写申请说明" placeholder-class="ph" maxlength="500" />
        </view>
        <view class="form-actions">
          <view class="btn plain" @tap="closeStart">取消</view>
          <view class="btn" @tap="submitStart">提交申请</view>
        </view>
      </view>
    </view>

    <!-- 重新提交 -->
    <view v-if="showReapply" class="mask" @tap="closeReapply">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">重新提交</view>
        <view class="form-item">
          <text class="label">申请标题</text>
          <input v-model="reapplyForm.title" class="input" placeholder="必填" placeholder-class="ph" />
        </view>
        <view class="form-item column">
          <text class="label">申请内容</text>
          <textarea v-model="reapplyForm.content" class="textarea" placeholder="填写申请说明" placeholder-class="ph" maxlength="500" />
        </view>
        <view class="form-actions">
          <view class="btn plain" @tap="closeReapply">取消</view>
          <view class="btn" @tap="submitReapply">提交</view>
        </view>
      </view>
    </view>

    <!-- 终止申请 -->
    <view v-if="showTerminate" class="mask" @tap="closeTerminate">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">终止申请</view>
        <view class="sheet-sub">{{ current.title || '未命名申请' }}</view>
        <textarea v-model="terminateComment" class="textarea" placeholder="填写终止原因（可选）" placeholder-class="ph" maxlength="200" />
        <view class="form-actions">
          <view class="btn plain" @tap="closeTerminate">取消</view>
          <view class="btn danger" @tap="submitTerminate">确定终止</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listFlowApply, listEnabledDefs, startFlow, withdrawFlow, reapplyFlow, urgeFlow, terminateFlow, getFlowInstanceDetail } from '@/api/flow'
import { trimStr } from '@/utils/validate'

const FILTERS = [
  { label: '全部', value: '' },
  { label: '运行中', value: '1' },
  { label: '已通过', value: '2' },
  { label: '已驳回', value: '3' },
  { label: '已撤销', value: '4' },
  { label: '已终止', value: '5' }
]
const STATUS_TEXT = { 1: '运行中', 2: '已通过', 3: '已驳回', 4: '已撤销', 5: '已终止' }
const ACTION_TEXT = { 1: '发起', 2: '通过', 3: '驳回', 4: '转办', 5: '加签', 6: '撤回', 7: '撤销', 8: '终止', 9: '催办', 10: '重新提交', 11: '抄送' }

const keyword = ref('')
const status = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

const defs = ref([])
const showStart = ref(false)
const startForm = reactive({ flowDefId: 0, title: '', content: '' })

const showReapply = ref(false)
const reapplyForm = reactive({ flowInstanceId: 0, title: '', content: '' })

const showTerminate = ref(false)
const current = ref({})
const terminateComment = ref('')

function fmtTime(t) {
  if (!t) return '—'
  return String(t).replace('T', ' ').slice(0, 19)
}

function statusText(s) { return STATUS_TEXT[s] || '—' }
function statusClass(s) {
  if (s === '1') return 'run'
  if (s === '2') return 'ok'
  if (s === '3') return 'bad'
  return 'gray'
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listFlowApply({ pageNum: pageNum.value, pageSize, keyword: keyword.value || undefined, status: status.value || undefined })
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

async function openStart() {
  startForm.flowDefId = 0
  startForm.title = ''
  startForm.content = ''
  showStart.value = true
  if (!defs.value.length) {
    try {
      const data = await listEnabledDefs()
      defs.value = data.data || []
      if (defs.value.length) startForm.flowDefId = defs.value[0].flowDefId
    } catch (e) {}
  } else {
    startForm.flowDefId = defs.value[0].flowDefId
  }
}

function closeStart() { showStart.value = false }

async function submitStart() {
  if (!startForm.flowDefId) {
    uni.showToast({ title: '请选择流程', icon: 'none' })
    return
  }
  const title = trimStr(startForm.title)
  if (!title) {
    uni.showToast({ title: '请填写申请标题', icon: 'none' })
    return
  }
  if (title.length > 100) {
    uni.showToast({ title: '申请标题不能超过 100 个字符', icon: 'none' })
    return
  }
  try {
    await startFlow({ flowDefId: startForm.flowDefId, title, content: trimStr(startForm.content), ccUserIds: [] })
    closeStart()
    uni.showToast({ title: '提交成功', icon: 'success' })
    setTimeout(() => fetchList(true), 400)
  } catch (e) {}
}

function onWithdraw(a) {
  uni.showModal({
    title: '提示',
    content: `确定撤回申请「${a.title}」吗？`,
    confirmText: '撤回',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await withdrawFlow(a.flowInstanceId)
        uni.showToast({ title: '已撤回', icon: 'success' })
        fetchList(true)
      } catch (e) {}
    }
  })
}

function onUrge(a) {
  uni.showModal({
    title: '提示',
    content: `确定催办申请「${a.title}」吗？`,
    confirmText: '催办',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await urgeFlow(a.flowInstanceId)
        uni.showToast({ title: '已催办', icon: 'success' })
      } catch (e) {}
    }
  })
}

function openReapply(a) {
  reapplyForm.flowInstanceId = a.flowInstanceId
  reapplyForm.title = a.title || ''
  reapplyForm.content = a.content || ''
  showReapply.value = true
}

function closeReapply() { showReapply.value = false }

async function submitReapply() {
  const title = trimStr(reapplyForm.title)
  if (!title) {
    uni.showToast({ title: '请填写申请标题', icon: 'none' })
    return
  }
  if (title.length > 100) {
    uni.showToast({ title: '申请标题不能超过 100 个字符', icon: 'none' })
    return
  }
  try {
    await reapplyFlow(reapplyForm.flowInstanceId, { title, content: trimStr(reapplyForm.content) })
    closeReapply()
    uni.showToast({ title: '已重新提交', icon: 'success' })
    fetchList(true)
  } catch (e) {}
}

function openTerminate(a) {
  current.value = a
  terminateComment.value = ''
  showTerminate.value = true
}

function closeTerminate() { showTerminate.value = false }

async function submitTerminate() {
  const comment = trimStr(terminateComment.value)
  if (!comment) {
    uni.showToast({ title: '请填写终止原因', icon: 'none' })
    return
  }
  try {
    await terminateFlow(current.value.flowInstanceId, { comment })
    closeTerminate()
    uni.showToast({ title: '已终止', icon: 'success' })
    fetchList(true)
  } catch (e) {}
}

async function showDetail(a) {
  if (!$hasPerm('flow:instance:query')) return
  try {
    const data = await getFlowInstanceDetail(a.flowInstanceId)
    const d = data.data || {}
    const inst = d.instance || {}
    const records = (d.records || []).map((r) => `${r.operatorName || '—'} ${ACTION_TEXT[r.action] || r.action || ''}${r.comment ? '：' + r.comment : ''}`)
    const content = [
      `流程：${inst.defName || a.defName || '—'}`,
      `状态：${STATUS_TEXT[inst.status] || statusText(a.status)}`,
      inst.content ? `申请内容：${inst.content}` : '',
      records.length ? `审批记录：\n${records.slice(-6).join('\n')}` : ''
    ].filter(Boolean).join('\n')
    uni.showModal({ title: inst.title || a.title || '申请详情', content, showCancel: false, confirmText: '关闭' })
  } catch (e) {}
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '发起申请' })
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
.tool-bar { display: flex; justify-content: flex-end; margin-top: 16rpx; }
.tool-btn { padding: 0 22rpx; height: 64rpx; line-height: 64rpx; background: #67c23a; color: #fff; font-size: 26rpx; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.filter-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 12rpx; margin-top: 16rpx; }
.filter { padding: 10rpx 24rpx; font-size: 24rpx; color: #5a6678; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.filter.active { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.apply-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.apply-item { padding: 26rpx 0; }
.apply-item + .apply-item { border-top: 1rpx solid #f0f3f7; }
.apply-line { display: flex; align-items: center; gap: 12rpx; }
.apply-title { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.run { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.status.ok { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.bad { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.status.gray { color: #8a94a6; background: #f0f3f7; }
.apply-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.def { color: #409eff; }
.time { margin-left: auto; }
.apply-ops { display: flex; flex-wrap: wrap; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 22rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 88vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; }
.sheet-sub { margin-top: 12rpx; font-size: 24rpx; color: #8a94a6; text-align: center; }
.form-item { display: flex; align-items: center; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.form-item.column { flex-direction: column; align-items: stretch; gap: 12rpx; }
.label { width: 140rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.form-item.column .label { width: auto; }
.input { flex: 1; height: 64rpx; font-size: 26rpx; }
.textarea { width: 100%; height: 180rpx; padding: 20rpx; box-sizing: border-box; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.hint { font-size: 24rpx; color: #9aa6b8; padding: 8rpx 0; }
.chips { display: flex; flex-wrap: wrap; gap: 12rpx; }
.chip { font-size: 24rpx; padding: 10rpx 22rpx; border-radius: 18rpx; color: #5a6678; background: #f5f7fa; }
.chip.active { color: #409eff; background: rgba(64, 158, 255, 0.12); }
.form-actions { display: flex; gap: 20rpx; margin-top: 32rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
.btn.danger { background: #f56c6c; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
