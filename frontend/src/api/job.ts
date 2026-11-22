import request from '@/utils/request';

// 查询定时任务列表
export function listJob(params) {
  return request.get('/monitor/job/list', { params });
}

// 内置任务执行器列表（调用目标下拉）
export function listExecutors() {
  return request.get('/monitor/job/executors');
}

// 获取定时任务详情
export function getJob(jobId) {
  return request.get('/monitor/job/' + jobId);
}

// 新增定时任务
export function addJob(data) {
  return request.post('/monitor/job', data);
}

// 修改定时任务
export function updateJob(data) {
  return request.put('/monitor/job', data);
}

// 删除定时任务
export function delJob(jobId) {
  return request.delete('/monitor/job/' + jobId);
}

// 修改定时任务状态
export function changeJobStatus(jobId, status) {
  return request.put('/monitor/job/changeStatus', { jobId, status });
}

// 执行定时任务
export function runJob(jobId) {
  return request.put('/monitor/job/run', { jobId });
}

// 查询定时任务列表
export function listJobLog(params) {
  return request.get('/monitor/jobLog/list', { params });
}

// 删除定时任务
export function delJobLog(jobLogId) {
  return request.delete('/monitor/jobLog/' + jobLogId);
}

// 清空定时任务
export function cleanJobLog() {
  return request.delete('/monitor/jobLog/clean');
}
