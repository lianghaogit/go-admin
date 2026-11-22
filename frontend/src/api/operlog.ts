import request from '@/utils/request';

// 查询操作日志列表
export function listOperLog(params) {
  return request.get('/monitor/operlog/list', { params });
}

// 删除操作日志
export function delOperLog(operId) {
  return request.delete('/monitor/operlog/' + operId);
}

// 清空操作日志
export function cleanOperLog() {
  return request.delete('/monitor/operlog/clean');
}

// 导出操作日志
export function exportOperLog(params) {
  return request.get('/monitor/operlog/export', { params, responseType: 'blob' });
}
