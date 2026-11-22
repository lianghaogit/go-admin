import { useI18n } from 'vue-i18n';

// 后端菜单/路由标题为中文，国际化时按「中文标题 → key」映射翻译；
// 未收录的标题（用户自定义菜单）保持原文
const zhToKey: Record<string, string> = {
  首页: 'home',
  系统管理: 'system',
  用户管理: 'user',
  角色管理: 'role',
  菜单管理: 'menu',
  部门管理: 'dept',
  岗位管理: 'post',
  字典管理: 'dict',
  参数设置: 'config',
  参数配置: 'config',
  日志管理: 'logmgmt',
  通知公告: 'notice',
  系统监控: 'monitor',
  在线用户: 'online',
  定时任务: 'job',
  服务监控: 'server',
  缓存监控: 'cache',
  数据监控: 'db',
  登录日志: 'loginlog',
  操作日志: 'operlog',
  系统工具: 'tool',
  代码生成: 'gen',
  接口文档: 'apiDoc',
  文件上传: 'upload',
  个人中心: 'profile',
  接口权限: 'api',
  角色分配: 'role',
  分配用户: 'user',
  流程中心: 'flowCenter',
  流程定义: 'flowDef',
  发起申请: 'flowApply',
  待办审批: 'flowTodo',
  已办任务: 'flowDone',
  抄送我的: 'flowCc'
};

/**
 * 菜单/面包屑标题国际化：
 * - 当前语言为 en 且标题命中映射 → 返回对应英文
 * - 其余情况返回原文（中文标题照常显示）
 */
export function useMenuTitle() {
  const { t, locale } = useI18n();
  return (title: unknown): string => {
    const s = String(title ?? '');
    if (!s) return '';
    if (locale.value === 'en' && zhToKey[s]) {
      return t(`menu.${zhToKey[s]}`);
    }
    return s;
  };
}
