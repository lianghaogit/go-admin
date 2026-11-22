import request from '@/utils/request';

// 查询岗位列表
export function listPost(params) {
  return request.get('/system/post/list', { params });
}

// 查询全部岗位列表
export function listAllPost() {
  return request.get('/system/post/all');
}

// 获取岗位详情
export function getPost(postId) {
  return request.get('/system/post/' + postId);
}

// 新增岗位
export function addPost(data) {
  return request.post('/system/post', data);
}

// 修改岗位
export function updatePost(data) {
  return request.put('/system/post', data);
}

// 删除岗位
export function delPost(postId) {
  return request.delete('/system/post/' + postId);
}
