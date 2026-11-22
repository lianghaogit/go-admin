import { defineStore } from 'pinia';

// 多页签标签管理 + 页面缓存（对标 RuoYi tagsView）
export const useTagsViewStore = defineStore('tagsView', {
  state: () => ({
    visitedViews: [],
    cachedViews: [],
    refreshKey: 0
  }),
  actions: {
    // 添加标签（登录页/无标题路由不收录；首页固定不可关）并登记缓存
    addView(route) {
      if (!route || route.path === '/login' || !route.meta || !route.meta.title) {
        return;
      }
      const isHome = route.path === '/index' || route.path === '/';
      if (this.visitedViews.some((v) => v.path === route.path)) return;
      this.visitedViews.push({
        path: route.path,
        name: route.name,
        title: route.meta.title,
        affix: isHome,
        meta: { title: route.meta.title, icon: route.meta.icon }
      });
      // 登记组件缓存（keep-alive include 匹配组件 name）
      const compName = this.getComponentName(route.path);
      if (compName && !this.cachedViews.includes(compName)) {
        this.cachedViews.push(compName);
      }
    },
    // 删除标签（首页不可删）并同步移除缓存
    delView(view) {
      this.visitedViews = this.visitedViews.filter((v) => {
        if (v.path === view.path) {
          return v.affix;
        }
        return true;
      });
      this.removeCached(view.path);
    },
    // 关闭其他标签（保留当前 + 固定首页）
    delOthersViews(view) {
      const keep = this.visitedViews
        .filter((v) => v.affix || v.path === view.path)
        .map((v) => v.path);
      this.visitedViews = this.visitedViews.filter((v) => keep.includes(v.path));
      this.removeCachedExcept(keep);
    },
    // 关闭全部标签（仅保留固定首页）
    delAllViews() {
      const keep = this.visitedViews.filter((v) => v.affix).map((v) => v.path);
      this.visitedViews = this.visitedViews.filter((v) => v.affix);
      this.removeCachedExcept(keep);
    },
    // 关闭右侧标签
    delRightTags(view) {
      const index = this.visitedViews.findIndex((v) => v.path === view.path);
      if (index === -1) return;
      const keep = this.visitedViews.filter((v, i) => v.affix || i <= index).map((v) => v.path);
      this.visitedViews = this.visitedViews.filter((v) => keep.includes(v.path));
      this.removeCachedExcept(keep);
    },
    // 刷新当前视图（AppMain 通过 key 重新挂载）
    refreshView() {
      this.refreshKey++;
    },
    // 移除指定路径对应的组件缓存
    removeCached(path) {
      const compName = this.getComponentName(path);
      if (compName) {
        this.cachedViews = this.cachedViews.filter((n) => n !== compName);
      }
    },
    // 只保留 keep 集合内路径对应的缓存
    removeCachedExcept(keepPaths) {
      const keepNames = keepPaths.map((p) => this.getComponentName(p)).filter(Boolean);
      this.cachedViews = this.cachedViews.filter((n) => keepNames.includes(n));
    },
    // 路由路径 -> 组件 name（与 views 页面 defineOptions name 一致）
    getComponentName(path) {
      const map = {
        '/index': 'Dashboard',
        '/system/menu': 'SystemMenu',
        '/system/log/loginlog': 'SystemLoginLog',
        '/system/dept': 'SystemDept',
        '/system/dict': 'SystemDict',
        '/system/log/operlog': 'SystemOperLog',
        '/system/post': 'SystemPost',
        '/system/config': 'SystemConfig',
        '/system/notice': 'SystemNotice',
        '/monitor/online': 'MonitorOnline',
        '/monitor/job': 'MonitorJob',
        '/monitor/server': 'MonitorServer',
        '/monitor/cache': 'MonitorCache',
        '/monitor/db': 'MonitorDb',
        '/system/role': 'SystemRole',
        '/system/api': 'SystemApi',
        '/tool/gen': 'ToolGen',
        '/system/user': 'SystemUser',
        '/tool/api-doc': 'ToolApiDoc',
        '/tool/upload': 'ToolUpload',
        '/profile': 'Profile'
      };
      return map[path] || '';
    }
  }
});
