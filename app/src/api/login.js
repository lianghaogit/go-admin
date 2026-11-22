// 登录/用户/菜单接口（对齐后端 go-admin 的 /api 路由）
import { request } from '@/utils/request'

// 获取验证码（captchaEnabled / img(base64 PNG) / uuid）
export function captchaImage() {
  return request({ url: '/api/captchaImage', method: 'GET' })
}

// 登录（username / password / code / uuid）
export function login(data) {
  return request({ url: '/api/login', method: 'POST', data })
}

// 当前用户信息 + 角色 + 权限
export function getInfo() {
  return request({ url: '/api/getInfo', method: 'GET' })
}

// 动态菜单（菜单树）
export function getRouters() {
  return request({ url: '/api/getRouters', method: 'GET' })
}

// 退出登录
export function logout() {
  return request({ url: '/api/logout', method: 'POST' })
}
