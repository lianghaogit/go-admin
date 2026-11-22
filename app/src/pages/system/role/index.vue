<template>
  <view class="page">
    <!-- 搜索栏 + 新增 -->
    <view class="search-bar">
      <input
        v-model="keyword"
        class="search-input"
        placeholder="输入角色名搜索"
        placeholder-class="ph"
        confirm-type="search"
        @confirm="onSearch"
      />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
      <view v-if="$hasPerm('system:role:add')" class="add-btn" @tap="openForm()">＋ 新增</view>
    </view>

    <!-- 列表 -->
    <view v-if="list.length" class="role-list">
      <view v-for="r in list" :key="r.roleId" class="role-item">
        <view class="avatar" :style="{ background: avatarColor(r.roleId) }">{{ (r.roleName || 'R').charAt(0) }}</view>
        <view class="role-main" @tap="userStore.hasPerm('system:role:edit') && openForm(r)">
          <view class="role-line">
            <text class="role-name">{{ r.roleName }}</text>
            <text class="status" :class="r.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(r)">{{ r.status === '0' ? '正常' : '停用' }}</text>
          </view>
          <view class="role-line sub">
            <text class="role-key">{{ r.roleKey }}</text>
          </view>
          <view class="role-line sub">
            <text class="role-meta">排序 {{ r.roleSort }} · {{ scopeText(r.dataScope) }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('system:role:edit')" class="op assign" @tap="openAuthUser(r)">人</view>
        <view v-if="$hasPerm('system:role:edit')" class="op" @tap="openForm(r)">✎</view>
        <view v-if="$hasPerm('system:role:remove')" class="op danger" @tap="onDelete(r)">✕</view>
      </view>

      <view v-if="loading" class="load-hint">
        <text class="load-spin"></text>
        <text>加载中…</text>
      </view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>

    <!-- 空态 -->
    <view v-else-if="!loading" class="empty">暂无角色数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.roleName || 'R').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.roleId ? '编辑角色' : '新增角色' }}</text>
            <text class="sheet-sub">{{ form.roleId ? '调整角色权限与可见范围' : '定义一组权限并分配给用户' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">角色名称</text>
            <input v-model="form.roleName" class="field-input" placeholder="必填，如 部门主管" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">权限字符</text>
            <input v-model="form.roleKey" class="field-input" placeholder="必填，如 manager" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">显示排序</text>
            <input v-model="form.roleSort" class="field-input" type="number" placeholder="数字越小越靠前" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">备注</text>
            <input v-model="form.remark" class="field-input" placeholder="选填" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">角色状态</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">角色状态</text>
              <text class="field-desc">停用后持有该角色的用户将失去对应权限</text>
            </view>
            <switch :checked="form.status === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '正常' : '停用' }}</text>
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">数据权限</text>
          </view>
          <view class="field">
            <text class="field-label">数据范围</text>
            <picker :range="scopeOptions" range-key="label" :value="scopeIndex" @change="onScopeChange">
              <view class="field-select">
                <text class="field-select-text">{{ scopeOptions[scopeIndex].label }}</text>
                <text class="field-select-arrow">›</text>
              </view>
            </picker>
          </view>
          <view v-if="form.dataScope === '2'" class="field">
            <text class="field-label">自定义部门</text>
            <TreePicker v-model:checkedKeys="checkedDeptIds" :tree-data="deptTreeData" value-key="deptId" label-key="deptName" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">菜单权限</text>
          </view>
          <view class="field">
            <TreePicker v-model:checkedKeys="checkedMenuIds" :tree-data="menuTreeData" value-key="menuId" label-key="menuName" />
          </view>
        </view>
        <view v-if="formLoading" class="auth-loading">权限数据加载中…</view>

        <view class="form-actions">
          <view class="btn plain" @tap="closeForm">取消</view>
          <view class="btn" @tap="submitForm">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listRole, createRole, updateRole, deleteRole, getRole, roleMenuTree } from '@/api/role'
import { menuTree } from '@/api/menu'
import { deptTree } from '@/api/dept'
import { useUserStore } from '@/store/user'
import { trimStr, isNonNegInt, CODE_RE } from '@/utils/validate'
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
const form = reactive({ roleId: 0, roleName: '', roleKey: '', roleSort: 1, status: '0', dataScope: '1', remark: '' })

// 权限配置：菜单树 / 自定义部门
const menuTreeData = ref([])
const checkedMenuIds = ref([])
const deptTreeData = ref([])
const checkedDeptIds = ref([])
const formLoading = ref(false)
const authLoaded = ref(false)
const deptLoaded = ref(false)

const SCOPE_TEXT = {
  '1': '全部数据权限',
  '2': '自定义数据权限',
  '3': '本部门数据权限',
  '4': '本部门及以下数据权限',
  '5': '仅本人数据权限'
}
function scopeText(v) {
  return SCOPE_TEXT[v] || '全部数据权限'
}

const scopeOptions = ['1', '2', '3', '4', '5'].map((v) => ({ value: v, label: SCOPE_TEXT[v] }))
const scopeIndex = computed(() => Math.max(0, scopeOptions.findIndex((o) => o.value === form.dataScope)))

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
    const data = await listRole({
      pageNum: pageNum.value,
      pageSize,
      roleName: keyword.value || undefined
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

// ---- 表单 ----
async function openForm(r) {
  if (r) {
    Object.assign(form, {
      roleId: r.roleId,
      roleName: r.roleName || '',
      roleKey: r.roleKey || '',
      roleSort: r.roleSort ?? 1,
      status: r.status || '0',
      dataScope: r.dataScope || '1',
      remark: r.remark || ''
    })
  } else {
    Object.assign(form, { roleId: 0, roleName: '', roleKey: '', roleSort: 1, status: '0', dataScope: '1', remark: '' })
  }
  // 先清空并标记未加载：加载失败时禁止保存，避免用空数组覆盖原权限
  menuTreeData.value = []
  checkedMenuIds.value = []
  deptTreeData.value = []
  checkedDeptIds.value = []
  authLoaded.value = false
  deptLoaded.value = false
  showForm.value = true
  formLoading.value = true
  try {
    const roleId = form.roleId
    const [menuRes, deptRes] = await Promise.allSettled([roleId ? roleMenuTree(roleId) : menuTree(), deptTree()])
    // 菜单权限是保存的必要数据，加载失败则禁止保存，避免空数组覆盖
    if (menuRes.status !== 'fulfilled') return
    menuTreeData.value = roleId ? menuRes.value.data.menus || [] : menuRes.value.data || []
    checkedMenuIds.value = roleId ? menuRes.value.data.checkedKeys || [] : []
    if (deptRes.status === 'fulfilled') {
      deptTreeData.value = deptRes.value.data || []
      deptLoaded.value = true
    }
    // 编辑时始终回填已选部门，避免切换到自定义范围时丢失原部门
    if (roleId) {
      const detail = await getRole(roleId)
      checkedDeptIds.value = (detail.data && detail.data.deptIds) || []
    }
    authLoaded.value = true
  } catch (e) {
    // 保持 authLoaded=false，保存时会被拦截并提示重试
  } finally {
    formLoading.value = false
  }
}

function closeForm() {
  showForm.value = false
}

function onSwitch(e) {
  form.status = e.detail.value ? '0' : '1'
}

function onScopeChange(e) {
  form.dataScope = scopeOptions[Number(e.detail.value)].value
}

async function submitForm() {
  const roleName = trimStr(form.roleName)
  const roleKey = trimStr(form.roleKey)
  if (!roleName) {
    uni.showToast({ title: '请输入角色名称', icon: 'none' })
    return
  }
  if (roleName.length > 30) {
    uni.showToast({ title: '角色名称不能超过 30 个字符', icon: 'none' })
    return
  }
  if (!roleKey) {
    uni.showToast({ title: '请输入权限字符', icon: 'none' })
    return
  }
  if (!CODE_RE.test(roleKey) || roleKey.length > 100) {
    uni.showToast({ title: '权限字符需以字母开头，仅含字母、数字、下划线或中划线', icon: 'none' })
    return
  }
  if (!isNonNegInt(form.roleSort)) {
    uni.showToast({ title: '显示排序需为不小于 0 的整数', icon: 'none' })
    return
  }
  if (formLoading.value || !authLoaded.value) {
    uni.showToast({ title: '权限数据未加载完成，请稍后重试', icon: 'none' })
    return
  }
  if (form.dataScope === '2' && !deptLoaded.value) {
    uni.showToast({ title: '部门数据未加载，无法配置自定义数据范围', icon: 'none' })
    return
  }
  form.roleName = roleName
  form.roleKey = roleKey
  const payload = {
    roleName,
    roleKey,
    roleSort: Number(form.roleSort),
    status: form.status,
    dataScope: form.dataScope,
    remark: trimStr(form.remark),
    // 提交已加载的完整权限，绝不能传空数组，否则后端会清空关联
    menuIds: checkedMenuIds.value,
    deptIds: form.dataScope === '2' ? checkedDeptIds.value : []
  }
  try {
    const data = form.roleId ? await updateRole({ ...payload, roleId: form.roleId }) : await createRole(payload)
    if (data.code === 200) {
      closeForm()
      uni.showToast({ title: form.roleId ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
      if (form.roleId) {
        const idx = list.value.findIndex((it) => it.roleId === form.roleId)
        if (idx >= 0) {
          const extra = {
            roleName: payload.roleName,
            roleKey: payload.roleKey,
            roleSort: payload.roleSort,
            status: payload.status,
            dataScope: payload.dataScope,
            remark: payload.remark
          }
          list.value[idx] = { ...list.value[idx], ...extra }
        }
      } else {
        setTimeout(() => fetchList(true), 400)
      }
    } else {
      uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
    }
  } catch (e) {
    // request 层已统一提示
  }
}

// ---- 状态切换 ----
function toggleStatus(r) {
  if (!userStore.hasPerm('system:role:edit')) {
    uni.showToast({ title: '无权修改角色', icon: 'none' })
    return
  }
  if (r.roleId === 1) {
    uni.showToast({ title: '超级管理员角色不允许修改', icon: 'none' })
    return
  }
  openForm(r)
}

// ---- 删除 ----
function onDelete(r) {
  if (r.roleId === 1) {
    uni.showToast({ title: '超级管理员角色不允许删除', icon: 'none' })
    return
  }
  uni.showModal({
    title: '提示',
    content: `确定删除角色「${r.roleName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const data = await deleteRole(r.roleId)
        if (data.code === 200) {
          uni.showToast({ title: '删除成功', icon: 'success' })
          list.value = list.value.filter((it) => it.roleId !== r.roleId)
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

// ---- 分配用户 ----
function openAuthUser(r) {
  uni.navigateTo({ url: `/pages/system/role/authUser/index?roleId=${r.roleId}&roleName=${encodeURIComponent(r.roleName || '')}` })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '角色管理' })
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

.role-list {
  margin-top: 24rpx;
  background: #fff;
  border-radius: 18rpx;
  padding: 0 28rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.role-item {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 26rpx 0;
}

.role-item + .role-item {
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

.role-main {
  flex: 1;
  min-width: 0;
}

.role-line {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.role-name {
  font-size: 28rpx;
  color: #1f2d3d;
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

.role-line.sub {
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #9aa6b8;
}

.role-key {
  color: #8a94a6;
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

.op.assign {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
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
  width: 140rpx;
  font-size: 26rpx;
  color: #5a6678;
  flex-shrink: 0;
}

.input {
  flex: 1;
  height: 64rpx;
  font-size: 26rpx;
}

.picker-value {
  height: 64rpx;
  line-height: 64rpx;
  font-size: 26rpx;
  color: #1f2d3d;
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
/* ===== 角色详情弹层（新版） ===== */
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
  background: linear-gradient(135deg, #a99bff, #7b61ff);
  box-shadow: 0 10rpx 24rpx rgba(123, 97, 255, 0.26);
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
  background: linear-gradient(180deg, #a99bff, #7b61ff);
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

.field-select {
  display: flex;
  align-items: center;
  height: 84rpx;
  padding: 0 24rpx;
  background: #f5f7fb;
  border-radius: 14rpx;
}

.field-select-text {
  flex: 1;
  font-size: 27rpx;
  color: #1f2d3d;
}

.field-select-arrow {
  font-size: 30rpx;
  color: #b8c2d2;
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

.auth-loading {
  padding: 16rpx 32rpx;
  font-size: 23rpx;
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
  background: linear-gradient(135deg, #7b61ff, #5b45d6);
  box-shadow: 0 12rpx 26rpx rgba(91, 69, 214, 0.24);
}

.btn.plain {
  background: #f5f7fb;
  color: #5a6678;
  box-shadow: none;
}
</style>
