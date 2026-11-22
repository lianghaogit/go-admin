// 部门管理接口
import { request } from '@/utils/request'

// 部门树（无分页，data 为树形数组）
export function listDept() {
  return request({ url: '/api/system/dept/list', method: 'GET' })
}

// 部门下拉树（按当前用户数据权限过滤）
export function deptTree() {
  return request({ url: '/api/system/dept/treeselect', method: 'GET' })
}

// 新增部门：{ parentId, deptName, orderNum, leader, phone, email, status }
export function createDept(data) {
  return request({ url: '/api/system/dept', method: 'POST', data })
}

// 修改部门：{ deptId, parentId, deptName, orderNum, leader, phone, email, status }
export function updateDept(data) {
  return request({ url: '/api/system/dept', method: 'PUT', data })
}

// 删除部门
export function deleteDept(deptId) {
  return request({ url: `/api/system/dept/${deptId}`, method: 'DELETE' })
}
