<template>
  <view class="page">
    <!-- 面包屑 -->
    <view class="crumb fade-up">
      <text class="crumb-link" @tap="resetToTop">全部菜单</text>
      <text v-for="(c, i) in crumbs" :key="i" class="crumb-sep"> / </text>
      <text v-for="(c, i) in crumbs" :key="'c' + i" class="crumb-link" @tap="gotoCrumb(i)">{{ c.name }}</text>
    </view>

    <!-- 当前层级菜单 -->
    <view v-if="currentMenus.length" class="menu-list">
      <view
        v-for="m in currentMenus"
        :key="m.menuId"
        class="menu-item"
        @tap="onItem(m)"
        @longpress="onLongPress(m)"
      >
        <view class="menu-item-icon" :style="{ background: iconGradient(m.menuId) }">{{ iconChar(m) }}</view>
        <view class="menu-item-main">
          <text class="menu-item-name">{{ m.menuName }}</text>
          <text class="menu-item-path">{{ m.path || '—' }}</text>
        </view>
        <text class="menu-item-type">{{ m.menuType === 'M' ? '目录' : '页面' }}</text>
        <text class="menu-item-arrow">{{ m.menuType === 'M' ? '›' : '' }}</text>
      </view>
    </view>
    <view v-else class="empty">该目录下暂无菜单</view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { openPageMenu } from '@/utils/menuRouter'

const userStore = useUserStore()

// 当前展示节点
const currentNodes = ref([])
// 面包屑（目录链）
const crumbs = ref([])
// 当前下钻入口（供下拉刷新重载）
const entry = ref({ parentId: '', title: '' })

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

// 在树中按 id 找节点
function findNode(nodes, id) {
  for (const n of nodes || []) {
    if (String(n.menuId) === String(id)) return n
    const hit = findNode(n.children, id)
    if (hit) return hit
  }
  return null
}

function load(parentId, title) {
  if (!parentId) {
    currentNodes.value = (userStore.menus || []).filter((m) => m.menuType !== 'F')
    crumbs.value = []
  } else {
    const node = findNode(userStore.menus, parentId)
    if (!node) {
      currentNodes.value = []
      return
    }
    currentNodes.value = (node.children || []).filter((m) => m.menuType !== 'F')
  }
  if (title) {
    uni.setNavigationBarTitle({ title: decodeURIComponent(title) })
  }
}

const currentMenus = computed(() => currentNodes.value)

function onItem(m) {
  if (m.menuType === 'M') {
    // 下钻
    crumbs.value.push({ id: m.menuId, name: m.menuName })
    currentNodes.value = (m.children || []).filter((x) => x.menuType !== 'F')
    uni.setNavigationBarTitle({ title: m.menuName })
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

function resetToTop() {
  crumbs.value = []
  currentNodes.value = (userStore.menus || []).filter((m) => m.menuType !== 'F')
  uni.setNavigationBarTitle({ title: '功能菜单' })
}

function gotoCrumb(i) {
  const target = crumbs.value[i]
  crumbs.value = crumbs.value.slice(0, i + 1)
  const node = findNode(userStore.menus, target.id)
  currentNodes.value = (node?.children || []).filter((m) => m.menuType !== 'F')
  uni.setNavigationBarTitle({ title: target.name })
}

onLoad((options) => {
  entry.value = { parentId: options.parentId || '', title: options.title || '' }
  load(entry.value.parentId, entry.value.title)
})

// 下拉刷新：重新拉取菜单后重载当前层级
onPullDownRefresh(async () => {
  try {
    await userStore.init()
    load(entry.value.parentId, entry.value.title)
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
    uni.showToast({ title: '刷新失败', icon: 'none' })
  } finally {
    uni.stopPullDownRefresh()
  }
})

// 长按菜单项：查看菜单详情（id/path/权限标识/类型）
function onLongPress(m) {
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
</script>

<style scoped>
.page {
  padding: 28rpx;
}

.crumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  padding: 20rpx 8rpx 24rpx;
  font-size: 24rpx;
  color: #9aa6b8;
}

.crumb-link {
  color: #409eff;
}

.crumb-sep {
  margin: 0 6rpx;
}

.menu-list {
  border-radius: 18rpx;
  background: #fff;
  padding: 0 28rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 26rpx 0;
}

.menu-item + .menu-item {
  border-top: 1rpx solid #f0f3f7;
}

.menu-item-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 18rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
}

.menu-item-main {
  flex: 1;
  min-width: 0;
}

.menu-item-name {
  display: block;
  font-size: 28rpx;
  color: #1f2d3d;
}

.menu-item-path {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #9aa6b8;
}

.menu-item-type {
  font-size: 22rpx;
  color: #8a94a6;
  flex-shrink: 0;
}

.menu-item-arrow {
  font-size: 32rpx;
  color: #c3ccd8;
}

.empty {
  padding: 80rpx 0;
  text-align: center;
  color: #9aa6b8;
  font-size: 26rpx;
}
/* ===== 新版菜单视觉（覆盖旧版） ===== */
.page {
  padding: 28rpx 28rpx 60rpx;
}

.crumb {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6rpx;
  padding: 18rpx 26rpx;
  border-radius: 16rpx;
  background: #fff;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
  font-size: 24rpx;
}

.crumb-link {
  color: #3a6fe0;
  padding: 2rpx 6rpx;
  border-radius: 8rpx;
}

.crumb-sep {
  color: #c3ccd8;
}

.menu-list {
  margin-top: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 0;
  background: transparent;
  box-shadow: none;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx 26rpx;
  border-radius: 22rpx;
  background: #fff;
  border: none;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.menu-item + .menu-item {
  border-top: none;
}

.menu-item-icon {
  width: 84rpx;
  height: 84rpx;
  border-radius: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34rpx;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 10rpx 20rpx rgba(52, 87, 201, 0.16);
}

.menu-item-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.menu-item-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1f2d3d;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-item-path {
  font-size: 22rpx;
  color: #9aa6b8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-item-type {
  flex-shrink: 0;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  background: #f2f5fb;
  font-size: 20rpx;
  color: #8a94a6;
}

.menu-item-arrow {
  font-size: 32rpx;
  color: #c3ccd8;
  flex-shrink: 0;
}

.empty {
  margin-top: 24rpx;
  padding: 90rpx 0;
  text-align: center;
  font-size: 26rpx;
  background: #fff;
  border-radius: 24rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}
</style>
