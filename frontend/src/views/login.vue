<template>
  <div class="login-page">
    <!-- 左侧品牌区 -->
    <div class="brand-panel">
      <div class="brand-inner">
        <div class="brand-logo">Go Admin</div>
        <div class="brand-slogan">{{ t('login.slogan') }}</div>
        <div class="brand-features">
          <div class="feature-item">
            <el-icon :size="16"><Lock /></el-icon>
            <span>{{ t('login.featureRbac') }}</span>
          </div>
          <div class="feature-item">
            <el-icon :size="16"><Grid /></el-icon>
            <span>{{ t('login.featureGen') }}</span>
          </div>
          <div class="feature-item">
            <el-icon :size="16"><Odometer /></el-icon>
            <span>{{ t('login.featureMonitor') }}</span>
          </div>
        </div>
        <div class="brand-decor"></div>
      </div>
    </div>

    <!-- 右侧登录表单 -->
    <div class="form-panel">
      <div class="login-card">
        <div class="login-title">
          <h2>{{ t('login.title') }}</h2>
          <p>{{ t('login.subtitle') }}</p>
        </div>
        <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" size="large">
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              :placeholder="t('login.username')"
              :prefix-icon="User"
              clearable
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              :placeholder="t('login.password')"
              :prefix-icon="Lock"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item v-if="captchaEnabled" prop="code">
            <div class="captcha-row">
              <el-input
                v-model="loginForm.code"
                :placeholder="t('login.captcha')"
                :prefix-icon="Key"
                @keyup.enter="handleLogin"
              />
              <img
                :src="captchaImg"
                class="captcha-img"
                :title="t('login.captchaTip')"
                alt="captcha"
                @click="refreshCaptcha"
              />
            </div>
          </el-form-item>
          <div class="login-options">
            <el-checkbox v-model="rememberMe">{{ t('login.rememberMe') }}</el-checkbox>
          </div>
          <el-form-item>
            <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
              {{ t('login.login') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { User, Lock, Key, Grid, Odometer } from '@element-plus/icons-vue';
import { getCaptchaImage } from '@/api/auth';
import { useUserStore } from '@/store/user';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const loginFormRef = ref();
const loading = ref(false);
const captchaEnabled = ref(false);
const captchaImg = ref('');
const rememberMe = ref(false);
const loginForm = reactive({ username: '', password: '', code: '', uuid: '' });

const loginRules = {
  username: [{ required: true, message: t('login.inputUsername'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.inputPassword'), trigger: 'blur' }],
  code: [{ required: true, message: t('login.inputCaptcha'), trigger: 'blur' }]
};

// 获取验证码（后端返回是否开启）
async function refreshCaptcha() {
  const res = await getCaptchaImage();
  captchaEnabled.value = res.data.captchaEnabled;
  if (captchaEnabled.value) {
    loginForm.uuid = res.data.uuid;
    captchaImg.value = res.data.img;
  }
}

// 提交登录表单
async function handleLogin() {
  await loginFormRef.value.validate();
  loading.value = true;
  try {
    await userStore.login(loginForm);
    // 记住用户名（localStorage 持久化）
    if (rememberMe.value) {
      localStorage.setItem('go-admin-remember-user', loginForm.username);
    } else {
      localStorage.removeItem('go-admin-remember-user');
    }
    ElMessage.success(t('login.loginSuccess'));
    const redirect = route.query.redirect;
    router.push(typeof redirect === 'string' && redirect ? decodeURIComponent(redirect) : '/');
  } catch (e) {
    // 登录失败刷新验证码
    if (captchaEnabled.value) refreshCaptcha();
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  refreshCaptcha();
  const saved = localStorage.getItem('go-admin-remember-user');
  if (saved) {
    loginForm.username = saved;
    rememberMe.value = true;
  }
});
</script>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  background: var(--go-bg-page, #fff);
}

/* 左侧品牌区 */
.brand-panel {
  flex: 1;
  background: linear-gradient(135deg, #1f2d3d 0%, #2b4c6f 60%, #3a7bd5 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  min-width: 0;
}

.brand-inner {
  color: #fff;
  padding: 40px;
  position: relative;
  z-index: 2;
  max-width: 460px;
}

.brand-logo {
  font-size: 40px;
  font-weight: 700;
  letter-spacing: 4px;
}

.brand-slogan {
  margin-top: 14px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.75);
  letter-spacing: 1px;
}

.brand-features {
  margin-top: 40px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
}

/* 装饰圆 */
.brand-decor {
  position: absolute;
  width: 420px;
  height: 420px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  bottom: -160px;
  right: -120px;
}

.brand-decor::after {
  content: '';
  position: absolute;
  width: 260px;
  height: 260px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.05);
  top: -180px;
  left: -90px;
}

/* 右侧表单区 */
.form-panel {
  width: 480px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--go-bg-page, #fff);
}

.login-card {
  width: 340px;
  max-width: 100%;
  padding: 24px 8px;
}

.login-title {
  margin-bottom: 28px;
}

.login-title h2 {
  font-size: 24px;
  color: var(--go-text-primary, #303133);
  letter-spacing: 1px;
}

.login-title p {
  color: var(--go-text-secondary, #909399);
  font-size: 13px;
  margin-top: 8px;
}

.login-btn {
  width: 100%;
}

.login-options {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.captcha-row {
  display: flex;
  width: 100%;
  gap: 10px;
}

.captcha-img {
  width: 112px;
  height: 40px;
  cursor: pointer;
  border-radius: 2px;
  flex-shrink: 0;
}

/* 窄屏降级为居中卡片 */
@media (max-width: 768px) {
  .brand-panel {
    display: none;
  }
  .form-panel {
    width: 100%;
  }
}
</style>
