import request from '@/utils/request';

// ---------- 字典类型 ----------
export function listType(params) {
  return request.get('/system/dict/type/list', { params });
}

// 获取字典详情
export function getType(dictId) {
  return request.get('/system/dict/type/' + dictId);
}

// 新增字典
export function addType(data) {
  return request.post('/system/dict/type', data);
}

// 修改字典
export function updateType(data) {
  return request.put('/system/dict/type', data);
}

// 删除字典
export function delType(dictId) {
  return request.delete('/system/dict/type/' + dictId);
}

// ---------- 字典数据 ----------
export function listData(params) {
  return request.get('/system/dict/data/list', { params });
}

// 按字典类型查询（下拉用）
export function getDicts(dictType) {
  return request.get('/system/dict/data/type/' + dictType);
}

// 获取字典详情
export function getData(dictCode) {
  return request.get('/system/dict/data/' + dictCode);
}

// 新增字典
export function addData(data) {
  return request.post('/system/dict/data', data);
}

// 修改字典
export function updateData(data) {
  return request.put('/system/dict/data', data);
}

// 删除字典
export function delData(dictCode) {
  return request.delete('/system/dict/data/' + dictCode);
}
