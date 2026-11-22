// token 存取（uni storage，与 Web 端 localStorage 'Admin-Token' 语义一致）
const TOKEN_KEY = 'Admin-Token'

export function getToken() {
  return uni.getStorageSync(TOKEN_KEY) || ''
}

export function setToken(token) {
  uni.setStorageSync(TOKEN_KEY, token)
}

export function removeToken() {
  uni.removeStorageSync(TOKEN_KEY)
}
