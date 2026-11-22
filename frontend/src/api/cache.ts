import request from '@/utils/request';

// 获取缓存监控信息
export function getCacheInfo() {
  return request.get('/monitor/cache/info');
}

// 查询缓存键值
export function getCacheKey(key) {
  return request.get('/monitor/cache/key', { params: { key } });
}

// 清理缓存
export function clearCache() {
  return request.delete('/monitor/cache/clear');
}
