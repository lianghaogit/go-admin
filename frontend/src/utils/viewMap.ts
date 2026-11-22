// 路由路径 -> 页面组件映射（桌面布局窗口按 path 解析自身组件）
// 与 store/tagsView.js getComponentName 的 path 集合保持一致
import Dashboard from '@/views/dashboard/index.vue';
import SystemUser from '@/views/system/user/index.vue';
import SystemRole from '@/views/system/role/index.vue';
import SystemMenu from '@/views/system/menu/index.vue';
import SystemDept from '@/views/system/dept/index.vue';
import SystemPost from '@/views/system/post/index.vue';
import SystemDict from '@/views/system/dict/index.vue';
import SystemConfig from '@/views/system/config/index.vue';
import SystemNotice from '@/views/system/notice/index.vue';
import SystemApi from '@/views/system/api/index.vue';
import SystemLoginLog from '@/views/system/loginlog/index.vue';
import SystemOperLog from '@/views/system/operlog/index.vue';
import MonitorOnline from '@/views/monitor/online/index.vue';
import MonitorJob from '@/views/monitor/job/index.vue';
import MonitorServer from '@/views/monitor/server/index.vue';
import MonitorCache from '@/views/monitor/cache/index.vue';
import MonitorDb from '@/views/monitor/db/index.vue';
import ToolGen from '@/views/tool/gen/index.vue';
import ToolApiDoc from '@/views/tool/api-doc/index.vue';
import ToolUpload from '@/views/tool/upload/index.vue';
import Profile from '@/views/profile/index.vue';

const viewMap = {
  '/index': Dashboard,
  '/system/user': SystemUser,
  '/system/role': SystemRole,
  '/system/menu': SystemMenu,
  '/system/dept': SystemDept,
  '/system/post': SystemPost,
  '/system/dict': SystemDict,
  '/system/config': SystemConfig,
  '/system/notice': SystemNotice,
  '/system/api': SystemApi,
  '/system/log/loginlog': SystemLoginLog,
  '/system/log/operlog': SystemOperLog,
  '/monitor/online': MonitorOnline,
  '/monitor/job': MonitorJob,
  '/monitor/server': MonitorServer,
  '/monitor/cache': MonitorCache,
  '/monitor/db': MonitorDb,
  '/tool/gen': ToolGen,
  '/tool/api-doc': ToolApiDoc,
  '/tool/upload': ToolUpload,
  '/profile': Profile
};

// 将组件路径解析为异步组件（找不到返回 null）
export function resolveView(path) {
  return viewMap[path] || null;
}
