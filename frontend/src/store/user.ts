import { defineStore } from 'pinia';
import { login, getInfo, getRouters, logout } from '@/api/auth';
import { setToken, removeToken } from '@/utils/auth';

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    name: '',
    avatar: '',
    roles: [],
    permissions: [],
    menus: []
  }),
  actions: {
    // 登录
    async login(userInfo) {
      const res = await login(userInfo);
      setToken(res.data.token);
      this.token = res.data.token;
    },
    // 获取用户信息（角色、权限）
    async getInfo() {
      const res = await getInfo();
      this.name = res.data.user.nickName;
      this.avatar = res.data.user.avatar || '';
      this.roles = res.data.roles;
      this.permissions = res.data.permissions;
      return res.data;
    },
    // 获取动态路由菜单
    async getRouters() {
      const res = await getRouters();
      this.menus = res.data;
      return res.data;
    },
    // 退出登录
    async logout() {
      try {
        await logout();
      } catch (e) {
        // 忽略退出接口异常
      }
      removeToken();
      this.token = '';
      this.name = '';
      this.avatar = '';
      this.roles = [];
      this.permissions = [];
      this.menus = [];
      // 整页刷新：重置动态路由与全部内存态（避免多账号切换时动态路由重复注册）
      window.location.href = '/login';
    }
  }
});
