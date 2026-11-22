// 字典管理接口
import { request } from '@/utils/request'

// 字典类型列表（分页 + 名称/类型关键词）
export function listDictType(query) {
  return request({ url: '/api/system/dict/type/list', method: 'GET', data: query })
}

// 新增字典类型：{ dictName, dictType, status, remark }
export function createDictType(data) {
  return request({ url: '/api/system/dict/type', method: 'POST', data })
}

// 修改字典类型：{ dictId, dictName, dictType, status, remark }
export function updateDictType(data) {
  return request({ url: '/api/system/dict/type', method: 'PUT', data })
}

// 删除字典类型
export function deleteDictType(dictId) {
  return request({ url: `/api/system/dict/type/${dictId}`, method: 'DELETE' })
}

// ---- 字典数据 ----
// 按类型查字典数据（分页）
export function listDictData(query) {
  return request({ url: '/api/system/dict/data/list', method: 'GET', data: query })
}

// 新增字典数据：{ dictSort, dictLabel, dictValue, dictType, status }
export function createDictData(data) {
  return request({ url: '/api/system/dict/data', method: 'POST', data })
}

// 修改字典数据：{ dictCode, ... }
export function updateDictData(data) {
  return request({ url: '/api/system/dict/data', method: 'PUT', data })
}

// 删除字典数据
export function deleteDictData(dictCode) {
  return request({ url: `/api/system/dict/data/${dictCode}`, method: 'DELETE' })
}
