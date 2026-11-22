import request from '@/utils/request';

// 查询接口列表
export function listApi(params) {
  return request.get('/system/api/list', { params });
}

// 获取接口详情
export function getApi(apiId) {
  return request.get('/system/api/' + apiId);
}

// 新增接口
export function addApi(data) {
  return request.post('/system/api', data);
}

// 修改接口
export function updateApi(data) {
  return request.put('/system/api', data);
}

// 删除接口
export function delApi(apiId) {
  return request.delete('/system/api/' + apiId);
}

// 修改接口状态
export function changeApiStatus(data) {
  return request.put('/system/api/changeStatus', data);
}
