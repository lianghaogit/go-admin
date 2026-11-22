import request from '@/utils/request';

// 查询登录日志列表
export function listLoginLog(params) {
  return request.get('/monitor/logininfor/list', { params });
}

// 删除登录日志
export function delLoginLog(infoId) {
  return request.delete('/monitor/logininfor/' + infoId);
}

// 清空登录日志
export function cleanLoginLog() {
  return request.delete('/monitor/logininfor/clean');
}

// 导出登录日志（携带过滤条件）
export function exportLoginLog(params) {
  return request.get('/monitor/logininfor/export', { params, responseType: 'blob' });
}
