// 接口权限接口
import { request } from '@/utils/request'

// 接口列表（分页 + 接口名称/路径关键词）
export function listApi(query) {
  return request({ url: '/api/system/api/list', method: 'GET', data: query })
}

// 接口详情
export function getApi(apiId) {
  return request({ url: `/api/system/api/${apiId}`, method: 'GET' })
}

// 新增接口：{ apiName, path, method, perms, status, remark }
export function createApi(data) {
  return request({ url: '/api/system/api', method: 'POST', data })
}

// 修改接口：{ apiId, ... }
export function updateApi(data) {
  return request({ url: '/api/system/api', method: 'PUT', data })
}

// 启停接口：{ apiId, status }
export function changeApiStatus(apiId, status) {
  return request({ url: '/api/system/api/changeStatus', method: 'PUT', data: { apiId, status } })
}

// 删除接口
export function deleteApi(apiId) {
  return request({ url: `/api/system/api/${apiId}`, method: 'DELETE' })
}
