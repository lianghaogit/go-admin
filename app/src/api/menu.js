// 菜单管理接口
import { request } from '@/utils/request'

// 菜单树（无分页，data 为树形数组）
export function listMenu() {
  return request({ url: '/api/system/menu/list', method: 'GET' })
}

// 菜单下拉树（仅结构与名称，用于角色授权）
export function menuTree() {
  return request({ url: '/api/system/menu/treeselect', method: 'GET' })
}

// 新增菜单：{ menuName, parentId, orderNum, path, component, menuType, perms, icon, visible, status }
export function createMenu(data) {
  return request({ url: '/api/system/menu', method: 'POST', data })
}

// 修改菜单：{ menuId, ... }
export function updateMenu(data) {
  return request({ url: '/api/system/menu', method: 'PUT', data })
}

// 删除菜单
export function deleteMenu(menuId) {
  return request({ url: `/api/system/menu/${menuId}`, method: 'DELETE' })
}
