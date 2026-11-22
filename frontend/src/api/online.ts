import request from '@/utils/request';

// 查询在线用户列表
export function listOnline(params) {
  return request.get('/monitor/online/list', { params });
}

// 强制在线用户
export function forceLogout(tokenHash) {
  return request.delete('/monitor/online/' + tokenHash);
}

// 强制所有用户下线
export function forceLogoutAll() {
  return request.delete('/monitor/online/all');
}
