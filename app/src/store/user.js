// 用户状态（pinia）：token / 用户信息 / 角色 / 权限 / 菜单
import { defineStore } from 'pinia'
import { login as loginApi, getInfo as getInfoApi, logout as logoutApi, getRouters as getRoutersApi } from '@/api/login'
import { getToken, setToken, removeToken } from '@/utils/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: getToken(),
    userInfo: null,
    roles: [],
    permissions: [],
    menus: []
  }),
  actions: {
    // 登录：写 token
    async login(loginForm) {
      const res = await loginApi(loginForm)
      setToken(res.data.token)
      this.token = res.data.token
    },

    // 拉取用户信息/角色/权限
    async getInfo() {
      const res = await getInfoApi()
      this.userInfo = res.data.user
      this.roles = res.data.roles || []
      this.permissions = res.data.permissions || []
    },

    // 拉取动态菜单树
    async getRouters() {
      const res = await getRoutersApi()
      this.menus = res.data || []
    },

    // 登录后的初始化（信息 + 菜单）
    async init() {
      await this.getInfo()
      await this.getRouters()
    },

    // 权限判断（admin / *:*:* 放行；支持 hasPerm('system:user:list')）
    hasPerm(perm) {
      if (this.roles.includes('admin')) return true
      return this.permissions.includes('*:*:*') || this.permissions.includes(perm)
    },

    // 退出登录
    async logout() {
      try {
        await logoutApi()
      } catch (e) {
        // 登出接口失败不阻塞本地清理
      }
      this.reset()
    },

    reset() {
      removeToken()
      this.token = ''
      this.userInfo = null
      this.roles = []
      this.permissions = []
      this.menus = []
    }
  }
})
