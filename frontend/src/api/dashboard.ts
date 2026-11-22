import request from '@/utils/request';

// 首页统计汇总
export function getDashStats() {
  return request({ url: '/dashboard/stats', method: 'get' });
}

// 近 7 天登录趋势
export function getLoginTrend() {
  return request({ url: '/dashboard/loginTrend', method: 'get' });
}

// 近 7 天操作类型分布
export function getOperTypeStats() {
  return request({ url: '/dashboard/operTypeStats', method: 'get' });
}

// 近 7 天登录热力（天 × 小时）
export function getLoginHeatmap() {
  return request({ url: '/dashboard/loginHeatmap', method: 'get' });
}
