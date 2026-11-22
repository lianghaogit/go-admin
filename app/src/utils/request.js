// ===== 全局请求封装（参考 RuoYi-App 的 request.js）=====
// 后端地址支持在「我的」页动态配置（utils/config.js），便于真机/小程序使用局域网 IP
import { getToken, removeToken } from '@/utils/auth'
import { getBaseUrl } from '@/utils/config'

// 401/登录态失效统一处理：清 token 回登录页
function handleUnauthorized() {
  removeToken()
  uni.showToast({ title: '登录已过期，请重新登录', icon: 'none' })
  setTimeout(() => {
    uni.reLaunch({ url: '/pages/login/index' })
  }, 600)
}

export function request(options) {
  return new Promise((resolve, reject) => {
    const token = getToken()
    uni.request({
      url: getBaseUrl() + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      timeout: 15000,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: 'Bearer ' + token } : {}),
        ...(options.header || {})
      },
      success: (res) => {
        // HTTP 层 401（JWT 中间件直接拒绝）
        if (res.statusCode === 401) {
          handleUnauthorized()
          reject(new Error('未授权'))
          return
        }
        const data = res.data
        // 业务层统一 code 约定（200 成功 / 401 未登录 / 500 失败）
        if (data && typeof data === 'object' && data.code !== undefined) {
          if (data.code === 200) {
            resolve(data)
          } else if (data.code === 401) {
            handleUnauthorized()
            reject(data)
          } else {
            // 业务失败统一提示，避免页面未 catch 时点击“没反应”
            uni.showToast({ title: (data && data.msg) || '操作失败', icon: 'none' })
            reject(data)
          }
        } else {
          // 非标准结构透传
          resolve(data)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '网络异常，请检查后端服务', icon: 'none' })
        reject(err)
      }
    })
  })
}
