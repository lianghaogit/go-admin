import router from '@/router';
import { getToken } from '@/utils/auth';
import { useUserStore } from '@/store/user';
import Layout from '@/layout/index.vue';
// 多级目录透传组件（目录 > 目录 > 菜单 场景避免嵌套 Layout 双栏）
import ParentView from '@/views/layout/parentView.vue';
import { startProgress, finishProgress, failProgress } from '@/utils/progress';
import { ElMessage } from 'element-plus';

// 免登录白名单
const whiteList = ['/login', '/404'];

// 动态路由是否已注册（独立标记：零角色用户也必须置位，否则每次导航都会重复拉取并回跳）
let routesLoaded = false;

// 预加载所有视图组件（相对本文件所在 src/ 目录）
const viewModules = import.meta.glob('./views/**/*.vue');

// 根据后端 component 路径加载组件
function loadComponent(component) {
  const path = './views/' + component + '.vue';
  if (!viewModules[path]) {
    // 组件路径不匹配时降级 404 并给出诊断提示（便于排查菜单 component 配置错误）
    console.warn(`[permission] 菜单组件路径未找到: ${component}`);
    return () => import('@/views/error/404.vue');
  }
  return viewModules[path];
}

// 后端菜单 -> 前端路由（RuoYi 风格动态路由，支持多级目录）
function buildRoutes(menus) {
  const routes = [];
  menus.forEach((menu) => {
    if (menu.menuType === 'M') {
      // 目录：递归构建子路由（支持 目录 > 目录 > 菜单 多级）
      const children = buildChildren(menu.children || [], '/' + menu.path);
      routes.push({
        path: '/' + menu.path,
        component: Layout,
        redirect: children.length ? '/' + menu.path + '/' + children[0].path : undefined,
        name: menu.menuName,
        meta: { title: menu.menuName, icon: menu.icon },
        children
      });
    } else if (menu.menuType === 'C') {
      // 菜单
      routes.push({
        path: '/' + menu.path,
        component: Layout,
        children: [
          {
            path: '',
            name: menu.menuName,
            component: loadComponent(menu.component),
            meta: { title: menu.menuName, icon: menu.icon }
          }
        ]
      });
    }
  });
  routes.push({ path: '/:pathMatch(.*)*', redirect: '/404' });
  return routes;
}

// 递归构建子路由：C 菜单为叶子；M 目录用 ParentView 透传（避免嵌套 Layout 双栏），其子菜单继续递归
function buildChildren(list, base) {
  const out = [];
  for (const ch of list || []) {
    if (ch.menuType === 'F') continue;
    if (ch.menuType === 'M') {
      const sub = buildChildren(ch.children || [], base + '/' + ch.path);
      out.push({
        path: ch.path,
        component: ParentView,
        // 目录直达用绝对路径重定向（相对路径在嵌套路由下解析错误）
        redirect: sub.length ? base + '/' + ch.path + '/' + sub[0].path : undefined,
        name: ch.menuName,
        meta: { title: ch.menuName, icon: ch.icon },
        children: sub
      });
    } else {
      out.push({
        path: ch.path,
        name: ch.menuName,
        component: loadComponent(ch.component),
        meta: { title: ch.menuName, icon: ch.icon }
      });
    }
  }
  return out;
}

// 全局路由守卫（带顶部进度条）
router.beforeEach(async (to, from, next) => {
  startProgress();
  const userStore = useUserStore();
  if (getToken()) {
    if (to.path === '/login') {
      next({ path: '/' });
    } else if (routesLoaded) {
      next();
    } else {
      try {
        await userStore.getInfo();
        const menus = await userStore.getRouters();
        buildRoutes(menus).forEach((r) => router.addRoute(r));
        routesLoaded = true;
        next({ ...to, replace: true });
      } catch (err) {
        await userStore.logout();
        ElMessage.error(err.message || '获取用户信息失败，请重新登录');
        next('/login');
      }
    }
  } else {
    if (whiteList.includes(to.path)) {
      next();
    } else {
      next(`/login?redirect=${encodeURIComponent(to.fullPath)}`);
    }
  }
});

router.afterEach((to) => {
  finishProgress();
  // 动态页面标题（对标 RuoYi：模块名 - 系统名）
  const baseTitle = 'Go Admin';
  document.title = to.meta?.title ? `${to.meta.title} - ${baseTitle}` : baseTitle;
});

router.onError(() => {
  failProgress();
});

export default router;
