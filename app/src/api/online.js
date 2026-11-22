// 在线用户接口
import { request } from '@/utils/request'

// 在线用户列表
export function listOnline(query) {
  return request({ url: '/api/monitor/online/list', method: 'GET', data: query })
}

// 强退单个在线用户（tokenHash 为会话主键）
export function forceLogout(tokenHash) {
  return request({ url: `/api/monitor/online/${tokenHash}`, method: 'DELETE' })
}

// 强退全部在线用户
export function forceLogoutAll() {
  return request({ url: '/api/monitor/online/all', method: 'DELETE' })
}
