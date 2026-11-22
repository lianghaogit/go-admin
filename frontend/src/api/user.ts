import request from '@/utils/request';
import type { PageResult, DataResult, PlainResult } from '@/types/api';

// 用户行数据（列表/详情返回；可选字段按后端实际返回）
export interface UserItem {
  userId?: number;
  deptId?: number;
  userName?: string;
  nickName?: string;
  avatar?: string;
  phone?: string;
  email?: string;
  status?: string;
  remark?: string;
  createTime?: string;
  dept?: { deptId?: number; deptName?: string };
  posts?: Array<{ postId?: number; postName?: string }>;
  roles?: Array<{ roleId?: number; roleName?: string; roleKey?: string }>;
}

// 查询用户列表
export function listUser(params): Promise<PageResult<UserItem>> {
  return request.get('/system/user/list', { params });
}

// 获取用户详情
export function getUser(userId): Promise<DataResult<UserItem>> {
  return request.get('/system/user/' + userId);
}

// 新增用户
export function addUser(data): Promise<PlainResult> {
  return request.post('/system/user', data);
}

// 修改用户
export function updateUser(data): Promise<PlainResult> {
  return request.put('/system/user', data);
}

// 删除用户
export function delUser(userIds): Promise<PlainResult> {
  // 支持单个或多个（逗号分隔批量删除）
  return request.delete('/system/user/' + userIds);
}

// 修改用户状态
export function changeUserStatus(userId, status): Promise<PlainResult> {
  return request.put('/system/user/changeStatus', { userId, status });
}

// 分配角色给用户
export function authRole(data): Promise<PlainResult> {
  // 分配角色 { userId, roleIds }
  return request.put('/system/user/authRole', data);
}

// 重置用户密码
export function resetUserPwd(userId, password): Promise<PlainResult> {
  return request.put('/system/user/resetPwd', { userId, password });
}

// ---------- 导入导出 ----------
export function exportUser(params) {
  return request.get('/system/user/export', { params, responseType: 'blob', timeout: 30000 });
}

// 下载用户导入模板
export function importUserTemplate() {
  return request.get('/system/user/importTemplate', { responseType: 'blob', timeout: 30000 });
}

// 导入用户
export function importUser(data) {
  return request.post('/system/user/importData', data, { timeout: 60000 });
}

// ---------- 个人中心 ----------
export function getProfile() {
  return request.get('/system/user/profile');
}

// 修改个人资料
export function updateProfile(data) {
  return request.put('/system/user/profile', data);
}

// 更新头像
export function updateAvatar(avatar) {
  return request.put('/system/user/profile/avatar', { avatar });
}

// 修改个人密码
export function updatePwd(oldPassword, newPassword) {
  return request.put('/system/user/profile/updatePwd', { oldPassword, newPassword });
}
