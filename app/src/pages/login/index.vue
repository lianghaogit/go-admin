<template>
  <view class="login-page">
    <view class="bg-orb bg-a"></view>
    <view class="bg-orb bg-b"></view>

    <view class="login-brand">
      <view class="brand-mark">G</view>
      <view class="login-title">Go Admin</view>
      <view class="login-subtitle">移动工作台 · 一站式管理</view>
    </view>

    <view class="login-card">
      <view class="form-item">
        <text class="form-label">账号</text>
        <input
          v-model="form.username"
          class="form-input"
          placeholder="请输入账号"
          placeholder-class="form-placeholder"
          :maxlength="50"
        />
      </view>

      <view class="form-item">
        <text class="form-label">密码</text>
        <view class="form-input-wrap">
          <input
            v-model="form.password"
            class="form-input"
            :password="!showPwd"
            placeholder="请输入密码"
            placeholder-class="form-placeholder"
            :maxlength="50"
          />
          <text class="pwd-toggle" @tap="showPwd = !showPwd">{{ showPwd ? '隐藏' : '显示' }}</text>
        </view>
      </view>

      <view class="form-item">
        <text class="form-label">验证码</text>
        <view class="captcha-wrap">
          <input
            v-model="form.code"
            class="form-input captcha-input"
            placeholder="请输入验证码"
            placeholder-class="form-placeholder"
            :maxlength="6"
          />
          <!-- #ifdef H5 -->
          <!-- H5 端 uni-image 对动态 data URL 渲染不稳定，改用 CSS 背景直接显示 -->
          <view
            v-if="captchaImg"
            class="captcha-img"
            :style="{ backgroundImage: 'url(' + captchaImg + ')' }"
            @tap="loadCaptcha"
          />
          <view v-else class="captcha-img captcha-loading" @tap="loadCaptcha">点击加载</view>
          <!-- #endif -->
          <!-- #ifndef H5 -->
          <image
            v-if="captchaImg"
            class="captcha-img"
            :src="captchaImg"
            mode="aspectFill"
            @tap="loadCaptcha"
          />
          <view v-else class="captcha-img captcha-loading" @tap="loadCaptcha">点击加载</view>
          <!-- #endif -->
        </view>
      </view>

      <view class="form-item remember-row">
        <view class="remember" @tap="remember = !remember">
          <view class="remember-box" :class="{ checked: remember }">
            <text v-if="remember" class="remember-check">✓</text>
          </view>
          <text class="remember-text">记住账号</text>
        </view>
      </view>

      <button class="login-btn" :loading="loading" :disabled="loading" @tap="handleLogin">
        {{ loading ? '登录中…' : '登 录' }}
      </button>

      <view class="server-entry" @tap="openServer">
        <text class="server-label">服务器</text>
        <text class="server-url">{{ baseUrl }}</text>
        <text class="server-edit">设置</text>
      </view>
    </view>

    <!-- 服务器地址设置：真机/小程序默认 localhost 不可用，登录前即可修改 -->
    <view v-if="showServer" class="mask" @tap="showServer = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">服务器地址</view>
        <view class="server-form-item">
          <input v-model="serverUrl" class="server-input" placeholder="http://192.168.1.10:8080" placeholder-class="form-placeholder" />
        </view>
        <view class="server-tip">真机调试请填写后端所在电脑的局域网 IP；保存后会重新加载验证码。</view>
        <view class="server-actions">
          <view class="server-btn plain" @tap="resetServer">恢复默认</view>
          <view class="server-btn" @tap="saveServer">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { reactive, ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { captchaImage } from '@/api/login'
import { useUserStore } from '@/store/user'
import { getBaseUrl, setBaseUrl, resetBaseUrl } from '@/utils/config'

const userStore = useUserStore()

const form = reactive({ username: '', password: '', code: '' })
const showPwd = ref(false)
const remember = ref(true)
const loading = ref(false)

// 服务器地址设置（登录前即可改，避免真机 localhost 无法登录）
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
  uni.showToast({ title: '已保存，重新加载验证码', icon: 'none' })
  loadCaptcha()
}

function resetServer() {
  resetBaseUrl()
  serverUrl.value = getBaseUrl()
  uni.showToast({ title: '已恢复默认地址', icon: 'none' })
}

let uuid = ''
const captchaImg = ref('')

// 加载验证码（base64 图 + uuid；验证码一次性，失败后必须重新获取）
async function loadCaptcha() {
  try {
    const res = await captchaImage()
    const data = res.data || {}
    if (data.captchaEnabled === false) {
      // 后端关闭验证码时直接置空，走免验证码登录
      uuid = ''
      captchaImg.value = ''
      return
    }
    uuid = data.uuid
    // 后端返回的 img 已带 data:image/png;base64, 前缀（幂等兼容两种返回）
    const img = data.img || ''
    captchaImg.value = img.startsWith('data:') ? img : 'data:image/png;base64,' + img
    form.code = ''
  } catch (e) {
    uni.showToast({ title: '验证码加载失败，请检查后端服务', icon: 'none' })
  }
}

function failText(e) {
  if (e && typeof e === 'object') return e.msg || e.message || '请求失败'
  return String(e || '请求失败')
}

async function handleLogin() {
  if (!form.username.trim()) return uni.showToast({ title: '请输入账号', icon: 'none' })
  if (!form.password) return uni.showToast({ title: '请输入密码', icon: 'none' })
  if (uuid && !form.code) return uni.showToast({ title: '请输入验证码', icon: 'none' })
  loading.value = true
  try {
    await userStore.login({
      username: form.username.trim(),
      password: form.password,
      code: form.code.trim(),
      uuid
    })
    // 记住账号
    if (remember.value) {
      uni.setStorageSync('go-admin-app-username', form.username.trim())
    } else {
      uni.removeStorageSync('go-admin-app-username')
    }
    // 拉取用户信息 + 菜单（权限）
    await userStore.init()
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/index/index' })
    }, 400)
  } catch (e) {
    uni.showToast({ title: failText(e), icon: 'none' })
    // 验证码一次性，失败即重新获取
    loadCaptcha()
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  // 已登录则直接进首页
  if (userStore.token) {
    uni.switchTab({ url: '/pages/index/index' })
    return
  }
  loadCaptcha()
  // 回填记住的账号
  form.username = uni.getStorageSync('go-admin-app-username') || ''
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  padding: 120rpx 48rpx 60rpx;
  box-sizing: border-box;
  background: linear-gradient(180deg, #e8f1ff 0%, #f5f7fa 45%);
}

.login-header {
  margin-bottom: 72rpx;
}

.login-title {
  font-size: 52rpx;
  font-weight: 700;
  color: #1f2d3d;
  letter-spacing: 2rpx;
}

.login-subtitle {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #8a94a6;
}

.login-card {
  padding: 48rpx 40rpx 40rpx;
  border-radius: 18rpx;
  background: #fff;
  box-shadow: 0 12rpx 40rpx rgba(31, 45, 61, 0.08);
}

.form-item {
  margin-bottom: 36rpx;
}

.form-label {
  display: block;
  margin-bottom: 12rpx;
  font-size: 26rpx;
  color: #5a6a7a;
}

.form-input {
  height: 88rpx;
  padding: 0 24rpx;
  border-radius: 18rpx;
  background: #f6f8fa;
  font-size: 28rpx;
  color: #1f2d3d;
  box-sizing: border-box;
}

.form-placeholder {
  color: #9aa6b8;
}

.form-input-wrap {
  display: flex;
  align-items: center;
  background: #f6f8fa;
  border-radius: 18rpx;
}

.form-input-wrap .form-input {
  flex: 1;
  background: transparent;
}

.pwd-toggle {
  padding: 0 24rpx;
  font-size: 24rpx;
  color: #409eff;
}

.captcha-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.captcha-input {
  flex: 1;
}

.captcha-img {
  width: 200rpx;
  height: 88rpx;
  border-radius: 18rpx;
  border: 1rpx solid #e4e9f0;
  flex-shrink: 0;
  background-size: cover;
  background-position: center;
}

.captcha-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22rpx;
  color: #8a94a6;
  background: #f6f8fa;
}

.remember-row {
  margin-top: -8rpx;
}

.remember {
  display: inline-flex;
  align-items: center;
  gap: 10rpx;
}

.remember-box {
  width: 34rpx;
  height: 34rpx;
  border: 2rpx solid #c3ccd8;
  border-radius: 18rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.remember-box.checked {
  background: #409eff;
  border-color: #409eff;
}

.remember-check {
  font-size: 22rpx;
  color: #fff;
  line-height: 1;
}

.remember-text {
  font-size: 24rpx;
  color: #6b7686;
}

.login-btn {
  margin-top: 20rpx;
  height: 92rpx;
  line-height: 92rpx;
  border-radius: 18rpx;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(90deg, #409eff, #337ecc);
  border: none;
}

.login-btn[disabled] {
  opacity: 0.7;
  color: #fff;
  background: linear-gradient(90deg, #409eff, #337ecc);
}

.server-entry {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 28rpx;
  padding: 16rpx 24rpx;
  background: #f6f8fa;
  border-radius: 18rpx;
}

.server-label {
  font-size: 22rpx;
  color: #8a94a6;
  flex-shrink: 0;
}

.server-url {
  flex: 1;
  min-width: 0;
  font-size: 22rpx;
  color: #5a6a7a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-edit {
  font-size: 22rpx;
  color: #409eff;
  flex-shrink: 0;
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

.server-form-item {
  background: #f6f8fa;
  border-radius: 18rpx;
  padding: 0 20rpx;
}

.server-input {
  height: 80rpx;
  font-size: 26rpx;
  color: #1f2d3d;
}

.server-tip {
  margin-top: 16rpx;
  font-size: 22rpx;
  color: #8a94a6;
  line-height: 1.6;
}

.server-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 36rpx;
}

.server-btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  text-align: center;
  background: #409eff;
  color: #fff;
  font-size: 28rpx;
  border-radius: 18rpx;
}

.server-btn.plain {
  background: #f5f7fa;
  color: #5a6678;
}
/* ===== 新版登录视觉（覆盖旧版） ===== */
.login-page {
  position: relative;
  overflow: hidden;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60rpx 40rpx;
  box-sizing: border-box;
  background: linear-gradient(180deg, #eef3ff 0%, #f4f6fb 45%, #f4f6fb 100%);
}

.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(2rpx);
}

.bg-a {
  width: 460rpx;
  height: 460rpx;
  top: -180rpx;
  right: -160rpx;
  background: radial-gradient(circle at 30% 30%, rgba(106, 164, 255, 0.35), rgba(106, 164, 255, 0));
}

.bg-b {
  width: 360rpx;
  height: 360rpx;
  bottom: -120rpx;
  left: -120rpx;
  background: radial-gradient(circle at 60% 40%, rgba(123, 97, 255, 0.25), rgba(123, 97, 255, 0));
}

.login-brand {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 44rpx;
}

.brand-mark {
  width: 112rpx;
  height: 112rpx;
  border-radius: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 52rpx;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #4a7df0, #3457c9);
  box-shadow: 0 16rpx 32rpx rgba(52, 87, 201, 0.28);
}

.login-title {
  margin-top: 22rpx;
  font-size: 42rpx;
  font-weight: 700;
  color: #1f2d3d;
  letter-spacing: 1rpx;
}

.login-subtitle {
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #8a94a6;
}

.login-card {
  position: relative;
  width: 100%;
  padding: 44rpx 36rpx 36rpx;
  border-radius: 30rpx;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 22rpx 60rpx rgba(30, 48, 80, 0.1);
}

.form-item {
  margin-bottom: 26rpx;
}

.form-label {
  display: block;
  margin-bottom: 12rpx;
  font-size: 24rpx;
  color: #5a6678;
}

.form-input-wrap {
  position: relative;
}

.form-input {
  width: 100%;
  height: 92rpx;
  padding: 0 26rpx;
  border-radius: 16rpx;
  background: #f5f7fb;
  border: 1rpx solid #edf1f7;
  font-size: 28rpx;
  color: #1f2d3d;
  box-sizing: border-box;
}

.form-input-wrap .form-input {
  padding-right: 110rpx;
}

.pwd-toggle {
  position: absolute;
  right: 24rpx;
  top: 50%;
  transform: translateY(-50%);
  font-size: 24rpx;
  color: #3a6fe0;
}

.captcha-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.captcha-input {
  flex: 1;
  min-width: 0;
}

.captcha-img {
  width: 210rpx;
  height: 92rpx;
  border-radius: 14rpx;
  flex-shrink: 0;
  overflow: hidden;
  background-color: #eef2f9;
  background-size: cover;
  background-position: center;
}

.captcha-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22rpx;
  color: #8a94a6;
}

.remember-row {
  margin-bottom: 20rpx;
}

.remember-text {
  font-size: 24rpx;
  color: #8a94a6;
}

.login-btn {
  margin-top: 12rpx;
  height: 100rpx;
  line-height: 100rpx;
  border-radius: 20rpx;
  font-size: 32rpx;
  font-weight: 600;
  letter-spacing: 4rpx;
  color: #fff;
  background: linear-gradient(135deg, #4a7df0, #3457c9);
  box-shadow: 0 16rpx 32rpx rgba(52, 87, 201, 0.28);
  border: none;
}

.login-btn::after {
  border: none;
}

.server-entry {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 26rpx;
  padding: 18rpx 22rpx;
  border-radius: 14rpx;
  background: #f5f7fb;
}

.server-label {
  font-size: 22rpx;
  color: #8a94a6;
  flex-shrink: 0;
}

.server-url {
  flex: 1;
  min-width: 0;
  font-size: 22rpx;
  color: #5a6678;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-edit {
  font-size: 22rpx;
  color: #3a6fe0;
  flex-shrink: 0;
}
</style>
