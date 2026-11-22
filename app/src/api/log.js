// 日志接口（登录日志 / 操作日志）
import { request } from '@/utils/request'

// 登录日志列表
export function listLoginLog(query) {
  return request({ url: '/api/monitor/logininfor/list', method: 'GET', data: query })
}

// 操作日志列表
export function listOperLog(query) {
  return request({ url: '/api/monitor/operlog/list', method: 'GET', data: query })
}

// 删除单条登录日志
export function deleteLoginLog(infoId) {
  return request({ url: `/api/monitor/logininfor/${infoId}`, method: 'DELETE' })
}

// 清空登录日志
export function cleanLoginLog() {
  return request({ url: '/api/monitor/logininfor/clean', method: 'DELETE' })
}

// 批量删除登录日志（ids 用逗号拼接）
export function deleteLoginLogs(ids) {
  return request({ url: `/api/monitor/logininfor/${ids}`, method: 'DELETE' })
}

// 删除单条操作日志
export function deleteOperLog(operId) {
  return request({ url: `/api/monitor/operlog/${operId}`, method: 'DELETE' })
}

// 清空操作日志
export function cleanOperLog() {
  return request({ url: '/api/monitor/operlog/clean', method: 'DELETE' })
}

// 批量删除操作日志（ids 用逗号拼接）
export function deleteOperLogs(ids) {
  return request({ url: `/api/monitor/operlog/${ids}`, method: 'DELETE' })
}
