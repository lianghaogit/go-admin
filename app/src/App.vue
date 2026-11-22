<script>
import { useUserStore } from '@/store/user'

export default {
  onLaunch: function () {
    // 启动时若已有 token，拉取用户信息/菜单恢复登录态
    const userStore = useUserStore()
    if (userStore.token) {
      userStore.init().catch(() => {
        userStore.reset()
        uni.reLaunch({ url: '/pages/login/index' })
      })
    }
  },
  onShow: function () {
    // 未登录直接回登录页
    const userStore = useUserStore()
    if (!userStore.token) {
      uni.reLaunch({ url: '/pages/login/index' })
    }
  },
  onHide: function () {
  },
}
</script>

<style>
/* ===== 全局设计基座 ===== */
page {
  background-color: #f4f6fb;
  font-size: 28rpx;
  color: #263243;
  -webkit-font-smoothing: antialiased;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Helvetica Neue", "Microsoft YaHei", sans-serif;
}

/* 统一卡片阴影 / 圆角观感 */
view,
text,
input {
  box-sizing: border-box;
}

/* 全局触摸反馈 */
.menu-item,
.menu-cell,
.user-item,
.crumb-link,
.search-btn,
.btn,
.logout-btn,
.role-item,
.func-cell,
.tool-btn,
.op,
.list-item,
.overview-cell,
.stat-card,
.quick-tile {
  transition: transform 0.16s ease, opacity 0.16s ease;
}

.menu-item:active,
.menu-cell:active,
.user-item:active,
.crumb-link:active,
.search-btn:active,
.btn:active,
.logout-btn:active,
.role-item:active,
.func-cell:active,
.tool-btn:active,
.op:active,
.list-item:active,
.overview-cell:active,
.stat-card:active,
.quick-tile:active {
  opacity: 0.86;
  transform: scale(0.97);
}

/* 遮罩淡入 + 底部弹层上滑（作用于所有页面同名结构） */
.mask {
  animation: mask-in 0.2s ease both;
}

.sheet {
  animation: sheet-up 0.26s cubic-bezier(0.22, 0.9, 0.32, 1) both;
}

@keyframes mask-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes sheet-up {
  from { transform: translateY(24%); opacity: 0.6; }
  to { transform: translateY(0); opacity: 1; }
}

/* 卡片进入的轻微上浮，用于核心页面板块 */
.fade-up {
  animation: fade-up 0.32s ease both;
}

@keyframes fade-up {
  from { opacity: 0; transform: translateY(12rpx); }
  to { opacity: 1; transform: translateY(0); }
}

/* 全局空状态与加载提示统一观感 */
.empty {
  color: #9aa6b8;
}
</style>
