import request from '@/utils/request';

// 系统接口文档
export function listApiDoc() {
  return request.get('/tool/api-doc/list', { timeout: 30000 });
}

// 通用文件上传
export function uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);
  return request.post('/common/upload', formData, { timeout: 60000 });
}
