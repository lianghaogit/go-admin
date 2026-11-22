import request from '@/utils/request';

// ================= 流程定义 =================

// 查询流程定义分页列表
export function listFlowDef(params) {
  return request.get('/flow/def/list', { params });
}

// 查询全部流程定义（不分页）
export function listFlowDefAll() {
  return request.get('/flow/def/all');
}

// 启用中的流程定义（发起申请下拉）
export function listEnabledFlowDef() {
  return request.get('/flow/def/all');
}

// 获取流程定义详情（含节点）
export function getFlowDef(flowDefId) {
  return request.get('/flow/def/' + flowDefId);
}

// 新增流程定义
export function addFlowDef(data) {
  return request.post('/flow/def', data);
}

// 修改流程定义
export function updateFlowDef(data) {
  return request.put('/flow/def', data);
}

// 删除流程定义
export function delFlowDef(flowDefId) {
  return request.delete('/flow/def/' + flowDefId);
}

// 启停流程定义
export function changeFlowDefStatus(flowDefId, status) {
  return request.put('/flow/def/changeStatus', { flowDefId, status });
}

// ================= 发起申请 =================

// 我的申请列表
export function listFlowApply(params) {
  return request.get('/flow/apply/list', { params });
}

// 发起申请
export function addFlowApply(data) {
  return request.post('/flow/apply', data);
}

// 撤回申请
export function withdrawFlowApply(flowInstanceId) {
  return request.put('/flow/apply/withdraw/' + flowInstanceId);
}

// 重新提交
export function reapplyFlowApply(flowInstanceId, data) {
  return request.put('/flow/apply/reapply/' + flowInstanceId, data);
}

// 催办
export function urgeFlowApply(flowInstanceId) {
  return request.put('/flow/apply/urge/' + flowInstanceId);
}

// 终止申请（管理员）
export function terminateFlowApply(flowInstanceId, data) {
  return request.put('/flow/apply/terminate/' + flowInstanceId, data);
}

// ================= 待办审批 =================

// 我的待办列表
export function listFlowTodo(params) {
  return request.get('/flow/todo/list', { params });
}

// 审批通过
export function approveFlowTask(data) {
  return request.put('/flow/todo/approve', data);
}

// 审批驳回
export function rejectFlowTask(data) {
  return request.put('/flow/todo/reject', data);
}

// 转办
export function transferFlowTask(data) {
  return request.put('/flow/todo/transfer', data);
}

// 加签
export function addSignFlowTask(data) {
  return request.put('/flow/todo/addSign', data);
}

// ================= 已办任务 =================

// 我已办列表
export function listFlowDone(params) {
  return request.get('/flow/done/list', { params });
}

// ================= 抄送我的 =================

// 抄送列表
export function listFlowCc(params) {
  return request.get('/flow/cc/list', { params });
}

// 标记抄送已读
export function readFlowCc(flowTaskId) {
  return request.put('/flow/cc/read/' + flowTaskId);
}

// ================= 实例 =================

// 全部实例列表（管理员）
export function listFlowInstance(params) {
  return request.get('/flow/instance/list', { params });
}

// 实例详情（实例+节点+任务+记录时间线）
export function getFlowInstance(flowInstanceId) {
  return request.get('/flow/instance/' + flowInstanceId);
}
