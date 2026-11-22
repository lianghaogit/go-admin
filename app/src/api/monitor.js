// 服务/缓存/数据监控接口
import { request } from '@/utils/request'

// 服务监控（CPU/内存/磁盘/系统）
export function serverInfo() {
  return request({ url: '/api/monitor/server/info', method: 'GET' })
}

// 缓存监控统计
export function cacheInfo() {
  return request({ url: '/api/monitor/cache/info', method: 'GET' })
}

// 缓存键详情
export function cacheKeyDetail(key) {
  return request({ url: '/api/monitor/cache/key', method: 'GET', data: { key } })
}

// 清空缓存
export function cacheClear() {
  return request({ url: '/api/monitor/cache/clear', method: 'DELETE' })
}

// 删除单个缓存键
export function cacheRemoveKey(key) {
  return request({ url: '/api/monitor/cache/key', method: 'DELETE', data: { key } })
}

// 数据监控（SQLite）
export function dbInfo() {
  return request({ url: '/api/monitor/db/info', method: 'GET' })
}
