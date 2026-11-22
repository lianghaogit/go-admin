// 参数配置接口
import { request } from '@/utils/request'

// 参数列表（分页 + 参数名称/键名）
export function listConfig(query) {
  return request({ url: '/api/system/config/list', method: 'GET', data: query })
}

// 新增参数：{ configName, configKey, configValue, configType, remark }
export function createConfig(data) {
  return request({ url: '/api/system/config', method: 'POST', data })
}

// 修改参数：{ configId, ... }
export function updateConfig(data) {
  return request({ url: '/api/system/config', method: 'PUT', data })
}

// 删除参数
export function deleteConfig(configId) {
  return request({ url: `/api/system/config/${configId}`, method: 'DELETE' })
}
