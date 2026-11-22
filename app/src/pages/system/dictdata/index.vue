<template>
  <view class="page">
    <view class="head-bar">
      <view class="head-info">
        <text class="head-name">{{ dictName }}</text>
        <text class="head-type">{{ dictType }}</text>
      </view>
      <view v-if="$hasPerm('system:dict:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>

    <view v-if="list.length" class="data-list">
      <view v-for="d in list" :key="d.dictCode" class="data-item">
        <view class="avatar" :style="{ background: avatarColor(d.dictCode) }">{{ (d.dictLabel || 'V').charAt(0) }}</view>
        <view class="data-main">
          <view class="data-line">
            <text class="data-label">{{ d.dictLabel }}</text>
            <text class="status" :class="d.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(d)">{{ d.status === '0' ? '正常' : '停用' }}</text>
          </view>
          <view class="data-line sub">
            <text class="data-value">{{ d.dictValue }}</text>
            <text class="data-sort">排序 {{ d.dictSort }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('system:dict:edit')" class="op" @tap="openForm(d)">✎</view>
        <view v-if="$hasPerm('system:dict:remove')" class="op danger" @tap="onDelete(d)">✕</view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无字典数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.dictLabel || 'V').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.dictCode ? '编辑字典数据' : '新增字典数据' }}</text>
            <text class="sheet-sub">{{ dictName ? `所属字典：${dictName}` : '为当前字典添加一项取值' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">数据标签</text>
            <input v-model="form.dictLabel" class="field-input" placeholder="必填，如 男" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">数据键值</text>
            <input v-model="form.dictValue" class="field-input" placeholder="必填，如 0" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">显示排序</text>
            <input v-model="form.dictSort" class="field-input" type="number" placeholder="数字越小越靠前" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">数据状态</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">数据状态</text>
              <text class="field-desc">停用后该项不再出现在下拉选项中</text>
            </view>
            <switch :checked="form.status === '0'" color="#7b61ff" style="transform: scale(0.7)" @change="onSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '正常' : '停用' }}</text>
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
import { ref, reactive } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listDictData, createDictData, updateDictData, deleteDictData } from '@/api/dict'
import { trimStr, isNonNegInt } from '@/utils/validate'

const dictType = ref('')
const dictName = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showForm = ref(false)
const form = reactive({ dictCode: 0, dictLabel: '', dictValue: '', dictSort: 1, dictType: '', status: '0' })

const PALETTE = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#9c6ade', '#00b8a9', '#fa8c5f', '#5b8ff9']
function avatarColor(id) {
  return PALETTE[Number(id) % PALETTE.length]
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listDictData({ pageNum: pageNum.value, pageSize, dictType: dictType.value })
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

function openForm(d) {
  if (d) {
    Object.assign(form, { dictCode: d.dictCode, dictLabel: d.dictLabel || '', dictValue: d.dictValue || '', dictSort: d.dictSort ?? 1, dictType: d.dictType || dictType.value, status: d.status || '0' })
  } else {
    Object.assign(form, { dictCode: 0, dictLabel: '', dictValue: '', dictSort: 1, dictType: dictType.value, status: '0' })
  }
  showForm.value = true
}

function closeForm() { showForm.value = false }

function onSwitch(e) { form.status = e.detail.value ? '0' : '1' }

async function submitForm() {
  const dictLabel = trimStr(form.dictLabel)
  const dictValue = trimStr(form.dictValue)
  if (!dictLabel) {
    uni.showToast({ title: '请输入数据标签', icon: 'none' })
    return
  }
  if (dictLabel.length > 100) {
    uni.showToast({ title: '数据标签不能超过 100 个字符', icon: 'none' })
    return
  }
  if (!dictValue) {
    uni.showToast({ title: '请输入数据键值', icon: 'none' })
    return
  }
  if (/\s/.test(dictValue)) {
    uni.showToast({ title: '数据键值不能包含空格', icon: 'none' })
    return
  }
  if (!isNonNegInt(form.dictSort)) {
    uni.showToast({ title: '显示排序需为不小于 0 的整数', icon: 'none' })
    return
  }
  Object.assign(form, { dictLabel, dictValue })
  const payload = { dictLabel, dictValue, dictSort: Number(form.dictSort) || 0, dictType: form.dictType, status: form.status }
  const data = form.dictCode ? await updateDictData({ ...payload, dictCode: form.dictCode }) : await createDictData(payload)
  if (data.code === 200) {
    closeForm()
    uni.showToast({ title: form.dictCode ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
    if (form.dictCode) {
      const idx = list.value.findIndex((it) => it.dictCode === form.dictCode)
      if (idx >= 0) list.value[idx] = { ...list.value[idx], ...payload }
    } else {
      setTimeout(() => fetchList(true), 400)
    }
  } else {
    uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
  }
}

function toggleStatus(d) {
  const next = d.status === '0' ? '1' : '0'
  openForm({ ...d, status: next })
}

function onDelete(d) {
  uni.showModal({
    title: '提示',
    content: `确定删除字典数据「${d.dictLabel}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      const data = await deleteDictData(d.dictCode)
      if (data.code === 200) {
        uni.showToast({ title: '删除成功', icon: 'success' })
        list.value = list.value.filter((it) => it.dictCode !== d.dictCode)
        total.value = Math.max(0, total.value - 1)
        if (!list.value.length) fetchList(true)
      } else {
        uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
      }
    }
  })
}

onLoad((q) => {
  dictType.value = q.dictType || ''
  dictName.value = q.dictName || ''
  uni.setNavigationBarTitle({ title: dictName.value ? `字典数据·${dictName.value}` : '字典数据' })
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
.head-bar { display: flex; align-items: center; justify-content: space-between; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.head-info { display: flex; align-items: center; gap: 14rpx; }
.head-name { font-size: 28rpx; font-weight: 600; color: #1f2d3d; }
.head-type { font-size: 22rpx; color: #8a94a6; background: #f5f7fa; border-radius: 18rpx; padding: 4rpx 14rpx; }
.add-btn { padding: 0 24rpx; height: 64rpx; line-height: 64rpx; background: #67c23a; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.data-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.data-item { display: flex; align-items: center; gap: 16rpx; padding: 26rpx 0; }
.data-item + .data-item { border-top: 1rpx solid #f0f3f7; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.data-main { flex: 1; min-width: 0; }
.data-line { display: flex; align-items: center; gap: 12rpx; }
.data-label { font-size: 28rpx; color: #1f2d3d; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.data-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.data-value { color: #8a94a6; }
.data-sort::before { content: '·'; margin: 0 8rpx; }
.op { width: 60rpx; height: 60rpx; line-height: 60rpx; text-align: center; font-size: 28rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); border-radius: 50%; flex-shrink: 0; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 86vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; margin-bottom: 28rpx; }
.form-item { display: flex; align-items: center; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.label { width: 140rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.input { flex: 1; height: 64rpx; font-size: 26rpx; }
.switch-text { font-size: 24rpx; color: #8a94a6; }
.form-actions { display: flex; gap: 20rpx; margin-top: 40rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 弹层精修（新版） ===== */
.sheet { padding: 0; max-height: 90vh; border-radius: 28rpx 28rpx 0 0; overflow-y: auto; background: #fff; }
.sheet-head { position: sticky; top: 0; z-index: 5; display: flex; align-items: center; gap: 20rpx; padding: 30rpx 32rpx 24rpx; background: #fff; border-bottom: 1rpx solid #f2f5fb; }
.sheet-avatar { width: 88rpx; height: 88rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 700; color: #fff; flex-shrink: 0; background: linear-gradient(135deg, #a99bff, #7b61ff); box-shadow: 0 10rpx 24rpx rgba(123, 97, 255, 0.26); }
.sheet-head-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8rpx; }
.sheet-head .sheet-title { font-size: 32rpx; font-weight: 700; color: #1c2738; text-align: left; margin: 0; }
.sheet-sub { font-size: 22rpx; color: #9aa6b8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sheet-close { width: 56rpx; height: 56rpx; line-height: 56rpx; text-align: center; font-size: 32rpx; color: #9aa6b8; background: #f3f6fb; border-radius: 50%; flex-shrink: 0; }
.group { padding: 28rpx 32rpx 0; }
.group-head { display: flex; align-items: center; gap: 12rpx; margin-bottom: 20rpx; }
.group-bar { width: 6rpx; height: 26rpx; border-radius: 4rpx; background: linear-gradient(180deg, #a99bff, #7b61ff); }
.group-title { font-size: 26rpx; font-weight: 600; color: #1c2738; }
.field { display: flex; flex-direction: column; gap: 12rpx; margin-bottom: 24rpx; }
.field-label { font-size: 24rpx; color: #6b7a90; }
.field-input { width: 100%; height: 84rpx; padding: 0 24rpx; box-sizing: border-box; background: #f5f7fb; border-radius: 14rpx; font-size: 27rpx; color: #1c2738; }
.field-select { display: flex; align-items: center; justify-content: space-between; height: 84rpx; padding: 0 24rpx; background: #f5f7fb; border-radius: 14rpx; }
.field-select-text { flex: 1; min-width: 0; font-size: 27rpx; color: #1c2738; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.field-select-arrow { font-size: 30rpx; color: #b9c3d3; }
.field-card { display: flex; align-items: center; gap: 16rpx; padding: 24rpx; margin-bottom: 24rpx; background: #f5f7fb; border-radius: 16rpx; }
.field-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.field-label-inline { font-size: 26rpx; font-weight: 500; color: #1c2738; }
.field-desc { font-size: 21rpx; color: #9aa6b8; }
.status-pill { font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 20rpx; color: #3fae7a; background: rgba(63, 174, 122, 0.12); flex-shrink: 0; }
.status-pill.off { color: #f56c6c; background: rgba(245, 108, 108, 0.12); }
.form-actions { position: sticky; bottom: 0; z-index: 5; display: flex; gap: 20rpx; padding: 20rpx 32rpx 40rpx; background: #fff; border-top: 1rpx solid #f2f5fb; }
.btn { flex: 1; height: 92rpx; line-height: 92rpx; text-align: center; color: #fff; font-size: 28rpx; font-weight: 600; border-radius: 18rpx; background: linear-gradient(135deg, #7b61ff, #5b45d6); box-shadow: 0 12rpx 26rpx rgba(91, 69, 214, 0.24); }
.btn.plain { background: #f3f6fb; color: #6b7a90; box-shadow: none; }
</style>
