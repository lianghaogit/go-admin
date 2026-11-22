// 定时任务接口
import { request } from '@/utils/request'

// 任务列表（分页 + jobName/jobGroup/status）
export function listJob(query) {
  return request({ url: '/api/monitor/job/list', method: 'GET', data: query })
}

// 内置任务执行器（调用目标下拉）
export function listJobExecutors() {
  return request({ url: '/api/monitor/job/executors', method: 'GET' })
}

// 任务详情
export function getJob(jobId) {
  return request({ url: `/api/monitor/job/${jobId}`, method: 'GET' })
}

// 新增任务
export function createJob(data) {
  return request({ url: '/api/monitor/job', method: 'POST', data })
}

// 修改任务
export function updateJob(data) {
  return request({ url: '/api/monitor/job', method: 'PUT', data })
}

// 启停任务：{ jobId, status }
export function changeJobStatus(data) {
  return request({ url: '/api/monitor/job/changeStatus', method: 'PUT', data })
}

// 立即执行一次：{ jobId }
export function runJob(jobId) {
  return request({ url: '/api/monitor/job/run', method: 'PUT', data: { jobId } })
}

// 删除任务
export function deleteJob(jobId) {
  return request({ url: `/api/monitor/job/${jobId}`, method: 'DELETE' })
}

// 执行日志列表（分页 + jobName/status）
export function listJobLog(query) {
  return request({ url: '/api/monitor/jobLog/list', method: 'GET', data: query })
}

// 删除单条执行日志
export function deleteJobLog(id) {
  return request({ url: `/api/monitor/jobLog/${id}`, method: 'DELETE' })
}

// 清空执行日志
export function cleanJobLog() {
  return request({ url: '/api/monitor/jobLog/clean', method: 'DELETE' })
}
