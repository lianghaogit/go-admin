import { useUserStore } from '@/store/user';

// 按钮权限指令 v-hasPermi="['system:user:add']"
export default {
  mounted(el, binding) {
    const { value } = binding;
    const userStore = useUserStore();
    const permissions = userStore.permissions;
    if (Array.isArray(value) && value.length) {
      const hasPermission = value.some(
        (p) => permissions.includes(p) || permissions.includes('*:*:*')
      );
      if (!hasPermission) {
        // 无权限时从 DOM 移除该按钮
        if (el.parentNode) {
          el.parentNode.removeChild(el);
        }
      }
    } else {
      throw new Error(`请设置操作权限标签值，如 v-hasPermi="['system:user:add']"`);
    }
  }
};
