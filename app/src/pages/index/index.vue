<template>
  <view class="page">
    <!-- 用户欢迎卡 -->
    <view class="hero fade-up">
      <view class="hero-orb hero-orb-a"></view>
      <view class="hero-orb hero-orb-b"></view>
      <view class="hero-top">
        <view class="avatar-ring">
          <image v-if="avatarUrl" :src="avatarUrl" class="hero-avatar-img" mode="aspectFill" />
          <view v-else class="hero-avatar">{{ avatarChar }}</view>
        </view>
        <view class="hero-info">
          <view class="hero-greet">{{ greeting }}，{{ displayName }}</view>
          <view class="hero-sub">{{ today }} · 工作台已就绪</view>
        </view>
        <view class="hero-role">{{ roleTag }}</view>
      </view>
      <view class="hero-stats">
        <view class="hero-stat">
          <text class="hero-num">{{ menuCount }}</text>
          <text class="hero-label">功能菜单</text>
        </view>
        <view class="hero-div"></view>
        <view class="hero-stat">
          <text class="hero-num">{{ permCount }}</text>
          <text class="hero-label">权限项</text>
        </view>
        <view class="hero-div"></view>
        <view class="hero-stat">
          <text class="hero-num">{{ userStore.roles.length || 0 }}</text>
          <text class="hero-label">角色</text>
        </view>
      </view>
    </view>

    <!-- 系统概览 -->
    <view class="section fade-up">
      <view class="section-head">
        <view class="section-bar"></view>
        <text class="section-title">系统概览</text>
        <text class="section-note">实时数据</text>
      </view>
      <view class="overview-card">
        <view v-for="s in overviewItems" :key="s.label" class="overview-cell">
          <text class="overview-num" :style="{ color: s.color }">{{ s.value }}</text>
          <text class="overview-label">{{ s.label }}</text>
        </view>
      </view>
    </view>

    <!-- 功能菜单宫格（动态菜单，按权限下发） -->
    <view class="section fade-up">
      <view class="section-head">
        <view class="section-bar"></view>
        <text class="section-title">功能菜单</text>
        <text class="section-more" @tap="goAllMenus">全部 ›</text>
      </view>
      <view v-if="topMenus.length" class="menu-grid">
        <view
          v-for="m in topMenus"
          :key="m.menuId"
          class="menu-cell"
          @tap="openMenu(m)"
          @longpress="onMenuLongPress(m)"
        >
          <view class="menu-icon" :style="{ background: iconGradient(m.menuId) }">{{ iconChar(m) }}</view>
          <text class="menu-name">{{ m.menuName }}</text>
          <text class="menu-sub">{{ m.menuType === 'M' ? '目录' : '页面' }}</text>
        </view>
      </view>
      <view v-else class="empty-card">暂无菜单权限</view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { openPageMenu } from '@/utils/menuRouter'
import { getBaseUrl } from '@/utils/config'
import { dashStats } from '@/api/dashboard'

const userStore = useUserStore()
const stats = ref({})

async function loadStats() {
  if (!userStore.token) return
  try {
    const data = await dashStats()
    stats.value = data.data || {}
  } catch (e) {}
}

const avatarUrl = computed(() => {
  const avatar = userStore.userInfo?.avatar
  if (!avatar) return ''
  return /^http(s)?:\/\//.test(avatar) ? avatar : getBaseUrl() + avatar
})

const avatarChar = computed(() => {
  const u = userStore.userInfo || {}
  const name = u.nickName || u.userName || 'U'
  return name.charAt(0).toUpperCase()
})
const displayName = computed(() => {
  const u = userStore.userInfo || {}
  return u.nickName || u.userName || '用户'
})
const today = computed(() => {
  const d = new Date()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return mm + '-' + dd
})

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 11) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const roleTag = computed(() => {
  const roles = userStore.roles || []
  if (!roles.length) return '普通用户'
  return roles[0] === 'admin' ? '超级管理员' : roles[0]
})

// 顶层菜单：目录(M) 与 页面(C) 均展示（按钮 F 不展示）
const topMenus = computed(() => (userStore.menus || []).filter((m) => m.menuType !== 'F'))
const menuCount = computed(() => topMenus.value.length)
const permCount = computed(() => {
  const p = userStore.permissions || []
  return p.length === 1 && p[0] === '*:*:*' ? '全部' : p.length
})

const overviewItems = computed(() => [
  { label: '用户', value: stats.value.userCount || 0, color: '#3a6fe0' },
  { label: '角色', value: stats.value.roleCount || 0, color: '#7b61ff' },
  { label: '部门', value: stats.value.deptCount || 0, color: '#00a88f' },
  { label: '菜单', value: stats.value.menuCount || 0, color: '#e6a23c' },
  { label: '公告', value: stats.value.noticeCount || 0, color: '#f56c6c' },
  { label: '今日登录', value: stats.value.todayLogin || 0, color: '#3a6fe0' },
  { label: '在线用户', value: stats.value.onlineCount || 0, color: '#00a88f' }
])

const GRADIENTS = [
  'linear-gradient(135deg, #6aa4ff, #3a6fe0)',
  'linear-gradient(135deg, #7fd3a8, #3fae7a)',
  'linear-gradient(135deg, #ffc46b, #e6a23c)',
  'linear-gradient(135deg, #ff9a9a, #f56c6c)',
  'linear-gradient(135deg, #a99bff, #7b61ff)',
  'linear-gradient(135deg, #5fd3cd, #00a88f)',
  'linear-gradient(135deg, #ffb08a, #f0805a)',
  'linear-gradient(135deg, #9fb2c8, #7b8ca3)'
]
function iconGradient(id) {
  return GRADIENTS[Number(id) % GRADIENTS.length]
}
function iconChar(m) {
  return (m.menuName || '项').charAt(0)
}

function goAllMenus() {
  uni.navigateTo({ url: '/pages/menu/index?title=' + encodeURIComponent('全部菜单') })
}

function openMenu(m) {
  if (m.menuType === 'M') {
    // 目录 → 进入菜单列表下钻
    uni.navigateTo({ url: '/pages/menu/index?parentId=' + m.menuId + '&title=' + encodeURIComponent(m.menuName) })
  } else {
    // 页面菜单 → 已落地的业务页直接跳转，未落地保持提示
    if (openPageMenu(m)) return
    uni.showModal({
      title: m.menuName,
      content: '已授权页面：' + (m.path || '—') + '\n该模块功能请在 Web 端操作。',
      showCancel: false,
      confirmText: '知道了'
    })
  }
}

// 长按宫格菜单：查看菜单详情
function onMenuLongPress(m) {
  uni.showModal({
    title: m.menuName,
    content:
      'ID：' + m.menuId +
      '\n类型：' + (m.menuType === 'M' ? '目录' : '页面') +
      '\n路径：' + (m.path || '—') +
      '\n权限：' + (m.perms || '—'),
    showCancel: false,
    confirmText: '知道了'
  })
}

onShow(() => {
  // 登录态恢复/退出后刷新（tab 页每次显示都校验数据）
  if (userStore.token && !userStore.userInfo) {
    userStore.init().catch(() => {})
  }
  loadStats()
})

// 下拉刷新：重新拉取用户信息/菜单
onPullDownRefresh(async () => {
  try {
    if (userStore.token) {
      await userStore.init()
      await loadStats()
    }
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
    uni.showToast({ title: '刷新失败', icon: 'none' })
  } finally {
    uni.stopPullDownRefresh()
  }
})
</script>

<style scoped>
.page {
  padding: 28rpx 28rpx 60rpx;
}

/* ===== 欢迎卡 ===== */
.hero {
  position: relative;
  overflow: hidden;
  padding: 40rpx 32rpx 0;
  border-radius: 28rpx;
  background: linear-gradient(135deg, #4a7df0 0%, #3457c9 60%, #2f49a8 100%);
  box-shadow: 0 18rpx 40rpx rgba(52, 87, 201, 0.28);
}

.hero-orb {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
}

.hero-orb-a {
  width: 240rpx;
  height: 240rpx;
  top: -110rpx;
  right: -60rpx;
}

.hero-orb-b {
  width: 140rpx;
  height: 140rpx;
  bottom: 40rpx;
  left: -50rpx;
  background: rgba(255, 255, 255, 0.08);
}

.hero-top {
  position: relative;
  display: flex;
  align-items: center;
  gap: 22rpx;
}

.avatar-ring {
  width: 104rpx;
  height: 104rpx;
  padding: 6rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.35);
  flex-shrink: 0;
}

.hero-avatar,
.hero-avatar-img {
  width: 92rpx;
  height: 92rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  color: #3457c9;
  font-size: 40rpx;
  font-weight: 700;
}

.hero-avatar-img {
  background: #e9efff;
}

.hero-info {
  flex: 1;
  min-width: 0;
}

.hero-greet {
  font-size: 34rpx;
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hero-sub {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.75);
}

.hero-role {
  flex-shrink: 0;
  max-width: 200rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  border: 1rpx solid rgba(255, 255, 255, 0.32);
  font-size: 20rpx;
  color: #fff;
}

.hero-stats {
  position: relative;
  display: flex;
  align-items: center;
  margin-top: 34rpx;
  padding: 24rpx 0 28rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.16);
}

.hero-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}

.hero-num {
  font-size: 40rpx;
  font-weight: 700;
  color: #fff;
  line-height: 1.1;
}

.hero-label {
  font-size: 20rpx;
  color: rgba(255, 255, 255, 0.7);
}

.hero-div {
  width: 1rpx;
  height: 36rpx;
  background: rgba(255, 255, 255, 0.2);
}

/* ===== 板块标题 ===== */
.section {
  margin-top: 36rpx;
}

.section-head {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-bottom: 20rpx;
  padding-left: 4rpx;
}

.section-bar {
  width: 8rpx;
  height: 28rpx;
  border-radius: 999rpx;
  background: linear-gradient(180deg, #6aa4ff, #3a6fe0);
}

.section-title {
  flex: 1;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2d3d;
}

.section-note {
  font-size: 22rpx;
  color: #9aa6b8;
}

.section-more {
  font-size: 24rpx;
  color: #3a6fe0;
}

/* ===== 系统概览 ===== */
.overview-card {
  display: flex;
  flex-wrap: wrap;
  background: #fff;
  border-radius: 24rpx;
  padding: 8rpx 0;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.overview-cell {
  width: 25%;
  padding: 24rpx 8rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
}

.overview-num {
  font-size: 38rpx;
  font-weight: 700;
  line-height: 1.1;
}

.overview-label {
  font-size: 21rpx;
  color: #8a94a6;
}

/* ===== 功能菜单宫格 ===== */
.menu-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18rpx;
}

.menu-cell {
  padding: 28rpx 12rpx 24rpx;
  border-radius: 24rpx;
  background: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.menu-icon {
  width: 92rpx;
  height: 92rpx;
  border-radius: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 38rpx;
  font-weight: 600;
  color: #fff;
  box-shadow: 0 10rpx 20rpx rgba(52, 87, 201, 0.18);
}

.menu-name {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #1f2d3d;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.menu-sub {
  margin-top: 6rpx;
  padding: 2rpx 14rpx;
  border-radius: 999rpx;
  background: #f2f5fb;
  font-size: 20rpx;
  color: #9aa6b8;
}

.empty-card {
  padding: 70rpx 0;
  text-align: center;
  color: #9aa6b8;
  font-size: 26rpx;
  background: #fff;
  border-radius: 24rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}
</style>
