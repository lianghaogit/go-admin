import request from '@/utils/request';

// 查询参数配置列表
export function listConfig(params) {
  return request.get('/system/config/list', { params });
}

// 获取参数配置详情
export function getConfig(configId) {
  return request.get('/system/config/' + configId);
}

// 新增参数配置
export function addConfig(data) {
  return request.post('/system/config', data);
}

// 修改参数配置
export function updateConfig(data) {
  return request.put('/system/config', data);
}

// 删除参数配置
export function delConfig(configId) {
  return request.delete('/system/config/' + configId);
}
