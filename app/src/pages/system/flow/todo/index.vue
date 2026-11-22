<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入申请标题搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>

    <view v-if="list.length" class="todo-list">
      <view v-for="t in list" :key="t.flowTaskId" class="todo-item">
        <view class="todo-line" @tap="showDetail(t)">
          <text class="todo-title">{{ t.title || '未命名申请' }}</text>
          <text v-if="isOverdue(t)" class="overdue">已超时</text>
        </view>
        <view class="todo-line sub" @tap="showDetail(t)">
          <text class="def">{{ t.defName || '—' }}</text>
          <text class="node">{{ t.nodeName || '—' }}</text>
        </view>
        <view class="todo-line sub" @tap="showDetail(t)">
          <text>发起人 {{ t.applicantName || '—' }}</text>
          <text class="time">{{ fmtTime(t.createTime) }}</text>
        </view>
        <view class="todo-ops">
          <text v-if="$hasPerm('flow:todo:approve')" class="op" @tap="openAction(t, 'approve')">通过</text>
          <text v-if="$hasPerm('flow:todo:reject')" class="op danger" @tap="openAction(t, 'reject')">驳回</text>
          <text v-if="$hasPerm('flow:todo:transfer')" class="op warn" @tap="openAction(t, 'transfer')">转办</text>
          <text v-if="$hasPerm('flow:todo:addSign')" class="op info" @tap="openAction(t, 'addSign')">加签</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无待办审批</view>
    <view v-else class="empty">加载中…</view>

    <!-- 审批弹层 -->
    <view v-if="showAction" class="mask" @tap="closeAction">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">{{ actionTitle }}</view>
        <view class="sheet-sub">{{ current.title || '未命名申请' }}</view>
        <view v-if="needUser" class="picker">
          <view class="picker-search">
            <input v-model="userKeyword" class="picker-input" placeholder="搜索用户名选择人员" placeholder-class="ph" confirm-type="search" @confirm="loadUsers" />
            <view class="picker-btn" @tap="loadUsers">搜索</view>
          </view>
          <scroll-view scroll-y class="picker-list">
            <view v-for="u in users" :key="u.userId" class="picker-item" :class="{ active: pickedUser && pickedUser.userId === u.userId }" @tap="pickedUser = u">
              <view class="picker-main">
                <text class="picker-name">{{ u.nickName || u.userName }}</text>
                <text class="picker-uname">@{{ u.userName }}</text>
              </view>
              <text v-if="u.dept && u.dept.deptName" class="picker-dept">{{ u.dept.deptName }}</text>
            </view>
            <view v-if="!users.length" class="hint">暂无用户{{ userKeyword ? '，换个关键词试试' : '' }}</view>
          </scroll-view>
        </view>
        <textarea v-model="comment" class="textarea" :placeholder="needUser ? '填写说明（可选）' : '填写审批意见（可选）'" placeholder-class="ph" maxlength="200" />
        <view class="form-actions">
          <view class="btn plain" @tap="closeAction">取消</view>
          <view class="btn" :class="{ danger: actionType === 'reject' }" @tap="submitAction">确定{{ submitLabel }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listFlowTodo, approveFlowTask, rejectFlowTask, transferFlowTask, addSignFlowTask, getFlowInstanceDetail } from '@/api/flow'
import { listUser } from '@/api/user'
import { trimStr } from '@/utils/validate'

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showAction = ref(false)
const actionType = ref('approve')
const comment = ref('')
const current = ref({})
const users = ref([])
const userKeyword = ref('')
const pickedUser = ref(null)

const ACTION_TITLE = { approve: '审批通过', reject: '审批驳回', transfer: '转办任务', addSign: '加签审批' }
const SUBMIT_LABEL = { approve: '通过', reject: '驳回', transfer: '转办', addSign: '加签' }
const needUser = computed(() => actionType.value === 'transfer' || actionType.value === 'addSign')
const actionTitle = computed(() => ACTION_TITLE[actionType.value] || '审批')
const submitLabel = computed(() => SUBMIT_LABEL[actionType.value] || '提交')

function fmtTime(t) {
  if (!t) return '—'
  return String(t).replace('T', ' ').slice(0, 19)
}

function isOverdue(t) {
  if (!t.overdueAt) return false
  return new Date(t.overdueAt).getTime() < Date.now()
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listFlowTodo({ pageNum: pageNum.value, pageSize, keyword: keyword.value || undefined })
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
  try {
    const data = await getFlowInstanceDetail(t.instanceId)
    const d = data.data || {}
    const inst = d.instance || {}
    const records = (d.records || []).map((r) => `${r.operatorName || '—'} ${ACTION_TEXT[r.action] || r.action || ''}${r.comment ? '：' + r.comment : ''}`)
    const content = [
      `流程：${inst.defName || t.defName || '—'}`,
      `发起人：${inst.applicantName || t.applicantName || '—'}`,
      `当前节点：${t.nodeName || '—'}`,
      inst.content ? `申请内容：${inst.content}` : '',
      records.length ? `审批记录：\n${records.slice(-6).join('\n')}` : ''
    ].filter(Boolean).join('\n')
    uni.showModal({ title: inst.title || t.title || '申请详情', content, showCancel: false, confirmText: '关闭' })
  } catch (e) {}
}

const ACTION_TEXT = { 1: '发起', 2: '通过', 3: '驳回', 4: '转办', 5: '加签', 6: '撤回', 7: '撤销', 8: '终止', 9: '催办', 10: '重新提交', 11: '抄送' }

function openAction(t, type) {
  current.value = t
  actionType.value = type
  comment.value = ''
  users.value = []
  userKeyword.value = ''
  pickedUser.value = null
  showAction.value = true
  if (type === 'transfer' || type === 'addSign') loadUsers()
}

async function loadUsers() {
  try {
    const data = await listUser({ pageNum: 1, pageSize: 20, username: userKeyword.value || undefined })
    users.value = data.rows || []
  } catch (e) {}
}

function closeAction() { showAction.value = false }

async function submitAction() {
  const t = current.value
  if (!t) return
  const comment = trimStr(comment.value)
  if (actionType.value === 'reject' && !comment) {
    uni.showToast({ title: '驳回时请填写审批意见', icon: 'none' })
    return
  }
  const payload = { taskId: t.flowTaskId, comment }
  if (needUser.value) {
    if (!pickedUser.value) {
      uni.showToast({ title: '请选择人员', icon: 'none' })
      return
    }
    payload.toUserId = pickedUser.value.userId
  }
  try {
    if (actionType.value === 'approve') await approveFlowTask(payload)
    else if (actionType.value === 'reject') await rejectFlowTask(payload)
    else if (actionType.value === 'transfer') await transferFlowTask(payload)
    else await addSignFlowTask(payload)
    const doneText = { approve: '已通过', reject: '已驳回', transfer: '已转办', addSign: '已加签' }[actionType.value]
    uni.showToast({ title: doneText, icon: 'success' })
    closeAction()
    list.value = list.value.filter((it) => it.flowTaskId !== t.flowTaskId)
    total.value = Math.max(0, total.value - 1)
    if (!list.value.length) fetchList(true)
  } catch (e) {}
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '待办审批' })
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
.todo-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.todo-item { padding: 26rpx 0; }
.todo-item + .todo-item { border-top: 1rpx solid #f0f3f7; }
.todo-line { display: flex; align-items: center; gap: 12rpx; }
.todo-title { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.overdue { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; color: #f56c6c; background: rgba(245, 108, 108, 0.1); flex-shrink: 0; }
.todo-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.def { color: #409eff; }
.node::before { content: '·'; margin: 0 8rpx; color: #9aa6b8; }
.time { margin-left: auto; }
.todo-ops { display: flex; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 26rpx; border-radius: 18rpx; color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.op.warn { color: #e6a23c; background: rgba(230, 162, 60, 0.12); }
.op.info { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.picker { margin-top: 20rpx; }
.picker-search { display: flex; align-items: center; gap: 12rpx; }
.picker-input { flex: 1; min-width: 0; height: 64rpx; padding: 0 18rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 24rpx; }
.picker-btn { padding: 0 20rpx; height: 64rpx; line-height: 64rpx; background: #409eff; color: #fff; font-size: 24rpx; border-radius: 18rpx; flex-shrink: 0; }
.picker-list { max-height: 300rpx; margin-top: 12rpx; }
.picker-item { display: flex; align-items: center; gap: 12rpx; padding: 16rpx 18rpx; border-radius: 18rpx; background: #f7f9fc; margin-bottom: 10rpx; }
.picker-item.active { background: rgba(64, 158, 255, 0.12); }
.picker-main { flex: 1; min-width: 0; }
.picker-name { font-size: 26rpx; color: #1f2d3d; }
.picker-uname { margin-left: 8rpx; font-size: 22rpx; color: #9aa6b8; }
.picker-dept { font-size: 22rpx; color: #8a94a6; flex-shrink: 0; }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; }
.sheet-sub { margin-top: 12rpx; font-size: 24rpx; color: #8a94a6; text-align: center; }
.textarea { width: 100%; height: 180rpx; margin-top: 24rpx; padding: 20rpx; box-sizing: border-box; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.form-actions { display: flex; gap: 20rpx; margin-top: 32rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #67c23a; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
.btn.danger { background: #f56c6c; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
