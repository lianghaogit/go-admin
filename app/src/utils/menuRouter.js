// 菜单「页面」类 path → App 本地业务页路由映射
// 后端菜单 path 为相对路径（如 user / role，父目录 system 仅用于前端路由组装），
// 故 PAGE_MAP 以菜单节点自身的 m.path 为键匹配；未登记的保持「请在 Web 端操作」提示
const PAGE_MAP = {
  user: '/pages/system/user/index',
  role: '/pages/system/role/index',
  menu: '/pages/system/menu/index',
  dept: '/pages/system/dept/index',
  dict: '/pages/system/dict/index',
  post: '/pages/system/post/index',
  config: '/pages/system/config/index',
  notice: '/pages/system/notice/index',
  loginlog: '/pages/system/loginlog/index',
  operlog: '/pages/system/operlog/index',
  online: '/pages/system/online/index',
  api: '/pages/system/api/index',
  job: '/pages/system/job/index',
  server: '/pages/system/server/index',
  cache: '/pages/system/cache/index',
  db: '/pages/system/db/index',
  apply: '/pages/system/flow/apply/index',
  done: '/pages/system/flow/done/index',
  cc: '/pages/system/flow/cc/index',
  todo: '/pages/system/flow/todo/index',
  def: '/pages/system/flow/def/index',
  gen: '/pages/system/tool/gen/index',
  'api-doc': '/pages/system/tool/api-doc/index',
  upload: '/pages/system/tool/upload/index'
}

// 打开页面菜单：能映射则跳转并返回 true；否则返回 false（由调用方弹提示）
export function openPageMenu(m) {
  const page = PAGE_MAP[m && m.path]
  if (page) {
    uni.navigateTo({ url: page })
    return true
  }
  return false
}
