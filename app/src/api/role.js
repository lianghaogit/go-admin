// 角色管理接口（对齐后端 /api/system/role 路由）
import { request } from '@/utils/request'

// 角色列表（分页 + 角色名/状态）
// query: { pageNum, pageSize, roleName, status }
export function listRole(query) {
  return request({ url: '/api/system/role/list', method: 'GET', data: query })
}

// 新增角色：{ roleName, roleKey, roleSort, status, dataScope, remark, menuIds, deptIds }
export function createRole(data) {
  return request({ url: '/api/system/role', method: 'POST', data })
}

// 修改角色：{ roleId, roleName, roleKey, roleSort, status, dataScope, remark, menuIds, deptIds }
export function updateRole(data) {
  return request({ url: '/api/system/role', method: 'PUT', data })
}

// 角色下拉（无分页）
export function listAllRoles() {
  return request({ url: '/api/system/role/all', method: 'GET' })
}

// 角色详情（含 deptIds）
export function getRole(roleId) {
  return request({ url: `/api/system/role/${roleId}`, method: 'GET' })
}

// 角色菜单权限树：data = { menus, checkedKeys }
export function roleMenuTree(roleId) {
  return request({ url: `/api/system/menu/roleMenuTreeselect/${roleId}`, method: 'GET' })
}

// 删除角色
export function deleteRole(roleId) {
  return request({ url: `/api/system/role/${roleId}`, method: 'DELETE' })
}

// 角色已分配用户（分页 + username）
export function listRoleAllocatedUsers(roleId, query) {
  return request({ url: '/api/system/role/authUser/allocatedList', method: 'GET', data: { roleId, ...query } })
}

// 角色未分配用户（分页 + username）
export function listRoleUnallocatedUsers(roleId, query) {
  return request({ url: '/api/system/role/authUser/unallocatedList', method: 'GET', data: { roleId, ...query } })
}

// 批量分配用户：{ roleId, userIds }
export function assignRoleUsers(roleId, userIds) {
  return request({ url: '/api/system/role/authUser', method: 'PUT', data: { roleId, userIds } })
}

// 批量取消分配：{ roleId, userIds }
export function cancelRoleUsers(roleId, userIds) {
  return request({ url: '/api/system/role/authUser', method: 'DELETE', data: { roleId, userIds } })
}
