import request from '@/utils/request';

// 获取数据库监控信息
export function getDbInfo() {
  return request.get('/monitor/db/info');
}
