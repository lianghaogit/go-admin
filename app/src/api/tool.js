// 系统工具接口（代码生成 / 接口文档 / 文件上传）
import { request } from '@/utils/request'
import { getToken } from '@/utils/auth'
import { getBaseUrl } from '@/utils/config'

// 已导入的生成表列表
export function listGenTables(query) {
  return request({ url: '/api/tool/gen/list', method: 'GET', data: query })
}

// 数据库中可导入的业务表
export function listDbTables(keyword) {
  return request({ url: '/api/tool/gen/db/list', method: 'GET', data: { tableName: keyword || undefined } })
}

// 导入数据库表（tables 逗号分隔）
export function importGenTables(tables) {
  return request({ url: `/api/tool/gen/importTable?tables=${encodeURIComponent(tables.join(','))}`, method: 'POST' })
}

// 删除生成表配置
export function deleteGenTable(tableId) {
  return request({ url: `/api/tool/gen/${tableId}`, method: 'DELETE' })
}

// 预览生成代码（返回 文件名 → 内容）
export function previewGenTable(tableId) {
  return request({ url: `/api/tool/gen/preview/${tableId}`, method: 'GET' })
}

// 生成配置详情（含列配置）
export function getGenTable(tableId) {
  return request({ url: `/api/tool/gen/${tableId}`, method: 'GET' })
}

// 保存生成配置（表信息 + 列配置全量）
export function saveGenTable(data) {
  return request({ url: '/api/tool/gen', method: 'PUT', data })
}

// 系统接口文档
export function apiDocList() {
  return request({ url: '/api/tool/api-doc/list', method: 'GET' })
}

// 通用文件上传（multipart，字段名 file）
export function uploadFile(filePath, fileName) {
  return new Promise((resolve, reject) => {
    const token = getToken()
    uni.uploadFile({
      url: getBaseUrl() + '/api/common/upload',
      filePath,
      name: 'file',
      ...(fileName ? { fileName } : {}),
      header: token ? { Authorization: 'Bearer ' + token } : {},
      success: (res) => {
        let body = res.data
        try { body = typeof body === 'string' ? JSON.parse(body) : body } catch (e) {}
        if (body && body.code === 200) {
          resolve(body)
        } else {
          uni.showToast({ title: (body && body.msg) || '上传失败', icon: 'none' })
          reject(body)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '上传失败，请检查网络', icon: 'none' })
        reject(err)
      }
    })
  })
}
