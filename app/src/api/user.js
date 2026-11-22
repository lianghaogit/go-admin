// 用户管理接口（对齐后端 /api/system/user 路由）
import { request } from '@/utils/request'

// 用户列表（分页 + 关键词/状态）
// query: { pageNum, pageSize, username, phone, status }
export function listUser(query) {
  return request({ url: '/api/system/user/list', method: 'GET', data: query })
}

// 用户详情（含 dept / roles / posts，用于编辑回填）
export function getUser(userId) {
  return request({ url: `/api/system/user/${userId}`, method: 'GET' })
}

// 当前用户个人信息
export function getProfile() {
  return request({ url: '/api/system/user/profile', method: 'GET' })
}

// 修改个人信息：{ nickName, email, phonenumber, remark }
export function updateProfile(data) {
  return request({ url: '/api/system/user/profile', method: 'PUT', data })
}

// 修改密码：{ oldPassword, newPassword }
export function updatePwd(data) {
  return request({ url: '/api/system/user/profile/updatePwd', method: 'PUT', data })
}

// 重置用户密码：{ userId, password }
export function resetUserPwd(userId, password) {
  return request({ url: '/api/system/user/resetPwd', method: 'PUT', data: { userId, password } })
}

// 更换头像：{ avatar: '/uploads/...' }
export function updateAvatar(avatar) {
  return request({ url: '/api/system/user/profile/avatar', method: 'PUT', data: { avatar } })
}

// 新增用户：{ username, password, nickName, email, phonenumber, status, remark, deptId, roleIds, postIds }
export function createUser(data) {
  return request({ url: '/api/system/user', method: 'POST', data })
}

// 修改用户：{ userId, nickName, email, phonenumber, status, remark, deptId, roleIds, postIds }
export function updateUser(data) {
  return request({ url: '/api/system/user', method: 'PUT', data })
}

// 删除用户
export function deleteUser(userId) {
  return request({ url: `/api/system/user/${userId}`, method: 'DELETE' })
}

// 修改用户状态：{ userId, status }
export function changeUserStatus(userId, status) {
  return request({ url: '/api/system/user/changeStatus', method: 'PUT', data: { userId, status } })
}
