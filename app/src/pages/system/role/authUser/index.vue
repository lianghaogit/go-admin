<template>
  <view class="page">
    <view class="tabs">
      <text class="tab" :class="{ active: tab === 0 }" @tap="switchTab(0)">已分配用户</text>
      <text class="tab" :class="{ active: tab === 1 }" @tap="switchTab(1)">未分配用户</text>
    </view>

    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入用户名搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>

    <view class="select-bar">
      <text class="select-all" @tap="toggleAll">{{ allSelected ? '取消全选' : '全选' }}</text>
      <text class="selected-count">已选 {{ selected.length }}</text>
    </view>

    <view v-if="list.length" class="user-list">
      <view v-for="u in list" :key="u.userId" class="user-item" @tap="toggleSelect(u.userId)">
        <view class="checkbox" :class="{ on: selected.includes(u.userId) }">{{ selected.includes(u.userId) ? '✓' : '' }}</view>
        <view class="avatar">{{ (u.nickName || u.userName || 'U').charAt(0) }}</view>
        <view class="user-main">
          <view class="user-line">
            <text class="user-name">{{ u.nickName || u.userName }}</text>
            <text class="user-uname">@{{ u.userName }}</text>
          </view>
          <view class="user-line sub">
            <text v-if="u.dept && u.dept.deptName" class="user-meta">{{ u.dept.deptName }}</text>
            <text v-if="u.phonenumber" class="user-meta">{{ u.phonenumber }}</text>
            <text v-if="u.email" class="user-meta">{{ u.email }}</text>
          </view>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">{{ tab === 0 ? '该角色暂无已分配用户' : '没有可分配的用户' }}</view>
    <view v-else class="empty">加载中…</view>

    <view class="footer">
      <view class="btn" :class="{ danger: tab === 0 }" @tap="submit">
        {{ tab === 0 ? '取消分配' : '分配所选用户' }}（{{ selected.length }}）
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listRoleAllocatedUsers, listRoleUnallocatedUsers, assignRoleUsers, cancelRoleUsers } from '@/api/role'

const roleId = ref(0)
const roleName = ref('')
const tab = ref(0)
const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const selected = ref([])

const allSelected = computed(() => list.value.length > 0 && selected.value.length === list.value.length)

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const query = { pageNum: pageNum.value, pageSize, username: keyword.value || undefined }
    const data = tab.value === 0 ? await listRoleAllocatedUsers(roleId.value, query) : await listRoleUnallocatedUsers(roleId.value, query)
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

function switchTab(t) {
  if (tab.value === t) return
  tab.value = t
  selected.value = []
  keyword.value = ''
  fetchList(true)
}

function onSearch() { selected.value = []; fetchList(true) }
function onReset() { keyword.value = ''; selected.value = []; fetchList(true) }

function toggleSelect(id) {
  const idx = selected.value.indexOf(id)
  if (idx >= 0) selected.value.splice(idx, 1)
  else selected.value.push(id)
}

function toggleAll() {
  selected.value = allSelected.value ? [] : list.value.map((u) => u.userId)
}

function submit() {
  if (!selected.value.length) {
    uni.showToast({ title: '请先选择用户', icon: 'none' })
    return
  }
  const isCancel = tab.value === 0
  uni.showModal({
    title: '提示',
    content: isCancel ? `确定取消选中 ${selected.value.length} 个用户的角色吗？` : `确定将选中 ${selected.value.length} 个用户分配给该角色吗？`,
    confirmText: isCancel ? '取消分配' : '分配',
    confirmColor: isCancel ? '#f56c6c' : '#409eff',
    success: async (res) => {
      if (!res.confirm) return
      try {
        if (isCancel) await cancelRoleUsers(roleId.value, selected.value)
        else await assignRoleUsers(roleId.value, selected.value)
        uni.showToast({ title: isCancel ? '已取消分配' : '分配成功', icon: 'success' })
        selected.value = []
        fetchList(true)
      } catch (e) {}
    }
  })
}

onLoad((options) => {
  roleId.value = Number((options && options.roleId) || 0)
  roleName.value = decodeURIComponent((options && options.roleName) || '')
  uni.setNavigationBarTitle({ title: roleName.value ? `${roleName.value} · 分配用户` : '分配用户' })
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
.page { padding: 28rpx 28rpx 180rpx; }
.tabs { display: flex; gap: 16rpx; }
.tab { flex: 1; height: 76rpx; line-height: 76rpx; text-align: center; font-size: 26rpx; color: #5a6678; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.tab.active { color: #fff; background: #409eff; }
.search-bar { display: flex; align-items: center; gap: 12rpx; margin-top: 20rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; min-width: 0; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.select-bar { display: flex; align-items: center; justify-content: space-between; padding: 18rpx 6rpx 0; }
.select-all { font-size: 24rpx; color: #409eff; }
.selected-count { font-size: 24rpx; color: #8a94a6; }
.user-list { margin-top: 16rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.user-item { display: flex; align-items: center; gap: 18rpx; padding: 24rpx 0; }
.user-item + .user-item { border-top: 1rpx solid #f0f3f7; }
.checkbox { width: 36rpx; height: 36rpx; line-height: 36rpx; text-align: center; border-radius: 50%; border: 2rpx solid #c3ccd8; color: #fff; font-size: 22rpx; flex-shrink: 0; }
.checkbox.on { background: #409eff; border-color: #409eff; }
.avatar { width: 68rpx; height: 68rpx; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 600; color: #fff; background: #409eff; flex-shrink: 0; }
.user-main { flex: 1; min-width: 0; }
.user-line { display: flex; align-items: baseline; gap: 10rpx; }
.user-name { font-size: 28rpx; color: #1f2d3d; }
.user-uname { font-size: 22rpx; color: #9aa6b8; }
.user-line.sub { margin-top: 6rpx; font-size: 22rpx; color: #8a94a6; }
.user-meta + .user-meta::before { content: '·'; margin: 0 8rpx; }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 16rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.footer { position: fixed; left: 0; right: 0; bottom: 0; padding: 20rpx 28rpx calc(20rpx + env(safe-area-inset-bottom)); background: #fff; box-shadow: 0 -6rpx 20rpx rgba(31, 45, 61, 0.06); }
.btn { height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.danger { background: #f56c6c; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
