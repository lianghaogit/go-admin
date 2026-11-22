<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入公告标题搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('system:notice:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>

    <view v-if="list.length" class="notice-list">
      <view v-for="n in list" :key="n.noticeId" class="notice-item">
        <view class="avatar" :style="{ background: avatarColor(n.noticeId) }">{{ (n.noticeTitle || 'N').charAt(0) }}</view>
        <view class="notice-main" @tap="openDetail(n)">
          <view class="notice-line">
            <text class="notice-title">{{ n.noticeTitle }}</text>
            <text class="type-tag" :class="n.noticeType === '1' ? 'notify' : 'announce'">{{ n.noticeType === '1' ? '通知' : '公告' }}</text>
            <text class="status" :class="n.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(n)">{{ n.status === '0' ? '正常' : '关闭' }}</text>
          </view>
          <view class="notice-line sub">
            <text class="notice-meta">{{ n.createBy || '—' }}</text>
            <text class="notice-meta">{{ fmtTime(n.createTime) }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('system:notice:edit')" class="op" @tap="openForm(n)">✎</view>
        <view v-if="$hasPerm('system:notice:remove')" class="op danger" @tap="onDelete(n)">✕</view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无公告数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 详情弹层 -->
    <view v-if="detail" class="mask" @tap="detail = null">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar announce-avatar">{{ (detail.noticeTitle || 'N').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ detail.noticeTitle }}</text>
            <text class="sheet-sub">{{ detail.createBy || '—' }} · {{ fmtTime(detail.createTime) }}</text>
          </view>
          <view class="sheet-close" @tap="detail = null">×</view>
        </view>
        <view class="detail-body">
          <view class="detail-meta">
            <text class="type-tag" :class="detail.noticeType === '1' ? 'notify' : 'announce'">{{ detail.noticeType === '1' ? '通知' : '公告' }}</text>
            <text class="status" :class="detail.status === '0' ? 'on' : 'off'">{{ detail.status === '0' ? '正常' : '关闭' }}</text>
          </view>
          <view class="detail-content">{{ detail.noticeContent || '暂无内容' }}</view>
        </view>
        <view class="form-actions">
          <view class="btn plain" @tap="detail = null">关闭</view>
        </view>
      </view>
    </view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.noticeTitle || 'N').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.noticeId ? '编辑公告' : '新增公告' }}</text>
            <text class="sheet-sub">{{ form.noticeId ? '调整公告内容与发布状态' : '发布一条通知或公告' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">公告标题</text>
            <input v-model="form.noticeTitle" class="field-input" placeholder="必填，简洁描述公告主题" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">公告类型</text>
            <picker :range="typeOptions" range-key="label" :value="typeIndex" @change="onTypeChange">
              <view class="field-select">
                <text class="field-select-text">{{ typeOptions[typeIndex].label }}</text>
                <text class="field-select-arrow">›</text>
              </view>
            </picker>
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">公告内容</text>
          </view>
          <view class="field">
            <textarea v-model="form.noticeContent" class="field-textarea" placeholder="必填，填写公告正文" placeholder-class="ph" :maxlength="-1" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">发布状态</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">公告状态</text>
              <text class="field-desc">关闭后公告不再展示给用户</text>
            </view>
            <switch :checked="form.status === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '正常' : '关闭' }}</text>
          </view>
        </view>

        <view class="form-actions">
          <view class="btn plain" @tap="closeForm">取消</view>
          <view class="btn" @tap="submitForm">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listNotice, createNotice, updateNotice, deleteNotice } from '@/api/notice'
import { trimStr } from '@/utils/validate'

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showForm = ref(false)
const detail = ref(null)
const form = reactive({ noticeId: 0, noticeTitle: '', noticeType: '1', noticeContent: '', status: '0', remark: '' })

const typeOptions = [
  { value: '1', label: '通知' },
  { value: '2', label: '公告' }
]
const typeIndex = computed(() => Math.max(0, typeOptions.findIndex((o) => o.value === form.noticeType)))

const PALETTE = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#9c6ade', '#00b8a9', '#fa8c5f', '#5b8ff9']
function avatarColor(id) {
  return PALETTE[Number(id) % PALETTE.length]
}

function fmtTime(t) {
  if (!t) return ''
  return String(t).replace('T', ' ').slice(0, 19)
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listNotice({ pageNum: pageNum.value, pageSize, noticeTitle: keyword.value || undefined })
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

function openDetail(n) { detail.value = n }

function openForm(n) {
  if (n) {
    Object.assign(form, { noticeId: n.noticeId, noticeTitle: n.noticeTitle || '', noticeType: n.noticeType || '1', noticeContent: n.noticeContent || '', status: n.status || '0', remark: n.remark || '' })
  } else {
    Object.assign(form, { noticeId: 0, noticeTitle: '', noticeType: '1', noticeContent: '', status: '0', remark: '' })
  }
  showForm.value = true
}

function closeForm() { showForm.value = false }

function onSwitch(e) { form.status = e.detail.value ? '0' : '1' }

function onTypeChange(e) { form.noticeType = typeOptions[Number(e.detail.value)].value }

async function submitForm() {
  const noticeTitle = trimStr(form.noticeTitle)
  const noticeContent = trimStr(form.noticeContent)
  if (!noticeTitle) {
    uni.showToast({ title: '请输入公告标题', icon: 'none' })
    return
  }
  if (noticeTitle.length > 50) {
    uni.showToast({ title: '公告标题不能超过 50 个字符', icon: 'none' })
    return
  }
  if (!noticeContent) {
    uni.showToast({ title: '请输入公告内容', icon: 'none' })
    return
  }
  if (!['1', '2'].includes(form.noticeType)) {
    uni.showToast({ title: '请选择公告类型', icon: 'none' })
    return
  }
  Object.assign(form, { noticeTitle, noticeContent })
  const payload = { noticeTitle, noticeType: form.noticeType, noticeContent, status: form.status, remark: trimStr(form.remark) }
  const data = form.noticeId ? await updateNotice({ ...payload, noticeId: form.noticeId }) : await createNotice(payload)
  if (data.code === 200) {
    closeForm()
    uni.showToast({ title: form.noticeId ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
    if (form.noticeId) {
      const idx = list.value.findIndex((it) => it.noticeId === form.noticeId)
      if (idx >= 0) list.value[idx] = { ...list.value[idx], ...payload }
    } else {
      setTimeout(() => fetchList(true), 400)
    }
  } else {
    uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
  }
}

function toggleStatus(n) {
  const next = n.status === '0' ? '1' : '0'
  openForm({ ...n, status: next })
}

function onDelete(n) {
  uni.showModal({
    title: '提示',
    content: `确定删除公告「${n.noticeTitle}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      const data = await deleteNotice(n.noticeId)
      if (data.code === 200) {
        uni.showToast({ title: '删除成功', icon: 'success' })
        list.value = list.value.filter((it) => it.noticeId !== n.noticeId)
        total.value = Math.max(0, total.value - 1)
        if (!list.value.length) fetchList(true)
      } else {
        uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
      }
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '通知公告' })
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
.add-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #67c23a; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.notice-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.notice-item { display: flex; align-items: center; gap: 16rpx; padding: 26rpx 0; }
.notice-item + .notice-item { border-top: 1rpx solid #f0f3f7; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.notice-main { flex: 1; min-width: 0; }
.notice-line { display: flex; align-items: center; gap: 12rpx; }
.notice-title { font-size: 28rpx; color: #1f2d3d; max-width: 320rpx; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.type-tag { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.type-tag.notify { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.type-tag.announce { color: #e6a23c; background: rgba(230, 162, 60, 0.1); }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.notice-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.notice-meta + .notice-meta::before { content: '·'; margin: 0 8rpx; }
.op { width: 60rpx; height: 60rpx; line-height: 60rpx; text-align: center; font-size: 28rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); border-radius: 50%; flex-shrink: 0; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 86vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.detail-sheet { display: flex; flex-direction: column; }
.detail-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; }
.detail-meta { display: flex; align-items: center; justify-content: center; gap: 12rpx; margin: 20rpx 0; font-size: 22rpx; color: #9aa6b8; }
.detail-content { font-size: 26rpx; color: #5a6678; line-height: 1.7; max-height: 40vh; overflow-y: auto; }
.detail-close { margin-top: 40rpx; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; margin-bottom: 28rpx; }
.form-item { display: flex; align-items: center; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.form-item.col { flex-direction: column; align-items: flex-start; gap: 12rpx; }
.label { width: 140rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.input { flex: 1; height: 64rpx; font-size: 26rpx; }
.textarea { width: 100%; min-height: 200rpx; font-size: 26rpx; background: #f5f7fa; border-radius: 18rpx; padding: 16rpx 20rpx; box-sizing: border-box; }
.picker-value { height: 64rpx; line-height: 64rpx; font-size: 26rpx; color: #1f2d3d; }
.switch-text { font-size: 24rpx; color: #8a94a6; }
.form-actions { display: flex; gap: 20rpx; margin-top: 40rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 弹层精修（新版） ===== */
.sheet { padding: 0; max-height: 90vh; border-radius: 28rpx 28rpx 0 0; overflow-y: auto; background: #fff; }
.sheet-head { position: sticky; top: 0; z-index: 5; display: flex; align-items: center; gap: 20rpx; padding: 30rpx 32rpx 24rpx; background: #fff; border-bottom: 1rpx solid #f2f5fb; }
.sheet-avatar { width: 88rpx; height: 88rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 700; color: #fff; flex-shrink: 0; background: linear-gradient(135deg, #6aa4ff, #3a6fe0); box-shadow: 0 10rpx 24rpx rgba(58, 111, 224, 0.26); }
.sheet-head-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8rpx; }
.sheet-head .sheet-title { font-size: 32rpx; font-weight: 700; color: #1c2738; text-align: left; margin: 0; line-height: 1.35; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.sheet-sub { font-size: 22rpx; color: #9aa6b8; }
.sheet-close { width: 56rpx; height: 56rpx; line-height: 56rpx; text-align: center; font-size: 32rpx; color: #9aa6b8; background: #f3f6fb; border-radius: 50%; flex-shrink: 0; }
.group { padding: 28rpx 32rpx 0; }
.group-head { display: flex; align-items: center; gap: 12rpx; margin-bottom: 20rpx; }
.group-bar { width: 6rpx; height: 26rpx; border-radius: 4rpx; background: linear-gradient(180deg, #6aa4ff, #3a6fe0); }
.group-title { font-size: 26rpx; font-weight: 600; color: #1c2738; }
.field { display: flex; flex-direction: column; gap: 12rpx; margin-bottom: 24rpx; }
.field-label { font-size: 24rpx; color: #6b7a90; }
.field-input { width: 100%; height: 84rpx; padding: 0 24rpx; box-sizing: border-box; background: #f5f7fb; border-radius: 14rpx; font-size: 27rpx; color: #1c2738; }
.field-textarea { width: 100%; min-height: 210rpx; padding: 20rpx 24rpx; box-sizing: border-box; background: #f5f7fb; border-radius: 14rpx; font-size: 26rpx; line-height: 1.6; color: #1c2738; }
.field-select { display: flex; align-items: center; justify-content: space-between; height: 84rpx; padding: 0 24rpx; background: #f5f7fb; border-radius: 14rpx; }
.field-select-text { flex: 1; min-width: 0; font-size: 27rpx; color: #1c2738; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.field-select-arrow { font-size: 30rpx; color: #b9c3d3; }
.field-card { display: flex; align-items: center; gap: 16rpx; padding: 24rpx; margin-bottom: 24rpx; background: #f5f7fb; border-radius: 16rpx; }
.field-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.field-label-inline { font-size: 26rpx; font-weight: 500; color: #1c2738; }
.field-desc { font-size: 21rpx; color: #9aa6b8; }
.status-pill { font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 20rpx; color: #3fae7a; background: rgba(63, 174, 122, 0.12); flex-shrink: 0; }
.status-pill.off { color: #f56c6c; background: rgba(245, 108, 108, 0.12); }
.detail-body { padding: 28rpx 32rpx 0; }
.detail-meta { display: flex; align-items: center; gap: 12rpx; margin: 0 0 20rpx; }
.detail-content { padding: 24rpx; background: #f7f9fd; border-radius: 16rpx; font-size: 27rpx; color: #3f4c60; line-height: 1.8; max-height: 44vh; overflow-y: auto; white-space: pre-wrap; }
.form-actions { position: sticky; bottom: 0; z-index: 5; display: flex; gap: 20rpx; padding: 20rpx 32rpx 40rpx; background: #fff; border-top: 1rpx solid #f2f5fb; }
.btn { flex: 1; height: 92rpx; line-height: 92rpx; text-align: center; color: #fff; font-size: 28rpx; font-weight: 600; border-radius: 18rpx; background: linear-gradient(135deg, #4a7df0, #3457c9); box-shadow: 0 12rpx 26rpx rgba(52, 87, 201, 0.24); }
.btn.plain { background: #f3f6fb; color: #6b7a90; box-shadow: none; }
</style>
