<template>
  <view class="page">
    <!-- 个人档案 -->
    <view class="profile-hero fade-up">
      <view class="hero-orb"></view>
      <view class="profile-main">
        <view class="avatar-ring">
          <image v-if="avatarUrl" :src="avatarUrl" class="avatar-img" mode="aspectFill" />
          <view v-else class="avatar">{{ avatarChar }}</view>
        </view>
        <view class="profile-info">
          <view class="profile-name">{{ displayName }}</view>
          <view class="profile-account">@{{ userStore.userInfo?.userName || '—' }}</view>
        </view>
        <view class="profile-tag">{{ roleTag }}</view>
      </view>
      <view class="profile-stats">
        <view class="pstat">
          <text class="pnum">{{ userStore.roles.length || 0 }}</text>
          <text class="plabel">角色</text>
        </view>
        <view class="pdiv"></view>
        <view class="pstat">
          <text class="pnum">{{ permCount }}</text>
          <text class="plabel">权限项</text>
        </view>
        <view class="pdiv"></view>
        <view class="pstat">
          <text class="pnum">{{ userStore.permissions.length && userStore.permissions[0] === '*:*:*' ? '全' : '限' }}</text>
          <text class="plabel">权限范围</text>
        </view>
      </view>
    </view>

    <!-- 快捷操作 -->
    <view class="quick-row fade-up">
      <view class="quick-tile" @tap="openProfile">
        <view class="quick-icon q1">资</view>
        <text class="quick-text">资料</text>
      </view>
      <view class="quick-tile" @tap="changeAvatar">
        <view class="quick-icon q2">像</view>
        <text class="quick-text">头像</text>
      </view>
      <view class="quick-tile" @tap="openPwd">
        <view class="quick-icon q3">密</view>
        <text class="quick-text">密码</text>
      </view>
      <view class="quick-tile" @tap="openServer">
        <view class="quick-icon q4">服</view>
        <text class="quick-text">服务器</text>
      </view>
    </view>

    <!-- 我的权限 -->
    <view class="section fade-up">
      <view class="section-head">
        <view class="section-bar"></view>
        <text class="section-title">我的权限</text>
      </view>
      <view class="panel">
        <view class="perm-row">
          <text class="perm-label">角色</text>
          <text class="perm-value">{{ userStore.roles.length ? userStore.roles.join('、') : '无' }}</text>
        </view>
        <view class="perm-row">
          <text class="perm-label">权限</text>
          <text class="perm-value perm-value-sm">{{ permText }}</text>
        </view>
      </view>
    </view>

    <!-- 设置 -->
    <view class="section fade-up">
      <view class="section-head">
        <view class="section-bar"></view>
        <text class="section-title">设置</text>
      </view>
      <view class="panel">
        <view class="cell" @tap="goMenus">
          <view class="cell-icon c1">菜</view>
          <text class="cell-text">全部菜单</text>
          <text class="cell-arrow">›</text>
        </view>
        <view class="cell" @tap="goHome">
          <view class="cell-icon c2">首</view>
          <text class="cell-text">回到首页</text>
          <text class="cell-arrow">›</text>
        </view>
        <view class="cell" @tap="openServer">
          <view class="cell-icon c3">服</view>
          <view class="cell-main">
            <text class="cell-text">服务器地址</text>
            <text class="cell-sub">{{ baseUrl }}</text>
          </view>
          <text class="cell-arrow">›</text>
        </view>
        <view class="cell" @tap="clearCache">
          <view class="cell-icon c4">清</view>
          <text class="cell-text">清除本地缓存</text>
          <text class="cell-arrow">›</text>
        </view>
      </view>
    </view>

    <button class="logout-btn" @tap="handleLogout">退出登录</button>

    <!-- 个人资料 -->
    <view v-if="showProfile" class="mask" @tap="showProfile = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">个人资料</view>
        <view v-if="profileLoading" class="sheet-loading">加载中…</view>
        <view v-else>
          <view class="form-item">
            <text class="label">昵称</text>
            <input v-model="profileForm.nickName" class="input" placeholder="请输入昵称" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">手机号</text>
            <input v-model="profileForm.phonenumber" class="input" type="tel" placeholder="请输入手机号" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">邮箱</text>
            <input v-model="profileForm.email" class="input" placeholder="请输入邮箱" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">备注</text>
            <input v-model="profileForm.remark" class="input" placeholder="选填" placeholder-class="ph" />
          </view>
          <view class="form-actions">
            <view class="btn plain" @tap="showProfile = false">取消</view>
            <view class="btn" @tap="submitProfile">保存</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 修改密码 -->
    <view v-if="showPwd" class="mask" @tap="showPwd = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">修改密码</view>
        <view class="form-item">
          <text class="label">原密码</text>
          <input v-model="pwdForm.oldPassword" class="input" password placeholder="请输入原密码" placeholder-class="ph" />
        </view>
        <view class="form-item">
          <text class="label">新密码</text>
          <input v-model="pwdForm.newPassword" class="input" password placeholder="至少 6 位" placeholder-class="ph" />
        </view>
        <view class="form-item">
          <text class="label">确认密码</text>
          <input v-model="pwdForm.confirmPassword" class="input" password placeholder="再次输入新密码" placeholder-class="ph" />
        </view>
        <view class="form-actions">
          <view class="btn plain" @tap="showPwd = false">取消</view>
          <view class="btn" @tap="submitPwd">保存</view>
        </view>
      </view>
    </view>

    <!-- 服务器地址 -->
    <view v-if="showServer" class="mask" @tap="showServer = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">服务器地址</view>
        <view class="form-item">
          <text class="label">地址</text>
          <input v-model="serverUrl" class="input" placeholder="http://192.168.1.10:8080" placeholder-class="ph" />
        </view>
        <view class="server-tip">保存后立即生效；真机调试请填写后端所在电脑的局域网 IP。</view>
        <view class="form-actions">
          <view class="btn plain" @tap="resetServer">恢复默认</view>
          <view class="btn" @tap="saveServer">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { getProfile, updateProfile, updatePwd, updateAvatar } from '@/api/user'
import { uploadFile } from '@/api/tool'
import { getBaseUrl, setBaseUrl, resetBaseUrl } from '@/utils/config'
import { trimStr, isMobile, isEmail } from '@/utils/validate'

const userStore = useUserStore()

const avatarUrl = computed(() => {
  const avatar = userStore.userInfo?.avatar
  if (!avatar) return ''
  return /^http(s)?:\/\//.test(avatar) ? avatar : getBaseUrl() + avatar
})

const avatarChar = computed(() => {
  const name = userStore.userInfo?.userName || 'U'
  return name.charAt(0).toUpperCase()
})
const displayName = computed(() => {
  const u = userStore.userInfo || {}
  return u.nickName || u.userName || '用户'
})
const permText = computed(() => {
  const p = userStore.permissions || []
  if (!p.length) return '无'
  if (p.length === 1 && p[0] === '*:*:*') return '*:*:*（全部权限）'
  return p.join('、')
})
const permCount = computed(() => {
  const p = userStore.permissions || []
  return p.length === 1 && p[0] === '*:*:*' ? '全部' : p.length
})
// 角色标签：取首个角色名，admin 显示「超级管理员」
const roleTag = computed(() => {
  const roles = userStore.roles || []
  if (!roles.length) return '普通用户'
  return roles[0] === 'admin' ? '超级管理员' : roles[0]
})

function goHome() {
  uni.switchTab({ url: '/pages/index/index' })
}

function goMenus() {
  uni.navigateTo({ url: '/pages/menu/index?title=' + encodeURIComponent('全部菜单') })
}

function clearCache() {
  // 仅清理非登录态缓存（保留 token 与用户信息）
  uni.removeStorageSync('go-admin-app-username')
  uni.showToast({ title: '已清理', icon: 'success' })
}

// ---------- 个人资料 ----------
const showProfile = ref(false)
const profileLoading = ref(false)
const profileForm = reactive({ nickName: '', phonenumber: '', email: '', remark: '' })

async function openProfile() {
  showProfile.value = true
  profileLoading.value = true
  try {
    const res = await getProfile()
    const u = (res && res.data) || {}
    profileForm.nickName = u.nickName || ''
    profileForm.phonenumber = u.phonenumber || u.phone || ''
    profileForm.email = u.email || ''
    profileForm.remark = u.remark || ''
  } catch (e) {
    showProfile.value = false
  } finally {
    profileLoading.value = false
  }
}

async function submitProfile() {
  const nickName = trimStr(profileForm.nickName)
  const phonenumber = trimStr(profileForm.phonenumber)
  const email = trimStr(profileForm.email)
  if (!nickName) {
    uni.showToast({ title: '昵称不能为空', icon: 'none' })
    return
  }
  if (nickName.length > 30) {
    uni.showToast({ title: '昵称不能超过 30 个字符', icon: 'none' })
    return
  }
  if (phonenumber && !isMobile(phonenumber)) {
    uni.showToast({ title: '手机号格式不正确', icon: 'none' })
    return
  }
  if (email && !isEmail(email)) {
    uni.showToast({ title: '邮箱格式不正确', icon: 'none' })
    return
  }
  Object.assign(profileForm, { nickName, phonenumber, email })
  try {
    await updateProfile({
      nickName,
      phonenumber,
      email,
      remark: trimStr(profileForm.remark)
    })
    if (userStore.userInfo) {
      userStore.userInfo.nickName = nickName
      userStore.userInfo.phonenumber = phonenumber
      userStore.userInfo.email = email
      userStore.userInfo.remark = trimStr(profileForm.remark)
    }
    showProfile.value = false
    uni.showToast({ title: '已保存', icon: 'none' })
  } catch (e) {}
}

// ---------- 修改密码 ----------
const showPwd = ref(false)
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

function openPwd() {
  pwdForm.oldPassword = ''
  pwdForm.newPassword = ''
  pwdForm.confirmPassword = ''
  showPwd.value = true
}

async function submitPwd() {
  if (!pwdForm.oldPassword) {
    uni.showToast({ title: '请输入原密码', icon: 'none' })
    return
  }
  if (pwdForm.newPassword.length < 6 || pwdForm.newPassword.length > 30) {
    uni.showToast({ title: '新密码长度需为 6-30 位', icon: 'none' })
    return
  }
  if (pwdForm.newPassword === pwdForm.oldPassword) {
    uni.showToast({ title: '新密码不能与原密码相同', icon: 'none' })
    return
  }
  if (pwdForm.newPassword !== pwdForm.confirmPassword) {
    uni.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
    return
  }
  try {
    await updatePwd({ oldPassword: pwdForm.oldPassword, newPassword: pwdForm.newPassword })
    showPwd.value = false
    uni.showToast({ title: '密码已修改', icon: 'none' })
  } catch (e) {}
}

// ---------- 更换头像 ----------
function changeAvatar() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const path = (res.tempFilePaths || [])[0]
      if (!path) return
      try {
        const up = await uploadFile(path)
        const url = up.data?.url
        if (!url) return
        await updateAvatar(url)
        if (userStore.userInfo) userStore.userInfo.avatar = url
        uni.showToast({ title: '头像已更新', icon: 'success' })
      } catch (e) {}
    }
  })
}

// ---------- 服务器地址 ----------
const showServer = ref(false)
const serverUrl = ref(getBaseUrl())
const baseUrl = computed(() => serverUrl.value || '默认')

function openServer() {
  serverUrl.value = getBaseUrl()
  showServer.value = true
}

function saveServer() {
  if (!setBaseUrl(serverUrl.value)) {
    uni.showToast({ title: '请输入服务器地址', icon: 'none' })
    return
  }
  serverUrl.value = getBaseUrl()
  showServer.value = false
  uni.showToast({ title: '已保存，立即生效', icon: 'none' })
}

function resetServer() {
  resetBaseUrl()
  serverUrl.value = getBaseUrl()
  uni.showToast({ title: '已恢复默认地址', icon: 'none' })
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定退出登录吗？',
    success: async (res) => {
      if (!res.confirm) return
      await userStore.logout()
      uni.showToast({ title: '已退出', icon: 'none' })
      setTimeout(() => {
        uni.reLaunch({ url: '/pages/login/index' })
      }, 400)
    }
  })
}

onShow(() => {
  if (userStore.token && !userStore.userInfo) {
    userStore.init().catch(() => {})
  }
})

// 下拉刷新：重新拉取用户信息/角色/权限
onPullDownRefresh(async () => {
  try {
    if (userStore.token) {
      await userStore.init()
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

.user-card {
  display: flex;
  align-items: center;
  gap: 24rpx;
  padding: 36rpx 32rpx;
  border-radius: 18rpx;
  background: #fff;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.user-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #409eff, #337ecc);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 44rpx;
  font-weight: 700;
}

.user-avatar-img {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  flex-shrink: 0;
  background: #f5f7fa;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 34rpx;
  font-weight: 600;
  color: #1f2d3d;
}

.user-account {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #8a94a6;
}

.user-tag {
  padding: 8rpx 20rpx;
  border-radius: 18rpx;
  background: #ecf5ff;
  color: #409eff;
  font-size: 22rpx;
}

.section {
  margin-top: 32rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2d3d;
  margin-bottom: 20rpx;
}

.perm-card {
  padding: 28rpx 32rpx;
  border-radius: 18rpx;
  background: #fff;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.perm-row {
  display: flex;
  gap: 20rpx;
  padding: 10rpx 0;
}

.perm-row + .perm-row {
  border-top: 1rpx solid #f0f3f7;
  margin-top: 10rpx;
  padding-top: 20rpx;
}

.perm-label {
  width: 110rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  color: #8a94a6;
}

.perm-value {
  flex: 1;
  font-size: 26rpx;
  color: #1f2d3d;
}

.perm-value-sm {
  word-break: break-all;
}

.list-card {
  border-radius: 18rpx;
  background: #fff;
  padding: 0 32rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 30rpx 0;
}

.list-item + .list-item {
  border-top: 1rpx solid #f0f3f7;
}

.list-text {
  font-size: 28rpx;
  color: #1f2d3d;
}

.list-arrow {
  font-size: 34rpx;
  color: #c3ccd8;
}

.list-main {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}

.list-sub {
  font-size: 22rpx;
  color: #8a94a6;
  word-break: break-all;
}

.logout-btn {
  margin-top: 60rpx;
  height: 92rpx;
  line-height: 92rpx;
  border-radius: 18rpx;
  font-size: 30rpx;
  color: #f56c6c;
  background: #fff;
  border: 1rpx solid #fde2e2;
}

.logout-btn::after {
  border: none;
}

/* ===== 新版个人中心样式（覆盖旧版） ===== */
.profile-hero {
  position: relative;
  overflow: hidden;
  padding: 40rpx 32rpx 0;
  border-radius: 28rpx;
  background: linear-gradient(135deg, #4a7df0 0%, #3457c9 60%, #2f49a8 100%);
  box-shadow: 0 18rpx 40rpx rgba(52, 87, 201, 0.28);
}

.hero-orb {
  position: absolute;
  width: 260rpx;
  height: 260rpx;
  top: -120rpx;
  right: -80rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
}

.profile-main {
  position: relative;
  display: flex;
  align-items: center;
  gap: 22rpx;
}

.avatar-ring {
  width: 112rpx;
  height: 112rpx;
  padding: 6rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.35);
  flex-shrink: 0;
}

.avatar,
.avatar-img {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.92);
  color: #3457c9;
  font-size: 42rpx;
  font-weight: 700;
}

.avatar-img {
  background: #e9efff;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-account {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.72);
}

.profile-tag {
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

.profile-stats {
  position: relative;
  display: flex;
  align-items: center;
  margin-top: 32rpx;
  padding: 22rpx 0 26rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.16);
}

.pstat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}

.pnum {
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  line-height: 1.1;
}

.plabel {
  font-size: 20rpx;
  color: rgba(255, 255, 255, 0.7);
}

.pdiv {
  width: 1rpx;
  height: 32rpx;
  background: rgba(255, 255, 255, 0.2);
}

.quick-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16rpx;
  margin-top: 24rpx;
}

.quick-tile {
  padding: 24rpx 0;
  border-radius: 20rpx;
  background: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.quick-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  color: #fff;
}

.q1 { background: linear-gradient(135deg, #6aa4ff, #3a6fe0); }
.q2 { background: linear-gradient(135deg, #a99bff, #7b61ff); }
.q3 { background: linear-gradient(135deg, #ffc46b, #e6a23c); }
.q4 { background: linear-gradient(135deg, #5fd3cd, #00a88f); }

.quick-text {
  font-size: 22rpx;
  color: #5a6678;
}

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
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2d3d;
  margin-bottom: 0;
}

.panel {
  border-radius: 24rpx;
  background: #fff;
  padding: 0 32rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.perm-row {
  display: flex;
  gap: 20rpx;
  padding: 26rpx 0;
}

.perm-row + .perm-row {
  border-top: 1rpx solid #f2f5fb;
  margin-top: 0;
  padding-top: 26rpx;
}

.perm-label {
  width: 100rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  color: #8a94a6;
}

.perm-value {
  flex: 1;
  font-size: 26rpx;
  color: #1f2d3d;
}

.perm-value-sm {
  word-break: break-all;
  color: #5a6678;
  font-size: 24rpx;
}

.cell {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 26rpx 0;
}

.cell + .cell {
  border-top: 1rpx solid #f2f5fb;
}

.cell-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 18rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26rpx;
  color: #fff;
  flex-shrink: 0;
}

.c1 { background: linear-gradient(135deg, #6aa4ff, #3a6fe0); }
.c2 { background: linear-gradient(135deg, #7fd3a8, #3fae7a); }
.c3 { background: linear-gradient(135deg, #ffc46b, #e6a23c); }
.c4 { background: linear-gradient(135deg, #c0ccdd, #8a98ab); }

.cell-text {
  flex: 1;
  font-size: 28rpx;
  color: #1f2d3d;
}

.cell-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}

.cell-sub {
  font-size: 22rpx;
  color: #8a94a6;
  word-break: break-all;
}

.cell-arrow {
  font-size: 34rpx;
  color: #c3ccd8;
  flex-shrink: 0;
}

.logout-btn {
  margin-top: 56rpx;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 24rpx;
  font-size: 30rpx;
  color: #f56c6c;
  background: #fff;
  border: none;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.mask {
  position: fixed;
  inset: 0;
  background: rgba(31, 45, 61, 0.45);
  z-index: 99;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.sheet {
  width: 100%;
  max-height: 86vh;
  overflow-y: auto;
  background: #fff;
  border-radius: 16rpx 16rpx 0 0;
  padding: 32rpx 32rpx 48rpx;
  box-sizing: border-box;
}

.sheet-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2d3d;
  text-align: center;
  margin-bottom: 28rpx;
}

.sheet-loading {
  padding: 60rpx 0;
  text-align: center;
  color: #9aa6b8;
  font-size: 26rpx;
}

.form-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #f0f3f7;
}

.label {
  width: 140rpx;
  font-size: 26rpx;
  color: #5a6678;
  flex-shrink: 0;
}

.input {
  flex: 1;
  height: 64rpx;
  font-size: 26rpx;
}

.ph {
  color: #9aa6b8;
}

.server-tip {
  margin-top: 20rpx;
  font-size: 22rpx;
  color: #8a94a6;
  line-height: 1.6;
}

.form-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 40rpx;
}

.btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  text-align: center;
  background: #409eff;
  color: #fff;
  font-size: 28rpx;
  border-radius: 18rpx;
}

.btn.plain {
  background: #f5f7fa;
  color: #5a6678;
}
</style>
