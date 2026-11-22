<template>
  <el-dropdown trigger="click">
    <div class="user-info">
      <el-avatar :size="30" :src="userStore.avatar || undefined" class="user-avatar">
        <el-icon :size="18"><UserFilled /></el-icon>
      </el-avatar>
      <span class="user-name">{{ userStore.name }}</span>
      <el-icon :size="12"><ArrowDown /></el-icon>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item @click="goProfile">{{ t('navbar.profile') }}</el-dropdown-item>
        <el-dropdown-item divided @click="openPwdDialog">修改密码</el-dropdown-item>
        <el-dropdown-item divided @click="handleLogout">{{ t('navbar.logout') }}</el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>

  <!-- 修改密码 -->
  <el-dialog v-model="pwdVisible" :title="pwdTitle" width="440px" destroy-on-close>
    <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
      <el-form-item label="旧密码" prop="oldPassword">
        <el-input
          v-model="pwdForm.oldPassword"
          type="password"
          show-password
          placeholder="请输入旧密码"
        />
      </el-form-item>
      <el-form-item label="新密码" prop="newPassword">
        <el-input
          v-model="pwdForm.newPassword"
          type="password"
          show-password
          placeholder="请输入新密码"
        />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="pwdForm.confirmPassword"
          type="password"
          show-password
          placeholder="请再次输入新密码"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="pwdVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmitPwd">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { UserFilled, ArrowDown } from '@element-plus/icons-vue';
import { useUserStore } from '@/store/user';
import { updatePwd } from '@/api/user';

const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();

const pwdTitle = computed(() => '修改密码');

const pwdVisible = ref(false);
const submitting = ref(false);
const pwdFormRef = ref();

const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const pwdRules = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [
    { required: true, min: 5, max: 20, message: '密码长度为 5-20 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== pwdForm.newPassword) {
          callback(new Error('两次输入的密码不一致'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

// 打开修改密码弹窗
function openPwdDialog() {
  pwdForm.oldPassword = '';
  pwdForm.newPassword = '';
  pwdForm.confirmPassword = '';
  pwdVisible.value = true;
}

// 跳转个人中心
function goProfile() {
  router.push('/profile');
}

// 提交修改密码
async function handleSubmitPwd() {
  await pwdFormRef.value.validate();
  submitting.value = true;
  try {
    await updatePwd(pwdForm.oldPassword, pwdForm.newPassword);
    ElMessage.success('密码修改成功，请重新登录');
    pwdVisible.value = false;
    await userStore.logout();
    router.push('/login');
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false;
  }
}

// 退出登录
async function handleLogout() {
  await userStore.logout();
  router.push('/login');
}
</script>

<style scoped>
.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  outline: none;
  color: var(--go-navbar-text, #303133);
}

.user-avatar {
  background: var(--el-color-primary);
  color: #fff;
}

.user-name {
  font-size: 14px;
  color: var(--go-navbar-text, #303133);
}

/* 窄窗：只保留头像，隐藏用户名与下拉箭头，避免顶栏溢出 */
@media (max-width: 768px) {
  .user-name,
  .user-info .el-icon {
    display: none;
  }
}
</style>
