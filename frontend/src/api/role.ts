import request from '@/utils/request';

// 查询角色列表
export function listRole(params) {
  return request.get('/system/role/list', { params });
}

// 查询全部角色列表
export function listAllRole() {
  return request.get('/system/role/all');
}

// 获取角色详情
export function getRole(roleId) {
  return request.get('/system/role/' + roleId);
}

// 新增角色
export function addRole(data) {
  return request.post('/system/role', data);
}

// 修改角色
export function updateRole(data) {
  return request.put('/system/role', data);
}

// 删除角色
export function delRole(roleId) {
  return request.delete('/system/role/' + roleId);
}

// 获取角色菜单树
export function roleMenuTreeselect(roleId) {
  return request.get('/system/menu/roleMenuTreeselect/' + roleId);
}

// ---------- 角色分配用户 ----------
export function allocatedUserList(roleId, params) {
  return request.get('/system/role/authUser/allocatedList', { params: { ...params, roleId } });
}

// 查询未分配用户列表
export function unallocatedUserList(roleId, params) {
  return request.get('/system/role/authUser/unallocatedList', { params: { ...params, roleId } });
}

// 分配用户给角色
export function assignAuthUsers(roleId, userIds) {
  return request.put('/system/role/authUser', { roleId, userIds });
}

// 取消分配的用户
export function cancelAuthUsers(roleId, userIds) {
  return request.delete('/system/role/authUser', { data: { roleId, userIds } });
}
