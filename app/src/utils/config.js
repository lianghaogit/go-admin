// 后端服务地址配置：支持在「我的」页修改，解决真机/小程序 localhost 不可用问题
const BASE_URL_KEY = 'go_admin_base_url'
const DEFAULT_BASE_URL = 'http://localhost:8080'

export function getBaseUrl() {
  return uni.getStorageSync(BASE_URL_KEY) || DEFAULT_BASE_URL
}

export function setBaseUrl(url) {
  const v = String(url || '').trim().replace(/\/+$/, '')
  if (!v) return false
  uni.setStorageSync(BASE_URL_KEY, v)
  return true
}

export function resetBaseUrl() {
  uni.removeStorageSync(BASE_URL_KEY)
}
