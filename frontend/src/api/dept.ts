import request from '@/utils/request';

// 查询部门列表
export function listDept() {
  return request.get('/system/dept/list');
}

// 获取部门树
export function deptTreeSelect() {
  return request.get('/system/dept/treeselect');
}

// 获取部门详情
export function getDept(deptId) {
  return request.get('/system/dept/' + deptId);
}

// 新增部门
export function addDept(data) {
  return request.post('/system/dept', data);
}

// 修改部门
export function updateDept(data) {
  return request.put('/system/dept', data);
}

// 删除部门
export function delDept(deptId) {
  return request.delete('/system/dept/' + deptId);
}
