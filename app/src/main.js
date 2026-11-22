import { createSSRApp } from 'vue'
import * as Pinia from 'pinia'
import App from './App.vue'
import { useUserStore } from '@/store/user'

export function createApp() {
  const app = createSSRApp(App)
  app.use(Pinia.createPinia())
  // 全局按钮权限：模板中直接使用 $hasPerm('system:user:add')
  app.config.globalProperties.$hasPerm = (perm) => useUserStore().hasPerm(perm)
  return {
    app,
    Pinia,
  }
}
