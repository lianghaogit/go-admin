const TOKEN_KEY = 'Admin-Token';

// 获取登录令牌
export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

// 保存登录令牌
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

// 清除登录令牌
export function removeToken() {
  localStorage.removeItem(TOKEN_KEY);
}
