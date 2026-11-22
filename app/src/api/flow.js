// 工作流接口
import { request } from '@/utils/request'

// 启用中的流程定义（发起申请下拉）
export function listEnabledDefs() {
  return request({ url: '/api/flow/def/all', method: 'GET' })
}

// 流程定义列表
export function listFlowDefs(query) {
  return request({ url: '/api/flow/def/list', method: 'GET', data: query })
}

// 流程启停：{ flowDefId, status }
export function changeFlowDefStatus(data) {
  return request({ url: '/api/flow/def/changeStatus', method: 'PUT', data })
}

// 我的申请（keyword + status）
export function listFlowApply(query) {
  return request({ url: '/api/flow/apply/list', method: 'GET', data: query })
}

// 发起申请：{ flowDefId, title, content, ccUserIds }
export function startFlow(data) {
  return request({ url: '/api/flow/apply', method: 'POST', data })
}

// 撤回申请
export function withdrawFlow(id) {
  return request({ url: `/api/flow/apply/withdraw/${id}`, method: 'PUT' })
}

// 驳回后重新提交：{ title, content }
export function reapplyFlow(id, data) {
  return request({ url: `/api/flow/apply/reapply/${id}`, method: 'PUT', data })
}

// 催办
export function urgeFlow(id) {
  return request({ url: `/api/flow/apply/urge/${id}`, method: 'PUT' })
}

// 终止申请：{ comment }
export function terminateFlow(id, data) {
  return request({ url: `/api/flow/apply/terminate/${id}`, method: 'PUT', data })
}

// 我处理的（已办）
export function listFlowDone(query) {
  return request({ url: '/api/flow/done/list', method: 'GET', data: query })
}

// 抄送我的
export function listFlowCc(query) {
  return request({ url: '/api/flow/cc/list', method: 'GET', data: query })
}

// 抄送标记已读
export function markCcRead(id) {
  return request({ url: `/api/flow/cc/read/${id}`, method: 'PUT' })
}

// 我的待办（keyword 按标题过滤）
export function listFlowTodo(query) {
  return request({ url: '/api/flow/todo/list', method: 'GET', data: query })
}

// 审批通过：{ taskId, comment }
export function approveFlowTask(data) {
  return request({ url: '/api/flow/todo/approve', method: 'PUT', data })
}

// 驳回：{ taskId, comment }
export function rejectFlowTask(data) {
  return request({ url: '/api/flow/todo/reject', method: 'PUT', data })
}

// 转办：{ taskId, toUserId, comment }
export function transferFlowTask(data) {
  return request({ url: '/api/flow/todo/transfer', method: 'PUT', data })
}

// 加签：{ taskId, toUserId, comment }
export function addSignFlowTask(data) {
  return request({ url: '/api/flow/todo/addSign', method: 'PUT', data })
}

// 实例详情（含节点/任务/记录）
export function getFlowInstanceDetail(id) {
  return request({ url: `/api/flow/instance/${id}`, method: 'GET' })
}
