// 首页统计接口
import { request } from '@/utils/request'

// 系统概览：用户/角色/菜单/部门/公告/今日登录/在线数
export function dashStats() {
  return request({ url: '/api/dashboard/stats', method: 'GET' })
}
