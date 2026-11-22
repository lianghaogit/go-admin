import request from '@/utils/request';
import axios from 'axios';
import { getToken } from '@/utils/auth';

// 未导入的数据库表列表
export function listDbTables(params) {
  return request.get('/tool/gen/db/list', { params });
}

// 导入表（表名逗号分隔）
export function importGenTables(tables) {
  return request.post('/tool/gen/importTable?tables=' + encodeURIComponent(tables), {});
}

// 生成配置列表
export function listGenTable(params) {
  return request.get('/tool/gen/list', { params });
}

// 生成配置详情（含列）
export function getGenTable(tableId) {
  return request.get('/tool/gen/' + tableId);
}

// 保存生成配置
export function updateGenTable(data) {
  return request.put('/tool/gen', data);
}

// 删除生成配置
export function delGenTable(tableId) {
  return request.delete('/tool/gen/' + tableId);
}

// 预览生成代码
export function previewGenTable(tableId) {
  return request.get('/tool/gen/preview/' + tableId);
}

// 下载生成代码（zip，blob 直连绕过统一拦截器）
export function downloadGenCode(tableId) {
  return axios({
    url: '/api/tool/gen/code/' + tableId,
    method: 'get',
    responseType: 'blob',
    headers: { Authorization: 'Bearer ' + getToken() }
  });
}
