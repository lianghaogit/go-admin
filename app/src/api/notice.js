// 通知公告接口
import { request } from '@/utils/request'

// 公告列表（分页 + 标题关键词）
export function listNotice(query) {
  return request({ url: '/api/system/notice/list', method: 'GET', data: query })
}

// 新增公告：{ noticeTitle, noticeType, noticeContent, status, remark }
export function createNotice(data) {
  return request({ url: '/api/system/notice', method: 'POST', data })
}

// 修改公告：{ noticeId, ... }
export function updateNotice(data) {
  return request({ url: '/api/system/notice', method: 'PUT', data })
}

// 删除公告
export function deleteNotice(noticeId) {
  return request({ url: `/api/system/notice/${noticeId}`, method: 'DELETE' })
}
