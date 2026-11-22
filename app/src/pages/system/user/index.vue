<template>
  <view class="page">
    <!-- 搜索栏 + 新增 -->
    <view class="search-bar">
      <input
        v-model="keyword"
        class="search-input"
        placeholder="输入用户名搜索"
        placeholder-class="ph"
        confirm-type="search"
        @confirm="onSearch"
      />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('system:user:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>
    <view v-if="$hasPerm('system:user:export') || $hasPerm('system:user:import')" class="tool-bar">
      <view v-if="$hasPerm('system:user:export')" class="tool-btn" @tap="onExport">导出</view>
      <view v-if="$hasPerm('system:user:import')" class="tool-btn green" @tap="onImport">导入</view>
      <view v-if="$hasPerm('system:user:import')" class="tool-btn plain" @tap="onTemplate">下载模板</view>
    </view>

    <!-- 列表 -->
    <view v-if="list.length" class="user-list">
      <view v-for="u in list" :key="u.userId" class="user-item">
        <view class="avatar" :style="{ background: avatarColor(u.userId) }">{{ (u.nickName || u.userName || 'U').charAt(0) }}</view>
        <view class="user-main" @tap="userStore.hasPerm('system:user:edit') && openForm(u)">
          <view class="user-line">
            <text class="user-name">{{ u.nickName || u.userName }}</text>
            <text class="user-uname">@{{ u.userName }}</text>
          </view>
          <view class="user-line sub">
            <text v-if="u.dept && u.dept.deptName" class="user-meta">{{ u.dept.deptName }}</text>
            <text v-if="u.phonenumber" class="user-meta">{{ u.phonenumber }}</text>
            <text v-if="u.email" class="user-meta">{{ u.email }}</text>
          </view>
        </view>
        <view class="status" :class="u.status === '0' ? 'on' : 'off'" @tap="toggleStatus(u)">{{ u.status === '0' ? '正常' : '停用' }}</view>
        <view v-if="$hasPerm('system:user:edit')" class="op" @tap="openForm(u)">✎</view>
        <view v-if="$hasPerm('system:user:remove')" class="op danger" @tap="onDelete(u)">✕</view>
      </view>

      <view v-if="loading" class="load-hint">
        <text class="load-spin"></text>
        <text>加载中…</text>
      </view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>

    <!-- 空态 -->
    <view v-else-if="!loading" class="empty">暂无用户数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.nickName || form.username || 'U').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.userId ? '编辑用户' : '新增用户' }}</text>
            <text class="sheet-sub">{{ form.userId ? '调整账号信息与组织归属' : '创建一个新的系统账号' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">用户名</text>
            <input v-model="form.username" class="field-input" placeholder="必填，登录账号" placeholder-class="ph" :disabled="!!form.userId" />
          </view>
          <view v-if="!form.userId" class="field">
            <text class="field-label">登录密码</text>
            <input v-model="form.password" class="field-input" password placeholder="必填，至少 6 位" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">昵称</text>
            <input v-model="form.nickName" class="field-input" placeholder="显示名称" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">手机号</text>
            <input v-model="form.phonenumber" class="field-input" type="tel" maxlength="11" placeholder="11 位手机号" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">邮箱</text>
            <input v-model="form.email" class="field-input" placeholder="联系邮箱" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">备注</text>
            <input v-model="form.remark" class="field-input" placeholder="选填" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">账号状态</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">账号状态</text>
              <text class="field-desc">停用后该账号将无法登录系统</text>
            </view>
            <switch :checked="form.status === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '正常' : '停用' }}</text>
          </view>
          <view v-if="form.userId && $hasPerm('system:user:resetPwd')" class="field-card reset-card">
            <view class="reset-row">
              <input v-model="newPwd" class="reset-input" password placeholder="输入新密码（至少 6 位）" placeholder-class="ph" />
              <view class="reset-btn" @tap="onResetPwd">重置</view>
            </view>
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">组织与角色</text>
          </view>
          <view class="field">
            <text class="field-label">所属部门</text>
            <TreePicker
              :tree-data="deptTreeData"
              :checked-keys="form.deptId ? [form.deptId] : []"
              value-key="deptId"
              label-key="deptName"
              single
              @update:checkedKeys="onDeptChange"
            />
          </view>
          <view class="field">
            <text class="field-label">角色（至少选择一个）</text>
            <view class="chip-wrap">
              <view
                v-for="r in roleOptions"
                :key="r.roleId"
                class="chip"
                :class="{ on: form.roleIds.includes(r.roleId) }"
                @tap="toggleRole(r.roleId)"
              >{{ r.roleName }}</view>
            </view>
            <text v-if="!roleOptions.length" class="form-hint">暂无可分配角色</text>
          </view>
          <view class="field">
            <text class="field-label">岗位</text>
            <view class="chip-wrap">
              <view
                v-for="p in postOptions"
                :key="p.postId"
                class="chip"
                :class="{ on: form.postIds.includes(p.postId) }"
                @tap="togglePost(p.postId)"
              >{{ p.postName }}</view>
            </view>
            <text v-if="!postOptions.length" class="form-hint">暂无可分配岗位</text>
          </view>
        </view>
        <view v-if="formLoading" class="auth-loading">部门/角色/岗位加载中…</view>

        <view class="form-actions">
          <view class="btn plain" @tap="closeForm">取消</view>
          <view class="btn" @tap="submitForm">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listUser, createUser, updateUser, deleteUser, changeUserStatus, getUser, resetUserPwd } from '@/api/user'
import { listAllRoles } from '@/api/role'
import { deptTree } from '@/api/dept'
import { listAllPosts } from '@/api/post'
import { useUserStore } from '@/store/user'
import { getBaseUrl } from '@/utils/config'
import { getToken } from '@/utils/auth'
import { exportExcel, exportFile, fileStamp } from '@/utils/download'
import { trimStr, isMobile, isEmail } from '@/utils/validate'
import TreePicker from '@/components/TreePicker.vue'

const userStore = useUserStore()

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

// 表单弹层
const showForm = ref(false)
const form = reactive({ userId: 0, username: '', password: '', nickName: '', phonenumber: '', email: '', status: '0', remark: '', deptId: 0, roleIds: [], postIds: [] })
const newPwd = ref('')

// 部门/角色/岗位选项
const deptTreeData = ref([])
const roleOptions = ref([])
const postOptions = ref([])
const formLoading = ref(false)
const authLoaded = ref(false)
const deptLoaded = ref(false)
const rolesLoaded = ref(false)
const postsLoaded = ref(false)
const detailLoaded = ref(false)

const PALETTE = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#9c6ade', '#00b8a9', '#fa8c5f', '#5b8ff9']
function avatarColor(id) {
  return PALETTE[Number(id) % PALETTE.length]
}

async function fetchList(reset) {
  if (loading.value) return
  if (reset) {
    pageNum.value = 1
    finished.value = false
  } else if (finished.value) {
    return
  }
  loading.value = true
  try {
    const data = await listUser({
      pageNum: pageNum.value,
      pageSize,
      username: keyword.value || undefined
    })
    if (data.code === 200) {
      total.value = data.total
      const rows = data.rows || []
      list.value = reset ? rows : list.value.concat(rows)
      finished.value = list.value.length >= data.total
      if (rows.length) pageNum.value += 1
    } else {
      uni.showToast({ title: data.msg || '查询失败', icon: 'none' })
    }
  } catch (e) {
  } finally {
    loading.value = false
  }
}

function onSearch() { fetchList(true) }
function onReset() { keyword.value = ''; fetchList(true) }

// ---- 导出 / 导入 ----
function onExport() {
  exportExcel('/api/system/user/export', { username: keyword.value || undefined }, `用户数据_${fileStamp()}.xlsx`)
}

function onTemplate() {
  exportFile('/api/system/user/importTemplate', {}, '用户导入模板.xlsx', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
}

function onImport() {
  // #ifdef H5
  uni.chooseFile({
    count: 1,
    extension: ['.xlsx'],
    success: (res) => {
      const f = (res.tempFiles && res.tempFiles[0]) || {}
      const path = f.path || (res.tempFilePaths || [])[0]
      if (!path) return
      uni.showModal({
        title: '导入用户',
        content: `将导入「${f.name || '所选文件'}」，用户名重复时不覆盖。是否继续？`,
        success: (r) => { if (r.confirm) doImport(path) }
      })
    }
  })
  // #endif
  // #ifndef H5
  uni.showModal({ title: '提示', content: '用户导入请在浏览器/Web 端使用', showCancel: false })
  // #endif
}

function doImport(path) {
  uni.showLoading({ title: '导入中…', mask: true })
  const token = getToken()
  uni.uploadFile({
    url: getBaseUrl() + '/api/system/user/importData',
    filePath: path,
    name: 'file',
    formData: { updateSupport: 'false' },
    header: token ? { Authorization: 'Bearer ' + token } : {},
    success: (res) => {
      uni.hideLoading()
      let data = {}
      try { data = JSON.parse(res.data) } catch (e) {}
      if (data.code === 200) {
        const d = data.data || {}
        const fails = (d.failMessages || []).slice(0, 5).join('\n')
        uni.showModal({
          title: '导入完成',
          content: `成功 ${d.successNum || 0} 条，失败 ${d.failureNum || 0} 条${fails ? '\n' + fails : ''}`,
          showCancel: false
        })
        fetchList(true)
      } else {
        uni.showToast({ title: data.msg || '导入失败', icon: 'none' })
      }
    },
    fail: () => {
      uni.hideLoading()
      uni.showToast({ title: '上传失败，请检查网络', icon: 'none' })
    }
  })
}

// ---- 表单 ----
async function openForm(u) {
  if (u) {
    Object.assign(form, {
      userId: u.userId,
      username: u.userName,
      password: '',
      nickName: u.nickName || '',
      phonenumber: u.phonenumber || '',
      email: u.email || '',
      status: u.status || '0',
      remark: u.remark || '',
      deptId: u.deptId || 0,
      roleIds: [],
      postIds: []
    })
  } else {
    Object.assign(form, { userId: 0, username: '', password: '', nickName: '', phonenumber: '', email: '', status: '0', remark: '', deptId: 0, roleIds: [], postIds: [] })
  }
  newPwd.value = ''
  showForm.value = true
  formLoading.value = true
  authLoaded.value = false
  deptLoaded.value = false
  rolesLoaded.value = false
  postsLoaded.value = false
  detailLoaded.value = false
  try {
    const [deptRes, roleRes, postRes] = await Promise.allSettled([deptTree(), listAllRoles(), listAllPosts()])
    if (deptRes.status === 'fulfilled') {
      deptTreeData.value = deptRes.value.data || []
      deptLoaded.value = true
    }
    if (roleRes.status === 'fulfilled') {
      roleOptions.value = roleRes.value.data || []
      rolesLoaded.value = true
    }
    if (postRes.status === 'fulfilled') {
      postOptions.value = postRes.value.data || []
      postsLoaded.value = true
    }
    if (u && u.userId) {
      // 详情用于回填部门/角色/岗位；失败时不提交关联字段，避免把已有关系清空
      const detail = await getUser(u.userId)
      const d = detail.data || {}
      form.deptId = d.deptId || form.deptId
      form.roleIds = (d.roles || []).map((x) => x.roleId)
      form.postIds = (d.posts || []).map((x) => x.postId)
      detailLoaded.value = true
    }
    authLoaded.value = true
  } catch (e) {
    // 关键数据缺失时由提交校验拦截
  } finally {
    formLoading.value = false
  }
}

function onDeptChange(keys) {
  form.deptId = keys.length ? keys[0] : 0
}

function toggleRole(roleId) {
  const idx = form.roleIds.indexOf(roleId)
  if (idx >= 0) form.roleIds.splice(idx, 1)
  else form.roleIds.push(roleId)
}

function togglePost(postId) {
  const idx = form.postIds.indexOf(postId)
  if (idx >= 0) form.postIds.splice(idx, 1)
  else form.postIds.push(postId)
}

function closeForm() {
  showForm.value = false
  newPwd.value = ''
}

// ---- 重置密码 ----
function onResetPwd() {
  const pwd = trimStr(newPwd.value)
  if (!pwd || pwd.length < 6 || pwd.length > 30) {
    uni.showToast({ title: '新密码长度需为 6-30 位', icon: 'none' })
    return
  }
  if (pwd === form.password) {
    uni.showToast({ title: '新密码不能与原密码相同', icon: 'none' })
    return
  }
  uni.showModal({
    title: '提示',
    content: `确定重置用户「${form.nickName || form.username}」的密码吗？`,
    confirmText: '重置',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await resetUserPwd(form.userId, pwd)
        newPwd.value = ''
        uni.showToast({ title: '密码已重置', icon: 'success' })
      } catch (e) {}
    }
  })
}

function onSwitch(e) {
  form.status = e.detail.value ? '0' : '1'
}

async function submitForm() {
  if (formLoading.value || !authLoaded.value) {
    uni.showToast({ title: '基础数据未加载完成，请稍后重试', icon: 'none' })
    return
  }
  const username = trimStr(form.username)
  const nickName = trimStr(form.nickName)
  const phone = trimStr(form.phonenumber)
  const email = trimStr(form.email)
  if (!form.userId && !username) {
    uni.showToast({ title: '请输入用户名', icon: 'none' })
    return
  }
  if (username.length > 30) {
    uni.showToast({ title: '用户名不能超过 30 个字符', icon: 'none' })
    return
  }
  if (!form.userId && !form.password) {
    uni.showToast({ title: '请输入登录密码', icon: 'none' })
    return
  }
  if (!form.userId && (form.password.length < 6 || form.password.length > 30)) {
    uni.showToast({ title: '密码长度需为 6-30 位', icon: 'none' })
    return
  }
  if (!nickName) {
    uni.showToast({ title: '昵称不能为空', icon: 'none' })
    return
  }
  if (nickName.length > 30) {
    uni.showToast({ title: '昵称不能超过 30 个字符', icon: 'none' })
    return
  }
  if (phone && !isMobile(phone)) {
    uni.showToast({ title: '手机号格式不正确', icon: 'none' })
    return
  }
  if (email && !isEmail(email)) {
    uni.showToast({ title: '邮箱格式不正确', icon: 'none' })
    return
  }
  if (!deptLoaded.value) {
    uni.showToast({ title: '部门数据未加载，请重试', icon: 'none' })
    return
  }
  if (!form.deptId) {
    uni.showToast({ title: '请选择所属部门', icon: 'none' })
    return
  }
  if (!form.userId && !rolesLoaded.value) {
    uni.showToast({ title: '角色数据未加载，请重试', icon: 'none' })
    return
  }
  if (!form.userId && !form.roleIds.length) {
    uni.showToast({ title: '请至少选择一个角色', icon: 'none' })
    return
  }
  form.username = username
  form.nickName = nickName
  form.phonenumber = phone
  form.email = email
  const payload = {
    nickName,
    phonenumber: phone,
    email,
    status: form.status,
    remark: trimStr(form.remark),
    deptId: form.deptId
  }
  // 仅当数据成功加载（编辑时还需带回原值）才提交关联字段，否则保持后端原值
  if (!form.userId || detailLoaded.value) {
    if (rolesLoaded.value) payload.roleIds = form.roleIds
    if (postsLoaded.value) payload.postIds = form.postIds
  }
  try {
    if (form.userId) {
      const data = await updateUser({ ...payload, userId: form.userId })
      handleSaveResult(data, '修改成功')
    } else {
      const data = await createUser({ ...payload, username: form.username, password: form.password })
      handleSaveResult(data, '新增成功')
    }
  } catch (e) {
    // request 层已统一提示
  }
}

function findDeptName(id) {
  let name = ''
  function walk(nodes) {
    for (const n of nodes || []) {
      if (n.deptId === id) { name = n.deptName; return true }
      if (n.children && walk(n.children)) return true
    }
  }
  walk(deptTreeData.value)
  return name
}

function handleSaveResult(data, okMsg) {
  if (data.code === 200) {
    closeForm()
    uni.showToast({ title: okMsg, icon: 'success', duration: 1200 })
    if (form.userId) {
      // 编辑成功：就地更新当前列表项，保持分页与滚动位置
      const idx = list.value.findIndex((it) => it.userId === form.userId)
      if (idx >= 0) {
        const old = list.value[idx]
        const deptName = findDeptName(form.deptId) || (old.dept && old.dept.deptName) || ''
        list.value[idx] = {
          ...old,
          nickName: form.nickName,
          phonenumber: form.phonenumber,
          email: form.email,
          status: form.status,
          remark: form.remark,
          deptId: form.deptId,
          dept: { ...(old.dept || {}), deptId: form.deptId, deptName }
        }
      }
    } else {
      setTimeout(() => fetchList(true), 400)
    }
  } else {
    uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
  }
}

// ---- 状态切换 ----
function toggleStatus(u) {
  if (!userStore.hasPerm('system:user:edit')) {
    uni.showToast({ title: '无权修改用户', icon: 'none' })
    return
  }
  if (u.userName === 'admin') {
    uni.showToast({ title: '超级管理员不允许停用', icon: 'none' })
    return
  }
  const next = u.status === '0' ? '1' : '0'
  uni.showModal({
    title: '提示',
    content: `确定将用户「${u.nickName || u.userName}」${next === '0' ? '启用' : '停用'}吗？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await changeUserStatus(u.userId, next)
        if (data.code === 200) {
          uni.showToast({ title: '操作成功', icon: 'success' })
          u.status = next
        } else {
          uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
        }
      } catch (e) {
        // request 层已统一提示
      }
    }
  })
}

// ---- 删除 ----
function onDelete(u) {
  if (!userStore.hasPerm('system:user:remove')) {
    uni.showToast({ title: '无权删除用户', icon: 'none' })
    return
  }
  if (u.userName === 'admin') {
    uni.showToast({ title: '超级管理员不允许删除', icon: 'none' })
    return
  }
  uni.showModal({
    title: '提示',
    content: `确定删除用户「${u.nickName || u.userName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteUser(u.userId)
        if (data.code === 200) {
          uni.showToast({ title: '删除成功', icon: 'success' })
          list.value = list.value.filter((it) => it.userId !== u.userId)
          total.value = Math.max(0, total.value - 1)
          if (!list.value.length) fetchList(true)
        } else {
          uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
        }
      } catch (e) {
        // request 层已统一提示
      }
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '用户管理' })
  fetchList(true)
})

onPullDownRefresh(async () => {
  try {
    await fetchList(true)
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
    uni.showToast({ title: '刷新失败', icon: 'none' })
  } finally {
    uni.stopPullDownRefresh()
  }
})

onReachBottom(() => {
  fetchList(false)
})
</script>

<style scoped>
.page {
  padding: 28rpx 28rpx 60rpx;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 24rpx;
  background: #fff;
  border-radius: 18rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.search-input {
  flex: 1;
  min-width: 0;
  height: 68rpx;
  padding: 0 20rpx;
  background: #f5f7fa;
  border-radius: 18rpx;
  font-size: 26rpx;
}

.ph {
  color: #9aa6b8;
}

.search-btn {
  padding: 0 22rpx;
  height: 68rpx;
  line-height: 68rpx;
  background: #409eff;
  color: #fff;
  font-size: 26rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.search-btn.plain {
  background: #f5f7fa;
  color: #5a6678;
}

.add-btn {
  padding: 0 22rpx;
  height: 68rpx;
  line-height: 68rpx;
  background: #67c23a;
  color: #fff;
  font-size: 26rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.tool-bar {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
  margin-top: 16rpx;
}

.tool-btn {
  padding: 0 22rpx;
  height: 64rpx;
  line-height: 64rpx;
  background: #e6a23c;
  color: #fff;
  font-size: 26rpx;
  border-radius: 18rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.tool-btn.green { background: #67c23a; }

.tool-btn.plain { background: #fff; color: #5a6678; }

.user-list {
  margin-top: 24rpx;
  background: #fff;
  border-radius: 18rpx;
  padding: 0 28rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.user-item {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 26rpx 0;
}

.user-item + .user-item {
  border-top: 1rpx solid #f0f3f7;
}

.avatar {
  width: 76rpx;
  height: 76rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
}

.user-main {
  flex: 1;
  min-width: 0;
}

.user-line {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.user-name {
  font-size: 28rpx;
  color: #1f2d3d;
}

.user-uname {
  font-size: 22rpx;
  color: #9aa6b8;
}

.user-line.sub {
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #9aa6b8;
}

.user-meta + .user-meta::before {
  content: '·';
  margin: 0 8rpx;
}

.status {
  font-size: 20rpx;
  padding: 2rpx 14rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.status.on {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
}

.status.off {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.1);
}

.op {
  width: 60rpx;
  height: 60rpx;
  line-height: 60rpx;
  text-align: center;
  font-size: 28rpx;
  color: #409eff;
  background: rgba(64, 158, 255, 0.08);
  border-radius: 50%;
  flex-shrink: 0;
}

.op.danger {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.08);
}

.load-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 24rpx 0;
  font-size: 24rpx;
  color: #9aa6b8;
}

.load-spin {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #e6ebf2;
  border-top-color: #409eff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.load-hint.done {
  color: #c3ccd8;
}

.empty {
  margin-top: 24rpx;
  padding: 100rpx 0;
  text-align: center;
  color: #9aa6b8;
  font-size: 26rpx;
  background: #fff;
  border-radius: 18rpx;
}

/* 弹层 */
.mask {
  position: fixed;
  inset: 0;
  background: rgba(31, 45, 61, 0.45);
  z-index: 99;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.sheet {
  width: 100%;
  max-height: 86vh;
  overflow-y: auto;
  background: #fff;
  border-radius: 16rpx 16rpx 0 0;
  padding: 32rpx 32rpx 48rpx;
  box-sizing: border-box;
}

.sheet-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2d3d;
  text-align: center;
  margin-bottom: 28rpx;
}

.form-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #f0f3f7;
}

.label {
  width: 120rpx;
  font-size: 26rpx;
  color: #5a6678;
  flex-shrink: 0;
}

.input {
  flex: 1;
  height: 64rpx;
  font-size: 26rpx;
}

.switch-text {
  font-size: 24rpx;
  color: #8a94a6;
}

.form-block {
  display: block;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #f0f3f7;
}

.block-label {
  display: block;
  width: auto;
  margin-bottom: 14rpx;
}

.chip-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.chip {
  padding: 10rpx 24rpx;
  font-size: 24rpx;
  color: #5a6678;
  background: #f5f7fa;
  border: 1rpx solid #e6ebf2;
  border-radius: 28rpx;
}

.chip.on {
  color: #fff;
  background: #409eff;
  border-color: #409eff;
}

.form-hint {
  font-size: 24rpx;
  color: #9aa6b8;
}

.reset-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.reset-input {
  flex: 1;
  height: 64rpx;
  padding: 0 18rpx;
  background: #f5f7fa;
  border-radius: 18rpx;
  font-size: 26rpx;
}

.reset-btn {
  padding: 0 24rpx;
  height: 64rpx;
  line-height: 64rpx;
  background: #e6a23c;
  color: #fff;
  font-size: 24rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.auth-loading {
  padding: 14rpx 0;
  font-size: 24rpx;
  color: #9aa6b8;
}

.form-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 40rpx;
}

.btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  text-align: center;
  background: #409eff;
  color: #fff;
  font-size: 28rpx;
  border-radius: 18rpx;
}

.btn.plain {
  background: #f5f7fa;
  color: #5a6678;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
/* ===== 用户详情弹层（新版） ===== */
.sheet {
  padding: 0;
  max-height: 90vh;
  border-radius: 28rpx 28rpx 0 0;
}

.sheet-head {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 30rpx 32rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f2f5fb;
}

.sheet-avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #6aa4ff, #3a6fe0);
  box-shadow: 0 10rpx 24rpx rgba(58, 111, 224, 0.26);
  flex-shrink: 0;
}

.sheet-head-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.sheet-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2d3d;
  text-align: left;
  margin-bottom: 0;
}

.sheet-sub {
  font-size: 22rpx;
  color: #9aa6b8;
}

.sheet-close {
  width: 56rpx;
  height: 56rpx;
  line-height: 52rpx;
  text-align: center;
  border-radius: 50%;
  background: #f2f5fb;
  color: #8a94a6;
  font-size: 34rpx;
  flex-shrink: 0;
}

.group {
  padding: 10rpx 32rpx 6rpx;
}

.group + .group {
  margin-top: 10rpx;
  border-top: 12rpx solid #f5f7fb;
}

.group-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 18rpx 0 6rpx;
}

.group-bar {
  width: 8rpx;
  height: 26rpx;
  border-radius: 999rpx;
  background: linear-gradient(180deg, #6aa4ff, #3a6fe0);
}

.group-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #1f2d3d;
}

.field {
  margin-top: 18rpx;
}

.field-label {
  display: block;
  font-size: 23rpx;
  color: #8a94a6;
  margin-bottom: 10rpx;
}

.field-label-inline {
  margin-bottom: 0;
  color: #1f2d3d;
  font-size: 26rpx;
}

.field-input {
  width: 100%;
  height: 84rpx;
  padding: 0 24rpx;
  background: #f5f7fb;
  border-radius: 14rpx;
  font-size: 27rpx;
  color: #1f2d3d;
  box-sizing: border-box;
}

.field-card {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-top: 18rpx;
  padding: 20rpx 24rpx;
  background: #f7f9fc;
  border-radius: 16rpx;
}

.field-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.field-desc {
  font-size: 21rpx;
  color: #9aa6b8;
}

.status-pill {
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(63, 174, 122, 0.1);
  color: #3fae7a;
  font-size: 21rpx;
  flex-shrink: 0;
}

.status-pill.off {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.reset-card {
  padding: 16rpx 20rpx;
}

.reset-input {
  height: 72rpx;
  background: #fff;
  border: 1rpx solid #e8edf6;
  border-radius: 12rpx;
}

.reset-btn {
  height: 72rpx;
  line-height: 72rpx;
  border-radius: 12rpx;
  background: linear-gradient(135deg, #ffc46b, #e6a23c);
}

.chip {
  padding: 12rpx 28rpx;
  font-size: 24rpx;
  color: #5a6678;
  background: #f5f7fb;
  border: 1rpx solid transparent;
  border-radius: 999rpx;
}

.chip.on {
  color: #3a6fe0;
  background: rgba(58, 111, 224, 0.1);
  border-color: #3a6fe0;
  font-weight: 600;
}

.chip-wrap {
  gap: 12rpx;
}

.form-hint {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: #9aa6b8;
}

.form-actions {
  position: sticky;
  bottom: 0;
  margin: 0;
  padding: 20rpx 32rpx 40rpx;
  background: #fff;
  border-top: 1rpx solid #f2f5fb;
}

.btn {
  height: 92rpx;
  line-height: 92rpx;
  border-radius: 18rpx;
  font-size: 30rpx;
  font-weight: 600;
  background: linear-gradient(135deg, #4a7df0, #3457c9);
  box-shadow: 0 12rpx 26rpx rgba(52, 87, 201, 0.24);
}

.btn.plain {
  background: #f5f7fb;
  color: #5a6678;
  box-shadow: none;
}
</style>
