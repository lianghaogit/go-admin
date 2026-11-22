<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入接口名称搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view v-if="$hasPerm('system:api:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>

    <view v-if="list.length" class="api-list">
      <view v-for="a in list" :key="a.apiId" class="api-item">
        <view class="avatar" :style="{ background: avatarColor(a.apiId) }">{{ (a.apiName || 'A').charAt(0) }}</view>
        <view class="api-main" @tap="$hasPerm('system:api:edit') && openForm(a)">
          <view class="api-line">
            <text class="api-name">{{ a.apiName }}</text>
            <text class="method" :class="a.method.toLowerCase()">{{ a.method }}</text>
            <text
              class="status"
              :class="a.status === '0' ? 'on' : 'off'"
              @tap.stop="$hasPerm('system:api:changeStatus') && toggleStatus(a)"
            >{{ a.status === '0' ? '启用' : '停用' }}</text>
          </view>
          <view class="api-line sub">
            <text class="api-path">{{ a.path }}</text>
          </view>
          <view v-if="a.perms" class="api-line sub">
            <text class="api-perms">{{ a.perms }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('system:api:remove')" class="op danger" @tap="onDelete(a)">✕</view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无接口数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">{{ form.apiId ? '编辑接口' : '新增接口' }}</view>

        <view class="form-item">
          <text class="label">接口名称</text>
          <input v-model="form.apiName" class="input" placeholder="必填" placeholder-class="ph" />
        </view>
        <view class="form-item">
          <text class="label">接口路径</text>
          <input v-model="form.path" class="input" placeholder="如 /api/system/user" placeholder-class="ph" />
        </view>
        <view class="form-item">
          <text class="label">请求方法</text>
          <picker :range="methodOptions" :value="methodIndex" @change="onMethodChange">
            <view class="picker-value">{{ form.method || '请选择' }}</view>
          </picker>
        </view>
        <view class="form-item">
          <text class="label">权限标识</text>
          <input v-model="form.perms" class="input" placeholder="如 system:user:list" placeholder-class="ph" />
        </view>
        <view class="form-item">
          <text class="label">状态</text>
          <switch :checked="form.status === '0'" color="#409eff" style="transform: scale(0.75)" @change="onSwitch" />
          <text class="switch-text">{{ form.status === '0' ? '启用' : '停用' }}</text>
        </view>
        <view class="form-item">
          <text class="label">备注</text>
          <input v-model="form.remark" class="input" placeholder="备注" placeholder-class="ph" />
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
import { listApi, createApi, updateApi, deleteApi, changeApiStatus } from '@/api/api'
import { useUserStore } from '@/store/user'
import { trimStr } from '@/utils/validate'

const userStore = useUserStore()

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

const showForm = ref(false)
const form = reactive({ apiId: 0, apiName: '', path: '', method: 'GET', perms: '', status: '0', remark: '' })
const methodOptions = ['GET', 'POST', 'PUT', 'DELETE']
const methodIndex = computed(() => Math.max(0, methodOptions.indexOf(form.method)))

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
    const data = await listApi({ pageNum: pageNum.value, pageSize, apiName: keyword.value || undefined })
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

// ---- 表单 ----
function openForm(a) {
  if (!userStore.hasPerm(a ? 'system:api:edit' : 'system:api:add')) {
    uni.showToast({ title: '无权限操作接口规则', icon: 'none' })
    return
  }
  if (a) {
    Object.assign(form, {
      apiId: a.apiId,
      apiName: a.apiName || '',
      path: a.path || '',
      method: a.method || 'GET',
      perms: a.perms || '',
      status: a.status || '0',
      remark: a.remark || ''
    })
  } else {
    Object.assign(form, { apiId: 0, apiName: '', path: '', method: 'GET', perms: '', status: '0', remark: '' })
  }
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

function onMethodChange(e) {
  form.method = methodOptions[Number(e.detail.value)]
}

function onSwitch(e) {
  form.status = e.detail.value ? '0' : '1'
}

async function submitForm() {
  const apiName = trimStr(form.apiName)
  const path = trimStr(form.path)
  const perms = trimStr(form.perms)
  if (!apiName) {
    uni.showToast({ title: '请输入接口名称', icon: 'none' })
    return
  }
  if (apiName.length > 50) {
    uni.showToast({ title: '接口名称不能超过 50 个字符', icon: 'none' })
    return
  }
  if (!path) {
    uni.showToast({ title: '请输入接口路径', icon: 'none' })
    return
  }
  if (!/^\//.test(path)) {
    uni.showToast({ title: '接口路径需以 / 开头', icon: 'none' })
    return
  }
  if (/\s/.test(path)) {
    uni.showToast({ title: '接口路径不能包含空格', icon: 'none' })
    return
  }
  if (!form.method) {
    uni.showToast({ title: '请选择请求方法', icon: 'none' })
    return
  }
  Object.assign(form, { apiName, path, perms })
  const payload = {
    apiName,
    path,
    method: form.method,
    perms,
    status: form.status,
    remark: trimStr(form.remark)
  }
  try {
    const data = form.apiId ? await updateApi({ ...payload, apiId: form.apiId }) : await createApi(payload)
    if (data.code === 200) {
      closeForm()
      uni.showToast({ title: form.apiId ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
      if (form.apiId) {
        const idx = list.value.findIndex((it) => it.apiId === form.apiId)
        if (idx >= 0) list.value[idx] = { ...list.value[idx], ...payload }
      } else {
        setTimeout(() => fetchList(true), 400)
      }
    } else {
      uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
    }
  } catch (e) {
    // request 层已统一提示
  }
}

// ---- 启停 / 删除 ----
function toggleStatus(a) {
  if (a.apiId === 1) return
  const next = a.status === '0' ? '1' : '0'
  uni.showModal({
    title: '提示',
    content: `确定${next === '0' ? '启用' : '停用'}接口「${a.apiName}」吗？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await changeApiStatus(a.apiId, next)
        if (data.code === 200) {
          uni.showToast({ title: '操作成功', icon: 'success' })
          a.status = next
        } else {
          uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
        }
      } catch (e) {
        // request 层已统一提示
      }
    }
  })
}

function onDelete(a) {
  uni.showModal({
    title: '提示',
    content: `确定删除接口「${a.apiName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteApi(a.apiId)
        if (data.code === 200) {
          uni.showToast({ title: '删除成功', icon: 'success' })
          list.value = list.value.filter((it) => it.apiId !== a.apiId)
          total.value = Math.max(0, total.value - 1)
          if (!list.value.length) fetchList(true)
        } else {
          uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
        }
      } catch (e) {
        // request 层已统一提示
      }
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '接口权限' })
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

onReachBottom(() => {
  fetchList(false)
})
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 12rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; min-width: 0; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.add-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #67c23a; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.api-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.api-item { display: flex; align-items: center; gap: 16rpx; padding: 26rpx 0; }
.api-item + .api-item { border-top: 1rpx solid #f0f3f7; }
.avatar { width: 76rpx; height: 76rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: 600; color: #fff; flex-shrink: 0; }
.api-main { flex: 1; min-width: 0; }
.api-line { display: flex; align-items: center; gap: 12rpx; }
.api-name { font-size: 28rpx; color: #1f2d3d; }
.method { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; font-weight: 600; }
.method.get { color: #409eff; background: rgba(64, 158, 255, 0.1); }
.method.post { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.method.put { color: #e6a23c; background: rgba(230, 162, 60, 0.1); }
.method.delete { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.api-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #9aa6b8; }
.api-path { color: #8a94a6; word-break: break-all; }
.api-perms { color: #5a6678; }
.op { width: 60rpx; height: 60rpx; line-height: 60rpx; text-align: center; font-size: 28rpx; border-radius: 50%; flex-shrink: 0; }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
/* 弹层 */
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 86vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; margin-bottom: 28rpx; }
.form-item { display: flex; align-items: center; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.label { width: 140rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.input { flex: 1; height: 64rpx; font-size: 26rpx; }
.picker-value { flex: 1; height: 64rpx; line-height: 64rpx; font-size: 26rpx; color: #1f2d3d; }
.switch-text { font-size: 24rpx; color: #8a94a6; }
.form-actions { display: flex; gap: 20rpx; margin-top: 40rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
