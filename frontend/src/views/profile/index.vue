<template>
  <div class="app-container profile-page">
    <el-row :gutter="20">
      <!-- 左：个人资料 -->
      <el-col :xs="24" :md="10">
        <el-card shadow="never" class="profile-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">个人资料</span>
              <el-tag size="small" type="success">{{ userStore.name }}</el-tag>
            </div>
          </template>

          <!-- 头像区：hover 显示更换遮罩 -->
          <div class="avatar-area">
            <div class="avatar-wrap" @click="uploadRef?.click()">
              <el-avatar :size="96" :src="avatarUrl" @error="() => (avatarError = true)">
                <span class="avatar-fallback">{{ userStore.name?.slice(0, 1) }}</span>
              </el-avatar>
              <div class="avatar-mask">
                <el-icon :size="18"><UploadFilled /></el-icon>
                <span>更换头像</span>
              </div>
              <input
                ref="uploadRef"
                type="file"
                accept="image/*"
                style="display: none"
                @change="handleAvatarChange"
              />
            </div>
            <div class="avatar-info">
              <div class="avatar-name">{{ profileForm.nickName || userStore.name }}</div>
              <div class="avatar-sub">
                <el-tag
                  v-for="r in userStore.roles"
                  :key="r"
                  size="small"
                  class="role-tag"
                  type="primary"
                  >{{ r }}</el-tag
                >
                <span v-if="profileForm.deptName" class="dept-text">
                  {{ profileForm.deptName }}
                </span>
              </div>
            </div>
          </div>

          <el-divider />

          <el-form ref="profileRef" :model="profileForm" :rules="profileRules" label-width="90px">
            <el-form-item label="昵称" prop="nickName">
              <el-input v-model="profileForm.nickName" placeholder="请输入昵称" clearable />
            </el-form-item>
            <el-form-item label="手机号码" prop="phonenumber">
              <el-input v-model="profileForm.phonenumber" placeholder="请输入手机号码" clearable />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" placeholder="请输入邮箱" clearable />
            </el-form-item>
            <el-form-item label="备注">
              <el-input
                v-model="profileForm.remark"
                type="textarea"
                :rows="3"
                placeholder="请输入备注"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingProfile" @click="handleSaveProfile">
                保存
              </el-button>
              <el-button :disabled="savingProfile" @click="resetProfile">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 右：修改密码 + 登录信息 -->
      <el-col :xs="24" :md="14">
        <el-card shadow="never" class="profile-card">
          <template #header>
            <span class="card-title">修改密码</span>
          </template>
          <el-form ref="pwdRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
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
                placeholder="5-20 个字符"
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
            <el-form-item>
              <el-button type="primary" :loading="savingPwd" @click="handleSavePwd">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="profile-card login-info-card">
          <template #header>
            <span class="card-title">登录信息</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户名">{{ userStore.name }}</el-descriptions-item>
            <el-descriptions-item label="角色">
              <template v-if="userStore.roles.length">
                <el-tag
                  v-for="r in userStore.roles"
                  :key="r"
                  size="small"
                  class="role-tag"
                  type="primary"
                  >{{ r }}</el-tag
                >
              </template>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ parseTime(profileForm.createTime) || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'Profile' });

import { reactive, ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { UploadFilled } from '@element-plus/icons-vue';
import { useUserStore } from '@/store/user';
import { getProfile, updateProfile, updateAvatar, updatePwd } from '@/api/user';
import { uploadFile } from '@/api/tool/apiDoc';
import { parseTime } from '@/utils/format';

const userStore = useUserStore();

const uploadRef = ref();
const profileRef = ref();
const pwdRef = ref();
const savingProfile = ref(false);
const savingPwd = ref(false);
const avatarError = ref(false);

const avatarUrl = computed(() => (userStore.avatar ? userStore.avatar : ''));

const profileForm = reactive({
  nickName: '',
  phonenumber: '',
  email: '',
  remark: '',
  deptName: '',
  createTime: ''
});

const profileRules = {
  nickName: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  phonenumber: [{ pattern: /^1[3-9]\d{9}$/, message: '手机号码格式不正确', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }]
};

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
      validator: (_rule, value, callback) => {
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

// 获取个人资料（响应体为 { code, msg, data: {...} }）
async function fetchProfile() {
  try {
    const res = await getProfile();
    const u = res.data;
    Object.assign(profileForm, {
      nickName: u.nickName || '',
      phonenumber: u.phonenumber || '',
      email: u.email || '',
      remark: u.remark || '',
      deptName: u.dept?.deptName || '',
      createTime: u.createTime || ''
    });
    if (u.avatar) {
      userStore.avatar = u.avatar;
    }
    if (u.nickName) {
      userStore.name = u.nickName;
    }
  } catch (e) {
    // 拦截器已提示
  }
}

// 重置为已保存的个人资料
function resetProfile() {
  fetchProfile();
}

// 保存个人资料
async function handleSaveProfile() {
  await profileRef.value.validate();
  savingProfile.value = true;
  try {
    await updateProfile({
      nickName: profileForm.nickName,
      phonenumber: profileForm.phonenumber,
      email: profileForm.email,
      remark: profileForm.remark
    });
    userStore.name = profileForm.nickName;
    ElMessage.success('个人资料修改成功');
  } catch (e) {
    // 拦截器提示
  } finally {
    savingProfile.value = false;
  }
}

// 头像选择回调：上传 → 更新头像
async function handleAvatarChange(e) {
  const file = e.target.files && e.target.files[0];
  e.target.value = '';
  if (!file) return;
  if (!/^image\//.test(file.type)) {
    ElMessage.warning('请选择图片文件');
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning('图片不能超过 5MB');
    return;
  }
  const form = new FormData();
  form.append('file', file);
  try {
    const res = await uploadFile(form);
    // 上传接口平铺返回 { fileName, url }，url 为 /uploads/ 相对路径
    const url = res.url || res.fileName;
    if (!url) {
      ElMessage.error('头像上传失败');
      return;
    }
    await updateAvatar(url);
    userStore.avatar = url;
    avatarError.value = false;
    ElMessage.success('头像更新成功');
  } catch (err) {
    // 拦截器已提示
  }
}

// 修改密码
async function handleSavePwd() {
  await pwdRef.value.validate();
  savingPwd.value = true;
  try {
    await updatePwd(pwdForm.oldPassword, pwdForm.newPassword);
    ElMessage.success('密码修改成功，请重新登录');
    // logout 内部整页跳转登录页
    await userStore.logout();
  } catch (e) {
    // 拦截器提示
  } finally {
    savingPwd.value = false;
  }
}

fetchProfile();
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  font-weight: 600;
}

.profile-card {
  margin-bottom: 16px;
}

/* 头像区 */
.avatar-area {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 8px 0 4px;
}

.avatar-wrap {
  position: relative;
  cursor: pointer;
  border-radius: 50%;
}

.avatar-wrap .el-avatar {
  display: block;
  transition: filter 0.2s;
}

.avatar-wrap:hover .el-avatar {
  filter: brightness(0.75);
}

.avatar-mask {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.2s;
}

.avatar-wrap:hover .avatar-mask {
  opacity: 1;
}

.avatar-fallback {
  font-size: 36px;
  color: #fff;
}

.avatar-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.avatar-name {
  font-size: 18px;
  font-weight: 600;
}

.avatar-sub {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dept-text {
  color: var(--go-text-secondary);
  font-size: 13px;
}

.role-tag {
  margin-right: 2px;
}

.login-info-card {
  margin-top: 0;
}
</style>
