// 岗位管理接口
import { request } from '@/utils/request'

// 岗位列表（分页 + 岗位名称）
export function listPost(query) {
  return request({ url: '/api/system/post/list', method: 'GET', data: query })
}

// 岗位下拉（无分页）
export function listAllPosts() {
  return request({ url: '/api/system/post/all', method: 'GET' })
}

// 新增岗位：{ postCode, postName, postSort, status, remark }
export function createPost(data) {
  return request({ url: '/api/system/post', method: 'POST', data })
}

// 修改岗位：{ postId, ... }
export function updatePost(data) {
  return request({ url: '/api/system/post', method: 'PUT', data })
}

// 删除岗位
export function deletePost(postId) {
  return request({ url: `/api/system/post/${postId}`, method: 'DELETE' })
}
