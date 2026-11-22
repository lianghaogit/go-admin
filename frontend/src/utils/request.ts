import axios from 'axios';
import type { AxiosRequestConfig, AxiosInstance } from 'axios';
import { ElMessage } from 'element-plus';
import { getToken, removeToken } from '@/utils/auth';

// 注意：不在此处静态引入 router，避免循环依赖
// （router/index.ts -> layout/index.vue -> 组件 -> api -> 本文件 -> router）
// 401 跳转时再动态引入 router

// ApiClient 类型：响应拦截器已解包业务响应体（{ code, msg, rows, total, data, ... }），
// 调用侧 await request.get/post/request 拿到的直接是业务体，而非 AxiosResponse
export interface ApiClient {
  (config: AxiosRequestConfig): Promise<any>;
  request<T = any>(config: AxiosRequestConfig): Promise<T>;
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>;
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>;
  head<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>;
  options<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  interceptors: AxiosInstance['interceptors'];
  defaults: AxiosInstance['defaults'];
}

// 创建 axios 实例
const service = axios.create({
  baseURL: '/api',
  timeout: 15000
});

// 请求拦截器：注入 token
service.interceptors.request.use(
  (config) => {
    if (getToken()) {
      config.headers['Authorization'] = 'Bearer ' + getToken();
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器：统一处理业务码
service.interceptors.response.use(
  (response) => {
    // 二进制响应（文件下载/导出）原样返回
    if (response.config.responseType === 'blob') {
      return response;
    }
    const res = response.data;
    if (res.code === 200) {
      return res;
    }
    if (res.code === 401) {
      removeToken();
      // 整页刷新跳转登录：彻底清空 pinia/tagsView 等内存态（对比 router.push 只清 URL）
      window.location.href = '/login';
      ElMessage.error(res.msg || '登录已过期，请重新登录');
      return Promise.reject(new Error(res.msg || '登录已过期'));
    }
    ElMessage.error(res.msg || '请求失败');
    return Promise.reject(new Error(res.msg || '请求失败'));
  },
  (error) => {
    const msg = error.response?.data?.msg || error.message || '网络异常';
    if (error.response?.status === 401) {
      removeToken();
      window.location.href = '/login';
    }
    ElMessage.error(msg);
    return Promise.reject(error);
  }
);

export default service as unknown as ApiClient;
