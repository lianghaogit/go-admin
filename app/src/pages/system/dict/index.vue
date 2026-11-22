<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入字典类型搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('system:dict:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>

    <view v-if="list.length" class="dict-list">
      <view v-for="d in list" :key="d.dictId" class="dict-item">
        <view class="avatar" :style="{ background: avatarColor(d.dictId) }">{{ (d.dictName || 'D').charAt(0) }}</view>
        <view class="dict-main" @tap="goData(d)">
          <view class="dict-line">
            <text class="dict-name">{{ d.dictName }}</text>
            <text class="status" :class="d.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(d)">{{ d.status === '0' ? '正常' : '停用' }}</text>
          </view>
          <view class="dict-line sub">
            <text class="dict-type">{{ d.dictType }}</text>
            <text v-if="d.remark" class="dict-remark">{{ d.remark }}</text>
          </view>
          <view class="dict-line sub tip">点击查看字典数据 ›</view>
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
          <view class="sheet-avatar">{{ (form.dictName || 'D').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.dictId ? '编辑字典类型' : '新增字典类型' }}</text>
            <text class="sheet-sub">{{ form.dictId ? '调整字典名称与类型标识' : '为一组枚举值定义类型' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">字典名称</text>
            <input v-model="form.dictName" class="field-input" placeholder="必填，如 用户性别" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">字典类型</text>
            <input v-model="form.dictType" class="field-input" placeholder="必填，如 sys_user_sex" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">备注</text>
            <input v-model="form.remark" class="field-input" placeholder="选填，描述该字典用途" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">字典状态</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">字典状态</text>
              <text class="field-desc">停用后该字典数据不再被业务引用</text>
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
import { listDictType, createDictType, updateDictType, deleteDictType } from '@/api/dict'
import { trimStr, DICT_TYPE_RE } from '@/utils/validate'

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showForm = ref(false)
const form = reactive({ dictId: 0, dictName: '', dictType: '', status: '0', remark: '' })

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
    const data = await listDictType({ pageNum: pageNum.value, pageSize, dictType: keyword.value || undefined })
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

// 进入字典数据页
function goData(d) {
  uni.navigateTo({
    url: `/pages/system/dictdata/index?dictType=${encodeURIComponent(d.dictType)}&dictName=${encodeURIComponent(d.dictName || '')}`
  })
}

function openForm(d) {
  if (d) {
    Object.assign(form, { dictId: d.dictId, dictName: d.dictName || '', dictType: d.dictType || '', status: d.status || '0', remark: d.remark || '' })
  } else {
    Object.assign(form, { dictId: 0, dictName: '', dictType: '', status: '0', remark: '' })
  }
  showForm.value = true
}

function closeForm() { showForm.value = false }

function onSwitch(e) { form.status = e.detail.value ? '0' : '1' }

async function submitForm() {
  const dictName = trimStr(form.dictName)
  const dictType = trimStr(form.dictType)
  if (!dictName) {
    uni.showToast({ title: '请输入字典名称', icon: 'none' })
    return
  }
  if (dictName.length > 100) {
    uni.showToast({ title: '字典名称不能超过 100 个字符', icon: 'none' })
    return
  }
  if (!dictType) {
    uni.showToast({ title: '请输入字典类型', icon: 'none' })
    return
  }
  if (!DICT_TYPE_RE.test(dictType) || dictType.length > 100) {
    uni.showToast({ title: '字典类型需以字母开头，仅含字母、数字和下划线', icon: 'none' })
    return
  }
  Object.assign(form, { dictName, dictType })
  const payload = { dictName, dictType, status: form.status, remark: trimStr(form.remark) }
  const data = form.dictId ? await updateDictType({ ...payload, dictId: form.dictId }) : await createDictType(payload)
  if (data.code === 200) {
    closeForm()
    uni.showToast({ title: form.dictId ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
    if (form.dictId) {
      const idx = list.value.findIndex((it) => it.dictId === form.dictId)
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
    content: `确定删除字典类型「${d.dictName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      const data = await deleteDictType(d.dictId)
      if (data.code === 200) {
        uni.showToast({ title: '删除成功', icon: 'success' })
        list.value = list.value.filter((it) => it.dictId !== d.dictId)
        total.value = Math.max(0, total.value - 1)
        if (!list.value.length) fetchList(true)
      } else {
        uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
      }
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '字典管理' })
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
.dict-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.dict-item { display: flex; align-items: center; gap: 16rpx; padding: 26rpx 0; }
.dict-item + .dict-item { border-top: 1rpx solid #f0f3f7; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.dict-main { flex: 1; min-width: 0; }
.dict-line { display: flex; align-items: center; gap: 12rpx; }
.dict-name { font-size: 28rpx; color: #1f2d3d; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.dict-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.dict-type { color: #8a94a6; }
.dict-remark { color: #9aa6b8; }
.dict-remark::before { content: '·'; margin: 0 8rpx; }
.tip { color: #409eff; font-size: 20rpx; }
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
.sheet-sub { font-size: 22rpx; color: #9aa6b8; }
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
