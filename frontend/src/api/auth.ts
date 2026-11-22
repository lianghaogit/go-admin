import request from '@/utils/request';

// 用户登录
export function login(data) {
  return request.post('/login', data);
}

// 获取当前用户信息与权限
export function getInfo() {
  return request.get('/getInfo');
}

// 获取当前用户路由
export function getRouters() {
  return request.get('/getRouters');
}

// 用户退出
export function logout() {
  return request.post('/logout');
}

// 获取图形验证码
export function getCaptchaImage() {
  return request.get('/captchaImage');
}
