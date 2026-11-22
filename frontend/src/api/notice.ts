import request from '@/utils/request';

// 查询公告列表
export function listNotice(params) {
  return request.get('/system/notice/list', { params });
}

// 查询启用的公告列表
export function listActiveNotice() {
  return request.get('/system/notice/active');
}

// 获取公告详情
export function getNotice(noticeId) {
  return request.get('/system/notice/' + noticeId);
}

// 新增公告
export function addNotice(data) {
  return request.post('/system/notice', data);
}

// 修改公告
export function updateNotice(data) {
  return request.put('/system/notice', data);
}

// 删除公告
export function delNotice(noticeId) {
  return request.delete('/system/notice/' + noticeId);
}
