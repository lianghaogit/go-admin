import request from '@/utils/request';

// 查询菜单列表
export function listMenu() {
  return request.get('/system/menu/list');
}

// 获取菜单树
export function menuTreeselect() {
  return request.get('/system/menu/treeselect');
}

// 获取菜单详情
export function getMenu(menuId) {
  return request.get('/system/menu/' + menuId);
}

// 新增菜单
export function addMenu(data) {
  return request.post('/system/menu', data);
}

// 修改菜单
export function updateMenu(data) {
  return request.put('/system/menu', data);
}

// 删除菜单
export function delMenu(menuId) {
  return request.delete('/system/menu/' + menuId);
}
